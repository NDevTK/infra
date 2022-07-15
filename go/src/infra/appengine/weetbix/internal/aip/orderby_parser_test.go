// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package aip

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
)

func TestParseOrderBy(t *testing.T) {
	Convey("ParseOrderBy", t, func() {
		// Test examples from the AIP-132 spec.
		Convey("Values should be a comma separated list of fields", func() {
			result, err := ParseOrderBy("foo,bar")
			So(err, ShouldBeNil)
			So(result, ShouldResemble, []OrderBy{
				{
					Name: "foo",
				},
				{
					Name: "bar",
				},
			})

			result, err = ParseOrderBy("foo")
			So(err, ShouldBeNil)
			So(result, ShouldResemble, []OrderBy{
				{
					Name: "foo",
				},
			})
		})
		Convey("The default sort order is ascending", func() {
			result, err := ParseOrderBy("foo desc, bar")
			So(err, ShouldBeNil)
			So(result, ShouldResemble, []OrderBy{
				{
					Name:       "foo",
					Descending: true,
				},
				{
					Name: "bar",
				},
			})
		})
		Convey("Redundant space characters in the syntax are insignificant", func() {
			expectedResult := []OrderBy{
				{
					Name: "foo",
				},
				{
					Name:       "bar",
					Descending: true,
				},
			}
			result, err := ParseOrderBy("foo, bar desc")
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectedResult)

			result, err = ParseOrderBy("  foo  ,  bar desc  ")
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectedResult)

			result, err = ParseOrderBy("foo,bar desc")
			So(err, ShouldBeNil)
			So(result, ShouldResemble, expectedResult)
		})
		Convey("Subfields are specified with a . character", func() {
			result, err := ParseOrderBy("foo.bar, foo.foo desc")
			So(err, ShouldBeNil)
			So(result, ShouldResemble, []OrderBy{
				{
					Name: "foo.bar",
				},
				{
					Name:       "foo.foo",
					Descending: true,
				},
			})
		})
		Convey("Invalid input is rejected", func() {
			_, err := ParseOrderBy("`something")
			So(err, ShouldErrLike, "invalid ordering \"`something\"")
		})
		Convey("Empty order by", func() {
			result, err := ParseOrderBy("   ")
			So(err, ShouldBeNil)
			So(result, ShouldHaveLength, 0)
		})
	})
}
