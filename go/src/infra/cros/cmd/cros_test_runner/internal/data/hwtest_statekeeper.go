// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package data

import (
	test_api "go.chromium.org/chromiumos/config/go/test/api"
	lab_api "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"google.golang.org/protobuf/types/known/anypb"

	"infra/cros/cmd/cros_test_runner/internal/interfaces"
	"infra/cros/cmd/cros_test_runner/internal/tools/crostoolrunner"
)

// HwTestStateKeeper represents all the data hw test execution flow requires.
type HwTestStateKeeper struct {
	interfaces.StateKeeper

	// Set from input
	CftTestRequest *skylab_test_runner.CFTTestRequest

	// Dut related
	HostName         string
	DutTopology      *lab_api.DutTopology
	DutServerAddress *lab_api.IpEndpoint

	// Provsion related
	InstallMetadata *anypb.Any

	// Test related
	TestResponses *test_api.CrosTestResponse

	// Publish related
	GcsUrl              string
	StainlessUrl        string
	GcsPublishSrcDir    string
	CurrentInvocationId string
	TkoPublishSrcDir    string

	// Tools and their related dependencies
	Ctr                   *crostoolrunner.CrosToolRunner
	DockerKeyFileLocation string
}
