// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package buildbucket

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// TestValidateTaskName tests that task names are validated correctly.
func TestValidateTaskName(t *testing.T) {
	t.Parallel()
	Convey("validate", t, func() {
		So(ValidateTaskName(""), ShouldNotBeNil)
		So(ValidateTaskName("audit_rpm"), ShouldBeNil)
		So(ValidateTaskName("deep_recovery"), ShouldBeNil)
		So(ValidateTaskName("audit____"), ShouldNotBeNil)
	})
}
