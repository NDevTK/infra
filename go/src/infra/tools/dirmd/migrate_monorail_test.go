// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dirmd

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	. "go.chromium.org/luci/common/testing/assertions"

	dirmdpb "infra/tools/dirmd/proto"
)

// TODO(crbug.com/1505875) - Remove this once migration is complete.
func TestHandleMetadata(t *testing.T) {
	t.Parallel()

	// The code will translate the .textproto mapping into a map with lower-cased
	// keys so that the checks remain case insensitive.
	cm := map[string]int64{
		"test>component": 12345,
	}
	dir := "/some/path/to/dir/"

	Convey(`Invalid`, t, func() {
		Convey(`Monorail Nil`, func() {
			md := &dirmdpb.Metadata{
				TeamEmail: "team@sample.com",
			}
			md, err := HandleMetadata(md, cm, dir)
			So(md, ShouldBeNil)
			So(err, ShouldErrLike, MonorailMissingError)
		})

		Convey(`Monorail Component Nil`, func() {
			md := &dirmdpb.Metadata{
				Monorail: &dirmdpb.Monorail{
					Project: "chromium",
				},
			}
			md, err := HandleMetadata(md, cm, dir)
			So(md, ShouldBeNil)
			So(err, ShouldErrLike, MonorailMissingError)
		})
		Convey(`Missing Component`, func() {
			md := &dirmdpb.Metadata{
				Monorail: &dirmdpb.Monorail{
					Project:   "chromium",
					Component: "Random>Component",
				},
			}
			md, err := HandleMetadata(md, cm, dir)
			So(md, ShouldBeNil)
			So(err, ShouldErrLike, "Random>Component is missing from the provided mapping")
		})
	})

	Convey(`Valid`, t, func() {
		Convey(`Buganizer Defined`, func() {
			md := &dirmdpb.Metadata{
				Buganizer: &dirmdpb.Buganizer{
					ComponentId: 123,
				},
			}
			newMd, err := HandleMetadata(md, cm, dir)
			So(err, ShouldBeNil)
			So(newMd, ShouldResembleProto, md)
		})
		Convey(`Buganizer Public Defined`, func() {
			md := &dirmdpb.Metadata{
				BuganizerPublic: &dirmdpb.Buganizer{
					ComponentId: 123,
				},
			}
			newMd, err := HandleMetadata(md, cm, dir)
			So(err, ShouldBeNil)
			So(newMd, ShouldResembleProto, md)
		})
		Convey(`1:1 mapping`, func() {
			md := &dirmdpb.Metadata{
				Monorail: &dirmdpb.Monorail{
					Project:   "chromium",
					Component: "test>component",
				},
			}
			newMd, err := HandleMetadata(md, cm, dir)
			So(err, ShouldBeNil)

			expected := &dirmdpb.Metadata{
				Monorail: &dirmdpb.Monorail{
					Project:   "chromium",
					Component: "test>component",
				},
				Buganizer: &dirmdpb.Buganizer{
					ComponentId: 12345,
				},
			}
			So(newMd, ShouldResembleProto, expected)
		})
		Convey(`Case insensitive`, func() {
			md := &dirmdpb.Metadata{
				Monorail: &dirmdpb.Monorail{
					Project:   "chromium",
					Component: "tEsT>CompOnent",
				},
			}
			newMd, err := HandleMetadata(md, cm, dir)
			So(err, ShouldBeNil)

			expected := &dirmdpb.Metadata{
				Monorail: &dirmdpb.Monorail{
					Project:   "chromium",
					Component: "tEsT>CompOnent",
				},
				Buganizer: &dirmdpb.Buganizer{
					ComponentId: 12345,
				},
			}
			So(newMd, ShouldResembleProto, expected)
		})
		Convey(`Non Chromium Monorail Project`, func() {
			md := &dirmdpb.Metadata{
				Monorail: &dirmdpb.Monorail{
					Project:   "v8",
					Component: "Random>Component",
				},
			}
			md, err := HandleMetadata(md, cm, dir)
			So(err, ShouldBeNil)

			// Should remain unchanged.
			So(md, ShouldResembleProto, md)
		})
	})
}

func TestCanSkipMixin(t *testing.T) {
	t.Parallel()

	Convey(`Skip`, t, func() {
		Convey(`Buganizer Present`, func() {
			mixin := &dirmdpb.Metadata{
				Buganizer: &dirmdpb.Buganizer{
					ComponentId: 12345,
				},
			}
			So(canSkipMixin(mixin), ShouldBeTrue)
		})
		Convey(`Buganizer & Monorail Present`, func() {
			mixin := &dirmdpb.Metadata{
				Monorail: &dirmdpb.Monorail{
					Project:   "chromium",
					Component: "tEsT>CompOnent",
				},
				Buganizer: &dirmdpb.Buganizer{
					ComponentId: 12345,
				},
			}
			So(canSkipMixin(mixin), ShouldBeTrue)
		})
		Convey(`No Monorail`, func() {
			mixin := &dirmdpb.Metadata{}
			So(canSkipMixin(mixin), ShouldBeTrue)
		})
		Convey(`Non Chromium`, func() {
			mixin := &dirmdpb.Metadata{
				Monorail: &dirmdpb.Monorail{
					Project:   "v8",
					Component: "tEsT>CompOnent",
				},
			}
			So(canSkipMixin(mixin), ShouldBeTrue)
		})
	})

	Convey(`No Skip`, t, func() {
		Convey(`Monorail Only`, func() {
			mixin := &dirmdpb.Metadata{
				Monorail: &dirmdpb.Monorail{
					Project:   "chromium",
					Component: "tEsT>CompOnent",
				},
			}
			So(canSkipMixin(mixin), ShouldBeFalse)
		})
	})

	cm := map[string]int64{
		"test>component": 12345,
	}

	Convey(`Handle Mixins`, t, func() {
		Convey(`Nil md`, func() {
			mixins, err := HandleMixins(nil, cm, "/root")
			So(err, ShouldBeNil)
			So(len(mixins), ShouldEqual, 0)
		})
		Convey(`No Mixins`, func() {
			md := &dirmdpb.Metadata{
				Monorail: &dirmdpb.Monorail{
					Project:   "chromium",
					Component: "tEsT>CompOnent",
				},
			}

			mixins, err := HandleMixins(md, cm, "/root")
			So(err, ShouldBeNil)
			So(len(mixins), ShouldEqual, 0)
		})
		Convey(`Empty Mixins`, func() {
			md := &dirmdpb.Metadata{
				Mixins: make([]string, 0),
			}

			mixins, err := HandleMixins(md, cm, "/root")
			So(err, ShouldBeNil)
			So(len(mixins), ShouldEqual, 0)
		})
	})
}
