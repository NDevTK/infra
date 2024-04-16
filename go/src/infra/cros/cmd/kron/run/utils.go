// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package run

import (
	"context"
	"strconv"
	"sync"

	"google.golang.org/protobuf/encoding/protojson"

	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
	kronpb "go.chromium.org/chromiumos/infra/proto/go/test_platform/kron"
	suschpb "go.chromium.org/chromiumos/infra/proto/go/testplans"
	"go.chromium.org/luci/auth/client/authcli"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"

	"infra/cros/cmd/kron/buildbucket"
	"infra/cros/cmd/kron/builds"
	"infra/cros/cmd/kron/common"
	"infra/cros/cmd/kron/configparser"
	"infra/cros/cmd/kron/ctprequest"
	"infra/cros/cmd/kron/metrics"
	"infra/cros/cmd/kron/pubsub"
	"infra/cros/cmd/kron/totmanager"
)

// launchCTPRequests takes in build images, generates CTP requests, and launches
// them via BuildBucket.
func launchCTPRequests(fetchedBuilds []*builds.BuildPackage, suiteSchedulerConfigs *configparser.SuiteSchedulerConfigs, authOpts *authcli.Flags, projectID string, isProd, dryRun bool) error {
	// Build CTP Requests for all triggered configs.
	common.Stdout.Println("Generating CTP requests for all configs and builds")
	err := buildCTPRequests(fetchedBuilds, suiteSchedulerConfigs)
	if err != nil {
		return err
	}

	// CTP is bottlenecked by it's drone count. To combat this, combine tests
	// requests into one large CTP request.
	common.Stdout.Println("Grouping CTP request by config name")
	ctpRequests := combineCTPRequests(fetchedBuilds)

	// If staging reduce requests to 5 MAX.
	if !isProd {
		common.Stdout.Printf("Limiting CTP request to %d max because we are running in staging", common.StagingMaxRequests)
		ctpRequests = limitStagingRequests(ctpRequests)
	}

	// Initialize an authenticated BuildBucket client for scheduling.
	common.Stdout.Printf("Initializing BuildBucket scheduling client prod: %t dryRun: %t", isProd, dryRun)
	schedulerClient, err := buildbucket.InitScheduler(context.Background(), authOpts, isProd, dryRun)
	if err != nil {
		return err
	}
	// Initialize the Pub/Sub client for event message publishing.
	common.Stdout.Printf("Initializing client for pub sub topic %s", common.EventsPubSubTopic)
	publishClient, err := pubsub.InitPublishClient(context.Background(), projectID, common.EventsPubSubTopic)
	if err != nil {
		return err
	}

	// Introduce a wait group to hold for all spun out goroutines.
	var wg sync.WaitGroup

	// Schedule all requests via BuildBucket in parallel.
	common.Stdout.Println("Scheduling CTP request batches in parallel.")
	for configName, requestList := range ctpRequests {
		wg.Add(1)
		go scheduleBatchViaBB(requestList, configName, schedulerClient, publishClient, &wg)
	}

	common.Stdout.Println("Waiting for batched requests to finish scheduling...")
	wg.Wait()
	common.Stdout.Println("NEW_BUILD scheduling completed")

	return nil
}

