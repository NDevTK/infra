// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package run holds all of the internal logic for the execution steps of a
// SuiteScheduler run.
package run

import (
	"go.chromium.org/luci/auth/client/authcli"

	"infra/cros/cmd/kron/builds"
	"infra/cros/cmd/kron/common"
	"infra/cros/cmd/kron/configparser"
	"infra/cros/cmd/kron/process3d"
)

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
	releaseBuilds, err := builds.IngestBuildsFromPubSub(projectID, common.BuildsSubscription, isProd)
	if err != nil {
		return err

	}

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

	return launchCTPRequests(releaseBuilds, suiteSchedulerConfigs, authOpts, projectID, isProd, dryRun)
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
