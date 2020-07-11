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

func TestMappingReader(t *testing.T) {
	t.Parallel()

	Convey(`MappingReader`, t, func() {
		var r MappingReader

		Convey(`ReadAll(false)`, func() {
			r.Root = "testdata/root"
			err := r.ReadAll(false)
			So(err, ShouldBeNil)
			So(r.Mapping.Proto(), ShouldResembleProto, &dirmetapb.Mapping{
				Dirs: map[string]*dirmetapb.Metadata{
					".": {
						TeamEmail: "chromium-review@chromium.org",
						Os:        dirmetapb.OS_LINUX,
					},
					"subdir_with_owners": {
						TeamEmail: "team-email@chromium.org",
						// OS was not inherited
						Monorail: &dirmetapb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
					},
					// "subdir_with_owners/empty_subdir" is not present because it has
					// not metadata.
				},
			})
		})

		Convey(`ReadAll(true)`, func() {
			r.Root = "testdata/root"
			err := r.ReadAll(true)
			So(err, ShouldBeNil)
			So(r.Mapping.Proto(), ShouldResembleProto, &dirmetapb.Mapping{
				Dirs: map[string]*dirmetapb.Metadata{
					".": {
						TeamEmail: "chromium-review@chromium.org",
						Os:        dirmetapb.OS_LINUX,
					},
					"subdir_with_owners": {
						TeamEmail: "team-email@chromium.org",
						Os:        dirmetapb.OS_LINUX,
						Monorail: &dirmetapb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
					},
					"subdir_with_owners/empty_subdir": {
						TeamEmail: "team-email@chromium.org",
						Os:        dirmetapb.OS_LINUX,
						Monorail: &dirmetapb.Monorail{
							Project:   "chromium",
							Component: "Some>Component",
						},
					},
				},
			})
		})
	})
}
