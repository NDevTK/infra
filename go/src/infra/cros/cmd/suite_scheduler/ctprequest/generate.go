// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package ctprequest will build and return a CTP request to be handled by the CTP
// BuildBucket builder.
package ctprequest

import (
	"fmt"

	"google.golang.org/protobuf/types/known/durationpb"

	"go.chromium.org/chromiumos/infra/proto/go/chromiumos"
	requestpb "go.chromium.org/chromiumos/infra/proto/go/test_platform"
	infrapb "go.chromium.org/chromiumos/infra/proto/go/testplans"

	"infra/cros/cmd/suite_scheduler/configparser"
)

const (
	GSPrefix                  = "gs://chromeos-image-archive/"
	ContainerMetadataLocation = "/metadata/containers.jsonpb"

	MaxRetry = 3

	Fortnightly = int64(240)
	Weekly      = int64(230)
	Daily       = int64(200)
	PostBuild   = int64(170)

	// CTP requests with a qs account will not require a priority so apply the
	// default swarming value.
	Default = int64(140)
)

// priorityMap returns the proper swarming priority value for the given launch profile type.
var priorityMap = map[infrapb.SchedulerConfig_LaunchCriteria_LaunchProfile]int64{
	infrapb.SchedulerConfig_LaunchCriteria_NEW_BUILD:   PostBuild,
	infrapb.SchedulerConfig_LaunchCriteria_DAILY:       Daily,
	infrapb.SchedulerConfig_LaunchCriteria_WEEKLY:      Weekly,
	infrapb.SchedulerConfig_LaunchCriteria_FORTNIGHTLY: Fortnightly,
}

// getSwarmingDimensions reads the configs runOptions dimensions and formats
// them how the CTP request expects them.
func getSwarmingDimensions(config *infrapb.SchedulerConfig) []string {
	dims := []string{}

	for _, dim := range config.RunOptions.Dimensions {
		str := fmt.Sprintf("%s:%s", dim.Key, dim.Value)

		dims = append(dims, str)
	}
	return dims
}

// getSchedulingFields transforms SuSch SchedulerConfig_PoolOptions into ctp SchedulerConfig_LaunchCriteria_LaunchProfile.
func getSchedulingFields(PoolOptions *infrapb.SchedulerConfig_PoolOptions, launchType infrapb.SchedulerConfig_LaunchCriteria_LaunchProfile) *requestpb.Request_Params_Scheduling {

	schedParams := &requestpb.Request_Params_Scheduling{
		QsAccount: PoolOptions.QsAccount,
	}

	if PoolOptions.QsAccount == "" {
		// Get the priority for the run.
		priority := Default
		if val, ok := priorityMap[launchType]; ok {
			priority = val
		}
		schedParams.Priority = priority

	}

	// Because of the proto typing we need cast the pool to one of these values.
	// In the CTP run the key of managedPool versus unManagedPool matters.
	if managedPool, ok := requestpb.Request_Params_Scheduling_ManagedPool_value[PoolOptions.Pool]; ok {
		schedParams.Pool = &requestpb.Request_Params_Scheduling_ManagedPool_{ManagedPool: requestpb.Request_Params_Scheduling_ManagedPool(managedPool)}
	} else {
		schedParams.Pool = &requestpb.Request_Params_Scheduling_UnmanagedPool{UnmanagedPool: PoolOptions.Pool}

	}

	return schedParams
}

func getTimeoutSeconds(timeoutMins int32) int64 {
	return int64(timeoutMins) * 60
}

func getTags(board, model, build string, config *infrapb.SchedulerConfig) []string {
	tags := []string{
		fmt.Sprintf("build:%s", build),
		fmt.Sprintf("label-pool:%s", config.PoolOptions.Pool),
		fmt.Sprintf("ctp-fwd-task-name:%s", config.Name),
		fmt.Sprintf("suite:%s", config.Suite),
		fmt.Sprintf("analytics_name:%s", config.AnalyticsName),
	}

	if board != "" {
		tags = append(tags, fmt.Sprintf("label-board:%s", board))
	}
	if model != "" {
		tags = append(tags, fmt.Sprintf("label-model:%s", model))
	}

	return tags
}

