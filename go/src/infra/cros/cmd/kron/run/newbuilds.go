// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package run holds all of the internal logic for the execution steps of a
// SuiteScheduler run.
package run

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"

	kronpb "go.chromium.org/chromiumos/infra/proto/go/test_platform/kron"
	suschpb "go.chromium.org/chromiumos/infra/proto/go/testplans"
	"go.chromium.org/luci/auth/client/authcli"

	"infra/cros/cmd/kron/builds"
	"infra/cros/cmd/kron/cloudsql"
	"infra/cros/cmd/kron/common"
	"infra/cros/cmd/kron/configparser"
	"infra/cros/cmd/kron/pubsub"
	"infra/cros/cmd/kron/totmanager"
)

const (
	disallowPublishErrors = false
	publishEventsToPubSub = true
)

// CrOSNewBuildCommand implements NewBuildCommand.
type CrOSNewBuildCommand struct {
	authOpts *authcli.Flags
	isProd   bool
	isTest   bool
	dryRun   bool

	labConfigs            *configparser.LabConfigs
	suiteSchedulerConfigs *configparser.SuiteSchedulerConfigs

	projectID string

	// Used to store builds made in the finalize step to reduce duplicate
	// generation.
	kronBuilds []*kronpb.Build
}

// InitCrOSNewBuildCommand generates and returns a CrOS NEW_BUILD client which
// implements the NewBuildCommand interface. This client does not handle
// firmware, Android, nor multi-DUT configs.
func InitCrOSNewBuildCommand(authOpts *authcli.Flags, isProd, dryRun, isTest bool, labConfigs *configparser.LabConfigs, suiteSchedulerConfigs *configparser.SuiteSchedulerConfigs, projectID string) NewBuildCommand {
	return &CrOSNewBuildCommand{
		authOpts:              authOpts,
		isProd:                isProd,
		dryRun:                dryRun,
		isTest:                isTest,
		labConfigs:            labConfigs,
		suiteSchedulerConfigs: suiteSchedulerConfigs,
		projectID:             projectID,
	}
}

// Name returns the custom name of the command. This will be used in logging.
func (c *CrOSNewBuildCommand) Name() string {
	return "CrOSNewBuilds"
}

// publishBuild uploads each build information proto to our long term storage
// PSQL database and our Pub/Sub metrics pipeline.
//
// NOTE: We will attempt to write the build message to the PSQL DB before we try
// uploading to pubsub. Since the BuildUUID is a hash, we will not be able to
// upload the build twice.
func publishBuild(ctx context.Context, kronBuild *kronpb.Build, psClient pubsub.PublishClient, sqlClient cloudsql.Client) error {
	common.Stdout.Printf("Publishing build %s for build target %s and milestone %d to long term storage", kronBuild.BuildUuid, kronBuild.BuildTarget, kronBuild.Milestone)

	// Convert the build to a PSQL compatible type.
	psqlBuild, err := cloudsql.ConvertBuildToPSQLRow(kronBuild)
	if err != nil {
		return err
	}

	// Insert the row into Cloud SQL PSQL.
	_, err = sqlClient.Exec(ctx, cloudsql.InsertBuildsTemplate, cloudsql.RowToSlice(psqlBuild)...)
	if err != nil {
		return err
	}
	common.Stdout.Printf("Published build %s for build target %s and milestone %d to PSQL", kronBuild.BuildUuid, kronBuild.BuildTarget, kronBuild.Milestone)

	// Publish the build to Pub/Sub.
	data, err := protojson.Marshal(kronBuild)
	if err != nil {
		return err
	}
	err = psClient.PublishMessage(ctx, data)
	if err != nil {
		return err
	}
	common.Stdout.Printf("Published build %s for build target %s and milestone %d to pub sub", kronBuild.BuildUuid, kronBuild.BuildTarget, kronBuild.Milestone)

	return nil
}

