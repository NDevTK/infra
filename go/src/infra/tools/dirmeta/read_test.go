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

func TestReadMapping(t *testing.T) {
	t.Parallel()

	Convey(`ReadMapping`, t, func() {
		actual, err := ReadMapping("testdata/root")
		So(err, ShouldBeNil)
		So(actual.Proto(), ShouldResembleProto, &dirmetapb.Mapping{
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
}

func TestReadInherited(t *testing.T) {
	t.Parallel()

	Convey(`ReadInheritance`, t, func() {
		md, err := ReadInherited("testdata/inheritance", "testdata/inheritance/a/b")
		So(err, ShouldBeNil)
		So(md, ShouldResembleProto, &dirmetapb.Metadata{
			TeamEmail: "chromium-review@chromium.org",
			Monorail: &dirmetapb.Monorail{
				Project:   "chromium",
				Component: "Component",
			},
		})
	})
}
