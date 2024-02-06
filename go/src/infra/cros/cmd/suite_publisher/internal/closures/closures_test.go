// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package closures

import (
	"infra/cros/cmd/suite_publisher/test"
	"testing"

	"github.com/google/go-cmp/cmp"

	csuite "go.chromium.org/chromiumos/platform/dev-util/src/chromiumos/test/suite/centralizedsuite"
)

func TestClosures(t *testing.T) {
	for _, tc := range []struct {
		mappings csuite.Mappings
		id       string
		want     []*SuiteClosure
	}{
		{
			mappings: csuite.Mappings{
				"example_suite": csuite.NewSuite(test.ExampleSuite()),
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
			mappings: csuite.Mappings{
				"example_suite":       csuite.NewSuite(test.ExampleSuite()),
				"example_suite_set":   csuite.NewSuiteSet(test.ExampleSuiteSet()),
				"example_suite_set_b": csuite.NewSuiteSet(test.ExampleSuiteSetB()),
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
		s := tc.mappings[tc.id]
		got := Closures(s, tc.mappings)
		if diff := cmp.Diff(got, tc.want); diff != "" {
			t.Errorf("closures mismatch (-got +want):\n%s\n\n", diff)
			for _, closure := range got {
				t.Logf("\t%+v\n", closure)
			}
		}
	}
}
