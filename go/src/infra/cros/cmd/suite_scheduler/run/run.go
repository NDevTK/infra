// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package run holds all of the internal logic for the execution steps of a
// SuiteScheduler run.
package run

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"google.golang.org/protobuf/encoding/protojson"

	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
	v15 "go.chromium.org/chromiumos/infra/proto/go/test_platform/suite_scheduler/v15"
	"go.chromium.org/luci/auth/client/authcli"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"

	"infra/cros/cmd/suite_scheduler/buildbucket"
	"infra/cros/cmd/suite_scheduler/builds"
	"infra/cros/cmd/suite_scheduler/common"
	"infra/cros/cmd/suite_scheduler/configparser"
	"infra/cros/cmd/suite_scheduler/ctprequest"
	"infra/cros/cmd/suite_scheduler/metrics"
	"infra/cros/cmd/suite_scheduler/pubsub"
	"infra/cros/cmd/suite_scheduler/totmanager"
)

// fetchTriggeredDailyEvents returns all DAILY configs which are triggered at
// the current run's operating time. Logging is also wrapped within this function.
func fetchTriggeredDailyEvents(currTime common.SuSchTime, ingestedConfigs *configparser.SuiteSchedulerConfigs, configs *configparser.ConfigList) error {
	common.Stdout.Printf("Gathering DAILY configs triggered at hour %d\n", currTime.Hour)
	triggeredConfigs, err := ingestedConfigs.FetchDailyByHour(currTime.Hour)
	if err != nil {
		return err
	}

	common.Stdout.Printf("The following %d configs are triggered at hour %d:\n", len(triggeredConfigs), currTime.Hour)
	for _, config := range triggeredConfigs {
		common.Stdout.Printf("\t%s\n", config.Name)
	}

	*configs = append(*configs, triggeredConfigs...)
	return nil
}

// fetchTriggeredWeeklyEvents returns all WEEKLY configs which are triggered at
// the current run's operating time. Logging is also wrapped within this function.
func fetchTriggeredWeeklyEvents(currTime common.SuSchTime, ingestedConfigs *configparser.SuiteSchedulerConfigs, configs *configparser.ConfigList) error {
	common.Stdout.Printf("Gathering WEEKLY configs triggered at day %d hour %d\n", currTime.RegularDay, currTime.Hour)
	triggeredConfigs, err := ingestedConfigs.FetchWeeklyByDayHour(currTime.RegularDay, currTime.Hour)
	if err != nil {
		return err
	}

	common.Stdout.Printf("The following %d configs are triggered at day %d hour %d:\n", len(triggeredConfigs), currTime.RegularDay, currTime.Hour)
	for _, config := range triggeredConfigs {
		common.Stdout.Printf("\t%s\n", config.Name)
	}

	*configs = append(*configs, triggeredConfigs...)

	return nil
}

// fetchTriggeredFortnightlyEvents returns all FORTNIGHTLY configs which are triggered at
// the current run's operating time. Logging is also wrapped within this function.
func fetchTriggeredFortnightlyEvents(currTime common.SuSchTime, ingestedConfigs *configparser.SuiteSchedulerConfigs, configs *configparser.ConfigList) error {
	common.Stdout.Printf("Gathering FORTNIGHTLY configs triggered at day %d hour %d\n", currTime.FortnightDay, currTime.Hour)
	triggeredConfigs, err := ingestedConfigs.FetchFortnightlyByDayHour(currTime.FortnightDay, currTime.Hour)
	if err != nil {
		return err
	}

	common.Stdout.Printf("The following %d configs are triggered at day %d hour %d:\n", len(triggeredConfigs), currTime.FortnightDay, currTime.Hour)
	for _, config := range triggeredConfigs {
		common.Stdout.Printf("\t%s\n", config.Name)
	}

	*configs = append(*configs, triggeredConfigs...)
	return nil
}

// fetchTimedEvents gathers all timed event config which will are triggered at
// the provided time.
// NOTE: This function in conjunction with the SuSchTime struct handles
// fortnightly/weekly differences natively.
func fetchTimedEvents(currTime common.SuSchTime, ingestedConfigs *configparser.SuiteSchedulerConfigs) (configparser.ConfigList, error) {
	timedConfigs := configparser.ConfigList{}

	// Daily
	err := fetchTriggeredDailyEvents(currTime, ingestedConfigs, &timedConfigs)
	if err != nil {
		return nil, err
	}

	// Weekly
	err = fetchTriggeredWeeklyEvents(currTime, ingestedConfigs, &timedConfigs)
	if err != nil {
		return nil, err
	}

	// Fortnightly
	err = fetchTriggeredFortnightlyEvents(currTime, ingestedConfigs, &timedConfigs)
	if err != nil {
		return nil, err
	}

	return timedConfigs, nil
}

