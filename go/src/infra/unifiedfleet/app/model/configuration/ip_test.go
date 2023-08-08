// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configuration

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/appengine/gaetesting"
	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/gae/service/datastore"

	ufspb "infra/unifiedfleet/api/v1/models"
)

func TestBatchUpdateIPs(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	datastore.GetTestable(ctx).Consistent(true)
	Convey("happy path", t, func() {
		count := 10
		ips := mockIps(count)

		resp, err := BatchUpdateIPs(ctx, ips)

		So(err, ShouldBeNil)
		So(resp, ShouldHaveLength, len(ips))

		getRes, _, err := ListIPs(ctx, 10, "", nil, false)

		So(err, ShouldBeNil)
		So(getRes, ShouldResembleProto, ips)
	})
	Convey("happy path - Updates multiple batches of IPs", t, func() {
		count := 700
		ips := mockIps(count)

		resp, err := BatchUpdateIPs(ctx, ips)

		So(err, ShouldBeNil)
		So(resp, ShouldHaveLength, len(ips))

		getRes, _, err := ListIPs(ctx, 700, "", nil, false)

		So(err, ShouldBeNil)
		So(getRes, ShouldHaveLength, count)
	})
}

func mockIps(count int) []*ufspb.IP {
	protos := make([]*ufspb.IP, count)
	for i := 0; i < count; i++ {
		protos[i] = &ufspb.IP{
			Id:      fmt.Sprint(i),
			Ipv4:    1111,
			Vlan:    "vlan" + fmt.Sprint(i),
			Ipv4Str: "1111",
		}
	}
	return protos
}
