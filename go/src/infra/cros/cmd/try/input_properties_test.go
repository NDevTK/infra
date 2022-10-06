// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"context"
	"fmt"
	"testing"

	"infra/cros/internal/assert"
	"infra/cros/internal/cmd"

	"google.golang.org/protobuf/types/known/structpb"
)

const (
	invalidJSON = "{'is-this-valid-json?': False"
	// experiment_reasons have enums as integers
	validJSON = `{
		"buildbucket": {
			"bbagent_args": {
				"build": {
					"input": {
						"properties": {
							"$chromeos/my_module": {
								"my_prop": 100
							},
							"my_other_prop": 101
						}
					},
					"infra": {
						"buildbucket": {
							"experiment_reasons": {
								"chromeos.cros_artifacts.use_gcloud_storage": 1
							}
						}
					}
				}
			}
		}
	}`
	// experiment_reasons have enums as strings
	unmarshalErrorButInputPropsOK = `{
		"buildbucket": {
			"bbagent_args": {
				"build": {
					"input": {
						"properties": {
							"$chromeos/my_module": {
								"my_prop": 100
							},
							"my_other_prop": 101
						}
					},
					"infra": {
						"buildbucket": {
							"experiment_reasons": {
								"chromeos.cros_artifacts.use_gcloud_storage": "EXPERIMENT_REASON_BUILDER_CONFIG"
							}
						}
					}
				}
			}
		}
	}`
	// "input" is misspelled
	unmarshalErrorWithNoInputProperties = `{
		"buildbucket": {
			"bbagent_args": {
				"build": {
					"inputt": {
						"properties": {
							"$chromeos/my_module": {
								"my_prop": 100
							},
							"my_other_prop": 101
						}
					},
					"infra": {
						"buildbucket": {
							"experiment_reasons": {
								"chromeos.cros_artifacts.use_gcloud_storage": "EXPERIMENT_REASON_BUILDER_CONFIG"
							}
						}
					}
				}
			}
		}
	}`
)

// TestGetBuilderInputProps tests GetBuilderInputProps.
// The most interesting logic to test is where it permits certain json.UnmarshalTypeErrors.
func TestGetBuilderInputProps(t *testing.T) {
	t.Parallel()
	okInputProperties, err := structpb.NewStruct(map[string]interface{}{
		"$chromeos/my_module": map[string]interface{}{
			"my_prop": 100,
		},
		"my_other_prop": 101,
	})
	if err != nil {
		t.Fatal("Error constructing okInputProperties:", err)
	}
	m := tryRunBase{}
	for i, tc := range []struct {
		ledGetBuilderStdout string
		expectError         bool
		expectedInputProps  *structpb.Struct // Unchecked if expectError
	}{
		{validJSON, false, okInputProperties},
		{invalidJSON, true, nil},
		{unmarshalErrorButInputPropsOK, false, okInputProperties},
		{unmarshalErrorWithNoInputProperties, true, nil},
	} {
		m.cmdRunner = cmd.FakeCommandRunner{
			ExpectedCmd: []string{"led", "get-builder", "chromeos/release:release-main-orchestrator"},
			Stdout:      tc.ledGetBuilderStdout,
		}
		propsStruct, err := m.GetBuilderInputProps(context.Background(), "chromeos/release/release-main-orchestrator")
		if err != nil && !tc.expectError {
			t.Errorf("#%d: Unexpected error running GetBuilderInputProps: %+v", i, err)
		}
		if err == nil && tc.expectError {
			fmt.Println("yo", propsStruct.String())
			t.Errorf("#%d: Expected error running GetBuilderInputProps; got no error. props: %+v", i, propsStruct)
		}
		if !tc.expectError && propsStruct.String() != tc.expectedInputProps.String() {
			t.Errorf("#%d: Unexpected input props: got %+v; want %+v", i, propsStruct, tc.expectedInputProps)
		}
	}
}

func TestSetProperty(t *testing.T) {
	t.Parallel()
	s, err := structpb.NewStruct(map[string]interface{}{
		"$chromeos/my_module": map[string]interface{}{
			"my_prop": 100,
		},
		"my_other_prop": "101",
	})
	assert.NilError(t, err)

	setProperty(s, "$chromeos/my_module.my_prop", 200)
	setProperty(s, "$chromeos/my_module.new_prop.foo", []string{"a", "b", "c"})
	setProperty(s, "my_other_prop", "201")

	fields := s.GetFields()
	assert.IntsEqual(t, int(fields["$chromeos/my_module"].GetStructValue().GetFields()["my_prop"].GetNumberValue()), 200)

	myProp := fields["$chromeos/my_module"].GetStructValue().GetFields()["new_prop"].GetStructValue().GetFields()["foo"].GetListValue().AsSlice()
	myPropStr := make([]string, len(myProp))
	for i := range myProp {
		myPropStr[i] = myProp[i].(string)
	}
	assert.StringArrsEqual(t, myPropStr, []string{"a", "b", "c"})

	assert.StringsEqual(t, fields["my_other_prop"].GetStringValue(), "201")
}

func TestSetProperty_error(t *testing.T) {
	t.Parallel()
	s, err := structpb.NewStruct(map[string]interface{}{
		"$chromeos/my_module": map[string]interface{}{
			"my_prop": 100,
		},
		"my_other_prop": "101",
	})
	assert.NilError(t, err)

	invalidValue := struct {
		String string
		Number int
	}{
		"foo",
		123,
	}
	err = setProperty(s, "$chromeos/my_module.my_prop", invalidValue)
	assert.ErrorContains(t, err, "invalid type")

	err = setProperty(s, "totally_new_prop", invalidValue)
	assert.ErrorContains(t, err, "invalid type")

	err = setProperty(s, "my_other_prop.invalid_nest", 123)
	assert.ErrorContains(t, err, "not a struct")
}
