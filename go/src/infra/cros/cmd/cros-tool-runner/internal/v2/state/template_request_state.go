// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package state

import (
	"go.chromium.org/chromiumos/config/go/test/api"
)

// TemplateRequestRecorder is the interface to track the state of
// StartTemplatedContainerRequest.
type TemplateRequestRecorder interface {
	// Add container ID to Request mapping to the state.
	Add(containerId string, request *api.StartTemplatedContainerRequest)
	// Exist checks if a container ID exists in the state.
	Exist(containerId string) bool
	// Get Request by container ID from the state.
	Get(containerId string) *api.StartTemplatedContainerRequest
}

// templateRequestState is the implementation of TemplateRequestRecorder. It
// uses a map to track the state.
type templateRequestState struct {
	state map[string]*api.StartTemplatedContainerRequest
}

// newTemplateRequestState returns an instance of templateRequestState
func newTemplateRequestState() TemplateRequestRecorder {
	return &templateRequestState{state: make(map[string]*api.StartTemplatedContainerRequest)}
}

func (t *templateRequestState) Add(containerId string, request *api.StartTemplatedContainerRequest) {
	t.state[containerId] = request
}

func (t *templateRequestState) Exist(containerId string) bool {
	if _, ok := t.state[containerId]; ok {
		return true
	}
	return false
}

func (t *templateRequestState) Get(containerId string) *api.StartTemplatedContainerRequest {
	return t.state[containerId]
}
