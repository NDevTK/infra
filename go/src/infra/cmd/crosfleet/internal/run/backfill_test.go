// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package run

import (
	"context"
	"fmt"
	"testing"

	"infra/cmd/crosfleet/internal/buildbucket"
	"infra/cmd/crosfleet/internal/common"

	"github.com/google/go-cmp/cmp"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestRemoveBackfills(t *testing.T) {
	t.Parallel()
	backfillBuild := &buildbucketpb.Build{
		Tags: []*buildbucketpb.StringPair{
			{Key: "crosfleet-tool", Value: "backfill"}}}
	nonBackfillBuild := &buildbucketpb.Build{
		Tags: []*buildbucketpb.StringPair{
			{Key: "crosfleet-tool", Value: "suite"}}}

	wantFilteredBuilds := []*buildbucketpb.Build{nonBackfillBuild}
	gotFilteredBuilds := removeBackfills([]*buildbucketpb.Build{
		backfillBuild, nonBackfillBuild})
	if diff := cmp.Diff(wantFilteredBuilds, gotFilteredBuilds, common.CmpOpts); diff != "" {
		t.Errorf("unexpected diff (%s)", diff)
	}
}

var testBackfillTagsData = []struct {
	build    *buildbucketpb.Build
	wantTags map[string]string
}{
	{
		&buildbucketpb.Build{
			Id: 1,
			Tags: []*buildbucketpb.StringPair{
				{Key: "foo", Value: "bar"},
				{Key: "baz", Value: "lol"}}},
		map[string]string{
			"foo":            "bar",
			"baz":            "lol",
			"crosfleet-tool": "backfill",
			"backfill":       "1",
		},
	},
	{
		&buildbucketpb.Build{
			Id: 2,
			Tags: []*buildbucketpb.StringPair{
				{Key: "bar", Value: "foo"},
				{Key: "lol", Value: "baz"},
				{Key: "backfill", Value: "3"},
				{Key: "crosfleet-tool", Value: "suite"}}},
		map[string]string{
			"bar":            "foo",
			"lol":            "baz",
			"crosfleet-tool": "backfill",
			"backfill":       "2",
		},
	},
}

func TestBackfillTags(t *testing.T) {
	t.Parallel()
	for _, tt := range testBackfillTagsData {
		tt := tt
		t.Run(fmt.Sprintf("(%s)", tt.wantTags), func(t *testing.T) {
			t.Parallel()
			gotTags := backfillTags(tt.build)
			if diff := cmp.Diff(tt.wantTags, gotTags); diff != "" {
				t.Errorf("unexpected diff (%s)", diff)
			}
		})
	}
}

// End to end tests.
func TestBackfill_ByTags(t *testing.T) {
	r := backfillRun{
		buildTags: map[string]string{
			"suite":       "bvt-installer",
			"build":       "asurada-release/R112-15357.0.0",
			"label-board": "asurada",
		},
		allowDupes:       false,
		skipConfirmation: true,
	}
	ctx := context.Background()

	inputProps, err := structpb.NewStruct(map[string]interface{}{
		"request": "foo",
	})
	if err != nil {
		t.Error(err)
	}

	bb := &buildbucket.FakeClient{
		ExpectedGetAllBuildsWithTags: []*buildbucket.ExpectedGetWithTagsCall{
			{
				Tags: map[string]string{
					"build":       "asurada-release/R112-15357.0.0",
					"suite":       "bvt-installer",
					"label-board": "asurada",
				},
				Response: []*buildbucketpb.Build{
					{
						Builder: &buildbucketpb.BuilderID{
							Builder: "cros_test_platform",
						},
						Status: buildbucketpb.Status_FAILURE,
						Tags: []*buildbucketpb.StringPair{
							{
								Key:   "build",
								Value: "asurada-release/R112-15357.0.0",
							},
							{
								Key:   "suite",
								Value: "bvt-installer",
							},
							{
								Key:   "label-board",
								Value: "asurada",
							},
						},
						Input: &buildbucketpb.Build_Input{
							Properties: inputProps,
						},
					}, {
						Builder: &buildbucketpb.BuilderID{
							Builder: "cros_test_platform",
						},
						Status: buildbucketpb.Status_SUCCESS,
						Tags: []*buildbucketpb.StringPair{
							{
								Key:   "build",
								Value: "asurada-release/R112-15357.0.0",
							},
							{
								Key:   "suite",
								Value: "bvt-installer",
							},
							{
								Key:   "label-board",
								Value: "asurada",
							},
						},
						Input: &buildbucketpb.Build_Input{
							Properties: inputProps,
						},
					}, {
						Builder: &buildbucketpb.BuilderID{
							Builder: "cros_test_platform",
						},
						Status: buildbucketpb.Status_FAILURE,
						Tags: []*buildbucketpb.StringPair{
							{
								Key:   "backfill",
								Value: "0",
							},
							{
								Key:   "crosfleet-tool",
								Value: "backfill",
							},
							{
								Key:   "build",
								Value: "asurada-release/R112-15357.0.0",
							},
							{
								Key:   "suite",
								Value: "bvt-installer",
							},
							{
								Key:   "label-board",
								Value: "asurada",
							},
						},
						Input: &buildbucketpb.Build_Input{
							Properties: inputProps,
						},
					},
				},
			},
		},
		ExpectedAnyIncompleteBuildsWithTags: []*buildbucket.ExpectedGetWithTagsCall{
			{
				Tags: map[string]string{
					"backfill":       "0",
					"crosfleet-tool": "backfill",
					"build":          "asurada-release/R112-15357.0.0",
					"suite":          "bvt-installer",
					"label-board":    "asurada",
				},
				Response: []*buildbucketpb.Build{
					{
						Builder: &buildbucketpb.BuilderID{
							Builder: "cros_test_platform",
						},
						Status: buildbucketpb.Status_SCHEDULED,
						Tags: []*buildbucketpb.StringPair{
							{
								Key:   "build",
								Value: "asurada-release/R112-15357.0.0",
							},
							{
								Key:   "suite",
								Value: "bvt-installer",
							},
							{
								Key:   "label-board",
								Value: "asurada",
							},
						},
						Input: &buildbucketpb.Build_Input{
							Properties: inputProps,
						},
					},
				},
			},
		},
		ExpectedScheduleBuild: []*buildbucket.ExpectedScheduleCall{
			{
				Tags: map[string]string{
					"crosfleet-tool": "backfill",
					"backfill":       "0",
					"build":          "asurada-release/R112-15357.0.0",
					"suite":          "bvt-installer",
					"label-board":    "asurada",
				},
				Response: &buildbucketpb.Build{
					Id: 1,
				},
			},
		},
	}
	if err := r.innerRun(nil, nil, ctx, bb); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBackfill_ByTags_AllowDupes(t *testing.T) {
	r := backfillRun{
		buildTags: map[string]string{
			"suite":       "bvt-installer",
			"build":       "asurada-release/R112-15357.0.0",
			"label-board": "asurada",
		},
		allowDupes: true,
	}
	ctx := context.Background()

	inputProps, err := structpb.NewStruct(map[string]interface{}{
		"request": "foo",
	})
	if err != nil {
		t.Error(err)
	}

	bb := &buildbucket.FakeClient{
		ExpectedGetAllBuildsWithTags: []*buildbucket.ExpectedGetWithTagsCall{
			{
				Tags: map[string]string{
					"build":       "asurada-release/R112-15357.0.0",
					"suite":       "bvt-installer",
					"label-board": "asurada",
				},
				Response: []*buildbucketpb.Build{
					{
						Builder: &buildbucketpb.BuilderID{
							Builder: "cros_test_platform",
						},
						Status: buildbucketpb.Status_SCHEDULED,
						Tags: []*buildbucketpb.StringPair{
							{
								Key:   "build",
								Value: "asurada-release/R112-15357.0.0",
							},
							{
								Key:   "suite",
								Value: "bvt-installer",
							},
							{
								Key:   "label-board",
								Value: "asurada",
							},
						},
						Input: &buildbucketpb.Build_Input{
							Properties: inputProps,
						},
					},
				},
			},
		},
		ExpectedScheduleBuild: []*buildbucket.ExpectedScheduleCall{
			{
				Tags: map[string]string{
					"crosfleet-tool": "backfill",
					"backfill":       "0",
					"build":          "asurada-release/R112-15357.0.0",
					"suite":          "bvt-installer",
					"label-board":    "asurada",
				},
				Response: &buildbucketpb.Build{
					Id: 1,
				},
			},
		},
	}
	if err := r.innerRun(nil, nil, ctx, bb); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBackfill_ByTags_DryRun(t *testing.T) {
	r := backfillRun{
		buildTags: map[string]string{
			"suite":       "bvt-installer",
			"build":       "asurada-release/R112-15357.0.0",
			"label-board": "asurada",
		},
		// Simplify the test, allow dupes.
		allowDupes: true,
		dryrun:     true,
	}
	ctx := context.Background()

	inputProps, err := structpb.NewStruct(map[string]interface{}{
		"request": "foo",
	})
	if err != nil {
		t.Error(err)
	}

	bb := &buildbucket.FakeClient{
		ExpectedGetAllBuildsWithTags: []*buildbucket.ExpectedGetWithTagsCall{
			{
				Tags: map[string]string{
					"build":       "asurada-release/R112-15357.0.0",
					"suite":       "bvt-installer",
					"label-board": "asurada",
				},
				Response: []*buildbucketpb.Build{
					{
						Builder: &buildbucketpb.BuilderID{
							Builder: "cros_test_platform",
						},
						Status: buildbucketpb.Status_SCHEDULED,
						Tags: []*buildbucketpb.StringPair{
							{
								Key:   "build",
								Value: "asurada-release/R112-15357.0.0",
							},
							{
								Key:   "suite",
								Value: "bvt-installer",
							},
							{
								Key:   "label-board",
								Value: "asurada",
							},
						},
						Input: &buildbucketpb.Build_Input{
							Properties: inputProps,
						},
					},
				},
			},
		},
	}
	if err := r.innerRun(nil, nil, ctx, bb); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
