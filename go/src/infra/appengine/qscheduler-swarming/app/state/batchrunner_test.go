// Copyright 2019 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package state_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/appengine/gaetesting"
	"go.chromium.org/luci/common/clock/testclock"

	"infra/appengine/qscheduler-swarming/app/eventlog"
	"infra/appengine/qscheduler-swarming/app/state"
	"infra/appengine/qscheduler-swarming/app/state/nodestore"
	"infra/qscheduler/qslib/tutils"
	"infra/swarming"
)

func TestBatcherCancellations(t *testing.T) {
	Convey("Given a testing context with a scheduler pool, and a batcher for that pool", t, func() {
		ctx := gaetesting.TestingContext()
		ctx, _ = testclock.UseTime(ctx, time.Now())
		ctx = eventlog.Use(ctx, &eventlog.NullBQInserter{})
		poolID := "pool 1"
		store := nodestore.New(poolID)
		store.Create(ctx, time.Now())

		batcher := state.NewBatchRunnerForTest()
		batcher.Start(store)
		defer batcher.Close()

		Convey("with a leader request and a bunch of followers", func() {
			leaderctx, leadercancel := context.WithCancel(ctx)
			defer leadercancel()
			nFollowers := 5
			errs := make([]error, nFollowers+1)

			leaderWait := sync.WaitGroup{}
			leaderWait.Add(1)

			go func() {
				_, err := batcher.Notify(leaderctx, &swarming.NotifyTasksRequest{})
				errs[0] = err
				leaderWait.Done()
			}()
			batcher.TBatchWait(1)

			followerWait := sync.WaitGroup{}
			followerWait.Add(nFollowers)

			followerctx, followercancel := context.WithCancel(ctx)
			defer followercancel()
			addFollowers := func() {
				for i := 0; i < nFollowers; i++ {
					go func(i int) {
						_, err := batcher.Notify(followerctx, &swarming.NotifyTasksRequest{})
						errs[i+1] = err
						followerWait.Done()
					}(i)
				}
			}

			Convey("when the leader context is cancelled, the whole batch unwinds.", func() {
				addFollowers()
				batcher.TBatchWait(nFollowers)
				leadercancel()
				leaderWait.Wait()
				followerWait.Wait()
				for _, err := range errs {
					So(err, ShouldEqual, context.Canceled)
				}
			})

			Convey("when the follower contexts are cancelled prior to their being batched, only they are unwound.", func() {
				followercancel()
				batcher.TBatchStart()
				addFollowers()
				leaderWait.Wait()
				followerWait.Wait()
				So(errs[0], ShouldBeNil)

				for _, err := range errs[1:] {
					So(err, ShouldEqual, context.Canceled)
				}
			})
		})
	})
}

func TestBatcherBehavior(t *testing.T) {
	Convey("Given a testing context with a scheduler pool, and a batcher for that pool", t, func() {
		ctx := gaetesting.TestingContext()
		ctx, _ = testclock.UseTime(ctx, time.Now())
		ctx = eventlog.Use(ctx, &eventlog.NullBQInserter{})
		poolID := "pool 1"
		store := nodestore.New(poolID)
		store.Create(ctx, time.Now())

		batcher := state.NewBatchRunnerForTest()
		batcher.Start(store)
		defer batcher.Close()

		Convey("a batch of requests can run, with notifications coming before assignments.", func() {
			nTasks := 5
			labels := make([]string, nTasks)
			// Give each bot-task pair a unique dimension.
			for i := range labels {
				labels[i] = uuid.New().String()
			}
			assignements := make([]*swarming.AssignTasksResponse, nTasks)
			now := tutils.TimestampProto(time.Now())

			wg := sync.WaitGroup{}
			for i := 0; i < nTasks; i++ {
				wg.Add(2)
				// Run nTasks assignment requests concurrently.
				go func(i int) {
					req := &swarming.AssignTasksRequest{
						IdleBots: []*swarming.IdleBot{
							{
								BotId:      fmt.Sprintf("%d", i),
								Dimensions: []string{labels[i]},
							},
						},
						Time: now,
					}
					resp, err := batcher.Assign(ctx, req)
					if err != nil {
						panic(err)
					}
					assignements[i] = resp
					wg.Done()
				}(i)
				// Also run nTasks task notifications concurrently.
				go func(i int) {
					req := &swarming.NotifyTasksRequest{
						Notifications: []*swarming.NotifyTasksItem{
							{
								Task: &swarming.TaskSpec{
									EnqueuedTime: now,
									Id:           fmt.Sprintf("%d", i),
									State:        swarming.TaskState_PENDING,
									Slices: []*swarming.SliceSpec{
										{
											Dimensions: []string{labels[i]},
										},
									},
								},
								Time: now,
							},
						},
					}
					resp, err := batcher.Notify(ctx, req)
					if err != nil {
						panic(err)
					}
					if resp == nil {
						panic("unexpectedly nil response")
					}
					wg.Done()
				}(i)
			}
			batcher.TBatchWait(2 * nTasks)
			batcher.TBatchStart()
			wg.Wait()
			// All tasks should be assigned to their corresponding idle bot.
			for _, a := range assignements {
				So(a.Assignments, ShouldHaveLength, 1)
				So(a.Assignments[0].BotId, ShouldEqual, a.Assignments[0].TaskId)
			}
		})
	})
}
