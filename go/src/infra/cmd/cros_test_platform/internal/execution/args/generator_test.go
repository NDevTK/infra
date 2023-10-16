// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package args contains the logic for assembling all data required for
// creating an individual task request.
package args

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	testapi "go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/config"
	bbpb "go.chromium.org/luci/buildbucket/proto"
)

func TestHasVariant(t *testing.T) {
	Convey("Given a request that build has variant", t, func() {
		ctx := context.Background()
		inv := basicInvocation()
		setTestName(inv, "foo-name")
		var params test_platform.Request_Params
		var dummyWorkerConfig = &config.Config_SkylabWorker{}
		setBuild(&params, "foo-arc-r-postsubmit/R106-12222.0.0")
		setRequestKeyval(&params, "suite", "foo-suite")
		setRequestMaximumDuration(&params, 1000)
		setPrimayDeviceBoard(&params, "foo")
		setPrimayDeviceModel(&params, "model")
		setRunViaCft(&params, true)
		Convey("when generating a test runner request's args", func() {
			g := Generator{
				Invocation:       inv,
				Params:           &params,
				WorkerConfig:     dummyWorkerConfig,
				ParentRequestUID: "TestPlanRuns/12345678/foo",
			}
			got, err := g.GenerateArgs(ctx)
			So(err, ShouldBeNil)
			Convey("the container metadata key is correct has variant", func() {
				So(got.CFTTestRunnerRequest.PrimaryDut.ContainerMetadataKey, ShouldEqual, "foo-arc-r")
			})

		})
	})
}

func TestDisplayNameTagsForUnamedRequest(t *testing.T) {
	Convey("Given a request does not specify a display name", t, func() {
		ctx := context.Background()
		inv := basicInvocation()
		setTestName(inv, "foo-name")
		var params test_platform.Request_Params
		var dummyWorkerConfig = &config.Config_SkylabWorker{}
		setBuild(&params, "foo-build")
		setRequestKeyval(&params, "suite", "foo-suite")
		setRequestMaximumDuration(&params, 1000)
		Convey("when generating a test runner request's args", func() {
			g := Generator{
				Invocation:       inv,
				Params:           &params,
				WorkerConfig:     dummyWorkerConfig,
				ParentRequestUID: "TestPlanRuns/12345678/foo",
			}
			got, err := g.GenerateArgs(ctx)
			So(err, ShouldBeNil)
			Convey("the display name tag is generated correctly.", func() {
				So(got.SwarmingTags, ShouldContain, "display_name:foo-build/foo-suite/foo-name")
				So(got.ParentRequestUID, ShouldEqual, "TestPlanRuns/12345678/foo")
			})
		})
	})
}

func TestInventoryLabels(t *testing.T) {
	Convey("Given a request with board and model info", t, func() {
		ctx := context.Background()
		inv := basicInvocation()
		setTestName(inv, "foo-name")
		var params test_platform.Request_Params
		var dummyWorkerConfig = &config.Config_SkylabWorker{}
		setRequestMaximumDuration(&params, 1000)
		setPrimayDeviceBoard(&params, "coral")
		setPrimayDeviceModel(&params, "babytiger")
		Convey("when generating a test runner request's args", func() {
			g := Generator{
				Invocation:   inv,
				Params:       &params,
				WorkerConfig: dummyWorkerConfig,
			}
			got, err := g.GenerateArgs(ctx)
			So(err, ShouldBeNil)
			Convey("the SchedulableLabels is generated correctly", func() {
				So(*got.SchedulableLabels.Board, ShouldEqual, "coral")
				So(*got.SchedulableLabels.Model, ShouldEqual, "babytiger")
				So(len(got.SecondaryDevicesLabels), ShouldEqual, 0)
			})
		})
	})
}

