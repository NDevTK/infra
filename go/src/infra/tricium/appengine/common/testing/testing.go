// Copyright 2016 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package testing implements common testing functionality for the Tricium service modules.
package testing

import (
	"time"

	"github.com/luci/gae/impl/memory"
	ds "github.com/luci/gae/service/datastore"
	tq "github.com/luci/gae/service/taskqueue"
	"github.com/luci/luci-go/common/clock/testclock"
	"github.com/luci/luci-go/common/logging/memlogger"

	"golang.org/x/net/context"

	"infra/tricium/appengine/common"
)

// Testing is a high-level testing object.
type Testing struct {
}

// Context generates a correctly configured context with queues and clock.
func (t *Testing) Context() context.Context {
	ctx := memory.Use(memlogger.Use(context.Background()))
	ctx, _ = testclock.UseTime(ctx, testclock.TestTimeUTC.Round(time.Millisecond))
	tq.GetTestable(ctx).CreateQueue(common.ServiceQueue)
	tq.GetTestable(ctx).CreateQueue(common.LauncherQueue)
	tq.GetTestable(ctx).CreateQueue(common.DriverQueue)
	tq.GetTestable(ctx).CreateQueue(common.TrackerQueue)
	ds.GetTestable(ctx).Consistent(true)
	return ctx
}
