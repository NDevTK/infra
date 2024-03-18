// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common_builders

import (
	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
)

// CrosTestRunnerRequestBuilder wraps the construction of a CrosTestRunnerRequest
// and contains the top-level object that gets built.
type CrosTestRunnerRequestBuilder struct {
	crosTestRunnerRequest *api.CrosTestRunnerDynamicRequest
}

// CrosTestRunnerRequestConstructor defines the interface by which
// concrete constructors may define how to construct the CrosTestRunnerRequest
// from varying internal objects.
type CrosTestRunnerRequestConstructor interface {
	ConstructStartRequest(*api.CrosTestRunnerDynamicRequest)
	ConstructParams(*api.CrosTestRunnerDynamicRequest)
	ConstructOrderedTasks(*api.CrosTestRunnerDynamicRequest)
}

// Build initializes the CrosTestRunnerRequest, constructs each top-level field
// using a given constructor, then returns the resulting construction.
func (builder *CrosTestRunnerRequestBuilder) Build(constructor CrosTestRunnerRequestConstructor) *api.CrosTestRunnerDynamicRequest {
	builder.initializeBuilder()

	constructor.ConstructStartRequest(builder.crosTestRunnerRequest)
	constructor.ConstructParams(builder.crosTestRunnerRequest)
	constructor.ConstructOrderedTasks(builder.crosTestRunnerRequest)

	return builder.crosTestRunnerRequest
}

// initializeBuilder sets the CrosTestRunnerRequest to a default empty
// state in which important high-level fields are safe to reference.
func (builder *CrosTestRunnerRequestBuilder) initializeBuilder() {
	builder.crosTestRunnerRequest = &api.CrosTestRunnerDynamicRequest{
		Params: &api.CrosTestRunnerParams{
			TestSuites: []*api.TestSuite{},
			Keyvals:    make(map[string]string),
		},
		OrderedTasks: []*api.CrosTestRunnerDynamicRequest_Task{},
	}
}

// Concrete CrosTestRunnerRequestConstructor that translates a CftTestRequest
// into the expected values within a CrosTestRunnerRequest.
type CftCrosTestRunnerRequestConstructor struct {
	CrosTestRunnerRequestConstructor

	Cft *skylab_test_runner.CFTTestRequest
}

// ConstructStartRequest builds a CrosTestRunnerRequest_StartRequest from
// a CftTestRequest.
func (constructor *CftCrosTestRunnerRequestConstructor) ConstructStartRequest(crosTestRunnerRequest *api.CrosTestRunnerDynamicRequest) {
	crosTestRunnerRequest.StartRequest = &api.CrosTestRunnerDynamicRequest_Build{
		Build: &api.BuildMode{
			ParentBuildId:    constructor.Cft.GetParentBuildId(),
			ParentRequestUid: constructor.Cft.GetParentRequestUid(),
		},
	}
}

// ConstructParams builds a CrosTestRunnerParams from
// a CftTestRequest.
func (constructor *CftCrosTestRunnerRequestConstructor) ConstructParams(crosTestRunnerRequest *api.CrosTestRunnerDynamicRequest) {
	params := &api.CrosTestRunnerParams{
		Keyvals:           constructor.Cft.GetAutotestKeyvals(),
		ContainerMetadata: PatchContainerMetadata(constructor.Cft.GetContainerMetadata(), constructor.Cft.GetAutotestKeyvals()["build"]),
		TestSuites:        constructor.Cft.GetTestSuites(),
	}

	if params.Keyvals == nil {
		params.Keyvals = make(map[string]string)
	}

	if params.TestSuites == nil {
		params.TestSuites = []*api.TestSuite{}
	}

	constructor.addDevicesInfoToKeyvals(params.Keyvals)

	crosTestRunnerRequest.Params = params
}

// ConstructOrderedTasks builds a slice of CrosTestRunnerRequest_Task from
// a CftTestRequest.
func (constructor *CftCrosTestRunnerRequestConstructor) ConstructOrderedTasks(crosTestRunnerRequest *api.CrosTestRunnerDynamicRequest) {
	orderedTasks := &[]*api.CrosTestRunnerDynamicRequest_Task{}

	constructor.buildPrimaryDutProvision(orderedTasks)
	constructor.buildCompanionDutProvisions(orderedTasks)
	constructor.buildTestExecution(orderedTasks)
	constructor.buildPublishes(orderedTasks)

	crosTestRunnerRequest.OrderedTasks = *orderedTasks
}
