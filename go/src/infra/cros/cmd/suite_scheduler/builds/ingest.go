// Copyright 2023 The Chromium Authors
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
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	buildPB "go.chromium.org/chromiumos/infra/proto/go/chromiumos"
	suschPB "go.chromium.org/chromiumos/infra/proto/go/test_platform/suite_scheduler/v15"

	"infra/cros/cmd/suite_scheduler/common"
	"infra/cros/cmd/suite_scheduler/metrics"
	"infra/cros/cmd/suite_scheduler/pubsub"
)

const projectID = "google.com:suite-scheduler-staging"
const subscriptionName = "chromeos-builds-all"

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
	if buildTarget == "amd64-generic" {
		board = buildTarget
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

// extractModels returns all models that the image applies towards.
func extractModels(reportModels []*buildPB.BuildReport_BuildConfig_Model, board string) []string {
	models := []string{}

	for _, model := range reportModels {
		models = append(models, model.Name)
	}

	// If no models were provided use the board name in it's place. This is odd
	// but it is standard convention in the area. E.g. brya-brya.
	if len(models) == 0 {
		models = append(models, board)
	}

	return models
}

// generateBuildInfoPerModel returns a unique BuildInformation item per model.
func generateBuildInfoPerModel(buildTarget, version, imagePath, board, variant string, milestone, bbid int64, models []string) []*suschPB.BuildInformation {
	createTime := timestamppb.Now()

	rows := []*suschPB.BuildInformation{}
	for _, model := range models {
		buildRow := suschPB.BuildInformation{
			BuildUid:    &suschPB.UID{Id: uuid.NewString()},
			RunUid:      &suschPB.UID{Id: metrics.GetRunID().Id},
			CreateTime:  createTime,
			Bbid:        bbid,
			BuildTarget: buildTarget,
			Milestone:   milestone,
			Version:     version,
			ImagePath:   imagePath,
			Board:       board,
			Variant:     variant,
			Model:       model,
		}
		rows = append(rows, &buildRow)
	}

	return rows
}

// transformReportToSuSchBuilds takes a build report and returns all relevant
// builds in a SuiteScheduler parsable form.
func transformReportToSuSchBuilds(report *buildPB.BuildReport) ([]*suschPB.BuildInformation, error) {
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

	models := extractModels(report.Config.Models, board)

	rows := generateBuildInfoPerModel(report.Config.Target.Name, version, imagePath, board, variant, milestone, report.GetBuildbucketId(), models)

	return rows, nil
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
		return err
	}

	// Check for a successful release build. Ignore all types of reports.
	if !(buildReport.Type == buildPB.BuildReport_BUILD_TYPE_RELEASE && buildReport.Status.Value.String() == "SUCCESS") {
		// TODO(b/315340446): Switch to ACK() once we have this sending
		// information to long term storage.
		msg.Nack()
		return nil
	}

	common.Stdout.Printf("Processing build report for go/bbid/%d\n", buildReport.GetBuildbucketId())
	// Ingest the report and return all SuSch usable builds.
	rows, err := transformReportToSuSchBuilds(&buildReport)
	if err != nil {
		return err
	}
	common.Stdout.Printf("go/bbid/%d produced %d build images for SuiteScheduler\n", buildReport.GetBuildbucketId(), len(rows))

	// Store build locally for NEW_BUILD configs.
	// NOTE: We are using a channel here because this function will only be
	// called asynchronously via goroutines.
	for _, row := range rows {
		h.buildsChan <- row
	}

	return nil
}

type handler struct {
	buildsChan chan *suschPB.BuildInformation
}

// IngestBuildsFromPubSub connects to pubsub ingests all new build information
// from the releases Pub/Sub stream. Once read, all builds will be written into
// long term storage.
func IngestBuildsFromPubSub() ([]*suschPB.BuildInformation, error) {
	psHandler := handler{
		buildsChan: make(chan *suschPB.BuildInformation),
	}

	// Gather all newly processed build images.
	var wait sync.WaitGroup
	wait.Add(1)
	builds := []*suschPB.BuildInformation{}

	// Spin up a goroutine to handle the incoming messages to the channel
	// buffer.
	// NOTE: non-buffered channels in GO require that a ready consumer is
	// receiving before any messages can be passed in. If this is launched after
	// we begin sending messages into the channel the application will hang in a
	// deadlock.
	go func(builds *[]*suschPB.BuildInformation, wg *sync.WaitGroup, buildsChan chan *suschPB.BuildInformation) {
		defer wg.Done()
		for build := range buildsChan {
			*builds = append(*builds, build)
		}
	}(&builds, &wait, psHandler.buildsChan)

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
