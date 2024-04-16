// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package run

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	kronpb "go.chromium.org/chromiumos/infra/proto/go/test_platform/kron"
	suschpb "go.chromium.org/chromiumos/infra/proto/go/testplans"
	"go.chromium.org/luci/auth/client/authcli"

	"infra/cros/cmd/kron/builds"
	"infra/cros/cmd/kron/common"
	"infra/cros/cmd/kron/configparser"
	"infra/cros/cmd/kron/pubsub"
	"infra/cros/cmd/kron/totmanager"
)

const noBuildUUID = ""

// requiredBuild encapsulates the information needed to request a build from
// PSQL for TimedEvents configs.
type requiredBuild struct {
	buildTarget string
	board       string
	milestone   int
}

// determineRequiredBuilds takes in a SuiteScheduler config and returns
// what buildTargets, and at which milestones, we'll need to request from PSQL.
func determineRequiredBuilds(configs configparser.ConfigList, suiteSchedulerConfigs *configparser.SuiteSchedulerConfigs) ([]*requiredBuild, map[requiredBuild][]*suschpb.SchedulerConfig, error) {
	// The map is generated and returned so that we can quickly find the configs
	// based on their image needs. During a bust of many suites there may be
	// thousands of builds/configs so this will mitigate scaling issues.
	requiredBuildMap := map[requiredBuild][]*suschpb.SchedulerConfig{}

	// This list will be used by the Cloud SQL Query to fetch only the required
	// builds and not every build in the database.
	requiredBuilds := []*requiredBuild{}
	for _, config := range configs {
		// Retrieve the cached targetOptions for the current config.
		targetOptions, err := suiteSchedulerConfigs.FetchConfigTargetOptions(config.Name)
		if err != nil {
			return nil, nil, err
		}

		// Fetch the milestone numbers that the config tracks based on the
		// current ToT.
		milestones, err := totmanager.BranchesToMilestones(config.GetBranches())
		if err != nil {
			return nil, nil, err
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
					key := requiredBuild{
						buildTarget: string(buildTarget),
						board:       targetOption.Board,
						milestone:   milestone,
					}

					// If this is a new key add it to the map.
					if _, ok := requiredBuildMap[key]; !ok {
						requiredBuildMap[key] = []*suschpb.SchedulerConfig{}
						requiredBuilds = append(requiredBuilds, &key)
					}

					// Add the current config to the tracking map. This will
					// allow us to quickly access all configs which are targeted
					// by a fetch build.
					requiredBuildMap[key] = append(requiredBuildMap[key], config)
				}
			}
		}

	}

	return requiredBuilds, requiredBuildMap, nil
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

// attachedTriggeredConfigsToBuilds reads through the queried builds and
// attaches the config items required the build.
func attachedTriggeredConfigsToBuilds(buildImages []*builds.BuildPackage, buildsToConfigMap map[requiredBuild][]*suschpb.SchedulerConfig) ([]*builds.BuildPackage, error) {
	buildsABCD := []*builds.BuildPackage{}

	for _, build := range buildImages {
		// Generate a key to fit the provided map.
		key := requiredBuild{
			buildTarget: build.Build.BuildTarget,
			milestone:   int(build.Build.Milestone),
		}

		// If the build wasn't found in the provided map that means that we
		// received a build in the PSQL query that was not supposed to be there.
		if _, ok := buildsToConfigMap[key]; !ok {
			return nil, fmt.Errorf("build %s for milestone %d retrieved from PSQL but not required", key.buildTarget, key.milestone)
		}

		// If all the configs were removed in the filtering steps before this
		// function, this will be a noop and we'll continue along.
		for _, config := range buildsToConfigMap[key] {
			// Create a wrapper around the triggered config. The CTP requests
			// will be made and added later.
			configPackage := &builds.ConfigDetails{
				Config: config,
				Events: []*builds.EventWrapper{},
			}

			build.TriggeredConfigs = append(build.TriggeredConfigs, configPackage)
		}

	}

	return buildsABCD, nil
}

