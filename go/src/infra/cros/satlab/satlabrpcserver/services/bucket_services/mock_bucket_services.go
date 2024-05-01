// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package bucket_services

import (
	"context"

	"github.com/stretchr/testify/mock"

	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
)

// MockBucketServices This object is only for testing
//
// Object should provide the same functions that `IBucketServices` interfaces provide.
// TODO: I will write a generator for the interface later to generate this file
type MockBucketServices struct {
	mock.Mock
}

// IsBucketInAsia Mock the function instead of calling an API.
func (m *MockBucketServices) IsBucketInAsia(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

// GetMilestones Mock the function instead of calling an API.
func (m *MockBucketServices) GetMilestones(ctx context.Context, board string) ([]string, error) {
	args := m.Called(ctx, board)
	return args.Get(0).([]string), args.Error(1)
}

// GetBuilds Mock the function instead of calling an API.
func (m *MockBucketServices) GetBuilds(ctx context.Context, board string, milestone int32) ([]string, error) {
	args := m.Called(ctx, board, milestone)
	return args.Get(0).([]string), args.Error(1)
}

// ListTestplans list all testplan json in partner bucket under a `testplans` folder
func (m *MockBucketServices) ListTestplans(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

// GetTestPlan get the test plan's content from the given filename.
func (m *MockBucketServices) GetTestPlan(ctx context.Context, name string) (*test_platform.Request_TestPlan, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*test_platform.Request_TestPlan), args.Error(1)
}

func (m *MockBucketServices) UploadLog(ctx context.Context, name, path string) (string, error) {
	args := m.Called(ctx, name, path)
	return args.String(0), args.Error(1)
}
