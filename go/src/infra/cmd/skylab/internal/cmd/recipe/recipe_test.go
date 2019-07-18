// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package recipe

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
)

func TestSchedulingParam(t *testing.T) {
	Convey("Given a", t, func() {
		cases := []struct {
			name                  string
			inputPool             string
			inputAccount          string
			expectedAccount       string
			expectedManagedPool   test_platform.Request_Params_Scheduling_ManagedPool
			expectedUnmanagedPool string
		}{
			{
				name:            "quota account",
				inputAccount:    "foo account",
				expectedAccount: "foo account",
			},
			{
				name:                "long-named managed pool",
				inputPool:           "MANAGED_POOL_CQ",
				expectedManagedPool: test_platform.Request_Params_Scheduling_MANAGED_POOL_CQ,
			},
			{
				name:                "skylab-named managed pool",
				inputPool:           "DUT_POOL_CQ",
				expectedManagedPool: test_platform.Request_Params_Scheduling_MANAGED_POOL_CQ,
			},
			{
				name:                "short-named managed pool",
				inputPool:           "cq",
				expectedManagedPool: test_platform.Request_Params_Scheduling_MANAGED_POOL_CQ,
			},
			{
				name:                  "unmanaged pool",
				inputPool:             "foo-pool",
				expectedUnmanagedPool: "foo-pool",
			},
		}
		for _, c := range cases {
			Convey(c.name, func() {
				s := toScheduling(c.inputPool, c.inputAccount)
				Convey("then scheduling parameters are correct.", func() {
					So(s.GetManagedPool(), ShouldResemble, c.expectedManagedPool)
					So(s.GetQuotaAccount(), ShouldResemble, c.expectedAccount)
					So(s.GetUnmanagedPool(), ShouldResemble, c.expectedUnmanagedPool)
				})
			})
		}
	})
}
