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
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/appengine/gaetesting"

	"infra/appengine/qscheduler-swarming/app/eventlog"
	"infra/appengine/qscheduler-swarming/app/state"
	"infra/appengine/qscheduler-swarming/app/state/nodestore"
	"infra/qscheduler/qslib/tutils"
	swarming "infra/swarming"
)

/*func TestBatcherErrors(t *testing.T) {
	Convey("Given a testing context with a scheduler pool, and a batcher for that pool", t, func() {
		ctx := gaetesting.TestingContext()
		ctx = eventlog.Use(ctx, &eventlog.NullBQInserter{})
		poolID := "pool 1"
		store := nodestore.New(poolID)
		store.Create(ctx, time.Now())

		batcher := state.NewBatchRunnerForTest()
		batcher.Start(store)
		defer batcher.Close()

		Convey("an error in one operation should only affect that operation.", func() {
			var goodError error
			goodOperation := func(ctx context.Context, s *types.QScheduler, m scheduler.EventSink) error {
				return nil
			}

			var badError error
			badOperation := func(ctx context.Context, s *types.QScheduler, m scheduler.EventSink) error {
				return errors.New("a bad error occurred")
			}

			wg := sync.WaitGroup{}
			wg.Add(2)

			go func() {
				done := batcher.EnqueueOperation(ctx, goodOperation, 0)
				goodError = <-done
				wg.Done()
			}()

			go func() {
				done := batcher.EnqueueOperation(ctx, badOperation, 0)
				badError = <-done
				wg.Done()
			}()

			batcher.TBatchStart()
			batcher.TBatchWait(2)
			wg.Wait()

			So(badError, ShouldNotBeNil)
			So(badError.Error(), ShouldEqual, "a bad error occurred")

			So(goodError, ShouldBeNil)
		})
	})
}*/

func TestBatcherBehavior(t *testing.T) {
	Convey("Given a testing context with a scheduler pool, and a batcher for that pool", t, func() {
		ctx := gaetesting.TestingContext()
		ctx = eventlog.Use(ctx, &eventlog.NullBQInserter{})
		poolID := "pool 1"
		store := nodestore.New(poolID)
		store.Create(ctx, time.Now())

		batcher := state.NewBatchRunnerForTest()
		batcher.Start(store)
		defer batcher.Close()

		Convey("a batch of requests can run, with notifications coming before assignments.", func() {
			nTasks := 10
			labels := make([]string, nTasks)
			for i := range labels {
				labels[i] = uuid.New().String()
			}
			assignements := make([]*swarming.AssignTasksResponse, nTasks)
			now := tutils.TimestampProto(time.Now())

			wg := sync.WaitGroup{}
			for i := 0; i < nTasks; i++ {
				wg.Add(2)
				// Run 10 assignment requests concurrently.
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
				// Also run 10 notifications concurrently.
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
			batcher.TBatchStart()
			batcher.TBatchWait(20)
			wg.Wait()
			for _, a := range assignements {
				So(a.Assignments, ShouldHaveLength, 1)
				So(a.Assignments[0].BotId, ShouldEqual, a.Assignments[0].TaskId)
			}
		})
	})
}
