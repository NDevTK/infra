// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package run

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"

	kronpb "go.chromium.org/chromiumos/infra/proto/go/test_platform/kron"
	suschpb "go.chromium.org/chromiumos/infra/proto/go/testplans"
	"go.chromium.org/luci/auth/client/authcli"

	"infra/cros/cmd/kron/builds"
	"infra/cros/cmd/kron/common"
	"infra/cros/cmd/kron/configparser"
	"infra/cros/cmd/kron/metrics"
	"infra/cros/cmd/kron/pubsub"
	"infra/cros/cmd/kron/totmanager"
)

const noBuildUUID = ""

// determineRequiredBuilds takes in a SuiteScheduler config and returns
// what buildTargets, and at which milestones, we'll need to request from PSQL.
func determineRequiredBuilds(configs configparser.ConfigList, suiteSchedulerConfigs *configparser.SuiteSchedulerConfigs) (map[builds.RequiredBuild][]*suschpb.SchedulerConfig, error) {
	// The map is generated and returned so that we can quickly find the configs
	// based on their image needs. During a burst of many suites there may be
	// thousands of builds/configs so this will mitigate scaling issues.
	requiredBuildMap := map[builds.RequiredBuild][]*suschpb.SchedulerConfig{}

	for _, config := range configs {
		// Retrieve the cached targetOptions for the current config.
		targetOptions, err := suiteSchedulerConfigs.FetchConfigTargetOptions(config.Name)
		if err != nil {
			return nil, err
		}

		// Fetch the milestone numbers that the config tracks based on the
		// current ToT.
		milestones, err := totmanager.BranchesToMilestones(config.GetBranches())
		if err != nil {
			return nil, err
		}

		// Create a list of build targets based on the cached target options.
		//
		// NOTE: Target Options contain all the board/model/variant information
		// whereas build targets refer to the target of the release image build.
		// Build targets are typically in the form of board(-<variant>).
		for _, milestone := range milestones {
			for _, targetOption := range targetOptions {
				// Fetch the current board's cached build targets.
				buildTargets := configparser.GetBuildTargets(targetOption, targetOption.VariantsOnly)

				for _, buildTarget := range buildTargets {
					// Generate the requireBuild "key".
					key := builds.RequiredBuild{
						BuildTarget: string(buildTarget),
						Board:       targetOption.Board,
						Milestone:   milestone,
					}

					// If this is a new key add it to the map.
					if _, ok := requiredBuildMap[key]; !ok {
						requiredBuildMap[key] = []*suschpb.SchedulerConfig{}
					}

					// Add the current config to the tracking map. This will
					// allow us to quickly access all configs which are targeted
					// by a fetch build.
					requiredBuildMap[key] = append(requiredBuildMap[key], config)
				}
			}
		}

	}

	return requiredBuildMap, nil
}

// isBuildTooOld checks to make sure that the is not older than the cadence
// period length. This will ensure that testing will only occur on untested
// images an no duplication will occur.
func isBuildTooOld(buildCreateTime *timestamppb.Timestamp, cadence suschpb.SchedulerConfig_LaunchCriteria_LaunchProfile) bool {
	var maxAge time.Duration

	// Set the max age to the length of the cadence period.
	switch cadence {
	case suschpb.SchedulerConfig_LaunchCriteria_DAILY:
		maxAge = 1 * common.Day
	case suschpb.SchedulerConfig_LaunchCriteria_WEEKLY:
		maxAge = 1 * common.Week
	case suschpb.SchedulerConfig_LaunchCriteria_FORTNIGHTLY:
		maxAge = 1 * common.Fortnight

	}

	return time.Since(buildCreateTime.AsTime()) > maxAge
}

