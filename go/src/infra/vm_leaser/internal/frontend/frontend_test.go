// Copyright 2023 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
	"google.golang.org/protobuf/types/known/durationpb"
)

func testingContext() context.Context {
	ctx := context.Background()
	ctx = gologger.StdConfig.Use(ctx)
	ctx = logging.SetLevel(ctx, logging.Debug)
	return ctx
}

func TestComputeExpirationTime(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	Convey("Test computeExpirationTime", t, func() {
		Convey("Compute expiration time - no lease duration passed", func() {
			defaultExpTime := time.Now().Unix() + (DefaultLeaseDuration * 60)
			res, err := computeExpirationTime(ctx, nil)
			So(err, ShouldBeNil)
			So(res, ShouldBeBetweenOrEqual, defaultExpTime, defaultExpTime+1)
		})
		Convey("Compute expiration time - lease duration passed", func() {
			leaseDuration, err := time.ParseDuration("20m")
			So(err, ShouldBeNil)

			expTime := time.Now().Add(leaseDuration).Unix()
			logging.Errorf(ctx, "%s", durationpb.New(leaseDuration))
			res, err := computeExpirationTime(ctx, durationpb.New(leaseDuration))
			So(err, ShouldBeNil)
			So(res, ShouldBeBetweenOrEqual, expTime, expTime+1)
		})
	})
}