// TimedEvents fetches all configs which are are triggered at the current
// day:hour, fetches all relevant build images, and then schedules their
// subsequent CTP requests via BuildBucket.
// TODO(b/315340446 | b/319463660): This function cannot be completed till we have some sort
// of long term storage to fetch build image information from.
func TimedEvents() error {
	// Ingest lab configs into memory.
	// TODO(b/319273179): Implement option to pass in a local config as to bypass
	// network reliance.
	common.Stdout.Println("Fetch lab configs")
	labConfigs, err := configparser.FetchLabConfigs("")
	if err != nil {
		return err
	}

	// Ingest SuiteScheduler configs into memory.
	// TODO(b/319273179): Implement option to pass in a local config as to bypass
	// network reliance.
	common.Stdout.Println("Fetch SuSch configs")
	suiteSchedulerConfigs, err := configparser.FetchSchedulerConfigs("", labConfigs)
	if err != nil {
		return err
	}
	// Format time as SuSchTime for searching through the configs.
	// TODO(b/319463660): implement option for passed in time, similar to the
	// -start-time flag in the search CLI command.
	operatingTime := common.TimeToSuSchTime(time.Now())

	// Fetch all configs, from all TIMED_EVENT types, which are triggered at the
	// current operating time.
	timedEvents, err := fetchTimedEvents(operatingTime, suiteSchedulerConfigs)
	if err != nil {
		return err
	}

	// TODO(b/319463660): Calculate all images needed for triggered TIME_EVENT
	// configs.

	// TODO(b/315340446): Fetch the newest build image for each target option.

	// Build CTP requests
	ctpRequests := ctprequest.CTPRequests{}
	for _, event := range timedEvents {
		// TODO(b/315340446 | b/319463660): Once long term storage of build images is in place then
		// we can properly build CTP requests. Right now we are not passing in
		// target options nor build images.
		ctpRequests = append(ctpRequests, ctprequest.BuildAllCTPRequests(event, configparser.TargetOptions{})...)
	}

	// TODO(b/319463660): Schedule each CTP request via BB API.

	return nil
}

// publishBuilds uploads each build information proto to a pubsub queue
// to later be sent to BigQuery.
func publishBuilds(builds []*builds.BuildPackage) error {
	common.Stdout.Printf("Initializing client for pub sub topic %s", common.BuildsPubSubTopic)
	client, err := pubsub.InitPublishClient(context.Background(), common.StagingProjectID, common.BuildsPubSubTopic)
	if err != nil {
		return err
	}

	common.Stdout.Printf("Publishing %d builds to pub sub", len(builds))
	for _, build := range builds {

		data, err := protojson.Marshal(build.Build)
		if err != nil {
			return err
		}

		err = client.PublishMessage(context.Background(), data)
		if err != nil {
			return err
		}
		common.Stdout.Printf("Published %s build %s to Pub/Sub", build.Build.BuildTarget, build.Build.BuildUid.Id)
	}
	common.Stdout.Printf("Done publishing %d builds to pub sub", len(builds))

	return nil
}

// publishEvent sends the event message to Pub/Sub.
func publishEvent(client pubsub.PublishClient, event *v15.SchedulingEvent) error {
	data, err := protojson.Marshal(event)
	if err != nil {
		return err
	}

	err = client.PublishMessage(context.Background(), data)
	if err != nil {
		return err
	}

	return nil
}

