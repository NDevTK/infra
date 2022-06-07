// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package pbutil

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	rdbpb "go.chromium.org/luci/resultdb/proto/v1"

	pb "infra/appengine/weetbix/proto/v1"
)

func TestResultDB(t *testing.T) {
	Convey("TestResultStatusFromResultDB", t, func() {
		// Confirm Weetbix handles every test status defined by ResultDB.
		// This test is designed to break if ResultDB extends the set of
		// allowed values, without a corresponding update to Weetbix.
		for _, v := range rdbpb.TestStatus_value {
			rdbStatus := rdbpb.TestStatus(v)
			if rdbStatus == rdbpb.TestStatus_STATUS_UNSPECIFIED {
				continue
			}

			status := TestResultStatusFromResultDB(rdbStatus)
			So(status, ShouldNotEqual, pb.TestResultStatus_TEST_RESULT_STATUS_UNSPECIFIED)
		}
	})
	Convey("ExonerationReasonFromResultDB", t, func() {
		// Confirm Weetbix handles every exoneration reason defined by ResultDB.
		// This test is designed to break if ResultDB extends the set of
		// allowed values, without a corresponding update to Weetbix.
		for _, v := range rdbpb.ExonerationReason_value {
			rdbReason := rdbpb.ExonerationReason(v)
			if rdbReason == rdbpb.ExonerationReason_EXONERATION_REASON_UNSPECIFIED {
				continue
			}

			reason := ExonerationReasonFromResultDB(rdbReason)
			So(reason, ShouldNotEqual, pb.ExonerationReason_EXONERATION_REASON_UNSPECIFIED)
		}
	})
}
