// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package controller

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"

	ufspb "infra/unifiedfleet/api/v1/models"
)

func TestCreateDefaultWifi(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	Convey("CreateDefaultWifi", t, func() {
		Convey("Create new DefaultWifi - happy path", func() {
			wifi := &ufspb.DefaultWifi{Name: "zone_sfo36_os"}
			resp, err := CreateDefaultWifi(ctx, wifi)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp, ShouldResembleProto, wifi)
		})
		Convey("Create new DefaultWifi - already existing", func() {
			w1 := &ufspb.DefaultWifi{Name: "pool1"}
			_, _ = CreateDefaultWifi(ctx, w1)

			dup := &ufspb.DefaultWifi{Name: "pool1"}
			_, err := CreateDefaultWifi(ctx, dup)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "already exists")
		})
	})
}
