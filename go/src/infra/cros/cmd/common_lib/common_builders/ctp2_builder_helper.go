// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common_builders

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"google.golang.org/protobuf/proto"

	testapi "go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
	"go.chromium.org/luci/common/logging"

	"infra/cros/cmd/common_lib/common"
)

const (
	LabelPool                     = "label-pool"
	AnalyticsName                 = "analytics_name"
	DefaultChromeosBuildGcsBucket = "chromeos-image-archive"

	ChromeosBuild          = "chromeos_build"
	ChromeosBuildGcsBucket = "chromeos_build_gcs_bucket"
	RoFirmwareBuild        = "ro_firmware_build"
	RwFirmwareBuild        = "rw_firmware_build"
	LacrosGcsPath          = "lacros_gcs_path"
	AndroidImageVersion    = "android_image_version"
	GmsCorePackage         = "gms_core_package"
)

var (
	ExcludedChromeosBuildPrefixes  = []string{"staging", "dev"}
	ExcludedChromeosBuildPostfixes = []string{"main"}
)

// GroupV2Requests reduces the list of v2 requests by grouping
// by build and suite request.
func GroupV2Requests(ctx context.Context, v2s []*testapi.CTPRequest) []*testapi.CTPRequest {
	groups := map[string][]*testapi.CTPRequest{}
	for _, v2 := range v2s {
		// Safe guard against CTPRequests missing targets to schedule on.
		if len(v2.GetScheduleTargets()) == 0 || len(v2.GetScheduleTargets()[0].GetTargets()) == 0 {
			logging.Infof(ctx, "request missing targets to schedule on, dropping: %v", v2)
			continue
		}
		// Translator only supplies singular length schedule targets,
		// and only care about checking the primary target's gcspath when grouping.
		gcsPath := v2.GetScheduleTargets()[0].GetTargets()[0].GetSwTarget().GetLegacySw().GetGcsPath()
		build := common.GetMajorBuildFromGCSPath(gcsPath)
		if build == "" || v2.SuiteRequest == nil {
			continue
		}
		if _, ok := groups[build]; !ok {
			groups[build] = []*testapi.CTPRequest{}
		}

		foundMatch := false
		for _, suiteGroup := range groups[build] {
			if proto.Equal(suiteGroup.GetSuiteRequest(), v2.GetSuiteRequest()) {
				suiteGroup.ScheduleTargets = append(suiteGroup.ScheduleTargets, v2.GetScheduleTargets()...)
				foundMatch = true
				break
			}
		}
		if foundMatch {
			continue
		}
		groups[build] = append(groups[build], v2)
	}

	flatRequests := []*testapi.CTPRequest{}
	for _, buildRequests := range groups {
		flatRequests = append(flatRequests, buildRequests...)
	}

	return flatRequests
}

// buildCTPRequest converts a v1 ctp request into a v2 CTPRequest.
func buildCTPRequest(v1 *test_platform.Request) *testapi.CTPRequest {
	return &testapi.CTPRequest{
		SuiteRequest:    buildSuiteRequest(v1),
		ScheduleTargets: buildScheduleTargets(v1),
		SchedulerInfo:   buildSchedulerInfo(v1),
		Pool:            getSchedulingPool(v1),
		// Reuse translate flag from v1 to signal dynamic run in v2.
		RunDynamic: v1.GetParams().GetTranslateTrv2Request(),
	}
}

// buildSchedulerInfo produces the scheduling system to be used,
// as well as the qs account for qs scheduling.
func buildSchedulerInfo(v1 *test_platform.Request) *testapi.SchedulerInfo {
	return &testapi.SchedulerInfo{
		// TODO(cdelagarza): Update to upstream variable.
		Scheduler: testapi.SchedulerInfo_SCHEDUKE,
		QsAccount: v1.GetParams().GetScheduling().GetQsAccount(),
	}
}

