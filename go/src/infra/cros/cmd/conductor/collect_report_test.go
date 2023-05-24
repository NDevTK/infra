// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"reflect"
	"testing"

	bbpb "go.chromium.org/luci/buildbucket/proto"
)

func TestCollectReport(t *testing.T) {
	report := &CollectReport{}

	build := &bbpb.Build{
		Id:     12345,
		Status: bbpb.Status_FAILURE,
		Builder: &bbpb.BuilderID{
			Project: "chromeos",
			Bucket:  "release",
			Builder: "eve-release-main",
		},
	}
	report.recordBuild(build, "12345", false)
	build = &bbpb.Build{
		Id:     12346,
		Status: bbpb.Status_SUCCESS,
		Builder: &bbpb.BuilderID{
			Project: "chromeos",
			Bucket:  "release",
			Builder: "eve-release-main",
		},
	}
	report.recordBuild(build, "12345", true)
	build = &bbpb.Build{
		Id:     12347,
		Status: bbpb.Status_SUCCESS,
		Builder: &bbpb.BuilderID{
			Project: "chromeos",
			Bucket:  "release",
			Builder: "atlas-release-main",
		},
	}
	report.recordBuild(build, "12347", false)

	expectedReport := &CollectReport{
		BuilderInfo: map[string][]*BuilderRun{
			"eve-release-main": {
				{
					Builds: []*BuildInfo{
						{
							BBID:   "12345",
							Status: "FAILURE",
							Retry:  false,
						},
						{
							BBID:   "12346",
							Status: "SUCCESS",
							Retry:  true,
						},
					},
					RetryCount: 1,
					LastStatus: "SUCCESS",
				},
			},
			"atlas-release-main": {
				{
					Builds: []*BuildInfo{
						{
							BBID:   "12347",
							Status: "SUCCESS",
							Retry:  false,
						},
					},
					LastStatus: "SUCCESS",
				},
			},
		},
		RetryCount: 1,
	}
	if !reflect.DeepEqual(*report, *expectedReport) {
		t.Fatalf("mismatch on CollectReport: expected\n%+v\ngot\n%+v", *expectedReport, *report)
	}
}
