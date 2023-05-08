// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"infra/cros/satlab/satlabrpcserver/utils"
)

// MockDUTServices This object is only for testing
//
// Object should provide the same functions that `IDUTServices` interfaces provide.
type MockDUTServices struct {
	mock.Mock
}

// RunCommandOnIP send the command to the DUT device and then get the result back
func (m *MockDUTServices) RunCommandOnIP(ctx context.Context, IP, cmd string) (*utils.SSHResult, error) {
	args := m.Called(ctx, IP, cmd)
	return args.Get(0).(*utils.SSHResult), args.Error(1)
}

// RunCommandOnIPs send the command to DUT devices and then get the result back
func (m *MockDUTServices) RunCommandOnIPs(ctx context.Context, IPs []string, cmd string) []*utils.SSHResult {
	args := m.Called(ctx, IPs, cmd)
	return args.Get(0).([]*utils.SSHResult)
}
