// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package pagination

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
)

func TestPageSizeLimiter(t *testing.T) {
	t.Parallel()

	Convey(`PageSizeLimiter`, t, func() {
		psl := PageSizeLimiter{
			Max:     1000,
			Default: 10,
		}

		Convey(`Adjust works`, func() {
			So(psl.Adjust(0), ShouldEqual, 10)
			So(psl.Adjust(10000), ShouldEqual, 1000)
			So(psl.Adjust(500), ShouldEqual, 500)
			So(psl.Adjust(5), ShouldEqual, 5)
		})
	})
}

func TestValidatePageSize(t *testing.T) {
	t.Parallel()

	Convey(`ValidatePageSize`, t, func() {
		Convey(`Positive`, func() {
			So(ValidatePageSize(10), ShouldBeNil)
		})
		Convey(`Zero`, func() {
			So(ValidatePageSize(0), ShouldBeNil)
		})
		Convey(`Negative`, func() {
			So(ValidatePageSize(-10), ShouldErrLike, "negative")
		})
	})
}
