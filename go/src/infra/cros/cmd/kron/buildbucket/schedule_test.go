// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package buildbucket

import (
	"testing"

	"go.chromium.org/chromiumos/infra/proto/go/chromiumos"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
	ctppb "go.chromium.org/chromiumos/infra/proto/go/test_platform/cros_test_platform"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/kron"

	"infra/cros/cmd/kron/builds"
)

// TestMergeRequests verifies that the mergeRequests() is properly formatting
// the keys and not mutating the objects.
func TestMergeRequests(t *testing.T) {
	mockEvent1 := &builds.EventWrapper{
		Event: &kron.Event{
			RunUuid:    "",
			EventUuid:  "",
			ConfigName: "test1",
			SuiteName:  "testSuite1",
		},
		CtpRequest: &test_platform.Request{
			Params: &test_platform.Request_Params{
				HardwareAttributes: &test_platform.Request_Params_HardwareAttributes{},
				SoftwareAttributes: &test_platform.Request_Params_SoftwareAttributes{
					BuildTarget: &chromiumos.BuildTarget{
						Name: "testBoard1",
					},
				},
			},
		},
	}
	mockEvent2 := &builds.EventWrapper{
		Event: &kron.Event{
			RunUuid:    "",
			EventUuid:  "",
			ConfigName: "test2",
			SuiteName:  "testSuite2",
		},
		CtpRequest: &test_platform.Request{
			Params: &test_platform.Request_Params{
				HardwareAttributes: &test_platform.Request_Params_HardwareAttributes{},
				SoftwareAttributes: &test_platform.Request_Params_SoftwareAttributes{
					BuildTarget: &chromiumos.BuildTarget{
						Name: "testBoard2",
					},
				},
			},
		},
	}
	mockEvent3 := &builds.EventWrapper{
		Event: &kron.Event{
			EventUuid:  "uuid1",
			ConfigName: "test1",
			SuiteName:  "testSuite1",
		},
		CtpRequest: &test_platform.Request{
			Params: &test_platform.Request_Params{
				HardwareAttributes: &test_platform.Request_Params_HardwareAttributes{},
				SoftwareAttributes: &test_platform.Request_Params_SoftwareAttributes{
					BuildTarget: &chromiumos.BuildTarget{
						Name: "testBoard3",
					},
				},
			},
		},
	}
	mockEvent4 := &builds.EventWrapper{
		Event: &kron.Event{
			EventUuid:  "uuid1",
			ConfigName: "test1",
			SuiteName:  "testSuite1",
		},
		CtpRequest: &test_platform.Request{
			Params: &test_platform.Request_Params{
				HardwareAttributes: &test_platform.Request_Params_HardwareAttributes{},
				SoftwareAttributes: &test_platform.Request_Params_SoftwareAttributes{
					BuildTarget: &chromiumos.BuildTarget{
						Name: "testBoard3",
					},
				},
			},
		},
	}

	mockEvents := []*builds.EventWrapper{
		mockEvent1,
		mockEvent2,
		mockEvent3,
		mockEvent4,
	}

	expectedProps := &ctppb.CrosTestPlatformProperties{
		Requests: map[string]*test_platform.Request{
			"testBoard1.test1.testSuite1": mockEvent1.CtpRequest,
			"testBoard3.test1.testSuite1": mockEvent3.CtpRequest,
			"testBoard2.test2.testSuite2": mockEvent2.CtpRequest,
			// Test that UUID is added to dedupe a key already used
			"testBoard3.test1.testSuite1.uuid1": mockEvent4.CtpRequest,
		},
	}

	receivedProps := mergeRequests(mockEvents)

	for key, request := range receivedProps.Requests {

		expectedRequest, ok := expectedProps.Requests[key]
		if !ok {
			t.Errorf("key %s was provided in the request map but not expected", key)
			return
		}

		if request != expectedRequest {
			t.Errorf("given request for key %s does not match expected request.", key)
			return
		}
	}
}