// checkForMissingBuilds iterates the builds received from long term storage and
// checks it against the list of required builds for this trigger time's
// configs.
func checkForMissingBuilds(requiredBuildList []*requiredBuild, fetchedBuilds []*builds.BuildPackage) []requiredBuild {
	notFound := []requiredBuild{}
	for _, requiredBuild := range requiredBuildList {
		found := false
		for _, fetchedBuild := range fetchedBuilds {
			if requiredBuild.buildTarget == fetchedBuild.Build.BuildTarget && requiredBuild.milestone == int(fetchedBuild.Build.Milestone) {
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

// buildAndPublishUnschedulableEvents generates all events that would have been
// build for the config if it were able to be launched. The passed in scheduling
// decision will be used for the event type.
func buildAndPublishUnschedulableEvents(config *suschpb.SchedulerConfig, suiteSchedulerConfigs *configparser.SuiteSchedulerConfigs, buildInfo requiredBuild, schedulingDecision *kronpb.SchedulingDecision, eventPublishClient pubsub.PublishClient, buildUUID string) error {
	targetOptions, err := suiteSchedulerConfigs.FetchConfigTargetOptionsForBoard(config.Name, configparser.Board(buildInfo.board))
	if err != nil {
		return err
	}

	// Get get the branch target which this build matched with.
	events, err := buildConfigEventsPerModel(targetOptions.Models, config, buildInfo.board, buildUUID, schedulingDecision)
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
func logMissingBuilds(requiredBuildsMap map[requiredBuild][]*suschpb.SchedulerConfig, suiteSchedulerConfigs *configparser.SuiteSchedulerConfigs, eventPublishClient pubsub.PublishClient, requiredBuildList []*requiredBuild, fetchedBuilds []*builds.BuildPackage) (map[requiredBuild][]*suschpb.SchedulerConfig, error) {
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
				FailedReason: fmt.Sprintf("A build for buildTarget %s on milestone %d was not found in the PSQL query.", build.buildTarget, build.milestone),
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
func logStaleBuilds(fetchedBuilds []*builds.BuildPackage, requiredBuildsMap map[requiredBuild][]*suschpb.SchedulerConfig, suiteSchedulerConfigs *configparser.SuiteSchedulerConfigs, eventPublishClient pubsub.PublishClient) (map[requiredBuild][]*suschpb.SchedulerConfig, error) {
	for _, fetchedBuild := range fetchedBuilds {
		// Generate a key for the builds/config map.
		key := requiredBuild{
			buildTarget: fetchedBuild.Build.BuildTarget,
			board:       fetchedBuild.Build.Board,
			milestone:   int(fetchedBuild.Build.Milestone),
		}

		validatedConfigs := []*suschpb.SchedulerConfig{}

		// Check if the build is fresh enough for each of the configs which
		// required it.
		for _, config := range requiredBuildsMap[key] {
			// If the build is too old then generate an event message for
			// each of the would have been generated requests and mark them
			// as NO_PASSING_BUILD. Otherwise add it to the updated list of
			// compliant configs.
			if isBuildTooOld(fetchedBuild.Build.CreateTime, config.LaunchCriteria.LaunchProfile) {
				schedulingDecision := &kronpb.SchedulingDecision{
					Type:         kronpb.DecisionType_NO_PASSING_BUILD,
					Scheduled:    false,
					FailedReason: fmt.Sprintf("Build %s is too old for config %s on testing cadence %s", fetchedBuild.Build.GetBuildUuid(), config.Name, suschpb.SchedulerConfig_LaunchCriteria_LaunchProfile_name[int32(*config.GetLaunchCriteria().GetLaunchProfile().Enum())]),
				}
				err := buildAndPublishUnschedulableEvents(config, suiteSchedulerConfigs, key, schedulingDecision, eventPublishClient, fetchedBuild.Build.BuildUuid)
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

// TimedEvents fetches all configs which are are triggered at the current
// day:hour, fetches all relevant build images, and then schedules their
// subsequent CTP requests via BuildBucket.
// TODO(b/315340446 | b/319463660): This function cannot be completed till we have some sort
// of long term storage to fetch build image information from.
func TimedEvents(authOpts *authcli.Flags, isProd, dryRun bool) error {
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

	projectID := common.StagingProjectID
	if isProd {
		projectID = common.ProdProjectID
	}

	// Format time as KronTime for searching through the configs.
	// TODO(b/319463660): implement option for passed in time, similar to the
	// -start-time flag in the search CLI command.
	operatingTime := common.TimeToKronTime(time.Now())

	common.Stdout.Printf("Initializing client for pub sub topic %s", common.EventsPubSubTopic)
	eventPublishClient, err := pubsub.InitPublishClient(context.Background(), projectID, common.EventsPubSubTopic)
	if err != nil {
		return err
	}

	// Fetch all configs, from all TIMED_EVENT types, which are triggered at the
	// current operating time.
	common.Stdout.Printf("Fetching configs for time %s kron time %s\n", time.Now().String(), operatingTime.String())
	timedConfigs, err := fetchTimedEvents(operatingTime, suiteSchedulerConfigs)
	if err != nil {
		return err
	}

	common.Stdout.Println("Determining what buildTargets/Milestones to fetch from PSQL")
	requiredBuildList, requiredBuildsMap, err := determineRequiredBuilds(timedConfigs, suiteSchedulerConfigs)
	if err != nil {
		return err
	}

	// TODO(b/315340446): Fetch the newest build image for each target option
	// from long term storage.
	common.Stdout.Println("Fetching Builds from PSQL long term storage")
	fetchedBuilds := []*builds.BuildPackage{}

	common.Stdout.Println("Determining missing builds from PSQL query and logging lost events")
	requiredBuildsMap, err = logMissingBuilds(requiredBuildsMap, suiteSchedulerConfigs, eventPublishClient, requiredBuildList, fetchedBuilds)
	if err != nil {
		return err
	}

	// Find and handle stale builds
	common.Stdout.Println("Determining builds stale builds and logging lost events")
	requiredBuildsMap, err = logStaleBuilds(fetchedBuilds, requiredBuildsMap, suiteSchedulerConfigs, eventPublishClient)
	if err != nil {
		return err
	}

	common.Stdout.Println("Transforming builds and configs into a launch-able format")
	fetchedBuilds, err = attachedTriggeredConfigsToBuilds(fetchedBuilds, requiredBuildsMap)
	if err != nil {
		return err
	}

	common.Stdout.Println("Sending all builds to be launched to CTP")
	return launchCTPRequests(fetchedBuilds, suiteSchedulerConfigs, authOpts, projectID, isProd, dryRun)
}
