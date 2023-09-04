// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/chromiumos/config/go/test/api"
	. "go.chromium.org/luci/common/testing/assertions"

	"infra/vm_leaser/internal/constants"
)

func TestHandleLeaseVMError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	var allZones []string
	for _, a := range constants.AllQuotaZones {
		allZones = append(allZones, a...)
	}
	Convey("Test handleLeaseVMError", t, func() {
		Convey("handleLeaseVMError - no error; return original request", func() {
			req := &api.LeaseVMRequest{
				HostReqs: &api.VMRequirements{
					GceRegion:  "test-region",
					GceProject: "test-project",
				},
			}
			newReq := handleLeaseVMError(ctx, req, nil, nil)
			So(req, ShouldResembleProto, newReq)
		})
		Convey("handleLeaseVMError - QUOTA_EXCEEDED error; return request with new zone", func() {
			req := &api.LeaseVMRequest{
				HostReqs: &api.VMRequirements{
					GceRegion:  "test-region",
					GceProject: "test-project",
				},
			}
			err := errors.New("QUOTA_EXCEEDED error test")
			quotaExceededZones := map[string]bool{}
			newReq := handleLeaseVMError(ctx, req, err, quotaExceededZones)
			So(newReq.GetHostReqs().GetGceRegion(), ShouldNotEqual, "test-region")
			So(newReq.GetHostReqs().GetGceRegion(), ShouldBeIn, allZones)
		})
	})
}
