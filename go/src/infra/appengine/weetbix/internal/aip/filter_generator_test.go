// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package aip

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
)

func TestWhereClause(t *testing.T) {
	Convey("WhereClause", t, func() {
		table := NewTable().WithColumns(
			NewColumn().WithName("foo").WithDatabaseName("db_foo").FilterableImplicitly().Build(),
			NewColumn().WithName("bar").WithDatabaseName("db_bar").FilterableImplicitly().Build(),
			NewColumn().WithName("baz").WithDatabaseName("db_baz").Filterable().Build(),
			NewColumn().WithName("unfilterable").WithDatabaseName("unfilterable").Build(),
		).Build()

		Convey("Empty filter", func() {
			result, pars, err := table.WhereClause(&Filter{}, "p_")
			So(err, ShouldBeNil)
			So(pars, ShouldHaveLength, 0)
			So(result, ShouldEqual, "(TRUE)")
		})
		Convey("Simple filter", func() {
			Convey("has operator", func() {
				filter, err := ParseFilter("foo:somevalue")
				So(err, ShouldEqual, nil)

				result, pars, err := table.WhereClause(filter, "p_")
				So(err, ShouldBeNil)
				So(pars, ShouldResemble, []QueryParameter{
					{
						Name:  "p_0",
						Value: "%somevalue%",
					},
				})
				So(result, ShouldEqual, "(db_foo LIKE @p_0)")
			})
			Convey("equals operator", func() {
				filter, err := ParseFilter("foo = somevalue")
				So(err, ShouldEqual, nil)

				result, pars, err := table.WhereClause(filter, "p_")
				So(err, ShouldBeNil)
				So(pars, ShouldResemble, []QueryParameter{
					{
						Name:  "p_0",
						Value: "somevalue",
					},
				})
				So(result, ShouldEqual, "(db_foo = @p_0)")
			})
			Convey("not equals operator", func() {
				filter, err := ParseFilter("foo != somevalue")
				So(err, ShouldEqual, nil)

				result, pars, err := table.WhereClause(filter, "p_")
				So(err, ShouldBeNil)
				So(pars, ShouldResemble, []QueryParameter{
					{
						Name:  "p_0",
						Value: "somevalue",
					},
				})
				So(result, ShouldEqual, "(db_foo <> @p_0)")
			})
			Convey("implicit match operator", func() {
				filter, err := ParseFilter("somevalue")
				So(err, ShouldEqual, nil)

				result, pars, err := table.WhereClause(filter, "p_")
				So(err, ShouldBeNil)
				So(pars, ShouldResemble, []QueryParameter{
					{
						Name:  "p_0",
						Value: "%somevalue%",
					},
				})
				So(result, ShouldEqual, "(db_foo LIKE @p_0 OR db_bar LIKE @p_0)")
			})
			Convey("unsupported composite to LIKE", func() {
				filter, err := ParseFilter("foo:(somevalue)")
				So(err, ShouldEqual, nil)

				_, _, err = table.WhereClause(filter, "p_")
				So(err, ShouldErrLike, "composite expressions are not allowed as RHS to has (:) operator")
			})
			Convey("unsupported composite to equals", func() {
				filter, err := ParseFilter("foo=(somevalue)")
				So(err, ShouldEqual, nil)

				_, _, err = table.WhereClause(filter, "p_")
				So(err, ShouldErrLike, "composite expressions in arguments not implemented yet")
			})
			Convey("unsupported field LHS", func() {
				filter, err := ParseFilter("foo.baz=blah")
				So(err, ShouldEqual, nil)

				_, _, err = table.WhereClause(filter, "p_")
				So(err, ShouldErrLike, "fields not implemented yet")
			})
			Convey("unsupported field RHS", func() {
				filter, err := ParseFilter("foo=blah.baz")
				So(err, ShouldEqual, nil)

				_, _, err = table.WhereClause(filter, "p_")
				So(err, ShouldErrLike, "fields not implemented yet")
			})
			Convey("field on RHS of has", func() {
				filter, err := ParseFilter("foo:blah.baz")
				So(err, ShouldEqual, nil)

				_, _, err = table.WhereClause(filter, "p_")
				So(err, ShouldErrLike, "fields are not allowed on the RHS of has (:) operator")
			})
		})
		Convey("Complex filter", func() {
			filter, err := ParseFilter("implicit (foo=explicitone) OR -bar=explicittwo AND foo!=explicitthree OR baz:explicitfour")
			So(err, ShouldEqual, nil)

			result, pars, err := table.WhereClause(filter, "p_")
			So(err, ShouldBeNil)
			So(pars, ShouldResemble, []QueryParameter{
				{
					Name:  "p_0",
					Value: "%implicit%",
				},
				{
					Name:  "p_1",
					Value: "explicitone",
				},
				{
					Name:  "p_2",
					Value: "explicittwo",
				},
				{
					Name:  "p_3",
					Value: "explicitthree",
				},
				{
					Name:  "p_4",
					Value: "%explicitfour%",
				},
			})
			So(result, ShouldEqual, "((db_foo LIKE @p_0 OR db_bar LIKE @p_0) AND ((db_foo = @p_1) OR (NOT (db_bar = @p_2))) AND ((db_foo <> @p_3) OR (db_baz LIKE @p_4)))")
		})
	})
}
