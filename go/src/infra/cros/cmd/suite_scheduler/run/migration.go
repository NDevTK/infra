// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package run

import "infra/cros/cmd/suite_scheduler/builds"

var (
	// allowedConfigs is a quick access tool to check if the SuSch config is
	// being allowed through during the migration.
	//
	// A map was used here to reduce on search complexity.
	allowedConfigs = map[string]bool{
		"CFTNewBuild": true,
	}
)

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
		tempBuild.Requests = []*builds.ConfigDetails{}

		// Iterate through the requests and only add requests to the temp build
		// if their SuSch config is on the allowlist.
		for _, request := range build.Requests {
			if _, ok := allowedConfigs[request.Config.Name]; ok {
				tempBuild.Requests = append(tempBuild.Requests, request)
				hadAllowedConfig = true
			}
		}

		if hadAllowedConfig {
			filteredBuilds = append(filteredBuilds, &tempBuild)
		}
	}

	return filteredBuilds
}
