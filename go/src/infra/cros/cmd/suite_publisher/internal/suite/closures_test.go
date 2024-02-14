// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package suite

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"infra/cros/cmd/suite_publisher/test"
)

func TestClosures(t *testing.T) {
	for _, tc := range []struct {
		suites map[string]CentralizedSuite
		id     string
		want   []*SuiteClosure
	}{
		{
			suites: map[string]CentralizedSuite{
				"example_suite": NewSuite(test.ExampleSuite()),
			},
			id: "example_suite",
			want: []*SuiteClosure{
				{
					ID:    "example_suite",
					Child: "example_suite",
					Depth: 0,
				},
			},
		},
		{
			suites: map[string]CentralizedSuite{
				"example_suite":       NewSuite(test.ExampleSuite()),
				"example_suite_set":   NewSuiteSet(test.ExampleSuiteSet()),
				"example_suite_set_b": NewSuiteSet(test.ExampleSuiteSetB()),
			},
			id: "example_suite_set",
			want: []*SuiteClosure{
				{
					ID:    "example_suite_set",
					Child: "example_suite_set",
					Depth: 0,
				},
				{
					ID:    "example_suite_set",
					Child: "example_suite",
					Depth: 1,
				},
				{
					ID:    "example_suite_set",
					Child: "example_suite_set_b",
					Depth: 1,
				},
				{
					ID:    "example_suite_set",
					Child: "example_suite",
					Depth: 2,
				},
			},
		},
	} {
		s := tc.suites[tc.id]
		got, err := s.Closures(tc.suites)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(got, tc.want); diff != "" {
			t.Errorf("closures mismatch (-got +want):\n%s\n\n", diff)
			for _, closure := range got {
				t.Logf("\t%+v\n", closure)
			}
		}
	}
}