// publishBuildReports generates and publishes a Kron Build message to Pub/Sub
// for each of the successful release builds ingested.
//
// NOTE: To minimize function loss on flaky network issues, each publish action
// is hermetic and will not cause program halt on publishing errors.
//
// NOTE: This function is being handed to the pub/sub ingestion logic as the
// finalize() command.
func (c *CrOSNewBuildCommand) publishBuildReports(buildReports *[]*builds.BuildReportPackage) error {
	ctx := context.Background()
	// Exit early if no buildReports were received from the Pub/Sub queue. This
	// is not an error, it just means that all builds have completed for the day
	// or are in flight.
	if len(*buildReports) == 0 {
		return nil
	}

	common.Stdout.Printf("Initializing client for pub sub topic %s on project %s", common.BuildsPubSubTopic, c.projectID)
	psClient, err := pubsub.InitPublishClient(ctx, c.projectID, common.BuildsPubSubTopic)
	if err != nil {
		return err
	}

	// Initialize PSQL client for long term storage insertion.
	sqlClient, err := cloudsql.InitBuildsClient(ctx, c.isProd, true)
	if err != nil {
		return err
	}

	// Transform build reports to Kron Builds and publish to Pub/Sub.
	publishedReports := []*builds.BuildReportPackage{}
	for _, report := range *buildReports {
		kronBuild, err := builds.TransformReportToKronBuild(report.Report)
		if err != nil {
			return err
		}

		// If we are running in test mode then do not publish build
		// information to the metrics pipeline and Nack all received messages.
		if c.isTest {
			report.Message.Nack()
			// Add to the list of Kron builds to be used in this Kron run.
			c.kronBuilds = append(c.kronBuilds, kronBuild)

			// publishedReports will be used to replace the passed in buildReports
			// at the end. This is because this functions is supposed to change the
			// values of the slice in-place rather than via return.
			publishedReports = append(publishedReports, report)
			continue
		}

		// Publish the kron build to the metrics pipeline and LTS PSQL storage.
		if err = publishBuild(ctx, kronBuild, psClient, sqlClient); err != nil {
			common.Stderr.Println(err)
			// If we failed to republish the message then we should nack the
			// build to be ingested again on the next Kron invocation.
			report.Message.Nack()
			continue
		}

		// The build was successfully retrieved by the metrics pipeline. Ack the
		// message to remove it from the Pub/Sub queue.
		report.Message.Ack()

		// Add to the list of Kron builds to be used in this Kron run.
		c.kronBuilds = append(c.kronBuilds, kronBuild)

		// publishedReports will be used to replace the passed in buildReports
		// at the end. This is because this functions is supposed to change the
		// values of the slice in-place rather than via return.
		publishedReports = append(publishedReports, report)
	}

	// If no kronBuilds were made then that means all publish attempts failed.
	// we will want to halt the run as a deep issue is going on.
	//
	// NOTE: We check to make sure that the provided buildReports value is non
	// nil at the start of the function. If that check is removed then this will
	// cause failures during the quieter periods of the day.
	if len(c.kronBuilds) == 0 {
		return fmt.Errorf("all builds failed to publish")
	}

	// Set the slice to the newly published slice.
	//
	// NOTE: the reason that we are not returning this values is because we are
	// modifying the slice in-place as this is being handled in a goroutine.
	*buildReports = publishedReports
	return nil
}

// FetchBuilds retrieves all builds currently sitting in the release team's
// completed build Pub/Sub queue. We then convert each valid report to a kron
// build for later use. Publishing to the metrics pipeline is performed here as
// well.
func (c *CrOSNewBuildCommand) FetchBuilds() ([]*kronpb.Build, error) {
	// Fetch BuildReports from the Release Pub/Sub firehose.
	common.Stdout.Println("Fetching builds from Pub/Sub.")

	// If we are in test mode then pull from the testing Pub/Sub subscription
	// where we do not ACK messages.
	subscriptionID := common.BuildsSubscription
	if c.isTest {
		subscriptionID = common.BuildsSubscriptionTesting
	}

	// NOTE: We are ignoring the response from this function because our helper
	// function gives us the list of Kron builds as a struct field.
	_, err := builds.IngestBuildsFromPubSub(c.projectID, subscriptionID, c.isProd, c.publishBuildReports)
	if err != nil {
		return nil, err

	}

	return c.kronBuilds, nil
}

