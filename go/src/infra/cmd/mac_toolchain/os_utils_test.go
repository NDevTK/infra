// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/common/errors"
)

func TestParseOSVersion(t *testing.T) {
	t.Parallel()

	Convey("check isMacOS13OrLater works", t, func() {
		var s MockSession
		ctx := useMockCmd(context.Background(), &s)

		Convey("check a version that is greater than 13", func() {
			s.ReturnOutput = []string{
				"13.0.1",
			}
			os13OrLater, err := isMacOS13OrLater(ctx)
			So(err, ShouldBeNil)
			So(os13OrLater, ShouldEqual, true)
		})

		Convey("check a version that is equal to 13", func() {
			s.ReturnOutput = []string{
				"13.0.0",
			}
			os13OrLater, err := isMacOS13OrLater(ctx)
			So(err, ShouldBeNil)
			So(os13OrLater, ShouldEqual, true)
		})

		Convey("check a version that less than 13", func() {
			s.ReturnOutput = []string{
				"12.1.2",
			}
			os13OrLater, err := isMacOS13OrLater(ctx)
			So(err, ShouldBeNil)
			So(os13OrLater, ShouldEqual, false)
		})

		Convey("invalid output should return false", func() {
			s.ReturnOutput = []string{
				"invalid",
			}
			os13OrLater, err := isMacOS13OrLater(ctx)
			So(err, ShouldNotBeNil)
			So(os13OrLater, ShouldEqual, false)
		})

		Convey("error output should return false", func() {
			s.ReturnError = []error{errors.Reason("random Error").Err()}
			os13OrLater, err := isMacOS13OrLater(ctx)
			So(err, ShouldNotBeNil)
			So(os13OrLater, ShouldEqual, false)
		})
	})
}
