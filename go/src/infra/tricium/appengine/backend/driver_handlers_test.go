// Copyright 2016 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/common/clock"
	ds "go.chromium.org/luci/gae/service/datastore"
	tq "go.chromium.org/luci/gae/service/taskqueue"
	"google.golang.org/api/pubsub/v1"

	admin "infra/tricium/api/admin/v1"
	"infra/tricium/appengine/common"
	"infra/tricium/appengine/common/triciumtest"
)

var (
	msg = &pubsub.PubsubMessage{
		MessageId:   "58708071417623",
		PublishTime: "2017-02-28T19:39:28.104Z",
		Data:        "eyJ0YXNrX2lkIjoiMzQ5ZjBkODQ5MjI3Y2QxMCIsInVzZXJkYXRhIjoiQ0lDQWdJQ0E2TjBLRWdkaFltTmxaR1puR2hoSVpXeHNiMTlWWW5WdWRIVXhOQzR3TkY5NE9EWXROalE9In0=",
	}
	msgBB = &pubsub.PubsubMessage{
		MessageId:   "58708071417623",
		PublishTime: "2017-02-28T19:39:28.104Z",
		Data:        "eyJidWlsZCI6eyJpZCI6IjEyMzQifSwidXNlcmRhdGEiOiJDSUNBZ0lDQTZOMEtFZ2RoWW1ObFpHWm5HaGhJWld4c2IxOVZZblZ1ZEhVeE5DNHdORjk0T0RZdE5qUT0ifQ==",
	}
	msgBBV2 = &pubsub.PubsubMessage{
		MessageId:   "58708071417624",
		PublishTime: "2023-01-30T19:39:28.104Z",
		Data:        "eyJidWlsZFB1YnN1YiI6eyJidWlsZCI6eyJpZCI6IjEyMzQiLCAiYnVpbGRlciI6eyJwcm9qZWN0IjoicHJvamVjdCIsICJidWNrZXQiOiJidWNrZXQiLCAiYnVpbGRlciI6ImJ1aWxkZXIifX0sICJidWlsZExhcmdlRmllbGRzIjoiZUp5cVltaGlaQUFFQUFELy93UHZBUDQ9In0sICJ1c2VyRGF0YSI6IlEwbERRV2RKUTBFMlRqQkxSV2RrYUZsdFRteGFSMXB1UjJob1NWcFhlSE5pTVRsV1dXNVdkV1JJVlhoT1F6UjNUa1k1TkU5RVdYUk9hbEU5In0=",
		Attributes:  map[string]string{"version": "v2"},
	}
	buildID       = 1234             // matches the above pubsub message
	runID   int64 = 6042091272536064 // matches the above pubsub messages
)

func TestDecodePubsubMessage(t *testing.T) {
	Convey("Test Environment", t, func() {
		ctx := triciumtest.Context()
		Convey("Decodes pubsub message without error", func() {
			_, _, err := decodePubsubMessage(ctx, msg)
			So(err, ShouldBeNil)
		})
	})
}

func TestEnqueueCollectRequest(t *testing.T) {
	Convey("Test Environment", t, func() {
		ctx := triciumtest.Context()

		Convey("Enqueued task shouldn't start until after delay time is up", func() {
			So(enqueueCollectRequest(ctx, &admin.CollectRequest{}, 7*time.Minute), ShouldBeNil)

			task := tq.GetTestable(ctx).GetScheduledTasks()[common.DriverQueue]["5023444679101355902"]
			// ETA is the earliest time that the task should execute; an ETA of now
			// + delay means that the task should start after a delay. When ETA is set
			// on the task, then Delay is unset.
			So(task.ETA, ShouldEqual, clock.Now(ctx).Add(7*time.Minute))
			So(task.Delay, ShouldEqual, time.Duration(0))
		})
	})
}

func TestHandlePubSubMessage(t *testing.T) {
	Convey("Test Environment", t, func() {
		ctx := triciumtest.Context()

		Convey("Enqueues buildbucket collect task", func() {
			Convey("Buildbucket old message", func() {
				So(len(tq.GetTestable(ctx).GetScheduledTasks()[common.DriverQueue]), ShouldEqual, 0)
				received := &ReceivedPubSubMessage{ID: fmt.Sprintf("%d:%d", buildID, runID)}
				So(ds.Get(ctx, received), ShouldEqual, ds.ErrNoSuchEntity)
				err := handlePubSubMessage(ctx, msgBB)
				So(err, ShouldBeNil)
				So(ds.Get(ctx, received), ShouldBeNil)
				So(len(tq.GetTestable(ctx).GetScheduledTasks()[common.DriverQueue]), ShouldEqual, 1)
			})
			Convey("Buildbucket new message", func() {
				So(len(tq.GetTestable(ctx).GetScheduledTasks()[common.DriverQueue]), ShouldEqual, 0)
				received := &ReceivedPubSubMessage{ID: fmt.Sprintf("%d:%d", buildID, runID)}
				So(ds.Get(ctx, received), ShouldEqual, ds.ErrNoSuchEntity)
				err := handlePubSubMessage(ctx, msgBBV2)
				So(err, ShouldBeNil)
				So(ds.Get(ctx, received), ShouldBeNil)
				So(len(tq.GetTestable(ctx).GetScheduledTasks()[common.DriverQueue]), ShouldEqual, 1)
			})
		})

		Convey("Avoids duplicate processing", func() {
			Convey("Buildbucket old message", func() {
				So(handlePubSubMessage(ctx, msgBB), ShouldBeNil)
				So(handlePubSubMessage(ctx, msgBB), ShouldBeNil)
				So(len(tq.GetTestable(ctx).GetScheduledTasks()[common.DriverQueue]), ShouldEqual, 1)
			})
			Convey("Buildbucket new message", func() {
				So(handlePubSubMessage(ctx, msgBBV2), ShouldBeNil)
				So(handlePubSubMessage(ctx, msgBBV2), ShouldBeNil)
				So(len(tq.GetTestable(ctx).GetScheduledTasks()[common.DriverQueue]), ShouldEqual, 1)
			})
		})
	})
}
