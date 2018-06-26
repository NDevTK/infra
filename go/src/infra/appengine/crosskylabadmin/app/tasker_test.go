// Copyright 2018 The LUCI Authors.
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

package app

import (
	"fmt"
	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/gae/service/datastore"
	"go.chromium.org/luci/appengine/gaetesting"
	"golang.org/x/net/context"
)

func TestEnsureBackgroundTasks(t *testing.T) {
	Convey("In testing context", t, FailureHalts, func() {
		c := gaetesting.TestingContextWithAppID("dev~infra-crosskylabadmin")
		datastore.GetTestable(c).Consistent(true)
		fsc := &fakeSwarmingClient{
			pool:    swarmingBotPool,
			taskIDs: map[*SwarmingCreateTaskArgs]string{},
		}
		server := taskerServerImpl{
			swarmingClientFactory{
				swarmingClientHook: func(context.Context, string) (SwarmingClient, error) {
					return fsc, nil
				},
			},
		}

		Convey("with 2 known bots", func() {
			setKnownBots(c, fsc, []string{"dut_1", "dut_2"})

			Reset(func() {
				fsc.ResetTasks()
			})

			Convey("EnsureBackgroundTasks for unknown bot", func() {
				resp, err := server.EnsureBackgroundTasks(c, &fleet.EnsureBackgroundTasksRequest{
					Type:      fleet.TaskType_Reset,
					Selectors: makeBotSelectorForDuts([]string{"dut_3"}),
				})
				Convey("succeeds", func() {
					So(err, ShouldBeNil)
				})
				Convey("returns empty bot tasks", func() {
					So(resp.BotTasks, ShouldBeEmpty)
				})
			})

			taskURLFirst := []string{}
			Convey("EnsureBackgroundTasks(Reset) for known bot", func() {
				resp, err := server.EnsureBackgroundTasks(c, &fleet.EnsureBackgroundTasksRequest{
					Type:      fleet.TaskType_Reset,
					Selectors: makeBotSelectorForDuts([]string{"dut_1"}),
					TaskCount: 5,
					Priority:  10,
				})

				Convey("succeeds", func() {
					So(err, ShouldBeNil)
				})
				Convey("creates expected swarming tasks", func() {
					So(fsc.taskArgs, ShouldHaveLength, 5)
					for _, ta := range fsc.taskArgs {
						So(ta.DutID, ShouldEqual, "dut_1")
						So(ta.DutState, ShouldEqual, "needs_reset")
						So(ta.Priority, ShouldEqual, 10)

						cmd := strings.Join(ta.Cmd, " ")
						So(cmd, ShouldContainSubstring, "-task-name admin_reset")
					}

				})
				Convey("returns bot list with requested tasks", func() {
					So(resp.BotTasks, ShouldHaveLength, 1)
					botTasks := resp.BotTasks[0]
					So(botTasks.DutId, ShouldEqual, "dut_1")
					So(botTasks.Tasks, ShouldHaveLength, 5)
					for _, t := range botTasks.Tasks {
						So(t.Type, ShouldEqual, fleet.TaskType_Reset)
						So(t.TaskUrl, ShouldNotBeNil)
						taskURLFirst = append(taskURLFirst, t.TaskUrl)
					}
				})

				Convey("then another EnsureBackgroundTasks(Reset) with more tasks requested", func() {
					resp, err := server.EnsureBackgroundTasks(c, &fleet.EnsureBackgroundTasksRequest{
						Type:      fleet.TaskType_Reset,
						Selectors: makeBotSelectorForDuts([]string{"dut_1"}),
						TaskCount: 7,
						Priority:  10,
					})

					Convey("succeeds", func() {
						So(err, ShouldBeNil)
					})
					Convey("creates remaining swarming tasks", func() {
						// This includes the 5 created earlier.
						So(fsc.taskArgs, ShouldHaveLength, 7)
					})
					Convey("returns bot list containing tasks created earlier and the new tasks", func() {
						So(resp.BotTasks, ShouldHaveLength, 1)
						botTasks := resp.BotTasks[0]
						So(botTasks.DutId, ShouldEqual, "dut_1")
						So(botTasks.Tasks, ShouldHaveLength, 7)
						taskURLSecond := []string{}
						for _, t := range botTasks.Tasks {
							taskURLSecond = append(taskURLSecond, t.TaskUrl)
						}
						for _, t := range taskURLFirst {
							So(t, ShouldBeIn, taskURLSecond)
						}
					})
				})
			})
		})

		Convey("with a large number of known bots", func() {
			numDuts := 6 * maxConcurrentSwarmingCalls
			allDuts := make([]string, 0, numDuts)
			taskDuts := make([]string, 0, numDuts/2)
			for i := 0; i < numDuts; i++ {
				allDuts = append(allDuts, fmt.Sprintf("dut_%d", i))
				if i%2 == 0 {
					taskDuts = append(taskDuts, allDuts[i])
				}
			}
			setKnownBots(c, fsc, allDuts)

			Convey("EnsureBackgroundTasks(Repair) for some of the known bots", func() {
				resp, err := server.EnsureBackgroundTasks(c, &fleet.EnsureBackgroundTasksRequest{
					Type:      fleet.TaskType_Repair,
					Selectors: makeBotSelectorForDuts(taskDuts),
					TaskCount: 6,
					Priority:  9,
				})

				Convey("succeeds", func() {
					So(err, ShouldBeNil)
				})
				Convey("creates expected swarming tasks", func() {
					So(fsc.taskArgs, ShouldHaveLength, 6*len(taskDuts))
					gotDuts := map[string]int{}
					for _, ta := range fsc.taskArgs {
						So(ta.DutState, ShouldEqual, "needs_repair")
						So(ta.Priority, ShouldEqual, 9)
						gotDuts[ta.DutID] = gotDuts[ta.DutID] + 1
						cmd := strings.Join(ta.Cmd, " ")
						So(cmd, ShouldContainSubstring, "-task-name admin_repair")
					}
					So(gotDuts, ShouldHaveLength, len(taskDuts))
					for d, c := range gotDuts {
						So(d, ShouldBeIn, taskDuts)
						So(c, ShouldEqual, 6)
					}
				})
				Convey("returns bot list with requested tasks", func() {
					So(resp.BotTasks, ShouldHaveLength, len(taskDuts))
					gotDuts := map[string]bool{}
					for _, bt := range resp.BotTasks {
						So(bt.Tasks, ShouldHaveLength, 6)
						for _, t := range bt.Tasks {
							So(t.Type, ShouldEqual, fleet.TaskType_Repair)
							So(t.TaskUrl, ShouldNotBeNil)
						}
						So(bt.DutId, ShouldBeIn, taskDuts)
						gotDuts[bt.DutId] = true
					}
					So(gotDuts, ShouldHaveLength, len(taskDuts))
				})
			})

		})
	})
}

