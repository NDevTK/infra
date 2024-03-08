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
