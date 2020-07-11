// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dirmeta

import (
	dirmetapb "infra/tools/dirmeta/proto"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
)

func TestExpand(t *testing.T) {
	t.Parallel()

	Convey(`Expand`, t, func() {
		Convey(`Works`, func() {
			input := &Mapping{
				Dirs: map[string]*dirmetapb.Metadata{
					".": {
						TeamEmail: "team@example.com",
						// Will be inherited entirely.
						Wpt: &dirmetapb.WPT{Notify: true},

						// Will be inherited partially.
						Monorail: &dirmetapb.Monorail{
							Project: "chromium",
						},
					},
					"a": {
						TeamEmail: "team-email@chromium.org",
						Monorail: &dirmetapb.Monorail{
							Component: "Component",
						},
					},
				},
			}
			actual := input.Expand()
			So(actual.Proto(), ShouldResembleProto, &dirmetapb.Mapping{
				Dirs: map[string]*dirmetapb.Metadata{
					".": input.Dirs["."], // did not change
					"a": {
						TeamEmail: "team-email@chromium.org",
						Wpt:       &dirmetapb.WPT{Notify: true},
						Monorail: &dirmetapb.Monorail{
							Project:   "chromium",
							Component: "Component",
						},
					},
				},
			})
		})

		Convey(`Deep nesting`, func() {
			input := &Mapping{
				Dirs: map[string]*dirmetapb.Metadata{
					".":   {TeamEmail: "team@example.com"},
					"a":   {},
					"a/b": {},
				},
			}
			actual := input.Expand()
			So(actual.Dirs["a/b"].TeamEmail, ShouldEqual, "team@example.com")
		})

		Convey(`No root`, func() {
			input := &Mapping{
				Dirs: map[string]*dirmetapb.Metadata{
					"a": {TeamEmail: "a"},
					"b": {TeamEmail: "b"},
				},
			}
			actual := input.Expand()
			So(actual.Proto(), ShouldResembleProto, input.Proto())
		})
	})
}

func TestInherited(t *testing.T) {
	t.Parallel()

	Convey(`ancestorsAndSelf`, t, func() {
		Convey(`.`, func() {
			actual := ancestorsAndSelf(".")
			So(actual, ShouldResemble, []string{"."})
		})
		Convey(`a/b/c`, func() {
			actual := ancestorsAndSelf("a/b/c")
			So(actual, ShouldResemble, []string{
				".",
				"a",
				"a/b",
				"a/b/c",
			})
		})
	})

	Convey(`Inherited`, t, func() {
		Convey(`Empty`, func() {
			m := &Mapping{}
			So(m.Inherited("a"), ShouldResembleProto, &dirmetapb.Metadata{})
			So(m.Inherited("."), ShouldResembleProto, &dirmetapb.Metadata{})
		})
		Convey(`With just root`, func() {
			m := &Mapping{
				Dirs: map[string]*dirmetapb.Metadata{
					".": {TeamEmail: "root"},
				},
			}
			So(m.Inherited("a"), ShouldResembleProto, &dirmetapb.Metadata{TeamEmail: "root"})
			So(m.Inherited("."), ShouldResembleProto, &dirmetapb.Metadata{TeamEmail: "root"})
		})
		Convey(`With root and subdir`, func() {
			m := &Mapping{
				Dirs: map[string]*dirmetapb.Metadata{
					".": {TeamEmail: "root"},
					"a": {Wpt: &dirmetapb.WPT{Notify: true}},
				},
			}
			So(m.Inherited("a"), ShouldResembleProto, &dirmetapb.Metadata{
				TeamEmail: "root",
				Wpt:       &dirmetapb.WPT{Notify: true},
			})

			So(m.Inherited("a/b"), ShouldResembleProto, &dirmetapb.Metadata{
				TeamEmail: "root",
				Wpt:       &dirmetapb.WPT{Notify: true},
			})
		})
		Convey(`Long change`, func() {
			m := &Mapping{
				Dirs: map[string]*dirmetapb.Metadata{
					".":     {TeamEmail: "0"},
					"a":     {TeamEmail: "1"},
					"a/b":   {TeamEmail: "2"},
					"a/b/c": {TeamEmail: "3"},
				},
			}
			So(m.Inherited("a/b/c").TeamEmail, ShouldEqual, "3")
			So(m.Inherited("a/b/c/d").TeamEmail, ShouldEqual, "3")
		})
	})
}
