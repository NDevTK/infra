// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	Convey("Check isChromeOrChromiumProject", t, func() {
		So(isChromeOrChromiumProject("chrome"), ShouldEqual, true)
		So(isChromeOrChromiumProject("chromium"), ShouldEqual, true)
		So(isChromeOrChromiumProject("chromeos"), ShouldEqual, false)
		So(isChromeOrChromiumProject("chrome-100"), ShouldEqual, true)
		So(isChromeOrChromiumProject("chromium-100"), ShouldEqual, true)
		So(isChromeOrChromiumProject("turquoise"), ShouldEqual, false)
	})
}
