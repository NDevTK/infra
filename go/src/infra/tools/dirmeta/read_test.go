// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dirmeta

import (
	"testing"

	dirmetapb "infra/tools/dirmeta/proto"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
)

func TestReadFull(t *testing.T) {
	t.Parallel()

	Convey(`MappingReader`, t, func() {
		var r MappingReader

		Convey(`ReadFull`, func() {
			r.Root = "testdata/root"
			err := r.ReadFull()
			So(err, ShouldBeNil)
			So(r.Mapping.Proto(), ShouldResembleProto, &dirmetapb.Mapping{
				Dirs: map[string]*dirmetapb.Metadata{
					".": {
						TeamEmail: "chromium-review@chromium.org",
					},
					"subdir_with_owners": {
						TeamEmail: "team-email@chromium.org",
						Os:        dirmetapb.OS_IOS,
						Monorail: &dirmetapb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
					},
				},
			})
		})

		Convey(`ReadInheritance`, func() {
			r.Root = "testdata/inheritance"
			err := r.ReadTowards("testdata/inheritance/a/b")
			So(err, ShouldBeNil)
			So(r.Dirs, ShouldHaveLength, 3)
			So(r.Dirs, ShouldContainKey, ".")
			So(r.Dirs, ShouldContainKey, "a")
			So(r.Dirs, ShouldContainKey, "a/b")
		})
	})
}
