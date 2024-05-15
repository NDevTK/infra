// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package run

import (
	"io"
	"log"
	"testing"

	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
	kronpb "go.chromium.org/chromiumos/infra/proto/go/test_platform/kron"
	suschpb "go.chromium.org/chromiumos/infra/proto/go/testplans"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"

	"infra/cros/cmd/kron/common"
	"infra/cros/cmd/kron/metrics"
)

// SetUp sets the RunID and discards the stdout and stderr for cleaner test
// runs.
func SetUp() {
	_ = metrics.SetSuiteSchedulerRunID("")
	common.Stdout = log.New(io.Discard, "", log.Lshortfile|log.LstdFlags)
	common.Stderr = log.New(io.Discard, "", log.Lshortfile|log.LstdFlags)
}

func TestLimitStagingRequestsUnderMax(t *testing.T) {
	t.Parallel()

	requests := []*ctpEvent{
		{
			event:      &kronpb.Event{},
			ctpRequest: &test_platform.Request{},
			config:     &suschpb.SchedulerConfig{},
		},
		{
			event:      &kronpb.Event{},
			ctpRequest: &test_platform.Request{},
			config:     &suschpb.SchedulerConfig{},
		},
		{
			event:      &kronpb.Event{},
			ctpRequest: &test_platform.Request{},
			config:     &suschpb.SchedulerConfig{},
		},
	}

	limitedRequests := limitStagingRequests(requests)

	if len(limitedRequests) != len(requests) {
		t.Errorf("%d requests expected, got %d", len(limitedRequests), len(requests))
	}
}
func TestLimitStagingRequestsOverMax(t *testing.T) {
	t.Parallel()

	requests := []*ctpEvent{
		{
			event:      &kronpb.Event{},
			ctpRequest: &test_platform.Request{},
			config:     &suschpb.SchedulerConfig{},
		},
		{
			event:      &kronpb.Event{},
			ctpRequest: &test_platform.Request{},
			config:     &suschpb.SchedulerConfig{},
		},
		{
			event:      &kronpb.Event{},
			ctpRequest: &test_platform.Request{},
			config:     &suschpb.SchedulerConfig{},
		},
		{
			event:      &kronpb.Event{},
			ctpRequest: &test_platform.Request{},
			config:     &suschpb.SchedulerConfig{},
		},
		{
			event:      &kronpb.Event{},
			ctpRequest: &test_platform.Request{},
			config:     &suschpb.SchedulerConfig{},
		},
		{
			event:      &kronpb.Event{},
			ctpRequest: &test_platform.Request{},
			config:     &suschpb.SchedulerConfig{},
		},
	}

	limitedRequests := limitStagingRequests(requests)

	if len(limitedRequests) != common.StagingMaxRequests {
		t.Errorf("%d requests expected, got %d", len(limitedRequests), common.StagingMaxRequests)
	}
}

