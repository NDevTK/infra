// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package builds fetches and handles the build image information from the
// release team.
package builds

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	cloudPubsub "cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	buildPB "go.chromium.org/chromiumos/infra/proto/go/chromiumos"
	requestpb "go.chromium.org/chromiumos/infra/proto/go/test_platform"
	kronpb "go.chromium.org/chromiumos/infra/proto/go/test_platform/kron"
	suschpb "go.chromium.org/chromiumos/infra/proto/go/testplans"

	"infra/cros/cmd/kron/common"
	"infra/cros/cmd/kron/metrics"
	"infra/cros/cmd/kron/pubsub"
)

// extractMilestoneAndVersion returns the milestone and platform version from
// the build report's versions lists.
func extractMilestoneAndVersion(versions []*buildPB.BuildReport_BuildConfig_Version) (int64, string, error) {
	var err error
	milestone := int64(0)
	version := ""

	// Extract milestone and platform version from the versions list.
	for _, versionInfo := range versions {
		switch versionInfo.Kind {
		case buildPB.BuildReport_BuildConfig_VERSION_KIND_MILESTONE:
			milestone, err = strconv.ParseInt(versionInfo.Value, 10, 64)
			if err != nil {
				return 0, "", err
			}
		case buildPB.BuildReport_BuildConfig_VERSION_KIND_PLATFORM:
			version = versionInfo.Value

		}
	}

	return milestone, version, nil
}

// extractImagePath returns the GCS path for the image.zip from the report's
// artifact list.
func extractImagePath(artifacts []*buildPB.BuildReport_BuildArtifact) (string, error) {
	// Fetch the GCS path to the created image.
	for _, artifact := range artifacts {
		switch artifact.Type.String() {
		case "IMAGE_ZIP":
			return artifact.Uri.GetGcs(), nil
		}
	}

	return "", fmt.Errorf("no imagePath found in artifacts")
}

// extractBoardAndVariant will extract the board and potential variant from the
// build target.
func extractBoardAndVariant(buildTarget string) (string, string, error) {
	board := ""
	variant := ""
	// amd64-generic is a unique board which has a hyphen in its board name. If
	// more boards begin to follow this pattern we may want to turn this into a
	// tracking list.
	if buildTarget == "amd64-generic" || buildTarget == "fizz-labstation" {
		board = buildTarget
	} else if !strings.Contains(buildTarget, "-") && strings.HasSuffix(buildTarget, "64") {
		board = buildTarget[:len(buildTarget)-2]
		variant = "64"
	} else {
		before, after, didCut := strings.Cut(buildTarget, "-")
		board = before
		if didCut {
			variant = after
		}
	}

	if board == "" {
		return "", "", fmt.Errorf("no board provided in build target")
	}

	return board, variant, nil
}

// transformReportToKronBuild takes a build report and returns all relevant
// builds in a Kron parsable form.
func transformReportToKronBuild(report *buildPB.BuildReport) (*kronpb.Build, error) {
	milestone, version, err := extractMilestoneAndVersion(report.Config.Versions)
	if err != nil {
		return nil, fmt.Errorf("%d: %w", report.GetBuildbucketId(), err)
	}

	imagePath, err := extractImagePath(report.Artifacts)
	if err != nil {
		return nil, fmt.Errorf("%d: %w", report.GetBuildbucketId(), err)
	}

	board, variant, err := extractBoardAndVariant(report.Config.Target.Name)
	if err != nil {
		return nil, fmt.Errorf("%d: %w", report.GetBuildbucketId(), err)
	}

	return &kronpb.Build{
		BuildUuid:   uuid.NewString(),
		RunUuid:     metrics.GetRunID(),
		CreateTime:  common.TimestamppbNowWithoutNanos(),
		Bbid:        report.GetBuildbucketId(),
		BuildTarget: report.Config.Target.Name,
		Milestone:   milestone,
		Version:     version,
		ImagePath:   imagePath,
		Board:       board,
		Variant:     variant}, nil
}

// validateReport checks that all necessary fields are not nil.
func validateReport(report *buildPB.BuildReport) error {
	reportBBID := report.GetBuildbucketId()
	if report.Config == nil {
		return fmt.Errorf("report for go/bbid/%d contains a nil config", reportBBID)
	}

	if report.Config.Target == nil {
		return fmt.Errorf("report for go/bbid/%d contains a nil build target", reportBBID)
	}

	if report.Status == nil {
		return fmt.Errorf("report for go/bbid/%d contains a nil status field", reportBBID)
	}
	return nil
}

