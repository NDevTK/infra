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

package clients

import (
	"sort"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/appengine/gaetesting"

	"infra/appengine/crosskylabadmin/internal/tq"
)

func TestSuccessfulPushDuts(t *testing.T) {
	Convey("success", t, func() {
		ctx := gaetesting.TestingContext()
		tqt := tq.GetTestable(ctx)
		qn := "repair-bots"
		tqt.CreateQueue(qn)
		hosts := []string{"host1", "host2"}
		swarmingPool := "pool-a"
		err := PushRepairDUTs(ctx, hosts, "needs_repair", swarmingPool)
		So(err, ShouldBeNil)
		tasks := tqt.GetScheduledTasks()
		t, ok := tasks[qn]
		So(ok, ShouldBeTrue)
		var taskPaths, taskParams []string
		for _, v := range t {
			taskPaths = append(taskPaths, v.Path)
			taskParams = append(taskParams, string(v.Payload))
		}
		sort.Strings(taskPaths)
		sort.Strings(taskParams)
		expectedPaths := []string{"/internal/task/cros_repair/host1", "/internal/task/cros_repair/host2"}
		expectedParams := []string{"botID=host1&expectedState=needs_repair&swarmingPool=pool-a", "botID=host2&expectedState=needs_repair&swarmingPool=pool-a"}
		So(taskPaths, ShouldResemble, expectedPaths)
		So(taskParams, ShouldResemble, expectedParams)
	})
}

func TestSuccessfulPushLabstations(t *testing.T) {
	Convey("success", t, func() {
		ctx := gaetesting.TestingContext()
		tqt := tq.GetTestable(ctx)
		qn := "repair-labstations"
		tqt.CreateQueue(qn)
		hosts := []string{"host1", "host2"}
		err := PushRepairLabstations(ctx, hosts)
		So(err, ShouldBeNil)
		tasks := tqt.GetScheduledTasks()
		t, ok := tasks[qn]
		So(ok, ShouldBeTrue)
		var taskPaths, taskParams []string
		for _, v := range t {
			taskPaths = append(taskPaths, v.Path)
			taskParams = append(taskParams, string(v.Payload))
		}
		sort.Strings(taskPaths)
		sort.Strings(taskParams)
		expectedPaths := []string{"/internal/task/labstation_repair/host1", "/internal/task/labstation_repair/host2"}
		expectedParams := []string{"botID=host1", "botID=host2"}
		So(taskPaths, ShouldResemble, expectedPaths)
		So(taskParams, ShouldResemble, expectedParams)
	})
}

func TestSuccessfulPushAuditTasks(t *testing.T) {
	Convey("success", t, func() {
		ctx := gaetesting.TestingContext()
		tqt := tq.GetTestable(ctx)
		qn := "audit-bots"
		tqt.CreateQueue(qn)
		hosts := []string{"host1", "host2"}
		actions := []string{"action1", "action2"}
		err := PushAuditDUTs(ctx, hosts, actions, "Storage")
		So(err, ShouldBeNil)
		tasks := tqt.GetScheduledTasks()
		t, ok := tasks[qn]
		So(ok, ShouldBeTrue)
		var taskPaths, taskParams []string
		for _, v := range t {
			taskPaths = append(taskPaths, v.Path)
			taskParams = append(taskParams, string(v.Payload))
		}
		sort.Strings(taskPaths)
		sort.Strings(taskParams)
		expectedPaths := []string{"/internal/task/audit/host1/action1-action2", "/internal/task/audit/host2/action1-action2"}
		expectedParams := []string{"actions=action1%2Caction2&botID=host1&taskname=Storage", "actions=action1%2Caction2&botID=host2&taskname=Storage"}
		So(taskPaths, ShouldResemble, expectedPaths)
		So(taskParams, ShouldResemble, expectedParams)
	})
}

func TestUnknownQueuePush(t *testing.T) {
	Convey("no taskqueue", t, func() {
		ctx := gaetesting.TestingContext()
		tqt := tq.GetTestable(ctx)
		tqt.CreateQueue("no-repair-bots")
		err := PushRepairDUTs(ctx, []string{"host1", "host2"}, "some_state", "some_builder")
		So(err, ShouldNotBeNil)
	})
}