func TestSecondaryDevicesLabels(t *testing.T) {
	Convey("Given a request with secondary devices", t, func() {
		ctx := context.Background()
		inv := basicInvocation()
		setTestName(inv, "foo-name")
		var params test_platform.Request_Params
		var dummyWorkerConfig = &config.Config_SkylabWorker{}
		setRequestMaximumDuration(&params, 1000)
		setSecondaryDevice(&params, "nami", "", "")
		setSecondaryDevice(&params, "coral", "babytiger", "")
		Convey("when generating a test runner request's args", func() {
			g := Generator{
				Invocation:   inv,
				Params:       &params,
				WorkerConfig: dummyWorkerConfig,
			}
			got, err := g.GenerateArgs(ctx)
			So(err, ShouldBeNil)
			Convey("the SecondaryDevicesLabels is generated correctly", func() {
				So(len(got.SecondaryDevicesLabels), ShouldEqual, 2)
				So(*got.SecondaryDevicesLabels[0].Board, ShouldEqual, "nami")
				So(*got.SecondaryDevicesLabels[0].Model, ShouldEqual, "")
				So(*got.SecondaryDevicesLabels[1].Board, ShouldEqual, "coral")
				So(*got.SecondaryDevicesLabels[1].Model, ShouldEqual, "babytiger")
			})
		})
	})
}

func TestExperiments(t *testing.T) {
	Convey("Given a request with experiments", t, func() {
		ctx := context.Background()
		inv := basicInvocation()
		setTestName(inv, "foo-name")
		var params test_platform.Request_Params
		var dummyWorkerConfig = &config.Config_SkylabWorker{}
		setRequestMaximumDuration(&params, 1000)
		Convey("when generating a test runner request's args", func() {
			g := Generator{
				Invocation:   inv,
				Params:       &params,
				WorkerConfig: dummyWorkerConfig,
				Experiments:  []string{"exp1", "exp2"},
			}
			got, err := g.GenerateArgs(ctx)
			So(err, ShouldBeNil)
			Convey("the experiments field is propogated correctly", func() {
				So(len(got.Experiments), ShouldEqual, 2)
				So(got.Experiments, ShouldResemble, []string{"exp1", "exp2"})
			})
		})
	})
}

func TestGerritChanges(t *testing.T) {
	Convey("Given Gerrit Changes", t, func() {
		ctx := context.Background()
		inv := basicInvocation()
		setTestName(inv, "foo-name")
		var params test_platform.Request_Params
		var dummyWorkerConfig = &config.Config_SkylabWorker{}
		setRequestMaximumDuration(&params, 1000)
		Convey("when generating a test runner request's args", func() {
			gc := &bbpb.GerritChange{
				Host:     "a",
				Project:  "b",
				Change:   123,
				Patchset: 1,
			}
			g := Generator{
				Invocation:    inv,
				Params:        &params,
				WorkerConfig:  dummyWorkerConfig,
				GerritChanges: []*bbpb.GerritChange{gc},
			}
			got, err := g.GenerateArgs(ctx)
			So(err, ShouldBeNil)
			Convey("the GerritChanges are added correctly", func() {
				So(len(got.GerritChanges), ShouldEqual, 1)
				So(got.GerritChanges, ShouldResemble, []*bbpb.GerritChange{gc})
			})
		})
	})
}

// TestSwarmingPool ensures we correctly pass on the `SwarmingPool` arg
func TestSwarmingPool(t *testing.T) {
	Convey("Given SwarmingPool", t, func() {
		ctx := context.Background()
		inv := basicInvocation()
		setTestName(inv, "foo-name")
		var params test_platform.Request_Params
		var dummyWorkerConfig = &config.Config_SkylabWorker{}
		setRequestMaximumDuration(&params, 1000)
		Convey("when generating a test runner request's args", func() {
			g := Generator{
				Invocation:   inv,
				Params:       &params,
				WorkerConfig: dummyWorkerConfig,
				SwarmingPool: "OtherPool",
			}
			got, err := g.GenerateArgs(ctx)
			So(err, ShouldBeNil)
			Convey("the SwarmingPool should be added correctly", func() {
				So(got.SwarmingPool, ShouldEqual, "OtherPool")
			})
		})
	})
}

