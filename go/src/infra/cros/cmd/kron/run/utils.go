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
			targeted, err := totmanager.IsTargetedBranch(int(build.Build.Milestone), config.Branches)
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
				event.Event, err = metrics.GenerateEventMessage(requests.Config, nil, 0, wrappedBuild.Build.BuildUuid)
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

// combineCTPRequests fetches all CTP Requests from inside the puck packages and
// groups them by SuSch config.
func combineCTPRequests(releaseBuilds []*builds.BuildPackage) map[string][]*builds.EventWrapper {
	configMap := map[string][]*builds.EventWrapper{}

	// Iterate through all CTP Requests
	for _, build := range releaseBuilds {
		for _, request := range build.Requests {
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
