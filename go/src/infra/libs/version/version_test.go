// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package version

import (
	"testing"
	"testing/quick"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGEQ(t *testing.T) {
	t.Parallel()
	Convey("Test Greater Than or Equal to", t, func() {
		Convey("A: 1.2.3.4 B: 1.1.2.3.5-rc-4", func() {
			So(GEQ("1.2.3.4", "1.1.2.3.5-rc-4"), ShouldBeTrue)
		})
		Convey("A: 1.1.2.3.5-rc-4 B: 1.2.3.4", func() {
			So(GEQ("1.1.2.3.5-rc-4", "1.2.3.4"), ShouldBeFalse)
		})
		Convey("A: 1.1-debug.1.1 B: 1.1.1.0", func() {
			So(GEQ("1.1-debug.1.1", "1.1.1.0"), ShouldBeTrue)
		})
		Convey("A: 1.1.1.0 B: 1.1-debug.1.1", func() {
			So(GEQ("1.1.1.0", "1.1-debug.1.1"), ShouldBeFalse)
		})
		Convey("A: 10.12.33.1 B: 10.12.33.1-rc4", func() {
			So(GEQ("10.12.33.1", "10.12.33.1-rc4"), ShouldBeFalse)
		})
		Convey("A: 10.12.33.1-rc4 B: 10.12.33.1", func() {
			So(GEQ("10.12.33.1-rc4", "10.12.33.1"), ShouldBeTrue)
		})
		Convey("A:  B: 10.12.33.1", func() {
			So(GEQ("", "10.12.33.1"), ShouldBeFalse)
		})
		Convey("A: 10.12.33.1 B: ", func() {
			So(GEQ("10.12.33.1", ""), ShouldBeTrue)
		})
		Convey("A: Batman B: 10.12.33.1", func() {
			So(GEQ("Batman", "10.12.33.1"), ShouldBeFalse)
		})
		Convey("A: 10.12.33.1 B: Superman", func() {
			So(GEQ("10.12.33.1", "Superman"), ShouldBeTrue)
		})
		Convey("Transitive property test. If A >= B and B >= C then A>=C", func() {
			MaxCount := 100000 // 100k tests each time ??
			transitiveTest := func(a, b, c string) bool {
				if GEQ(a, b) && GEQ(b, c) {
					return GEQ(a, c)
				}
				// Ignore the cases where we can't check
				return true
			}
			// Run a test to figure out if it fails,
			err := quick.Check(transitiveTest, &quick.Config{MaxCount: MaxCount})
			So(err, ShouldBeNil)
		})
	})
}
