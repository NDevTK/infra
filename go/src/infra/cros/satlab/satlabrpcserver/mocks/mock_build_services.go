// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	moblabapipb "google.golang.org/genproto/googleapis/chromeos/moblab/v1beta1"

	"infra/cros/satlab/satlabrpcserver/services/build_services"
)

// MockBuildServices This object is only for testing
//
// Object should provide the same functions that `IBuildServices` interfaces provide.
// TODO: I will write a generator for the interface later to generate this file
type MockBuildServices struct {
	mock.Mock
}

// ListBuildTargets Mock the function instead of calling an API.
func (m *MockBuildServices) ListBuildTargets(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

// ListModels Mock the function instead of calling an API.
func (m *MockBuildServices) ListModels(ctx context.Context, board string) ([]string, error) {
	args := m.Called(ctx, board)
	return args.Get(0).([]string), args.Error(1)
}

// ListAvailableMilestones Mock the function instead of calling an API.
func (m *MockBuildServices) ListAvailableMilestones(ctx context.Context, board, model string) ([]string, error) {
	args := m.Called(ctx, board, model)
	return args.Get(0).([]string), args.Error(1)
}

// FindMostStableBuild Mock the function instead of calling an API.
func (m *MockBuildServices) FindMostStableBuild(ctx context.Context, board string) (string, error) {
	args := m.Called(ctx, board)
	return args.String(0), args.Error(1)
}

// ListBuildsForMilestone Mock the function instead of calling an API.
func (m *MockBuildServices) ListBuildsForMilestone(ctx context.Context, board, model string, milestone int32) ([]*build_services.BuildVersion, error) {
	args := m.Called(ctx, board, model, milestone)
	return args.Get(0).([]*build_services.BuildVersion), args.Error(1)
}

// CheckBuildStageStatus Mock the function instead of calling an API.
func (m *MockBuildServices) CheckBuildStageStatus(ctx context.Context, board, model, buildVersion, bucketName string) (bool, error) {
	args := m.Called(ctx, board, model, buildVersion, bucketName)
	return args.Bool(0), args.Error(1)
}

// StageBuild Mock the function instead of calling an API.
func (m *MockBuildServices) StageBuild(ctx context.Context, board, model, buildVersion, bucketName string) (*moblabapipb.BuildArtifact, error) {
	args := m.Called(ctx, board, model, buildVersion, bucketName)
	return args.Get(0).(*moblabapipb.BuildArtifact), args.Error(1)
}

// Close Mock the function instead of calling an API.
func (m *MockBuildServices) Close() error {
	return nil
}
