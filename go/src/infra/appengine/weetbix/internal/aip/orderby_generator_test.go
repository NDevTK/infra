// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package aip

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
)

func TestOrderByClause(t *testing.T) {
	Convey("OrderByClause", t, func() {
		table := NewTable().WithColumns(
			NewColumn().WithName("foo").WithDatabaseName("db_foo").Sortable().Build(),
			NewColumn().WithName("bar").WithDatabaseName("db_bar").Sortable().Build(),
			NewColumn().WithName("baz").WithDatabaseName("db_baz").Sortable().Build(),
			NewColumn().WithName("unsortable").WithDatabaseName("unsortable").Build(),
		).Build()

		Convey("Empty order by", func() {
			result, err := table.OrderByClause([]OrderBy{})
			So(err, ShouldBeNil)
			So(result, ShouldEqual, "")
		})
		Convey("Single order by", func() {
			result, err := table.OrderByClause([]OrderBy{
				{
					Name: "foo",
				},
			})
			So(err, ShouldBeNil)
			So(result, ShouldEqual, "ORDER BY db_foo\n")
		})
		Convey("Multiple order by", func() {
			result, err := table.OrderByClause([]OrderBy{
				{
					Name:       "foo",
					Descending: true,
				},
				{
					Name: "bar",
				},
				{
					Name:       "baz",
					Descending: true,
				},
			})
			So(err, ShouldBeNil)
			So(result, ShouldEqual, "ORDER BY db_foo DESC, db_bar, db_baz DESC\n")
		})
		Convey("Unsortable field in order by", func() {
			_, err := table.OrderByClause([]OrderBy{
				{
					Name:       "unsortable",
					Descending: true,
				},
			})
			So(err, ShouldErrLike, `no sortable field named "unsortable", valid fields are foo, bar, baz`)
		})
		Convey("Repeated field in order by", func() {
			_, err := table.OrderByClause([]OrderBy{
				{
					Name: "foo",
				},
				{
					Name: "foo",
				},
			})
			So(err, ShouldErrLike, `field appears in order_by multiple times: "foo"`)
		})
	})
}

func TestMergeWithDefaultOrder(t *testing.T) {
	Convey("MergeWithDefaultOrder", t, func() {
		defaultOrder := []OrderBy{
			{
				Name:       "foo",
				Descending: true,
			}, {
				Name: "bar",
			}, {
				Name:       "baz",
				Descending: true,
			},
		}
		Convey("Empty order", func() {
			result := MergeWithDefaultOrder(defaultOrder, nil)
			So(result, ShouldResemble, defaultOrder)
		})
		Convey("Non-empty order", func() {
			order := []OrderBy{
				{
					Name:       "other",
					Descending: true,
				},
				{
					Name: "baz",
				},
			}
			result := MergeWithDefaultOrder(defaultOrder, order)
			So(result, ShouldResemble, []OrderBy{
				{
					Name:       "other",
					Descending: true,
				},
				{
					Name: "baz",
				},
				{
					Name:       "foo",
					Descending: true,
				}, {
					Name: "bar",
				},
			})
		})
	})
}
