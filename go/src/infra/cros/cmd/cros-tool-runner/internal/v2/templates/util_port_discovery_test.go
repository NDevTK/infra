// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package templates

import (
	"errors"
	"testing"

	"go.chromium.org/chromiumos/config/go/test/api"
)

type mockTemplateProcessor struct {
	TemplateProcessor
	portDiscoverFunc func(*api.StartTemplatedContainerRequest) (*api.Container_PortBinding, error)
}

func (m *mockTemplateProcessor) discoverPort(req *api.StartTemplatedContainerRequest) (*api.Container_PortBinding, error) {
	return m.portDiscoverFunc(req)
}

func TestGetPortDiscoveryBindings_nilRequest(t *testing.T) {
	util := TemplateUtils

	bindings := util.getPortDiscoveryBindings(nil)

	if bindings == nil || len(bindings) != 0 {
		t.Fatalf("Expect bindings to be empty.")
	}
}

func TestGetPortDiscoveryBindings_invalidRequest(t *testing.T) {
	util := TemplateUtils

	bindings := util.getPortDiscoveryBindings(&api.StartTemplatedContainerRequest{Template: &api.Template{}})

	if bindings == nil || len(bindings) != 0 {
		t.Fatalf("Expect bindings to be empty.")
	}
}

func TestGetPortDiscoveryBindings_swallowError(t *testing.T) {
	util := templateUtils{templateRouter: &mockTemplateProcessor{
		portDiscoverFunc: func(req *api.StartTemplatedContainerRequest) (*api.Container_PortBinding, error) {
			return nil, errors.New("metadata does not exist")
		}}}

	bindings := util.getPortDiscoveryBindings(&api.StartTemplatedContainerRequest{})

	if bindings == nil || len(bindings) != 0 {
		t.Fatalf("Expect bindings to be empty.")
	}
}

func TestGetPortDiscoveryBindings_success(t *testing.T) {
	expectedBinding := &api.Container_PortBinding{ContainerPort: 42}
	util := templateUtils{templateRouter: &mockTemplateProcessor{
		portDiscoverFunc: func(req *api.StartTemplatedContainerRequest) (*api.Container_PortBinding, error) {
			return expectedBinding, nil
		}}}

	bindings := util.getPortDiscoveryBindings(&api.StartTemplatedContainerRequest{})

	if bindings == nil || len(bindings) != 1 {
		t.Fatalf("Expect one binding.")
	}

	if bindings[0] != expectedBinding {
		t.Fatalf("Result doesn't match\nexpect: %v\nactual: %v", expectedBinding, bindings[0])
	}
}