// filterUnmigratedConfigs checks all configs to be ran and removes configs
// which have not migrated to Kron yet.
//
// TODO(b/338128764): Remove once we are fully migrated to Kron.
func filterUnmigratedConfigs(buildToConfigsMap map[*kronpb.Build][]*suschpb.SchedulerConfig) map[*kronpb.Build][]*suschpb.SchedulerConfig {
	filteredMap := map[*kronpb.Build][]*suschpb.SchedulerConfig{}

	common.Stdout.Println("Filtering out SuSch configs not on migration allowlist.")
	for build, configList := range buildToConfigsMap {
		filteredList := filterConfigs(configList)

		// If the filtered list returns empty log a notice and continue.
		if len(filteredList) == 0 {
			common.Stdout.Printf("Build %s of buildTarget %s had all it's triggered configs filtered from by migration rules.", build.BuildUuid, build.BuildTarget)
			continue
		}

		filteredMap[build] = filteredList
	}
	return filteredMap
}

// FetchTriggeredConfigs takes in a list of kron builds and finds which
// SuiteScheduler Configs they trigger. This is then organized into a map to be
// used by the next stage in the pipeline.
func (c *CrOSNewBuildCommand) FetchTriggeredConfigs(kronBuilds []*kronpb.Build) (map[*kronpb.Build][]*suschpb.SchedulerConfig, error) {
	// Build the list of all configs triggered by the ingested build images.
	common.Stdout.Println("Gathering all configs triggered from retrieved build images.")

	// Group the list of configs by the kron build which triggered them. This
	// will save us time later on recomputing which config needs what builds.
	//
	// NOTE: While the build is unique as it is the map key, the configs may be
	// found in multiple map buckets. This is because each config likely targets
	// multiple build targets.
	buildToConfigsMap := map[*kronpb.Build][]*suschpb.SchedulerConfig{}
	for _, build := range kronBuilds {
		// Gather all configs which are triggered by the current builds
		// buildTarget.
		//
		// NOTE: This cache is formed at the beginning of the run when we ingest
		// the ToT SuiteScheduler configs.
		configs := c.suiteSchedulerConfigs.FetchNewBuildConfigsByBuildTarget(configparser.BuildTarget(build.BuildTarget))

		// Iterate through the triggered configs and verify that they should be
		// triggered in this run.
		for _, config := range configs {
			// If the build's milestone did not match the config's targeted
			// branches then do not add this config to the build's to run list.
			targeted, _, err := totmanager.IsTargetedBranch(int(build.Milestone), config.Branches)
			if err != nil {
				return nil, err
			}
			if !targeted {
				common.Stdout.Printf("Config %s did not match milestone %d for buildTarget %s on build %s\n", config.Name, build.Milestone, build.BuildTarget, build.BuildUuid)
				continue
			}
			common.Stdout.Printf("Config %s matched with build %s for buildTarget %s and milestone %d", config.Name, build.BuildUuid, build.BuildTarget, build.Milestone)

			// If this is the first entry to the map create a list that we can
			// append to.
			if _, ok := buildToConfigsMap[build]; !ok {
				buildToConfigsMap[build] = []*suschpb.SchedulerConfig{}
			}

			buildToConfigsMap[build] = append(buildToConfigsMap[build], config)
		}
	}

	buildToConfigsMap = filterUnmigratedConfigs(buildToConfigsMap)

	common.Stdout.Printf("%d builds being sent", len(buildToConfigsMap))
	return buildToConfigsMap, nil
}

// ScheduleRequests generates CTP Requests, batches them into BuildBucket
// requests, and Schedules them via the BuildBucket API.
func (c *CrOSNewBuildCommand) ScheduleRequests(kronBuildMap map[*kronpb.Build][]*suschpb.SchedulerConfig) error {
	return scheduleRequests(kronBuildMap, c.suiteSchedulerConfigs, c.authOpts, c.projectID, c.isProd, c.dryRun)
}
