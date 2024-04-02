// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package run holds all of the internal logic for the execution steps of a
// SuiteScheduler run.
package run

import (
	"context"
	"sync"
	"time"

	"go.chromium.org/luci/auth/client/authcli"

	"infra/cros/cmd/kron/buildbucket"
	"infra/cros/cmd/kron/builds"
	"infra/cros/cmd/kron/common"
	"infra/cros/cmd/kron/configparser"
	"infra/cros/cmd/kron/ctprequest"
	"infra/cros/cmd/kron/process3d"
	"infra/cros/cmd/kron/pubsub"
)

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
	// Format time as KronTime for searching through the configs.
	// TODO(b/319463660): implement option for passed in time, similar to the
	// -start-time flag in the search CLI command.
	operatingTime := common.TimeToKronTime(time.Now())

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

// NewBuilds fetches all builds from the release Pub/Sub queue, finds all
// triggered NEW_BUILD configs, then builds their respective CTP requests for
// BB.
func NewBuilds(authOpts *authcli.Flags, isProd, dryRun bool) error {
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

	// Get build information
	// TODO(b/315340446): Inside this client we need to not ACK the builds until
	// we can launch the BB task. This may require that we do a lot of heavy
	// lifting in the callback but overall it would be safer.
	common.Stdout.Println("Fetching builds from Pub/Sub.")
	releaseBuilds, err := builds.IngestBuildsFromPubSub(projectID, common.BuildsSubscription)
	if err != nil {
		return err

	}

	// TODO(b/315340446): Write build info to long term storage(database)

	// Build the list of all configs triggered by the ingested build images.
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

	if len(releaseBuilds) == 0 {
		common.Stderr.Println("No builds found")
		return nil
	}

	// CTP is bottlenecked by it's drone count. To combat this, combine tests
	// requests into one large CTP request.
	ctpRequests := combineCTPRequests(releaseBuilds)

	// If staging reduce requests to 5 MAX.
	if !isProd {
		ctpRequests = limitStagingRequests(ctpRequests)
	}

	// Initialize an authenticated BuildBucket client for scheduling.
	common.Stdout.Printf("Initializing BuildBucket scheduling client prod: %t dryrun: %t", isProd, dryRun)
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
	for configName, requestList := range ctpRequests {
		wg.Add(1)
		go scheduleBatchViaBB(requestList, configName, schedulerClient, publishClient, &wg)
	}

	common.Stdout.Println("Waiting for batched requests to finish scheduling...")
	wg.Wait()
	common.Stdout.Println("NEW_BUILD scheduling completed")

	return nil
}

// Process3d fetches all builds from the release Pub/Sub queue, checks
// if all builds have completed, then builds CTP request for all 3d configs and executes them.
func Process3d(authOpts *authcli.Flags, isProd, dryRun bool) error {
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

	process3d := process3d.NewProcess3d(projectID, common.BuildsSubscription3dTesting, suiteSchedulerConfigs.FetchAllNewBuild3dConfigs())
	err = process3d.Process3d()
	if err != nil {
		common.Stdout.Println("Error occurred while processing 3d configs")
		return err
	}
	return nil
}
