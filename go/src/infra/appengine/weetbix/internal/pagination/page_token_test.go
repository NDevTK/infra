// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package pagination

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPageToken(t *testing.T) {
	t.Parallel()

	Convey(`Token works`, t, func() {
		So(Token("v1", "v2"), ShouldResemble, "CgJ2MQoCdjI=")

		pos, err := ParseToken("CgJ2MQoCdjI=")
		So(err, ShouldBeNil)
		So(pos, ShouldResemble, []string{"v1", "v2"})

		Convey(`For fresh token`, func() {
			So(Token(), ShouldResemble, "")

			pos, err := ParseToken("")
			So(err, ShouldBeNil)
			So(pos, ShouldBeNil)
		})
	})
}
