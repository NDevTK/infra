// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package data

import (
	testapi "go.chromium.org/chromiumos/config/go/test/api"
	artifactpb "go.chromium.org/chromiumos/config/go/test/artifact"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"go.chromium.org/luci/luciexe/build"
	"google.golang.org/protobuf/types/known/anypb"
	vmlabapi "infra/libs/vmlab/api"

	"infra/cros/cmd/cros_test_runner/internal/interfaces"
	"infra/cros/cmd/cros_test_runner/internal/tools/crostoolrunner"
)

// HwTestStateKeeper represents all the data hw test execution flow requires.
type HwTestStateKeeper struct {
	interfaces.StateKeeper

	// Build related
	BuildState *build.State

	// Set from input
	CftTestRequest *skylab_test_runner.CFTTestRequest

	// Dut related
	HostName         string
	DutTopology      *labapi.DutTopology
	DutServerAddress *labapi.IpEndpoint
	// Only when DUT is a VM
	DutVmGceImage *vmlabapi.GceImage
	DutVm         *vmlabapi.VmInstance

	// Provision related
	InstallMetadata *anypb.Any
	ProvisionResp   *testapi.InstallResponse

	// Test related
	TestResponses *testapi.CrosTestResponse

	// Publish related
	GcsUrl              string
	StainlessUrl        string
	TesthausUrl         string
	GcsPublishSrcDir    string
	CurrentInvocationId string
	TkoPublishSrcDir    string
	TestResultForRdb    *artifactpb.TestResult

	// Build related
	SkylabResult *skylab_test_runner.Result

	// Tools and their related dependencies
	Ctr                   *crostoolrunner.CrosToolRunner
	DockerKeyFileLocation string
}