// buildSuiteRequest converts a v1 ctp request into a SuiteRequest.
func buildSuiteRequest(v1 *test_platform.Request) *testapi.SuiteRequest {
	return &testapi.SuiteRequest{
		SuiteRequest: &testapi.SuiteRequest_TestSuite{
			TestSuite: buildTestSuite(v1),
		},
		MaximumDuration: v1.GetParams().GetTime().GetMaximumDuration(),
		TestArgs:        getTestArgs(v1),
		AnalyticsName:   getAnalyticsName(v1),
		MaxInShard:      v1.GetTestPlan().GetMaxInShard(),
		DddSuite:        IsDDDSuite(v1),
		RetryCount:      GetRetryCount(v1),
	}
}

// buildTestSuite parses through the available options for
// setting up a test suite. This code follows the ctpv1
// method `_ctr_test_suite` which converts a test plan
// into a TestSuite object.
func buildTestSuite(v1 *test_platform.Request) *testapi.TestSuite {
	testplan := v1.TestPlan
	testSuite := &testapi.TestSuite{
		TotalShards: testplan.TotalShards,
	}
	if len(testplan.GetSuite()) == 0 {
		testSuite.Name = "adhoc"
	} else {
		testSuite.Name = testplan.GetSuite()[0].GetName()
	}

	if testplan.Test != nil && len(testplan.Test) > 0 {
		testCaseIds := []*testapi.TestCase_Id{}
		for _, test := range testplan.Test {
			switch harness := test.GetHarness().(type) {
			case *test_platform.Request_Test_Autotest_:
				testCaseIds = append(testCaseIds, &testapi.TestCase_Id{
					Value: harness.Autotest.Name,
				})
			}
		}
		testSuite.Spec = &testapi.TestSuite_TestCaseIds{
			TestCaseIds: &testapi.TestCaseIdList{
				TestCaseIds: testCaseIds,
			},
		}
	} else if testplan.TagCriteria != nil &&
		(testplan.TagCriteria.TagExcludes != nil ||
			testplan.TagCriteria.Tags != nil ||
			testplan.TagCriteria.TestNameExcludes != nil ||
			testplan.TagCriteria.TestNames != nil) {
		testSuite.Spec = &testapi.TestSuite_TestCaseTagCriteria_{
			TestCaseTagCriteria: testplan.TagCriteria,
		}
	} else {
		tags := []string{}
		for _, suite := range testplan.GetSuite() {
			tags = append(tags, fmt.Sprintf("suite:%s", suite.GetName()))
		}
		testSuite.Spec = &testapi.TestSuite_TestCaseTagCriteria_{
			TestCaseTagCriteria: &testapi.TestSuite_TestCaseTagCriteria{
				Tags: tags,
			},
		}
	}

	return testSuite
}

// getTestArgs returns the available test args.
func getTestArgs(v1 *test_platform.Request) string {
	testplan := v1.TestPlan
	if testplan.Test != nil && len(testplan.Test) > 0 {
		return testplan.GetTest()[0].GetAutotest().GetTestArgs()
	} else if testplan.Suite != nil && len(testplan.Suite) > 0 {
		return testplan.GetSuite()[0].GetTestArgs()
	}

	return ""
}

// buildScheduleTargets converts the v1 primary and
// secondary devices into v2 target objects.
func buildScheduleTargets(v1 *test_platform.Request) []*testapi.ScheduleTargets {
	targets := []*testapi.Targets{
		buildTarget(
			v1.GetParams().GetSoftwareAttributes(),
			v1.GetParams().GetHardwareAttributes(),
			v1.GetParams().GetFreeformAttributes(),
			v1.GetParams().GetSoftwareDependencies()),
	}
	for _, secondary := range v1.GetParams().GetSecondaryDevices() {
		targets = append(targets,
			buildTarget(
				secondary.GetSoftwareAttributes(),
				secondary.GetHardwareAttributes(),
				&test_platform.Request_Params_FreeformAttributes{},
				secondary.GetSoftwareDependencies()))
	}
	return []*testapi.ScheduleTargets{
		{
			Targets: targets,
		},
	}
}

