// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package mocks

import "github.com/stretchr/testify/mock"

type MockCPUTemperature struct {
	mock.Mock
}

// GetCurrentCPUTemperature Get the current temperature of CPU
func (m *MockCPUTemperature) GetCurrentCPUTemperature() (float32, error) {
	args := m.Called()
	return args.Get(0).(float32), args.Error(1)
}
