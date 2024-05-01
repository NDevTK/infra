// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package updaters_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/chromiumos/config/go/test/api"

	. "infra/cros/cmd/common_lib/dynamic_updates/updaters"
)

func TestInsertHelpers(t *testing.T) {
	Convey("InsertLeft", t, func() {
		taskList := []*api.CrosTestRunnerDynamicRequest_Task{
			getGenericTaskWithId("1"),
			getGenericTaskWithId("3"),
			getGenericTaskWithId("5"),
			getGenericTaskWithId("7"),
		}

		InsertTaskLeft(&taskList, getGenericTaskWithId("0"), 0)
		InsertTaskLeft(&taskList, getGenericTaskWithId("2"), 2)
		InsertTaskLeft(&taskList, getGenericTaskWithId("4"), 4)
		InsertTaskLeft(&taskList, getGenericTaskWithId("6"), 6)

		So(taskList, ShouldHaveLength, 8)
		taskHasId(taskList[0], "0")
		taskHasId(taskList[1], "1")
		taskHasId(taskList[2], "2")
		taskHasId(taskList[3], "3")
		taskHasId(taskList[4], "4")
		taskHasId(taskList[5], "5")
		taskHasId(taskList[6], "6")
		taskHasId(taskList[7], "7")
	})

	Convey("InsertRight", t, func() {
		taskList := []*api.CrosTestRunnerDynamicRequest_Task{
			getGenericTaskWithId("0"),
			getGenericTaskWithId("2"),
			getGenericTaskWithId("4"),
			getGenericTaskWithId("6"),
		}

		InsertTaskRight(&taskList, getGenericTaskWithId("1"), 0)
		InsertTaskRight(&taskList, getGenericTaskWithId("3"), 2)
		InsertTaskRight(&taskList, getGenericTaskWithId("5"), 4)
		InsertTaskRight(&taskList, getGenericTaskWithId("7"), 6)

		So(taskList, ShouldHaveLength, 8)
		taskHasId(taskList[0], "0")
		taskHasId(taskList[1], "1")
		taskHasId(taskList[2], "2")
		taskHasId(taskList[3], "3")
		taskHasId(taskList[4], "4")
		taskHasId(taskList[5], "5")
		taskHasId(taskList[6], "6")
		taskHasId(taskList[7], "7")
	})

	Convey("InsertLeftEmpty", t, func() {
		taskList := []*api.CrosTestRunnerDynamicRequest_Task{}

		InsertTaskLeft(&taskList, getGenericTaskWithId("0"), 0)

		So(taskList, ShouldHaveLength, 1)
		taskHasId(taskList[0], "0")
	})

	Convey("InsertRightEmpty", t, func() {
		taskList := []*api.CrosTestRunnerDynamicRequest_Task{}

		InsertTaskRight(&taskList, getGenericTaskWithId("0"), 0)

		So(taskList, ShouldHaveLength, 1)
		taskHasId(taskList[0], "0")
	})
}

func getGenericTaskWithId(id string) *api.CrosTestRunnerDynamicRequest_Task {
	return &api.CrosTestRunnerDynamicRequest_Task{
		Task: &api.CrosTestRunnerDynamicRequest_Task_Generic{
			Generic: &api.GenericTask{
				DynamicIdentifier: id,
			},
		},
	}
}

func taskHasId(task *api.CrosTestRunnerDynamicRequest_Task, id string) {
	So(task.GetGeneric().DynamicIdentifier, ShouldEqual, id)
}
