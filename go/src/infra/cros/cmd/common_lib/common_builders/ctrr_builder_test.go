// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common_builders_test

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	buildapi "go.chromium.org/chromiumos/config/go/build/api"
	"go.chromium.org/chromiumos/config/go/test/api"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	tpcommon "go.chromium.org/chromiumos/infra/proto/go/test_platform/common"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"

	"infra/cros/cmd/common_lib/common"
	builders "infra/cros/cmd/common_lib/common_builders"
)

func TestCrosTestRunnerRequestBuilder(t *testing.T) {
	builder := &builders.CrosTestRunnerRequestBuilder{}

	Convey("Empty CftTestRequest All Skipped", t, func() {
		constructor := &builders.CftCrosTestRunnerRequestConstructor{
			Cft: &skylab_test_runner.CFTTestRequest{
				StepsConfig: &tpcommon.CftStepsConfig{
					ConfigType: &tpcommon.CftStepsConfig_HwTestConfig{
						HwTestConfig: &tpcommon.HwTestConfig{
							SkipStartingDutService: true,
							SkipProvision:          true,
							SkipTestExecution:      true,
							SkipAllResultPublish:   true,
						},
					},
				},
			},
		}
		request := builder.Build(constructor)

		expected := &api.CrosTestRunnerDynamicRequest{
			StartRequest: &api.CrosTestRunnerDynamicRequest_Build{
				Build: &api.BuildMode{},
			},
			Params: &api.CrosTestRunnerParams{
				ContainerMetadata: &buildapi.ContainerMetadata{
					Containers: make(map[string]*buildapi.ContainerImageMap),
				},
				TestSuites: []*api.TestSuite{},
				Keyvals:    make(map[string]string),
			},
		}

		So(request.GetOrderedTasks(), ShouldHaveLength, 0)
		So(request.GetStartRequest(), ShouldResemble, expected.GetStartRequest())
		So(request.GetParams(), ShouldResemble, expected.GetParams())
	})

	Convey("Build Params and StartRequest", t, func() {
		constructor := &builders.CftCrosTestRunnerRequestConstructor{
			Cft: &skylab_test_runner.CFTTestRequest{
				ParentRequestUid: "parent",
				PrimaryDut: &skylab_test_runner.CFTTestRequest_Device{
					DutModel: &labapi.DutModel{
						BuildTarget: "test-board",
					},
				},
				AutotestKeyvals: map[string]string{
					"fizz": "buzz",
				},
				TestSuites: []*api.TestSuite{
					{
						Name: "test1",
					},
				},
				StepsConfig: &tpcommon.CftStepsConfig{
					ConfigType: &tpcommon.CftStepsConfig_HwTestConfig{
						HwTestConfig: &tpcommon.HwTestConfig{
							SkipStartingDutService: true,
							SkipProvision:          true,
							SkipTestExecution:      true,
							SkipAllResultPublish:   true,
						},
					},
				},
			},
		}
		request := builder.Build(constructor)

		expected := &api.CrosTestRunnerDynamicRequest{
			StartRequest: &api.CrosTestRunnerDynamicRequest_Build{
				Build: &api.BuildMode{
					ParentRequestUid: "parent",
				},
			},
			Params: &api.CrosTestRunnerParams{
				ContainerMetadata: &buildapi.ContainerMetadata{
					Containers: make(map[string]*buildapi.ContainerImageMap),
				},
				TestSuites: []*api.TestSuite{
					{
						Name: "test1",
					},
				},
				Keyvals: map[string]string{
					"fizz":          "buzz",
					"primary-board": "test-board",
				},
			},
		}

		So(request.GetOrderedTasks(), ShouldHaveLength, 0)
		So(request.GetStartRequest(), ShouldResemble, expected.GetStartRequest())
		So(request.GetParams(), ShouldResemble, expected.GetParams())
	})

	Convey("Builds Tasks", t, func() {
		constructor := &builders.CftCrosTestRunnerRequestConstructor{
			Cft: &skylab_test_runner.CFTTestRequest{
				ParentRequestUid: "parent",
				PrimaryDut: &skylab_test_runner.CFTTestRequest_Device{
					DutModel: &labapi.DutModel{
						BuildTarget: "test-board",
					},
				},
				AutotestKeyvals: map[string]string{
					"fizz": "buzz",
				},
				TestSuites: []*api.TestSuite{
					{
						Name: "test1",
					},
				},
			},
		}
		request := builder.Build(constructor)

		expected := &api.CrosTestRunnerDynamicRequest{
			StartRequest: &api.CrosTestRunnerDynamicRequest_Build{
				Build: &api.BuildMode{
					ParentRequestUid: "parent",
				},
			},
			Params: &api.CrosTestRunnerParams{
				ContainerMetadata: &buildapi.ContainerMetadata{
					Containers: make(map[string]*buildapi.ContainerImageMap),
				},
				TestSuites: []*api.TestSuite{
					{
						Name: "test1",
					},
				},
				Keyvals: map[string]string{
					"fizz":          "buzz",
					"primary-board": "test-board",
				},
			},
		}

		So(request.GetOrderedTasks(), ShouldHaveLength, 5)
		So(request.GetStartRequest(), ShouldResemble, expected.GetStartRequest())
		So(request.GetParams(), ShouldResemble, expected.GetParams())
	})

	Convey("Builds Tasks with Companions", t, func() {
		constructor := &builders.CftCrosTestRunnerRequestConstructor{
			Cft: &skylab_test_runner.CFTTestRequest{
				ParentRequestUid: "parent",
				PrimaryDut: &skylab_test_runner.CFTTestRequest_Device{
					DutModel: &labapi.DutModel{
						BuildTarget: "test-board",
					},
				},
				ContainerMetadata: &buildapi.ContainerMetadata{
					Containers: map[string]*buildapi.ContainerImageMap{
						"default": {},
					},
				},
				CompanionDuts: []*skylab_test_runner.CFTTestRequest_Device{
					{
						DutModel: &labapi.DutModel{
							BuildTarget: "test-board",
						},
					},
					{
						DutModel: &labapi.DutModel{
							BuildTarget: "test-board",
						},
					},
					{
						DutModel: &labapi.DutModel{
							BuildTarget: "test-board",
						},
					},
				},
				AutotestKeyvals: map[string]string{
					"fizz":  "buzz",
					"build": "Release/R123.0.0-123",
				},
				TestSuites: []*api.TestSuite{
					{
						Name: "test1",
					},
				},
			},
		}
		request := builder.Build(constructor)

		expected := &api.CrosTestRunnerDynamicRequest{
			StartRequest: &api.CrosTestRunnerDynamicRequest_Build{
				Build: &api.BuildMode{
					ParentRequestUid: "parent",
				},
			},
			Params: &api.CrosTestRunnerParams{
				ContainerMetadata: &buildapi.ContainerMetadata{
					Containers: map[string]*buildapi.ContainerImageMap{
						"default": {
							Images: map[string]*buildapi.ContainerImageInfo{
								"cros-fw-provision": common.CreateTestServicesContainer("cros-fw-provision", common.DefaultCrosFwProvisionSha),
							},
						},
					},
				},
				TestSuites: []*api.TestSuite{
					{
						Name: "test1",
					},
				},
				Keyvals: map[string]string{
					"fizz":             "buzz",
					"build":            "Release/R123.0.0-123",
					"primary-board":    "test-board",
					"companion-boards": "test-board,test-board,test-board",
				},
			},
		}

		So(request.GetOrderedTasks(), ShouldHaveLength, 11)
		So(request.GetStartRequest(), ShouldResemble, expected.GetStartRequest())
		So(request.GetParams(), ShouldResemble, expected.GetParams())
		So(request.GetParams().GetContainerMetadata().GetContainers()["default"].GetImages()["cros-fw-provision"].GetDigest(), ShouldEqual, fmt.Sprintf("sha256:%s", common.DefaultCrosFwProvisionSha))
	})
}