// buildTarget converts v1 software/hardware attributes
// and software dependencies into a v2 target object.
func buildTarget(
	softwareAttributes *test_platform.Request_Params_SoftwareAttributes,
	hardwareAttributes *test_platform.Request_Params_HardwareAttributes,
	freeformAttributes *test_platform.Request_Params_FreeformAttributes,
	softwareDeps []*test_platform.Request_Params_SoftwareDependency) *testapi.Targets {

	return &testapi.Targets{
		HwTarget: &testapi.HWTarget{
			Target: &testapi.HWTarget_LegacyHw{
				LegacyHw: &testapi.LegacyHW{
					Board:              softwareAttributes.GetBuildTarget().GetName(),
					Model:              hardwareAttributes.GetModel(),
					SwarmingDimensions: freeformAttributes.GetSwarmingDimensions(),
					Variant:            GetVariant(softwareDeps),
				},
			},
		},
		SwTarget: &testapi.SWTarget{
			SwTarget: &testapi.SWTarget_LegacySw{
				LegacySw: &testapi.LegacySW{
					Build:     GetBuildType(softwareDeps),
					GcsPath:   getImageGcsPath(softwareDeps),
					KeyValues: mapSoftwareDeps(softwareDeps),
				},
			},
		},
	}
}

// mapSoftwareDeps converts each software dependency
// into a corresponding key value pair.
func mapSoftwareDeps(softwareDeps []*test_platform.Request_Params_SoftwareDependency) []*testapi.KeyValue {
	m := map[string]string{}
	for _, softwareDep := range softwareDeps {
		switch dep := softwareDep.GetDep().(type) {
		case *test_platform.Request_Params_SoftwareDependency_ChromeosBuild:
			m[ChromeosBuild] = dep.ChromeosBuild
		case *test_platform.Request_Params_SoftwareDependency_ChromeosBuildGcsBucket:
			m[ChromeosBuildGcsBucket] = dep.ChromeosBuildGcsBucket
		case *test_platform.Request_Params_SoftwareDependency_RoFirmwareBuild:
			m[RoFirmwareBuild] = dep.RoFirmwareBuild
		case *test_platform.Request_Params_SoftwareDependency_RwFirmwareBuild:
			m[RwFirmwareBuild] = dep.RwFirmwareBuild
		case *test_platform.Request_Params_SoftwareDependency_LacrosGcsPath:
			m[LacrosGcsPath] = dep.LacrosGcsPath
		case *test_platform.Request_Params_SoftwareDependency_AndroidImageVersion:
			m[AndroidImageVersion] = dep.AndroidImageVersion
		case *test_platform.Request_Params_SoftwareDependency_GmsCorePackage:
			m[GmsCorePackage] = dep.GmsCorePackage
		}
	}

	keyValues := []*testapi.KeyValue{}
	for k, v := range m {
		keyValues = append(keyValues, &testapi.KeyValue{
			Key:   k,
			Value: v,
		})
	}
	return keyValues
}

// getSchedulingPool parses the v1 request tags for the label-pool.
func getSchedulingPool(v1 *test_platform.Request) string {
	return getTag(v1.GetParams().GetDecorations().GetTags(), LabelPool)
}

// getAnalyticsName parses the v1 request tags for the analytics_name.
func getAnalyticsName(v1 *test_platform.Request) string {
	return getTag(v1.GetParams().GetDecorations().GetTags(), AnalyticsName)
}

// IsDDDSuite will return if the suite is to run in ddd.
// TODO (b:327505895): For now, use the ddd prefix, but long term will move to a proper flag.
func IsDDDSuite(v1 *test_platform.Request) bool {
	return strings.HasPrefix(getAnalyticsName(v1), "ddd")
}