// checkForMissingBuilds iterates the builds received from long term storage and
// checks it against the list of required builds for this trigger time's
// configs.
func checkForMissingBuilds(requiredBuildList []*builds.RequiredBuild, fetchedBuilds []*kronpb.Build) []builds.RequiredBuild {
	notFound := []builds.RequiredBuild{}
	for _, requiredBuild := range requiredBuildList {
		found := false
		for _, fetchedBuild := range fetchedBuilds {
			if requiredBuild.BuildTarget == fetchedBuild.BuildTarget && requiredBuild.Milestone == int(fetchedBuild.Milestone) {
				found = true
				break
			}
		}
		if !found {
			notFound = append(notFound, *requiredBuild)
		}
	}
	return notFound
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

// buildAndPublishUnschedulableEvents generates all events that would have been
// build for the config if it were able to be launched. The passed in scheduling
// decision will be used for the event type.
func buildAndPublishUnschedulableEvents(config *suschpb.SchedulerConfig, suiteSchedulerConfigs *configparser.SuiteSchedulerConfigs, buildInfo builds.RequiredBuild, schedulingDecision *kronpb.SchedulingDecision, eventPublishClient pubsub.PublishClient, buildUUID string) error {
	targetOptions, err := suiteSchedulerConfigs.FetchConfigTargetOptionsForBoard(config.Name, configparser.Board(buildInfo.Board))
	if err != nil {
		return err
	}

	// Get get the branch target which this build matched with.
	events, err := buildConfigEventsPerModel(targetOptions.Models, config, buildInfo.Board, buildUUID, schedulingDecision)
	if err != nil {
		return err
	}
	for _, event := range events {
		err = publishEvent(eventPublishClient, event)
		if err != nil {
			return err
		}
	}

	return nil
}

// logMissingBuilds removes configs and builds from the tracking map and creates
// event messages with the BUILD_NOT_FOUND event type.
func logMissingBuilds(requiredBuildsMap map[builds.RequiredBuild][]*suschpb.SchedulerConfig, suiteSchedulerConfigs *configparser.SuiteSchedulerConfigs, eventPublishClient pubsub.PublishClient, requiredBuildList []*builds.RequiredBuild, fetchedBuilds []*kronpb.Build) (map[builds.RequiredBuild][]*suschpb.SchedulerConfig, error) {
	// Filter through the builds and search for missing builds. If a build
	// is missing then that means that we want to make an event and mark it as
	// DecisionType_BUILD_NOT_FOUND. This will then be sent to the metrics
	// pipeline for logging.
	missingBuilds := checkForMissingBuilds(requiredBuildList, fetchedBuilds)

	for _, build := range missingBuilds {
		for _, config := range requiredBuildsMap[build] {
			schedulingDecision := &kronpb.SchedulingDecision{
				Type:         kronpb.DecisionType_BUILD_NOT_FOUND,
				Scheduled:    false,
				FailedReason: fmt.Sprintf("A build for buildTarget %s on milestone %d was not found in the PSQL query.", build.BuildTarget, build.Milestone),
			}

			err := buildAndPublishUnschedulableEvents(config, suiteSchedulerConfigs, build, schedulingDecision, eventPublishClient, noBuildUUID)
			if err != nil {
				return nil, err
			}

			// Remove the build from the required builds map. This will remove
			// any confusion later when we build all the Config requests.
			delete(requiredBuildsMap, build)
		}
	}

	return requiredBuildsMap, nil
}

// logStaleBuilds checks each build against the dependant configs to ensure that
// the build image is fresh enough for scheduling. If a build image age is older
// than the period length of a configs cadence then it is considered stale as
// the build likely scheduled on the previous trigger time.
func logStaleBuilds(fetchedBuilds []*kronpb.Build, requiredBuildsMap map[builds.RequiredBuild][]*suschpb.SchedulerConfig, suiteSchedulerConfigs *configparser.SuiteSchedulerConfigs, eventPublishClient pubsub.PublishClient) (map[builds.RequiredBuild][]*suschpb.SchedulerConfig, error) {
	for _, fetchedBuild := range fetchedBuilds {
		// Generate a key for the builds/config map.
		key := builds.RequiredBuild{
			BuildTarget: fetchedBuild.BuildTarget,
			Board:       fetchedBuild.Board,
			Milestone:   int(fetchedBuild.Milestone),
		}

		validatedConfigs := []*suschpb.SchedulerConfig{}

		// Check if the build is fresh enough for each of the configs which
		// required it.
		for _, config := range requiredBuildsMap[key] {
			// If the build is too old then generate an event message for
			// each of the would have been generated requests and mark them
			// as NO_PASSING_BUILD. Otherwise add it to the updated list of
			// compliant configs.
			if isBuildTooOld(fetchedBuild.CreateTime, config.LaunchCriteria.LaunchProfile) {
				schedulingDecision := &kronpb.SchedulingDecision{
					Type:         kronpb.DecisionType_NO_PASSING_BUILD,
					Scheduled:    false,
					FailedReason: fmt.Sprintf("Build %s is too old for config %s on testing cadence %s", fetchedBuild.GetBuildUuid(), config.Name, suschpb.SchedulerConfig_LaunchCriteria_LaunchProfile_name[int32(*config.GetLaunchCriteria().GetLaunchProfile().Enum())]),
				}
				err := buildAndPublishUnschedulableEvents(config, suiteSchedulerConfigs, key, schedulingDecision, eventPublishClient, fetchedBuild.BuildUuid)
				if err != nil {
					return nil, err
				}
			} else {
				validatedConfigs = append(validatedConfigs, config)
			}
		}

		// If the build was too old for all it's configs remove it from the
		// tracking map. Otherwise, give it the updated list of configs.
		if len(validatedConfigs) > 0 {
			delete(requiredBuildsMap, key)
		} else {
			requiredBuildsMap[key] = validatedConfigs
		}
	}
	return requiredBuildsMap, nil
}

// CrOSTimedEventCommand is a TimedEventCommand which fetches all configs which
// are are triggered at the current day:hour, fetches all relevant build images,
// and then schedules their subsequent CTP requests via BuildBucket.
type CrOSTimedEventCommand struct {
	authOpts *authcli.Flags
	isProd   bool
	dryRun   bool

	labConfigs            *configparser.LabConfigs
	suiteSchedulerConfigs *configparser.SuiteSchedulerConfigs

	projectID string
}

// InitCrOSTimedEventCommand generates and returns a CrOS TIMED_EVENT client which
// implements the TimedEventCommand interface. This client does not handle
// firmware, Android, nor multi-DUT configs.
func InitCrOSTimedEventCommand(authOpts *authcli.Flags, isProd, dryRun bool, labConfigs *configparser.LabConfigs, suiteSchedulerConfigs *configparser.SuiteSchedulerConfigs, projectID string) TimedEventCommand {
	return &CrOSTimedEventCommand{
		authOpts:              authOpts,
		isProd:                isProd,
		dryRun:                dryRun,
		labConfigs:            labConfigs,
		suiteSchedulerConfigs: suiteSchedulerConfigs,
		projectID:             projectID,
	}
}
func (c *CrOSTimedEventCommand) Name() string {
	return "CrOSTimedEvents"
}

// FetchTriggeredConfigs gathers all CrOS TIMED_EVENT configs which are
// triggered at the passed in execution time.
func (c *CrOSTimedEventCommand) FetchTriggeredConfigs(executionTime common.KronTime) (map[builds.RequiredBuild][]*suschpb.SchedulerConfig, error) {
	// Fetch all configs, from all TIMED_EVENT types, which are triggered at the
	// current operating time.
	common.Stdout.Printf("Fetching configs for time %s kron time %s\n", time.Now().String(), executionTime.String())
	timedConfigs, err := fetchTimedEvents(executionTime, c.suiteSchedulerConfigs)
	if err != nil {
		return nil, err
	}

	common.Stdout.Println("Determining what buildTargets/Milestones to fetch from PSQL")
	requiredBuildMap, err := determineRequiredBuilds(timedConfigs, c.suiteSchedulerConfigs)
	if err != nil {
		return nil, err
	}

	return requiredBuildMap, nil
}

// convertToKronBuildMap converts the given map to being keyed by kron builds
// rather than required (requested) builds.
func convertToKronBuildMap(fetchedBuilds []*kronpb.Build, requiredBuildsMap map[builds.RequiredBuild][]*suschpb.SchedulerConfig) (map[*kronpb.Build][]*suschpb.SchedulerConfig, error) {
	kronBuildMap := map[*kronpb.Build][]*suschpb.SchedulerConfig{}
	for _, fetchedBuild := range fetchedBuilds {
		buildKey := builds.RequiredBuild{
			BuildTarget: fetchedBuild.GetBuildTarget(),
			Board:       fetchedBuild.GetBoard(),
			Milestone:   int(fetchedBuild.GetMilestone()),
		}

		// If the current fetched board is not in the map then we somehow have a
		// build fetched from LTS on accident.
		configs, ok := requiredBuildsMap[buildKey]
		if !ok {
			return nil, fmt.Errorf("build %s of board %s milestone %d fetched when not required", fetchedBuild.BuildUuid, fetchedBuild.Board, fetchedBuild.Milestone)
		}

		kronBuildMap[fetchedBuild] = configs
	}

	return kronBuildMap, nil
}

// FetchBuilds gathers all builds needed from long term storage.
func (c *CrOSTimedEventCommand) FetchBuilds(requiredBuildsMap map[builds.RequiredBuild][]*suschpb.SchedulerConfig) (map[*kronpb.Build][]*suschpb.SchedulerConfig, error) {
	common.Stdout.Printf("Initializing client for pub sub topic %s", common.EventsPubSubTopic)
	eventPublishClient, err := pubsub.InitPublishClient(context.Background(), c.projectID, common.EventsPubSubTopic)
	if err != nil {
		return nil, err
	}

	// TODO(b/315340446): Fetch the newest build image for each target option
	// from long term storage.
	common.Stdout.Println("Fetching Builds from PSQL long term storage")
	fetchedBuilds := []*kronpb.Build{}

	requiredBuildsList := []*builds.RequiredBuild{}
	for key := range requiredBuildsMap {
		requiredBuildsList = append(requiredBuildsList, &key)
	}

	common.Stdout.Println("Determining missing builds from PSQL query and logging lost events")
	requiredBuildsMap, err = logMissingBuilds(requiredBuildsMap, c.suiteSchedulerConfigs, eventPublishClient, requiredBuildsList, fetchedBuilds)
	if err != nil {
		return nil, err
	}

	// Find and handle stale builds
	common.Stdout.Println("Determining builds stale builds and logging lost events")
	requiredBuildsMap, err = logStaleBuilds(fetchedBuilds, requiredBuildsMap, c.suiteSchedulerConfigs, eventPublishClient)
	if err != nil {
		return nil, err
	}

	// Convert the builds map to a type compatible for the ScheduleRequests
	// function.
	kronBuildMap, err := convertToKronBuildMap(fetchedBuilds, requiredBuildsMap)
	if err != nil {
		return nil, err
	}

	return kronBuildMap, nil
}

// ScheduleRequests generates CTP Requests, batches them into BuildBucket
// requests, and Schedules them via the BuildBucket API.
func (c *CrOSTimedEventCommand) ScheduleRequests(kronBuildMap map[*kronpb.Build][]*suschpb.SchedulerConfig) error {
	return scheduleRequests(kronBuildMap, c.suiteSchedulerConfigs, c.authOpts, c.projectID, c.isProd, c.dryRun)
}
