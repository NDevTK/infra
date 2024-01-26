// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package test holds some helper functions for testing.
package test

import "go.chromium.org/chromiumos/config/go/test/api"

func ExampleSuite() *api.Suite {
	return &api.Suite{
		Id: &api.Suite_Id{Value: "example_suite"},
		Metadata: &api.Metadata{
			Owners: []*api.Contact{
				{Email: "example@chromium.org"},
				{Email: "example2@chromium.org"},
			},
			Criteria:     &api.Criteria{Value: "This is an example suite"},
			BugComponent: &api.BugComponent{Value: "b:123456"},
		},
		Tests: []*api.TestCase_Id{
			{Value: "example_test_0"},
			{Value: "example_test_1"},
			{Value: "example_test_2"},
		},
	}
}

func ExampleSuiteSet() *api.SuiteSet {
	return &api.SuiteSet{
		Id: &api.SuiteSet_Id{Value: "example_suite_set"},
		Metadata: &api.Metadata{
			Owners: []*api.Contact{
				{Email: "example@chromium.org"},
				{Email: "example2@chromium.org"},
			},
			Criteria:     &api.Criteria{Value: "This is an example suite set"},
			BugComponent: &api.BugComponent{Value: "b:123456"},
		},
		Suites: []*api.Suite_Id{
			{Value: "example_suite"},
		},
		SuiteSets: []*api.SuiteSet_Id{
			{Value: "example_suite_set_b"},
		},
	}
}
