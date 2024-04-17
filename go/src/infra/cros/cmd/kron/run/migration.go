// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package run

import (
	suschpb "go.chromium.org/chromiumos/infra/proto/go/testplans"

	"infra/cros/cmd/kron/builds"
	"infra/cros/cmd/kron/configparser"
)

var (
	allowedConfigs = map[string]bool{}
)

// isAllowed checks the migration rules to determine if a config has been
// migrated to Kron or not.
func isAllowed(config *suschpb.SchedulerConfig) bool {
	// Disallow partner configs.
	if config.GetRunOptions().GetBuilderId().GetProject() != "" && config.GetRunOptions().GetBuilderId().GetBucket() != "" && config.GetRunOptions().GetBuilderId().GetBuilder() != "" {
		return false
	}

	// Disallow multi-dut and firmware configs.
	if configparser.IsMultiDut(config) || configparser.IsFirmware(config) {
		return false
	}

	// Allow NEW_BUILD configs.
	if config.GetLaunchCriteria().GetLaunchProfile() == suschpb.SchedulerConfig_LaunchCriteria_NEW_BUILD {
		return true
	}

	// Allow Explicitly included configs.
	if _, ok := allowedConfigs[config.GetName()]; ok {
		return true
	}

	// Exclude everything else.
	return false
}

// filterConfigs iterates through the triggered SuSch Configs and scrubs out all
// configs which are not on the allowlist.
//
// TODO(b/319273876): Remove slow migration logic upon completion of
// transition.
func filterConfigs(buildPackages []*builds.BuildPackage) []*builds.BuildPackage {
	filteredBuilds := []*builds.BuildPackage{}

	hadAllowedConfig := false
	for _, build := range buildPackages {
		// Copy the build by value so that we can clear the requests field.
		tempBuild := *build
		tempBuild.TriggeredConfigs = []*builds.ConfigDetails{}

		// Iterate through the requests and only add requests to the temp build
		// if their SuSch config is on the allowlist.
		for _, triggeredConfig := range build.TriggeredConfigs {
			if isAllowed(triggeredConfig.Config) {
				tempBuild.TriggeredConfigs = append(tempBuild.TriggeredConfigs, triggeredConfig)
				hadAllowedConfig = true
			}
		}

		if hadAllowedConfig {
			filteredBuilds = append(filteredBuilds, &tempBuild)
		}
	}

	return filteredBuilds
}