// GetRetryCount returns the retry count from v1 request.
// By default it will be 0 which means no retry.
func GetRetryCount(v1 *test_platform.Request) int64 {
	if v1.GetParams().GetRetry().GetAllow() {
		return int64(v1.GetParams().GetRetry().GetMax())
	}

	return 0
}

// GetBuildType parses the software dependency's ChromeosBuild
// into the build type by taking the last part after removing postfixes.
func GetBuildType(softwareDeps []*test_platform.Request_Params_SoftwareDependency) string {
	for _, softwareDep := range softwareDeps {
		switch dep := softwareDep.GetDep().(type) {
		case *test_platform.Request_Params_SoftwareDependency_ChromeosBuild:
			chromeosBuildLeft := strings.Split(dep.ChromeosBuild, "/")[0]
			chromeosBuildParts := strings.Split(chromeosBuildLeft, "-")
			// Strip post-fixes.
			if slices.Contains(ExcludedChromeosBuildPostfixes, chromeosBuildParts[len(chromeosBuildParts)-1]) {
				chromeosBuildParts = chromeosBuildParts[:len(chromeosBuildParts)-1]
			}
			return chromeosBuildParts[len(chromeosBuildParts)-1]
		}
	}
	return ""
}

// GetVariant parses the software dependency's ChromeosBuild
// and removes the prefixes, postfixes, and base board.
func GetVariant(softwareDeps []*test_platform.Request_Params_SoftwareDependency) string {
	for _, softwareDep := range softwareDeps {
		switch dep := softwareDep.GetDep().(type) {
		case *test_platform.Request_Params_SoftwareDependency_ChromeosBuild:
			chromeosBuildLeft := strings.Split(dep.ChromeosBuild, "/")[0]
			chromeosBuildParts := strings.Split(chromeosBuildLeft, "-")
			// Strip post-fixes.
			if slices.Contains(ExcludedChromeosBuildPostfixes, chromeosBuildParts[len(chromeosBuildParts)-1]) {
				chromeosBuildParts = chromeosBuildParts[:len(chromeosBuildParts)-1]
			}
			// Strip pre-fixes.
			if slices.Contains(ExcludedChromeosBuildPrefixes, chromeosBuildParts[0]) {
				chromeosBuildParts = chromeosBuildParts[1:]
			}
			if len(chromeosBuildParts) <= 1 {
				return ""
			}
			// Remove base board and build type.
			chromeosBuildParts = chromeosBuildParts[1 : len(chromeosBuildParts)-1]
			return strings.Join(chromeosBuildParts, "-")
		}
	}
	return ""
}

// getImageGcsPath parses through the software dependencies
// to build out the gcs image path for this request.
func getImageGcsPath(softwareDeps []*test_platform.Request_Params_SoftwareDependency) string {
	chromeosBuild := ""
	chromeosBuildGcsBucket := ""
	for _, softwareDep := range softwareDeps {
		switch dep := softwareDep.GetDep().(type) {
		case *test_platform.Request_Params_SoftwareDependency_ChromeosBuild:
			chromeosBuild = dep.ChromeosBuild
		case *test_platform.Request_Params_SoftwareDependency_ChromeosBuildGcsBucket:
			chromeosBuildGcsBucket = dep.ChromeosBuildGcsBucket
		case *test_platform.Request_Params_SoftwareDependency_LacrosGcsPath:
			return dep.LacrosGcsPath
		}
	}
	if chromeosBuild == "" {
		return ""
	}
	if chromeosBuildGcsBucket == "" {
		chromeosBuildGcsBucket = DefaultChromeosBuildGcsBucket
	}
	return fmt.Sprintf("gs://%s/%s", chromeosBuildGcsBucket, chromeosBuild)
}

// getTag parses a list of tags in the format of "k:v".
func getTag(tags []string, targetTag string) string {
	for _, tag := range tags {
		splitTag := strings.Split(tag, ":")
		if len(splitTag) == 2 && splitTag[0] == targetTag {
			return splitTag[1]
		}
	}
	return ""
}
