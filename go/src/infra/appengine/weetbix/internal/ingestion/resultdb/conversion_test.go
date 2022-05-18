// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package resultdb

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
)

func TestInvocationFromTestResultName(t *testing.T) {
	Convey("Valid input", t, func() {
		result, err := InvocationFromTestResultName("invocations/build-1234/tests/a/results/b")
		So(err, ShouldBeNil)
		So(result, ShouldEqual, "build-1234")
	})
	Convey("Invalid input", t, func() {
		_, err := InvocationFromTestResultName("")
		So(err, ShouldErrLike, "invalid test result name")

		_, err = InvocationFromTestResultName("projects/chromium/resource/b")
		So(err, ShouldErrLike, "invalid test result name")

		_, err = InvocationFromTestResultName("invocations/build-1234")
		So(err, ShouldErrLike, "invalid test result name")

		_, err = InvocationFromTestResultName("invocations//")
		So(err, ShouldErrLike, "invalid test result name")

		_, err = InvocationFromTestResultName("invocations/")
		So(err, ShouldErrLike, "invalid test result name")

		_, err = InvocationFromTestResultName("invocations")
		So(err, ShouldErrLike, "invalid test result name")
	})
}