// scheduleBatchViaBB batch schedules all CTP requests for the given build and
// handles their response.
func scheduleBatchViaBB(buildRequest *builds.BuildPackage, schedulerClient buildbucket.Scheduler, publishClient pubsub.PublishClient, wg *sync.WaitGroup) {
	// Release the WaitGroups lock on this function.
	defer wg.Done()

	// Batch Schedule all requests in the provided build.
	buildRequest.ShouldAck = true
	common.Stdout.Printf("Scheduling %d requests to CTP for build %s of board %s", len(buildRequest.Requests), buildRequest.Build.BuildUid, buildRequest.Build.Board)
	for _, request := range buildRequest.Requests {
		for _, event := range request.Events {

			response, err := schedulerClient.Schedule(event.CtpRequest, buildRequest.Build.BuildUid.Id, event.Event.EventUid.Id, event.Event.ConfigName)
			if err != nil {
				common.Stderr.Println(err)
				event.Event.Decision = &v15.SchedulingDecision{
					Type:         v15.DecisionType_UNKNOWN,
					Scheduled:    false,
					FailedReason: err.Error(),
				}
				common.Stdout.Printf("Event %s failed to scheduled for unknown reason", event.Event.EventUid.Id)

				buildRequest.ShouldAck = false
				// TODO: Decide if we want to fast fail on one error or keep going
				// on.
			}

			// TODO(b/309683890): Add better support for failure/infra_failure/cancelled.
			if response.Status == buildbucketpb.Status_SCHEDULED {
				event.Event.Decision = &v15.SchedulingDecision{
					Type:      v15.DecisionType_SCHEDULED,
					Scheduled: true,
				}

				// TODO(b/309683890): update the metric proto to just use 1 bbid
				event.Event.Bbids = []int64{response.Id}

				common.Stdout.Printf("Event %s scheduled at http://go/bbid/%d using buildId %s", event.Event.EventUid.Id, response.Id, buildRequest.Build.BuildUid.Id)
			} else {
				event.Event.Decision = &v15.SchedulingDecision{
					Type:         v15.DecisionType_UNKNOWN,
					Scheduled:    false,
					FailedReason: buildbucketpb.Status_name[int32(response.Status.Number())],
				}

				common.Stdout.Printf("Event %s failed to scheduled for unknown reason", event.Event.EventUid.Id)

				// TODO: decide on if we should nack the message on one failure or
				// not.
			}

			err = publishEvent(publishClient, event.Event)
			if err != nil {
				common.Stderr.Println(err)
			}
		}
	}

	// TODO(b/319276542 | b/319464677): Consider removing the Ack/Nack logic here and moving
	// it to after the DB Insertion would take place. To solve the issue of
	// failed schedules, we could implement a backfill feature
	if buildRequest.ShouldAck {
		buildRequest.Message.Nack()
	} else {
		common.Stderr.Printf("Nacking build message for build %s because one or more failed\n", buildRequest.Build.BuildUid.Id)
		buildRequest.Message.Nack()
	}
}

// fetchTriggeredNewBuildConfigs attaches all configs to the builds which
// triggered their run.
func fetchTriggeredNewBuildConfigs(buildPackages []*builds.BuildPackage, suiteSchedulerConfigs *configparser.SuiteSchedulerConfigs) error {
	for _, build := range buildPackages {
		configs := suiteSchedulerConfigs.FetchNewBuildConfigsByBuildTarget(configparser.BuildTarget(build.Build.BuildTarget))
		for _, config := range configs {
			// If the build's milestone did not match the config's targeted
			// branches then do not add this config to the build's to run list.
			targeted, err := totmanager.IsTargetedBranch(int(build.Build.Milestone), config.Branches)
			if err != nil {
				return err
			}
			if !targeted {
				common.Stdout.Printf("Config %s did not match milestone %d for buildTarget %s\n", config.Name, build.Build.Milestone, build.Build.BuildTarget)
				continue
			}

			request := &builds.ConfigDetails{
				Config: config,
			}

			build.Requests = append(build.Requests, request)
		}
	}

	return nil
}

// buildCTPRequests iterates through all the provided BuildPackages and
// generates BuildBucket CTP requests for all triggered configs.
func buildCTPRequests(buildPackages []*builds.BuildPackage, suiteSchedulerConfigs *configparser.SuiteSchedulerConfigs) error {
	// Iterate through the wrapped builds and insert CTP request and their
	// associated metrics events into the package.
	for _, wrappedBuild := range buildPackages {
		// Iterate through all
		for _, requests := range wrappedBuild.Requests {
			// Fetch the target options requested for the current board on the
			// current configuration.
			boardTargetOption, err := suiteSchedulerConfigs.FetchConfigTargetOptionsForBoard(requests.Config.Name, configparser.Board(wrappedBuild.Build.Board))
			if err != nil {
				return err
			}

			// If provided, build a CTP request per model, otherwise leave the model
			// absent.
			ctpRequests := []*test_platform.Request{}
			if len(boardTargetOption.Models) > 0 {
				// Generate a CTP Request for each board/model combo.
				for _, model := range boardTargetOption.Models {
					ctpRequests = append(ctpRequests, ctprequest.BuildCTPRequest(requests.Config, wrappedBuild.Build.Board, model, wrappedBuild.Build.BuildTarget, strconv.FormatInt(wrappedBuild.Build.Milestone, 10), wrappedBuild.Build.Version))
				}
			} else {
				ctpRequests = append(ctpRequests, ctprequest.BuildCTPRequest(requests.Config, wrappedBuild.Build.Board, "", wrappedBuild.Build.BuildTarget, strconv.FormatInt(wrappedBuild.Build.Milestone, 10), wrappedBuild.Build.Version))
			}

			// Pair all generated CTP Requests inside an event message to be
			// uploaded to pubsub.
			events := []*builds.EventWrapper{}
			for _, ctpRequest := range ctpRequests {
				event := builds.EventWrapper{
					CtpRequest: ctpRequest,
				}
				event.Event, err = metrics.GenerateEventMessage(requests.Config, nil, nil)
				if err != nil {
					return err
				}
				events = append(events, &event)
			}

			requests.Events = append(requests.Events, events...)
		}
	}
	return nil
}

