// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tlw

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetServodVersion(t *testing.T) {
	t.Parallel()
	Convey("test get servod version", t, func() {
		s := ServoHost{ServodType: "servo_v4_with_ccd_cr50"}
		So(s.GetServodVersion(), ShouldEqual, "v4")
	})
}
