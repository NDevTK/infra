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

func TestReduce(t *testing.T) {
	t.Parallel()

	Convey(`Nearest ancestor`, t, func() {
		m := &Mapping{
			Dirs: map[string]*dirmetapb.Metadata{
				".": {TeamEmail: "0"},
			},
		}
		So(m.nearestAncestor("a/b/c").TeamEmail, ShouldEqual, "0")
	})

	Convey(`Reduce`, t, func() {
		Convey(`Works`, func() {
			input := &Mapping{
				Dirs: map[string]*dirmetapb.Metadata{
					".": {
						TeamEmail: "team@example.com",
						Monorail: &dirmetapb.Monorail{
							Project: "chromium",
						},
					},
					"a": {
						TeamEmail: "team@example.com", // redundant
						Wpt:       &dirmetapb.WPT{Notify: dirmetapb.Trinary_YES},
						Monorail: &dirmetapb.Monorail{
							Project:   "chromium", // redundant
							Component: "Component",
						},
					},
				},
			}
			actual := input.Reduce()
			So(actual.Proto(), ShouldResembleProto, &dirmetapb.Mapping{
				Dirs: map[string]*dirmetapb.Metadata{
					".": input.Dirs["."], // did not change
					"a": {
						Wpt: &dirmetapb.WPT{Notify: dirmetapb.Trinary_YES},
						Monorail: &dirmetapb.Monorail{
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
					"a":   {TeamEmail: "team@example.com"},
					"a/b": {TeamEmail: "team@example.com"},
				},
			}
			actual := input.Reduce()
			So(actual.Dirs, ShouldNotContainKey, "a")
			So(actual.Dirs, ShouldNotContainKey, "a/b")
		})
	})
}