func TestTaskerDummy(t *testing.T) {
	t.Parallel()
	Convey("In testing context", t, FailureHalts, func() {
		c := gaetesting.TestingContextWithAppID("dev~infra-crosskylabadmin")
		datastore.GetTestable(c).Consistent(true)
		fsc := &fakeSwarmingClient{
			pool:    swarmingBotPool,
			taskIDs: map[*SwarmingCreateTaskArgs]string{},
		}
		server := taskerServerImpl{
			swarmingClientFactory{
				swarmingClientHook: func(context.Context, string) (SwarmingClient, error) {
					return fsc, nil
				},
			},
		}

		Convey("TriggerRepairOnIdle returns internal error", func() {
			_, err := server.TriggerRepairOnIdle(c, nil)
			So(err, ShouldNotBeNil)
		})

		Convey("TriggerRepairOnRepairFailed returns internal error", func() {
			_, err := server.TriggerRepairOnRepairFailed(c, nil)
			So(err, ShouldNotBeNil)
		})
	})
}

func setKnownBots(c context.Context, fsc *fakeSwarmingClient, duts []string) {
	fsc.setAvailableDutIDs(duts)
	server := trackerServerImpl{
		swarmingClientFactory{
			swarmingClientHook: func(context.Context, string) (SwarmingClient, error) {
				return fsc, nil
			},
		},
	}
	resp, err := server.RefreshBots(c, &fleet.RefreshBotsRequest{})
	So(err, ShouldBeNil)
	So(resp.DutIds, ShouldHaveLength, len(duts))
}
