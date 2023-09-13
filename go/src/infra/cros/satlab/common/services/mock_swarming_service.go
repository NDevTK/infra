// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package services

import (
	"context"

	"github.com/stretchr/testify/mock"
	swarmingapi "go.chromium.org/luci/swarming/proto/api_v2"
)

// MockSwarmingService This object is only for testing
//
// Object should provide the same functions that `ISwarmingService` interfaces provide.
// TODO: I will write a generator for the interface later to generate this file
type MockSwarmingService struct {
	mock.Mock
}

func (m *MockSwarmingService) GetBot(ctx context.Context, hostname string) (*swarmingapi.BotInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).(*swarmingapi.BotInfo), args.Error(1)
}

func (m *MockSwarmingService) ListBotTasks(ctx context.Context, hostname, cursor string, pageSize int) (*TasksIterator, error) {
	args := m.Called(ctx, hostname, cursor, pageSize)
	return args.Get(0).(*TasksIterator), args.Error(1)
}

func (m *MockSwarmingService) ListBotEvents(ctx context.Context, hostname, cursor string, pageSize int) (*BotEventsIterator, error) {
	args := m.Called(ctx, hostname, cursor, pageSize)
	return args.Get(0).(*BotEventsIterator), args.Error(1)
}
