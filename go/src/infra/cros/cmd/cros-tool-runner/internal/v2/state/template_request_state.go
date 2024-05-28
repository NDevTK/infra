// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package state

import (
	"sync"

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
	state *sync.Map
}

// newTemplateRequestState returns an instance of templateRequestState
func newTemplateRequestState() TemplateRequestRecorder {
	return &templateRequestState{
		state: new(sync.Map),
	}
}

func (t *templateRequestState) Add(containerId string, request *api.StartTemplatedContainerRequest) {
	t.state.Store(containerId, request)
}

func (t *templateRequestState) Exist(containerId string) bool {
	if _, ok := t.state.Load(containerId); ok {
		return true
	}
	return false
}

func (t *templateRequestState) Get(containerId string) *api.StartTemplatedContainerRequest {
	request, ok := t.state.Load(containerId)
	if !ok {
		return nil
	}
	return request.(*api.StartTemplatedContainerRequest)
}