func getHardwareAttributes(model string) *requestpb.Request_Params_HardwareAttributes {
	if model != "" {
		return &requestpb.Request_Params_HardwareAttributes{Model: model}
	}

	return nil
}

func getRetryParams(retry bool) *requestpb.Request_Params_Retry {
	if retry {
		return &requestpb.Request_Params_Retry{
			Allow: true,
			Max:   MaxRetry,
		}
	}
	return nil

}

func getTestPlan(config *infrapb.SchedulerConfig) *requestpb.Request_TestPlan {
	testPlan := &requestpb.Request_TestPlan{
		Suite: []*requestpb.Request_Suite{
			{
				Name: config.Suite,
			},
		},
		TagCriteria: config.RunOptions.TagCriteria,
	}
	return testPlan
}

func formBuildImage(buildTarget, buildMilestone, buildVersion string) string {
	return buildTarget + "-release" + "/R" + buildMilestone + "-" + buildVersion
}

// BuildCTPRequest takes information from a SuSch config and builds the
// corresponding CTP request.
func BuildCTPRequest(config *infrapb.SchedulerConfig, board, model, buildTarget, buildMilestone, buildVersion string) *requestpb.Request {
	buildImage := formBuildImage(buildTarget, buildMilestone, buildVersion)

	request := &requestpb.Request{
		Params: &requestpb.Request_Params{
			HardwareAttributes: getHardwareAttributes(model),
			SoftwareAttributes: &requestpb.Request_Params_SoftwareAttributes{
				BuildTarget: &chromiumos.BuildTarget{
					Name: board,
				},
			},

			FreeformAttributes: &requestpb.Request_Params_FreeformAttributes{
				SwarmingDimensions: getSwarmingDimensions(config),
			},
			SoftwareDependencies: []*requestpb.Request_Params_SoftwareDependency{
				{
					Dep: &requestpb.Request_Params_SoftwareDependency_ChromeosBuild{
						ChromeosBuild: buildImage,
					},
				},
			},
			Scheduling: getSchedulingFields(config.PoolOptions, config.LaunchCriteria.LaunchProfile),
			Retry:      getRetryParams(config.RunOptions.Retry),
			// TODO(b:305792113): Get build information from the release-build pipeline.
			Metadata: &requestpb.Request_Params_Metadata{
				// Some gsURL
				TestMetadataUrl: GSPrefix + buildImage,
				// Some gsURL same as above
				DebugSymbolsArchiveUrl: GSPrefix + buildImage,

				ContainerMetadataUrl: GSPrefix + buildImage + ContainerMetadataLocation,
			},
			Time: &requestpb.Request_Params_Time{
				MaximumDuration: &durationpb.Duration{Seconds: getTimeoutSeconds(config.RunOptions.TimeoutMins)},
			},
			Decorations: &requestpb.Request_Params_Decorations{
				Tags: getTags(board, model, buildImage, config),
			},
			RunViaCft: config.RunOptions.RunViaCft,
		},
		TestPlan: getTestPlan(config),
	}

	return request
}

// BuildAllCTPRequests Generates all potential CTP options for the given configuration.
func BuildAllCTPRequests(config *infrapb.SchedulerConfig, targets configparser.TargetOptions) CTPRequests {
	requests := CTPRequests{}
	for _, target := range targets {
		if len(target.Models) > 0 {
			for _, model := range target.Models {
				request := BuildCTPRequest(config, string(target.Board), model, "", "", "")
				requests = append(requests, request)
			}
		} else {
			request := BuildCTPRequest(config, string(target.Board), "", "", "", "")
			requests = append(requests, request)

		}
	}

	return requests
}