// fetchTriggeredDailyEvents returns all DAILY configs which are triggered at
// the current run's operating time. Logging is also wrapped within this function.
func fetchTriggeredDailyEvents(currTime common.KronTime, ingestedConfigs *configparser.SuiteSchedulerConfigs, configs *configparser.ConfigList) error {
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
func fetchTriggeredWeeklyEvents(currTime common.KronTime, ingestedConfigs *configparser.SuiteSchedulerConfigs, configs *configparser.ConfigList) error {
	common.Stdout.Printf("Gathering WEEKLY configs triggered at day %d hour %d\n", currTime.WeeklyDay, currTime.Hour)
	triggeredConfigs, err := ingestedConfigs.FetchWeeklyByDayHour(currTime.WeeklyDay, currTime.Hour)
	if err != nil {
		return err
	}

	common.Stdout.Printf("The following %d configs are triggered at day %d hour %d:\n", len(triggeredConfigs), currTime.WeeklyDay, currTime.Hour)
	for _, config := range triggeredConfigs {
		common.Stdout.Printf("\t%s\n", config.Name)
	}

	*configs = append(*configs, triggeredConfigs...)

	return nil
}

// fetchTriggeredFortnightlyEvents returns all FORTNIGHTLY configs which are triggered at
// the current run's operating time. Logging is also wrapped within this function.
func fetchTriggeredFortnightlyEvents(currTime common.KronTime, ingestedConfigs *configparser.SuiteSchedulerConfigs, configs *configparser.ConfigList) error {
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
// NOTE: This function in conjunction with the KronTime struct handles
// fortnightly/weekly differences natively.
func fetchTimedEvents(currTime common.KronTime, ingestedConfigs *configparser.SuiteSchedulerConfigs) (configparser.ConfigList, error) {
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

// publishEvent sends the event message to Pub/Sub.
func publishEvent(client pubsub.PublishClient, event *kronpb.Event) error {
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

// fetchTriggeredNewBuildConfigs attaches all configs to the builds which
// triggered their run.
func fetchTriggeredNewBuildConfigs(buildPackages []*builds.BuildPackage, suiteSchedulerConfigs *configparser.SuiteSchedulerConfigs) error {
	for _, build := range buildPackages {
		configs := suiteSchedulerConfigs.FetchNewBuildConfigsByBuildTarget(configparser.BuildTarget(build.Build.BuildTarget))
		for _, config := range configs {
			// If the build's milestone did not match the config's targeted
			// branches then do not add this config to the build's to run list.
			targeted, _, err := totmanager.IsTargetedBranch(int(build.Build.Milestone), config.Branches)
			if err != nil {
				return err
			}
			if !targeted {
				common.Stdout.Printf("Config %s did not match milestone %d for buildTarget %s on build %s\n", config.Name, build.Build.Milestone, build.Build.BuildTarget, build.Build.BuildUuid)
				continue
			}
			common.Stdout.Printf("Config %s matched with build %s for buildTarget %s and milestone %d", config.Name, build.Build.BuildUuid, build.Build.BuildTarget, build.Build.Milestone)

			request := &builds.ConfigDetails{
				Config: config,
			}

			build.TriggeredConfigs = append(build.TriggeredConfigs, request)
		}
	}

	return nil
}

// wrapEvent wraps and returns a package containing a ctp request and it's newly
// generated event metrics message.
func wrapEvent(ctpRequest *test_platform.Request, config *suschpb.SchedulerConfig, buildUUID, board, model string) (*builds.EventWrapper, error) {
	var err error
	event := &builds.EventWrapper{
		CtpRequest: ctpRequest,
	}
	event.Event, err = metrics.GenerateEventMessage(config, nil, 0, buildUUID, board, model)
	if err != nil {
		return nil, err
	}

	return event, nil

}

// buildPerModelConfigs builds a CTP request per model (if it exists) for the
// given config.
func buildPerModelConfigs(models []string, config *suschpb.SchedulerConfig, build *kronpb.Build, branch string) ([]*builds.EventWrapper, error) {
	events := []*builds.EventWrapper{}
	// If provided, build a CTP request per model, otherwise leave the model
	// field absent.
	if len(models) > 0 {
		// Generate a CTP Request for each board/model combo.
		for _, model := range models {
			request := ctprequest.BuildCTPRequest(config, build.Board, model, build.BuildTarget, strconv.FormatInt(build.Milestone, 10), build.Version, branch)
			event, err := wrapEvent(request, config, build.BuildUuid, build.Board, model)
			if err != nil {
				return nil, err
			}
			events = append(events, event)
		}
	} else {
		request := ctprequest.BuildCTPRequest(config, build.Board, "", build.BuildTarget, strconv.FormatInt(build.Milestone, 10), build.Version, branch)
		event, err := wrapEvent(request, config, build.BuildUuid, build.Board, "")
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

// buildConfigEventsPerModel generates an event message for the current config,
// board, model(s).
func buildConfigEventsPerModel(models []string, config *suschpb.SchedulerConfig, board, buildUUID string, schedulingDecision *kronpb.SchedulingDecision) ([]*kronpb.Event, error) {
	events := []*kronpb.Event{}

	if len(models) > 0 {
		for _, model := range models {
			event, err := metrics.GenerateEventMessage(config, schedulingDecision, 0, buildUUID, board, model)
			if err != nil {
				return nil, err
			}
			events = append(events, event)
		}
	} else {
		event, err := metrics.GenerateEventMessage(config, schedulingDecision, 0, buildUUID, board, "")
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

// buildCTPRequests iterates through all the provided BuildPackages and
// generates BuildBucket CTP requests for all triggered configs.
func buildCTPRequests(buildPackages []*builds.BuildPackage, suiteSchedulerConfigs *configparser.SuiteSchedulerConfigs) error {
	// Iterate through the wrapped builds and insert CTP request and their
	// associated metrics events into the package.
	for _, wrappedBuild := range buildPackages {
		// Iterate through all
		for _, triggeredConfig := range wrappedBuild.TriggeredConfigs {
			// Fetch the target options requested for the current board on the
			// current configuration.
			boardTargetOption, err := suiteSchedulerConfigs.FetchConfigTargetOptionsForBoard(triggeredConfig.Config.Name, configparser.Board(wrappedBuild.Build.Board))
			if err != nil {
				return err
			}

			// Get get the branch target which this build matched with.
			_, branch, err := totmanager.IsTargetedBranch(int(wrappedBuild.Build.Milestone), triggeredConfig.Config.Branches)
			if err != nil {
				return err
			}

			events, err := buildPerModelConfigs(boardTargetOption.Models, triggeredConfig.Config, wrappedBuild.Build, suschpb.Branch_name[int32(branch)])
			if err != nil {
				return err
			}

			triggeredConfig.Events = append(triggeredConfig.Events, events...)
		}
	}
	return nil
}

// combineCTPRequests fetches all CTP Requests from inside the puck packages and
// groups them by SuSch config.
func combineCTPRequests(releaseBuilds []*builds.BuildPackage) map[string][]*builds.EventWrapper {
	configMap := map[string][]*builds.EventWrapper{}

	// Iterate through all CTP Requests
	for _, build := range releaseBuilds {
		for _, request := range build.TriggeredConfigs {
			if _, ok := configMap[request.Config.Name]; !ok {
				configMap[request.Config.Name] = []*builds.EventWrapper{}
			}

			configMap[request.Config.Name] = append(configMap[request.Config.Name], request.Events...)
		}
	}

	return configMap
}

// sendBatch sends the request batch to the BuildBucket client, attaches the
// scheduling responses to the events, and finally publishes all events to
// Pub/Sub.
func sendBatch(configName string, schedulerClient buildbucket.Scheduler, publishClient pubsub.PublishClient, requestBatch []*builds.EventWrapper) {
	// Schedule the builds.
	// TODO(b/309683890): Add better support for failure/infra_failure/cancelled.
	response, err := schedulerClient.Schedule(requestBatch, configName)
	if err != nil {
		for _, sentEvent := range requestBatch {
			sentEvent.Event.Decision = &kronpb.SchedulingDecision{
				Type:         kronpb.DecisionType_UNKNOWN,
				Scheduled:    false,
				FailedReason: err.Error(),
			}
			common.Stderr.Printf("Event %s failed to schedule: %s", sentEvent.Event.EventUuid, err)
		}
	}

	// Populate scheduling status field.
	for _, request := range requestBatch {
		if response.Status == buildbucketpb.Status_SCHEDULED {
			request.Event.Decision = &kronpb.SchedulingDecision{
				Type:      kronpb.DecisionType_SCHEDULED,
				Scheduled: true,
			}

			request.Event.Bbid = response.Id

			common.Stdout.Printf("Event %s for config %s scheduled at http://go/bbid/%d using buildId %s", request.Event.EventUuid, request.Event.ConfigName, response.Id, request.Event.BuildUuid)
		} else {
			request.Event.Decision = &kronpb.SchedulingDecision{
				Type:         kronpb.DecisionType_UNKNOWN,
				Scheduled:    false,
				FailedReason: buildbucketpb.Status_name[int32(response.Status.Number())],
			}

			common.Stdout.Printf("Event %s failed to schedule for unknown reason", request.Event.EventUuid)
		}

		// Publish the events that just got sent.
		err = publishEvent(publishClient, request.Event)
		if err != nil {
			common.Stderr.Println(err)
		}
	}
}

// scheduleBatchViaBB batches the requests according to common.MultirequestSize
// and sends it off to be scheduled via BuildBucket.
func scheduleBatchViaBB(requests []*builds.EventWrapper, configName string, schedulerClient buildbucket.Scheduler, publishClient pubsub.PublishClient, wg *sync.WaitGroup) {
	defer wg.Done()
	if len(requests) == 0 {
		common.Stdout.Println("No requests passed into scheduleBatchViaBB()")
		return
	}

	batchRequestList := []*builds.EventWrapper{}
	for _, request := range requests {
		if len(batchRequestList) == common.MultirequestSize {
			// Schedule the builds.
			sendBatch(configName, schedulerClient, publishClient, batchRequestList)

			//  Reset the running batch.
			batchRequestList = []*builds.EventWrapper{}
		}

		batchRequestList = append(batchRequestList, request)
	}

	// Send the remaining requests when
	// len(batchRequestList) % common.MultirequestSize != 0
	sendBatch(configName, schedulerClient, publishClient, batchRequestList)

}

// limitStagingRequests scrubs outs ctp requests to ensure that only
// common.StagingMaxRequests maximum requests can be sent to CTP-staging. This
// limits the pressure that kron places on our staging pools while allowing us a
// functional staging environment.
func limitStagingRequests(requestMap map[string][]*builds.EventWrapper) map[string][]*builds.EventWrapper {
	if requestMap == nil {
		return nil
	}

	returnMap := map[string][]*builds.EventWrapper{}
	count := 0
	for configName, requestList := range requestMap {
		if count == common.StagingMaxRequests {
			break
		}

		returnMap[configName] = []*builds.EventWrapper{}
		for _, request := range requestList {
			if count == common.StagingMaxRequests {
				break
			}

			returnMap[configName] = append(returnMap[configName], request)
			count += 1
		}
	}

	return returnMap
}
