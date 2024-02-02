// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package data

import (
	"container/list"

	buildapi "go.chromium.org/chromiumos/config/go/build/api"
	testapi "go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/luciexe/build"

	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
)

// PreLocalTestStateKeeper represents all the data pre local test execution flow requires.
type FilterStateKeeper struct {
	interfaces.StateKeeper

	CtpReq                  *testapi.CTPRequest
	CtpV2Response           *testapi.CTPv2Response
	InitialInternalTestPlan *testapi.InternalTestplan
	TestPlanStates          []*testapi.InternalTestplan
	Scheduler               testapi.SchedulerInfo_Scheduler

	// Build related
	BuildState *build.State

	// Container info queue
	ContainerInfoQueue *list.List

	// Dictionaries
	ContainerMetadataMap map[string]*buildapi.ContainerImageInfo
	ContainerInfoMap     *ContainerInfoMap

	// Tools and their related dependencies
	Ctr                   *crostoolrunner.CrosToolRunner
	DockerKeyFileLocation string

	MiddledOutResp *MiddleOutResponse
}