// TestAndroidProvisionState ensures we correctly pass android provisioning metadata through provision state
func TestAndroidProvisionState(t *testing.T) {
	Convey("Given a request with Android provision metadata in software deps", t, func() {
		ctx := context.Background()
		inv := basicInvocation()
		setTestName(inv, "foo-name")
		var params test_platform.Request_Params
		var dummyWorkerConfig = &config.Config_SkylabWorker{}
		setBuild(&params, "foo-build")
		setRequestKeyval(&params, "suite", "foo-suite")
		setRequestMaximumDuration(&params, 1000)
		setRunViaCft(&params, true)
		setSecondaryDevice(&params, "coral", "babytiger", "")
		Convey("when generating a cft test runner request's args with nil android provision metadata", func() {
			g := Generator{
				Invocation:       inv,
				Params:           &params,
				WorkerConfig:     dummyWorkerConfig,
				ParentRequestUID: "TestPlanRuns/12345678/foo",
			}
			_, err := g.GenerateArgs(ctx)
			So(err, ShouldNotBeNil)
		})
		params.SecondaryDevices = []*test_platform.Request_Params_SecondaryDevice{}
		setAndroidSecondaryDeviceWithAndroidProvisionMetadata(&params, "androidBoard", "Pixel6", "11", "latest_stable")
		Convey("when generating a cft test runner request's args with not nil android provision metadata", func() {
			g := Generator{
				Invocation:       inv,
				Params:           &params,
				WorkerConfig:     dummyWorkerConfig,
				ParentRequestUID: "TestPlanRuns/12345678/foo",
			}
			got, err := g.GenerateArgs(ctx)
			So(err, ShouldBeNil)
			Convey("provision state should have androidProvisionRequestMetadata as provision state when android metadata is passed", func() {

				companionDuts := got.CFTTestRunnerRequest.GetCompanionDuts()
				for _, companionDut := range companionDuts {
					var androidProvisionRequestMetadata testapi.AndroidProvisionRequestMetadata
					err = companionDut.ProvisionState.ProvisionMetadata.UnmarshalTo(&androidProvisionRequestMetadata)
					So(err, ShouldBeNil)
					cipdPackage := &testapi.CIPDPackage{
						AndroidPackage: 1,
						VersionOneof: &testapi.CIPDPackage_Ref{
							Ref: "latest_stable",
						},
					}
					So(androidProvisionRequestMetadata.GetAndroidOsImage().GetOsVersion(), ShouldEqual, "11")
					So(androidProvisionRequestMetadata.GetCipdPackages(), ShouldContain, cipdPackage)
				}
			})
		})
	})
}

func TestResultConfig(t *testing.T) {
	Convey("Given ResultConfigs Changes", t, func() {
		ctx := context.Background()
		inv := basicInvocation()
		setTestName(inv, "foo-name")
		var params test_platform.Request_Params
		setTestUploadVisibility(&params, test_platform.Request_Params_ResultsUploadConfig_TEST_RESULTS_VISIBILITY_CUSTOM_REALM)
		var dummyWorkerConfig = &config.Config_SkylabWorker{}
		setRequestMaximumDuration(&params, 1000)
		Convey("when generating a test runner request's args", func() {
			g := Generator{
				Invocation:   inv,
				Params:       &params,
				WorkerConfig: dummyWorkerConfig,
			}
			got, err := g.GenerateArgs(ctx)
			So(err, ShouldBeNil)
			Convey("the ResultsConfig is added correctly", func() {
				So(got.ResultsConfig.Mode, ShouldEqual, test_platform.Request_Params_ResultsUploadConfig_TEST_RESULTS_VISIBILITY_CUSTOM_REALM)
			})
		})
	})
}