// NewBuilds fetches all builds from the release Pub/Sub queue, finds all
// triggered NEW_BUILD configs, then builds their respective CTP requests for
// BB.
func NewBuilds(authOpts *authcli.Flags) error {
	// Ingest lab configs into memory.
	// TODO(b/319273179): Implement option to pass in a local config as to bypass
	// network reliance.
	common.Stdout.Println("Fetch lab configs")
	labConfigs, err := configparser.FetchLabConfigs("")
	if err != nil {
		return err
	}

	// Ingest SuiteScheduler configs into memory.
	// TODO(b/319273179): Implement option to pass in a local config as to bypass
	// network reliance.
	common.Stdout.Println("Fetch SuSch configs")
	suiteSchedulerConfigs, err := configparser.FetchSchedulerConfigs("", labConfigs)
	if err != nil {
		return err
	}

	// Get build information
	// TODO(b/315340446): Inside this client we need to not ACK the builds until
	// we can launch the BB task. This may require that we do a lot of heavy
	// lifting in the callback but overall it would be safer.
	common.Stdout.Println("Fetching builds from Pub/Sub.")
	releaseBuilds, err := builds.IngestBuildsFromPubSub()
	if err != nil {
		return err

	}

	// TODO(b/315340446): Write build info to long term storage(database)

	// Write all build information to pubsub
	common.Stdout.Println("Publishing fetched builds information to Pub/Sub.")
	err = publishBuilds(releaseBuilds)
	if err != nil {
		return err
	}

	// Build the list of all configs triggered by the ingested build images.
	//
	// TODO(TBD): For in run events, determine if we need to squash the
	// builds so that NEW_BUILD configs are only triggered by the newest images.
	// The created time is stored inside the build artifact type. Reach out to
	// release team to determine if this is required.
	// https://chromium.googlesource.com/chromiumos/infra/proto/+/refs/heads/main/src/chromiumos/build_report.proto#197
	common.Stdout.Println("Gathering all configs triggered from retrieved build images.")
	err = fetchTriggeredNewBuildConfigs(releaseBuilds, suiteSchedulerConfigs)
	if err != nil {
		return err
	}

	common.Stdout.Println("Filtering out SuSch configs not on migration allowlist.")
	releaseBuilds = filterConfigs(releaseBuilds)

	// Build CTP Requests for all triggered configs.
	err = buildCTPRequests(releaseBuilds, suiteSchedulerConfigs)
	if err != nil {
		return err
	}

	// TODO(b/319273876): Remove slow migration logic upon completion of
	// transition. Right now only the CFTNewBuild config on brya is
	// supported. Below checks ensure only one request can launch per run.
	if len(releaseBuilds) == 0 {
		return fmt.Errorf("no builds")
	}

	// TODO(b/319273876): For earliest stage of migration only launch 1 request.
	// Once we begin Acking messages from the chromeos build pipeline we can let
	// all requests flow through.
	releaseBuilds = releaseBuilds[:1]

	if len(releaseBuilds) > 1 {
		return fmt.Errorf("too many builds %d", len(releaseBuilds))
	}

	if len(releaseBuilds[0].Requests) == 0 {
		return fmt.Errorf("no configs")
	}

	if len(releaseBuilds[0].Requests[0].Events) == 0 {
		return fmt.Errorf("no requests")
	}

	if len(releaseBuilds[0].Requests) > 1 {
		return fmt.Errorf("too many requests %d", len(releaseBuilds[0].Requests))
	}

	// Initialize an authenticated BuildBucket client for scheduling.
	common.Stdout.Printf("Initializing BuildBucket scheduling client prod: %t dryrun: %t", false, false)
	schedulerClient, err := buildbucket.InitScheduler(context.Background(), authOpts, false, false)
	if err != nil {
		return err
	}

	// Initialize the Pub/Sub client for event message publishing.
	common.Stdout.Printf("Initializing client for pub sub topic %s", common.EventsPubSubTopic)
	publishClient, err := pubsub.InitPublishClient(context.Background(), common.StagingProjectID, common.EventsPubSubTopic)
	if err != nil {
		return err
	}

	// Introduce a wait group to hold for all spun out goroutines.
	var wg sync.WaitGroup

	// Schedule all requests via BuildBucket in parallel.
	// TODO(b/319273876): Remove slow migration logic upon completion of
	// transition.
	for _, wrappedBuild := range releaseBuilds {
		wg.Add(1)
		go scheduleBatchViaBB(wrappedBuild, schedulerClient, publishClient, &wg)
	}

	common.Stdout.Println("Waiting for batched requests to finish scheduling...")
	wg.Wait()
	common.Stdout.Println("NEW_BUILD scheduling completed")

	return nil
}
