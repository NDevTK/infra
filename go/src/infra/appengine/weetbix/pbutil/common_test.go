// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package pbutil

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/cv/api/bigquery/v1"

	pb "infra/appengine/weetbix/proto/v1"
)

func TestPresubmitRunMode(t *testing.T) {
	Convey("PresubmitRunModeFromString", t, func() {
		// Confirm a mapping exists for every mode defined by LUCI CV.
		for _, mode := range bigquery.Mode_name {
			if mode == "MODE_UNSPECIFIED" {
				continue
			}
			mode, err := PresubmitRunModeFromString(mode)
			So(err, ShouldBeNil)
			So(mode, ShouldNotEqual, pb.PresubmitRunMode_PRESUBMIT_RUN_MODE_UNSPECIFIED)
		}
	})
}