// processPSMessage is called within the Pubsub receive callback to process the
// contents of the message.
func (h *handler) processPSMessage(msg *cloudPubsub.Message) error {
	// Unmarshall the raw data into the BuildReport format.
	buildReport := buildPB.BuildReport{}
	// google.golang.org/protobuf/proto specifically needs to be used here to
	// handle the format of proto serialization being done from the recipes
	// builder.
	err := proto.Unmarshal(msg.Data, &buildReport)
	if err != nil {
		return err
	}
	if err := validateReport(&buildReport); err != nil {
		// Ack the invalid report because it will just sit in the queue otherwise.
		msg.Ack()
		common.Stderr.Println(err)
		return nil
	}

	// Check for a successful release build. Ignore all types of reports.
	if !(buildReport.Type == buildPB.BuildReport_BUILD_TYPE_RELEASE && buildReport.Status.Value.String() == "SUCCESS") {
		msg.Ack()
		return nil
	}

	common.Stdout.Printf("Processing build report for go/bbid/%d\n", buildReport.GetBuildbucketId())
	// Ingest the report and return all kron usable builds.
	kronBuild, err := transformReportToKronBuild(&buildReport)
	if err != nil {
		return err
	}

	// Store build locally for NEW_BUILD configs.
	// NOTE: We are using a channel here because this function will only be
	// called asynchronously via goroutines.
	wrappedBuild := &BuildPackage{
		Build:   kronBuild,
		Message: msg,
	}

	h.buildsChan <- wrappedBuild
	return nil
}

type handler struct {
	buildsChan chan *BuildPackage
}

type EventWrapper struct {
	Event      *kronpb.Event
	CtpRequest *requestpb.Request
}

type ConfigDetails struct {
	Config *suschpb.SchedulerConfig
	// NOTE: Events is a list because multiple requests can be made if the
	// config targets multiple models for the given build target.
	Events []*EventWrapper
}

type BuildPackage struct {
	Build            *kronpb.Build
	Message          *cloudPubsub.Message
	TriggeredConfigs []*ConfigDetails
}

// publishBuild uploads each build information proto to a pubsub queue
// to later be sent to BigQuery.
func publishBuild(build *BuildPackage, client pubsub.PublishClient) error {
	common.Stdout.Printf("Publishing build %s for build target %s and milestone %d to pub sub", build.Build.BuildUuid, build.Build.BuildTarget, build.Build.Milestone)
	data, err := protojson.Marshal(build.Build)
	if err != nil {
		return err
	}

	err = client.PublishMessage(context.Background(), data)
	if err != nil {
		return err
	}
	common.Stdout.Printf("Published build %s for build target %s and milestone %d to pub sub", build.Build.BuildUuid, build.Build.BuildTarget, build.Build.Milestone)

	return nil
}

// IngestBuildsFromPubSub connects to pubsub ingests all new build information
// from the releases Pub/Sub stream. Once read, all builds will be written into
// long term storage.
func IngestBuildsFromPubSub(projectID, subscriptionName string) ([]*BuildPackage, error) {
	psHandler := handler{
		buildsChan: make(chan *BuildPackage),
	}

	builds := []*BuildPackage{}

	common.Stdout.Printf("Initializing client for pub sub topic %s on project %s", common.BuildsPubSubTopic, projectID)
	client, err := pubsub.InitPublishClient(context.Background(), projectID, common.BuildsPubSubTopic)
	if err != nil {
		return nil, err
	}

	// Spin up a goroutine to handle the incoming messages to the channel
	// buffer.
	// NOTE: non-buffered channels in GO require that a ready consumer is
	// receiving before any messages can be passed in. If this is launched after
	// we begin sending messages into the channel the application will hang in a
	// deadlock.
	var wait sync.WaitGroup
	wait.Add(1)
	go func(builds *[]*BuildPackage, wg *sync.WaitGroup, buildsChan chan *BuildPackage, client pubsub.PublishClient) {
		defer wg.Done()
		for build := range buildsChan {
			//  We need to publish the messages here to pubsub.
			if err := publishBuild(build, client); err != nil {
				// If we failed to republish the message then we should nack the
				// build to be ingested again.
				build.Message.Nack()
				continue
			}

			*builds = append(*builds, build)

			// Ack the message once it has been correctly republished to our
			// metrics for analysis.
			// TODO: when we store build information in psql we will need move
			// where this ack call is made.
			build.Message.Ack()

		}
	}(&builds, &wait, psHandler.buildsChan, client)

	// Initialize the custom pubsub receiver. This custom handler implements a
	// timeout feature which will stop the pubsub Receive() call once no more
	// messages are incoming.
	common.Stdout.Println("Initializing Pub/Sub Receive Client")
	ctx := context.Background()
	receiveClient, err := pubsub.InitReceiveClientWithTimer(ctx, projectID, subscriptionName, psHandler.processPSMessage)
	if err != nil {
		return nil, err
	}

	// NOTE: This is a blocking receive call. This will end when the child
	// context in the ReceiveClient expires due to no messages incoming.
	common.Stdout.Println("Pulling messages from Pub/Sub Queue")
	err = receiveClient.PullMessages()
	// Close the channel be fore error handling, so that the goroutine finishes
	// and does not hang.
	close(psHandler.buildsChan)
	if err != nil {
		return nil, err
	}

	// Wait for the buffer receive function to end.
	wait.Wait()
	common.Stdout.Printf("Returning %d Builds from Pub/Sub feed\n", len(builds))

	return builds, nil
}
