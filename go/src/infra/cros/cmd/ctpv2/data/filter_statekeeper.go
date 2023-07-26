// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package data

import (
	"infra/cros/cmd/common_lib/interfaces"

	testapi "go.chromium.org/chromiumos/config/go/test/api"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
)

// PreLocalTestStateKeeper represents all the data pre local test execution flow requires.
type FilterStateKeeper struct {
	interfaces.StateKeeper

	CtpV2Req                *testapi.CTPv2Request
	CtpV2Response           *testapi.CTPv2Response
	InitialInternalTestPlan *testapi.InternalTestplan
	TestPlanStates          []*testapi.InternalTestplan

	// Tools and their related dependencies
	Ctr                   *crostoolrunner.CrosToolRunner
	DockerKeyFileLocation string
}
