// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package run

import (
	"testing"

	cloudPubsub "cloud.google.com/go/pubsub"

	"go.chromium.org/chromiumos/infra/proto/go/test_platform/kron"
	"go.chromium.org/chromiumos/infra/proto/go/testplans"

	"infra/cros/cmd/kron/builds"
	"infra/cros/cmd/kron/common"
)

// TestCombineCTPRequests verifies that combineCTPRequests() is properly
// generating the map and not mutating the objects.
func TestCombineCTPRequests(t *testing.T) {
	expectedEventsTestConfig := []*builds.EventWrapper{
		{
			Event: &kron.Event{
				EventUuid: "1",
			},
		},
		{
			Event: &kron.Event{
				EventUuid: "2",
			},
		},
		{
			Event: &kron.Event{
				EventUuid: "3",
			},
		},
		{
			Event: &kron.Event{
				EventUuid: "4",
			},
		},
	}
	expectedEventsTestConfig2 := []*builds.EventWrapper{
		{
			Event: &kron.Event{
				EventUuid: "5",
			},
		},
		{
			Event: &kron.Event{
				EventUuid: "6",
			},
		},
	}

	mockBuilds := []*builds.BuildPackage{
		{
			Build:   &kron.Build{},
			Message: &cloudPubsub.Message{},
			Requests: []*builds.ConfigDetails{
				{
					Config: &testplans.SchedulerConfig{
						Name: "TestConfig",
					},
					Events: []*builds.EventWrapper{
						expectedEventsTestConfig[0],
						expectedEventsTestConfig[1],
					},
				},
			},
		},
		{
			Build:   &kron.Build{},
			Message: &cloudPubsub.Message{},
			Requests: []*builds.ConfigDetails{
				{
					Config: &testplans.SchedulerConfig{
						Name: "TestConfig",
					},
					Events: []*builds.EventWrapper{
						expectedEventsTestConfig[2],
						expectedEventsTestConfig[3],
					},
				},
			},
		},
		{
			Build:   &kron.Build{},
			Message: &cloudPubsub.Message{},
			Requests: []*builds.ConfigDetails{
				{
					Config: &testplans.SchedulerConfig{
						Name: "TestConfig2",
					},
					Events: []*builds.EventWrapper{
						expectedEventsTestConfig2[0],
						expectedEventsTestConfig2[1],
					},
				},
			},
		},
	}

	expectedMap := map[string][]*builds.EventWrapper{
		"TestConfig": {
			expectedEventsTestConfig[0],
			expectedEventsTestConfig[1],
			expectedEventsTestConfig[2],
			expectedEventsTestConfig[3],
		},
		"TestConfig2": {
			expectedEventsTestConfig2[0],
			expectedEventsTestConfig2[1],
		},
	}

	requestMap := combineCTPRequests(mockBuilds)

	for key, eventList := range requestMap {
		// Ensure the key provided is in the expected return map
		expectedEventList, ok := expectedMap[key]
		if !ok {
			t.Errorf("key %s was provided but not in expectedMap", key)
			return
		}

		// Ensure that all events are in the expected map.
		for _, event := range eventList {
			inExpected := false
			for _, expectedEvent := range expectedEventList {
				inExpected = event == expectedEvent
				if inExpected {
					break
				}
			}

			if !inExpected {
				t.Errorf("event %s was not seen in expectedMap[%s]", event.Event.EventUuid, key)
				return
			}
		}
	}

	for key, expectedEventList := range expectedMap {
		// Ensure the key provided is in the expected return map
		eventList, ok := requestMap[key]
		if !ok {
			t.Errorf("key %s was expected but not provided", key)
			return
		}

		// Ensure that all events are in the expected map.
		for _, event := range expectedEventList {
			inExpected := false
			for _, expectedEvent := range eventList {
				inExpected = event == expectedEvent
				if inExpected {
					break
				}
			}

			if !inExpected {
				t.Errorf("event %s was not seen in in the return", event.Event.EventUuid)
				return
			}
		}
	}

}

func TestLimitStagingRequestsSingleConfigOverMax(t *testing.T) {
	mockRequestMap := map[string][]*builds.EventWrapper{
		"suite1": {
			{},
			{},
			{},
			{},
			{},
			{},
		},
	}

	limitedRequestMap := limitStagingRequests(mockRequestMap)

	requestCount := 0

	for _, requestList := range limitedRequestMap {
		for range requestList {
			requestCount += 1
		}
	}

	if requestCount != common.StagingMaxRequests {
		t.Errorf("%d requests expected, got %d", common.StagingMaxRequests, requestCount)
	}
}
func TestLimitStagingRequestsSingleConfigUnderMax(t *testing.T) {
	mockRequestMap := map[string][]*builds.EventWrapper{
		"suite1": {
			{},
			{},
			{},
			{},
		},
	}

	limitedRequestMap := limitStagingRequests(mockRequestMap)

	requestCount := 0

	for _, requestList := range limitedRequestMap {
		for range requestList {
			requestCount += 1
		}
	}

	if requestCount != 4 {
		t.Errorf("%d requests expected, got %d", 4, requestCount)
	}
}

func TestLimitStagingRequestsMultipleConfigsOverMax(t *testing.T) {
	mockRequestMap := map[string][]*builds.EventWrapper{
		"suite1": {
			{},
			{},
			{},
			{},
		},
		"suite2": {
			{},
			{},
			{},
			{},
		},
	}

	limitedRequestMap := limitStagingRequests(mockRequestMap)

	requestCount := 0

	suite1Seen := false
	suite2Seen := false
	for configName, requestList := range limitedRequestMap {
		if configName == "suite1" {
			suite1Seen = true
		}
		if configName == "suite2" {
			suite2Seen = true
		}

		for range requestList {
			requestCount += 1
		}
	}

	if requestCount != common.StagingMaxRequests {
		t.Errorf("%d requests expected, got %d", common.StagingMaxRequests, requestCount)
	}

	if !(suite1Seen && suite2Seen) {
		t.Errorf("Suite 1 and 2 were expected to be seen. suite1Seen:%t\tsuite2Seen:%t", suite1Seen, suite2Seen)
	}
}

func TestLimitStagingRequestsMultipleConfigsUnderMax(t *testing.T) {
	mockRequestMap := map[string][]*builds.EventWrapper{
		"suite1": {
			{},
			{},
			{},
		},
		"suite2": {
			{},
		},
	}

	limitedRequestMap := limitStagingRequests(mockRequestMap)

	requestCount := 0

	suite1Seen := false
	suite2Seen := false
	for configName, requestList := range limitedRequestMap {
		if configName == "suite1" {
			suite1Seen = true
		}
		if configName == "suite2" {
			suite2Seen = true
		}

		for range requestList {
			requestCount += 1
		}
	}

	if requestCount != 4 {
		t.Errorf("%d requests expected, got %d", 4, requestCount)
	}

	if !(suite1Seen && suite2Seen) {
		t.Errorf("Suite 1 and 2 were expected to be seen. suite1Seen:%t\tsuite2Seen:%t", suite1Seen, suite2Seen)
	}
}
