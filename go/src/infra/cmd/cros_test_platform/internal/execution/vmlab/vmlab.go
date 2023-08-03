// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package vmlab

import (
	"strings"

	"infra/libs/skylab/request"
)

const (
	vmLabLaunchExperimentName = "chromeos.cros_infra_config.vmlab.launch"
	testRunnerBuilderName     = "test_runner"
	testRunnerGceBuilderName  = "test_runner_gce"
)

// ShouldRun decides if VM test flow should be triggered based on eligibility
// and required data from the original Skylab request.
func ShouldRun(args *request.Args) bool {
	return args.CFTIsEnabled &&
		args.CFTTestRunnerRequest != nil &&
		eligible(args.SchedulableLabels.GetSelfServePools(), args.Experiments)
}

// eligible checks board name and experiments to decide if preconditions are met
// to run test on VMLab.
func eligible(pool []string, experiments []string) bool {
	return isExperimentEnabled(experiments) && isSupportedPool(pool)
}

// ConvertBuilderName converts the original test runner name to corresponding
// name of the VMLab version. See configs at http://shortn/_86vxOQ0XC6
// test_runner[-env] -> test_runner_gce[-env]
func ConvertBuilderName(originalBuilderName string) string {
	if !strings.HasPrefix(originalBuilderName, testRunnerBuilderName) {
		return originalBuilderName
	}
	if strings.HasPrefix(originalBuilderName, testRunnerGceBuilderName) {
		return originalBuilderName
	}
	return strings.Replace(originalBuilderName, testRunnerBuilderName, testRunnerGceBuilderName, 1)
}

func isSupportedPool(pool []string) bool {
	if pool == nil {
		return false
	}
	for _, e := range pool {
		if e == "vmlab" {
			return true
		}
	}
	return false
}

func isExperimentEnabled(experiments []string) bool {
	if experiments == nil {
		return false
	}
	for _, e := range experiments {
		if e == vmLabLaunchExperimentName {
			return true
		}
	}
	return false
}