func TestBuildPerModelConfigsMultipleModels(t *testing.T) {
	SetUp()
	models := []string{
		"model1",
		"model2",
		"model3",
	}

	testConfig := &suschpb.SchedulerConfig{
		Name:  "testConfig",
		Suite: "testSutie",
	}

	testBuild := &kronpb.Build{
		BuildUuid:   "123",
		RunUuid:     "abc",
		BuildTarget: "bt1",
		Milestone:   120,
		Version:     "9876",
		Board:       "board1",
	}

	testBranch := "CANARY"

	events, err := buildPerModelConfigs(models, testConfig, testBuild, testBranch)
	if err != nil {
		t.Error(err)
		return
	}

	for _, event := range events {
		if event.config.GetName() != testConfig.Name {
			t.Errorf("config name %s  expected, got %s", event.config.GetName(), testConfig.Name)
			return
		}

		if event.event.Board != testBuild.Board {
			t.Errorf("board %s  expected, got %s", event.event.GetBoard(), testBuild.GetBoard())
			return
		}
	}

	if len(events) != len(models) {
		t.Errorf("expected %d events got %d", len(models), len(events))
	}

	for _, model := range models {
		found := false
		for _, event := range events {
			if event.event.Model == model {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("model %s never seen in resulting events", model)
			return
		}
	}
}

func TestBuildPerModelConfigsNoModels(t *testing.T) {
	SetUp()

	models := []string{}

	testConfig := &suschpb.SchedulerConfig{
		Name:  "testConfig",
		Suite: "testSutie",
	}

	testBuild := &kronpb.Build{
		BuildUuid:   "123",
		RunUuid:     "abc",
		BuildTarget: "bt1",
		Milestone:   120,
		Version:     "9876",
		Board:       "board1",
	}

	testBranch := "CANARY"

	events, err := buildPerModelConfigs(models, testConfig, testBuild, testBranch)
	if err != nil {
		t.Error(err)
		return
	}

	for _, event := range events {
		if event.config.GetName() != testConfig.Name {
			t.Errorf("config name %s  expected, got %s", event.config.GetName(), testConfig.Name)
			return
		}

		if event.event.GetBoard() != testBuild.Board {
			t.Errorf("board %s expected, got %s", event.event.GetBoard(), testBuild.GetBoard())
			return
		}

		if event.event.GetModel() != "" {
			t.Errorf("empty model expected, got %s", event.event.GetBoard())
			return
		}
	}

	if len(events) != 1 {
		t.Errorf("expected %d events got %d", 1, len(events))
	}
}

func TestFillEventResponseScheduled(t *testing.T) {
	SetUp()
	events := []*kronpb.Event{
		{},
		{},
		{},
	}

	fakeBBResponse := &buildbucketpb.Build{
		Id:     123,
		Status: buildbucketpb.Status_SCHEDULED,
	}

	fillEventResponse(events, fakeBBResponse)

	for _, event := range events {
		if event.GetDecision().GetType() != kronpb.DecisionType_SCHEDULED && event.GetBbid() != fakeBBResponse.GetId() {
			t.Errorf("expected type %d bbid %d, got type %s bbid %d", kronpb.DecisionType_SCHEDULED, fakeBBResponse.GetId(), event.GetDecision().GetType(), event.GetBbid())
		}
	}
}

func TestFillEventResponseFailed(t *testing.T) {
	SetUp()
	events := []*kronpb.Event{
		{},
		{},
		{},
	}

	fakeBBResponse := &buildbucketpb.Build{
		Id:     123,
		Status: buildbucketpb.Status_FAILURE,
	}

	fillEventResponse(events, fakeBBResponse)

	for _, event := range events {
		if event.GetDecision().GetType() != kronpb.DecisionType_UNKNOWN && event.GetBbid() != 0 {
			t.Errorf("expected type %d bbid %d, got type %s bbid %d", kronpb.DecisionType_UNKNOWN, 0, event.GetDecision().GetType(), event.GetBbid())
		}
	}

	fakeBBResponse = &buildbucketpb.Build{
		Id:     123,
		Status: buildbucketpb.Status_INFRA_FAILURE,
	}

	fillEventResponse(events, fakeBBResponse)

	for _, event := range events {
		if event.GetDecision().GetType() != kronpb.DecisionType_UNKNOWN && event.GetBbid() != 0 {
			t.Errorf("expected type %d bbid %d, got type %s bbid %d", kronpb.DecisionType_UNKNOWN, 0, event.GetDecision().GetType(), event.GetBbid())
		}
	}

	fakeBBResponse = &buildbucketpb.Build{
		Id:     123,
		Status: buildbucketpb.Status_CANCELED,
	}

	fillEventResponse(events, fakeBBResponse)

	for _, event := range events {
		if event.GetDecision().GetType() != kronpb.DecisionType_UNKNOWN && event.GetBbid() != 0 {
			t.Errorf("expected type %d bbid %d, got type %s bbid %d", kronpb.DecisionType_UNKNOWN, 0, event.GetDecision().GetType(), event.GetBbid())
		}
	}

	fakeBBResponse = &buildbucketpb.Build{
		Id:     123,
		Status: buildbucketpb.Status_STATUS_UNSPECIFIED,
	}

	fillEventResponse(events, fakeBBResponse)

	for _, event := range events {
		if event.GetDecision().GetType() != kronpb.DecisionType_UNKNOWN && event.GetBbid() != 0 {
			t.Errorf("expected type %d bbid %d, got type %s bbid %d", kronpb.DecisionType_UNKNOWN, 0, event.GetDecision().GetType(), event.GetBbid())
		}
	}
}

func TestMapEventsByConfig(t *testing.T) {
	t.Parallel()

	config1 := &suschpb.SchedulerConfig{}
	config2 := &suschpb.SchedulerConfig{}

	fakeCtpRequests := []*ctpEvent{
		{
			event:      &kronpb.Event{},
			ctpRequest: &test_platform.Request{},
			config:     config1,
		},
		{
			event:      &kronpb.Event{},
			ctpRequest: &test_platform.Request{},
			config:     config2,
		},
		{
			event:      &kronpb.Event{},
			ctpRequest: &test_platform.Request{},
			config:     config2,
		},
	}

	configToEventsMap := mapEventsByConfig(fakeCtpRequests)

	for key, events := range configToEventsMap {
		if key == config1 && len(events) != 1 {
			t.Errorf("expected %d events for config1 got %d", 1, len(events))
			return
		}
		if key == config2 && len(events) != 2 {
			t.Errorf("expected %d events for config2 got %d", 2, len(events))
			return
		}
	}
}
