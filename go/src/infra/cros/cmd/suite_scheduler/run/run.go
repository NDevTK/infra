// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package run holds all of the internal logic for the execution steps of a
// SuiteScheduler run.
package run

import (
	"strconv"
	"time"

	v15 "go.chromium.org/chromiumos/infra/proto/go/test_platform/suite_scheduler/v15"
	infrapb "go.chromium.org/chromiumos/infra/proto/go/testplans"

	"infra/cros/cmd/suite_scheduler/builds"
	"infra/cros/cmd/suite_scheduler/common"
	"infra/cros/cmd/suite_scheduler/configparser"
	"infra/cros/cmd/suite_scheduler/ctprequest"
)

// newBuildRequest wraps the config with the image that triggered it. This
// makes for easier request building.
type newBuildRequest struct {
	config *infrapb.SchedulerConfig
	build  *v15.BuildInformation
}

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
// TODO(b/315340446): This function cannot be completed till we have some sort
// of long term storage to fetch build image information from.
func TimedEvents(suiteSchedulerConfigs *configparser.SuiteSchedulerConfigs) error {
	// Format time as SuSchTime for searching through the configs.
	// TODO(juahurta): implement option for passed in time, similar to the
	// -start-time flag in the search CLI command.
	operatingTime := common.TimeToSuSchTime(time.Now())

	timedEvents, err := fetchTimedEvents(operatingTime, suiteSchedulerConfigs)
	if err != nil {
		return err
	}

	// TODO(juahurta): Calculate all target options used from TimedEvents

	// TODO(b/315340446): Fetch the newest build image for each target option.

	// Build CTP requests
	// TODO(juahurta: Once long term storage of build images is in place then we
	// can properly build CTP requests. Right now we are not passing in target
	// options nor build images.
	ctpRequests := ctprequest.CTPRequests{}
	for _, event := range timedEvents {
		// TODO(juahurta): provide target options and build images for CTP requests.
		ctpRequests = append(ctpRequests, ctprequest.BuildAllCTPRequests(event, configparser.TargetOptions{})...)
	}

	// Schedule each CTP request via BB API.

	return nil
}

// NewBuilds fetches all builds from the release Pub/Sub queue, finds all
// triggered NEW_BUILD configs, then builds their respective CTP requests for
// BB.
func NewBuilds() error {
	// Ingest lab configs into memory.
	// TODO(juahurta): Implement option to pass in a local config as to bypass
	// network reliance.
	common.Stdout.Println("Fetch lab configs")
	labConfigs, err := configparser.FetchLabConfigs("")
	if err != nil {
		return err
	}

	// Ingest SuiteScheduler configs into memory.
	// TODO(juahurta): Implement option to pass in a local config as to bypass
	// network reliance.
	common.Stdout.Println("Fetch SuSch configs")
	suiteSchedulerConfigs, err := configparser.FetchSchedulerConfigs("", labConfigs)
	if err != nil {
		return err
	}

	common.Stdout.Println("Fetch Builds")

	// TODO(juahurta): Inside this client we need to not ACK the builds until we
	// can launch the BB task. This may require that we do a lot of heavy
	// lifting in the callback but overall it would be safer.
	// Get build information
	builds, err := builds.IngestBuildsFromPubSub()
	if err != nil {
		return err

	}
	// TODO(juahurta): For in run events. Squash the builds so that NEW_BUILD
	// configs are only triggered by the newest images. The created time is
	// stored inside the build artifact type.
	// https://chromium.googlesource.com/chromiumos/infra/proto/+/refs/heads/main/src/chromiumos/build_report.proto#197

	common.Stdout.Println("Finding all NEW_BUILD requests.")
	// This list will be used to wrap all NEW_BUILD configs triggered with their
	// respective build images.
	newBuildConfigs := []newBuildRequest{}

	// Build the list of all configs triggered by the ingested build images.
	for _, build := range builds {
		// Fetch all configs for which this build will will launch a NEW_BUILD event.
		configs := suiteSchedulerConfigs.FetchNewBuildConfigsByBuildTarget(configparser.BuildTarget(build.BuildTarget))
		for _, config := range configs {
			request := newBuildRequest{
				config: config,
				build:  build,
			}
			newBuildConfigs = append(newBuildConfigs, request)
		}
	}

	common.Stdout.Println("Building CTP requests for all NEW_BUILD configs triggered.")
	// Build CTP Requests for all triggered configs.
	ctpRequests := ctprequest.CTPRequests{}
	for _, request := range newBuildConfigs {
		ctpRequests = append(ctpRequests, ctprequest.BuildCTPRequest(request.config, request.build.Board, "", request.build.BuildTarget, strconv.FormatInt(request.build.Milestone, 10), request.build.Version))
	}

	// TODO(b/317084435): Send all CTP request to BB for scheduling.

	return nil
}
