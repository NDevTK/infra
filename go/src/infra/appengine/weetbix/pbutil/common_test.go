// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package pbutil

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/cv/api/bigquery/v1"
	cvv0 "go.chromium.org/luci/cv/api/v0"

	pb "infra/appengine/weetbix/proto/v1"
)

func TestCommon(t *testing.T) {
	Convey("PresubmitRunModeFromString", t, func() {
		// Confirm a mapping exists for every mode defined by LUCI CV.
		// This test is designed to break if LUCI CV extends the set of
		// allowed values, without a corresponding update to Weetbix.
		for _, mode := range bigquery.Mode_name {
			if mode == "MODE_UNSPECIFIED" {
				continue
			}
			mode, err := PresubmitRunModeFromString(mode)
			So(err, ShouldBeNil)
			So(mode, ShouldNotEqual, pb.PresubmitRunMode_PRESUBMIT_RUN_MODE_UNSPECIFIED)
		}
	})
	Convey("PresubmitRunStatusFromLUCICV", t, func() {
		// Confirm a mapping exists for every run status defined by LUCI CV.
		// This test is designed to break if LUCI CV extends the set of
		// allowed values, without a corresponding update to Weetbix.
		for _, v := range cvv0.Run_Status_value {
			runStatus := cvv0.Run_Status(v)
			if runStatus&cvv0.Run_ENDED_MASK == 0 {
				// Not a run ended status. Weetbix should not have to deal
				// with these, as Weetbix only ingests completed runs.
				continue
			}
			if runStatus == cvv0.Run_ENDED_MASK {
				// The run ended mask is itself not a valid status.
				continue
			}
			status, err := PresubmitRunStatusFromLUCICV(runStatus)
			So(err, ShouldBeNil)
			So(status, ShouldNotEqual, pb.PresubmitRunStatus_PRESUBMIT_RUN_STATUS_UNSPECIFIED)
		}
	})
}
