// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ethernethook

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCountSections(t *testing.T) {
	t.Parallel()
	e := extendedGSClient{}
	Convey("test count sections", t, func() {
		So(e.CountSections(""), ShouldEqual, 0)
		So(e.CountSections("gs://"), ShouldEqual, 0)
		So(e.CountSections("gs://a/b/c"), ShouldEqual, 3)
		So(e.CountSections("gs://a/b/c/"), ShouldEqual, 3)
	})
}
