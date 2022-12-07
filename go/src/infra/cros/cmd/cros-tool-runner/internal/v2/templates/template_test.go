// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package templates

import (
	"testing"

	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/errors"
)

// mockPortDiscoverer mocks defaultPortDiscoverer for testing
type mockPortDiscoverer struct {
	portDiscoverer
	portDiscoverFunc func(*api.StartTemplatedContainerRequest) (*api.Container_PortBinding, error)
}

func (m *mockPortDiscoverer) discoverPort(req *api.StartTemplatedContainerRequest) (*api.Container_PortBinding, error) {
	return m.portDiscoverFunc(req)
}

func getMockPortDiscovererWithSuccess(containerPort int32) portDiscoverer {
	return &mockPortDiscoverer{
		portDiscoverFunc: func(request *api.StartTemplatedContainerRequest) (*api.Container_PortBinding, error) {
			return &api.Container_PortBinding{ContainerPort: containerPort}, nil
		},
	}
}

func getMockPortDiscovererWithError(errMsg string) portDiscoverer {
	return &mockPortDiscoverer{
		portDiscoverFunc: func(request *api.StartTemplatedContainerRequest) (*api.Container_PortBinding, error) {
			return nil, errors.New(errMsg)
		},
	}
}

func check(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%v should match %v", a, b)
	}
}
