// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common_commands_test

import (
	"container/list"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"

	"infra/cros/cmd/common_lib/common_commands"
)

func TestPopQueueInstantiation_BadCast(t *testing.T) {
	Convey("Bad Type Cast", t, func() {
		queue := list.New()
		queue.PushBack(&skylab_test_runner.TestRequest{
			DynamicDeps: []*skylab_test_runner.DynamicDep{
				{
					Key:   "key",
					Value: "value",
				},
			},
		})
		err := common_commands.Instantiate_PopFromQueue(queue, func(element any) {
			_ = element.(*skylab_test_runner.ProvisionRequest)
		})
		So(err, ShouldNotBeNil)
	})
}

func TestPopQueueInstantiation_GoodCast(t *testing.T) {
	Convey("Good Type Cast", t, func() {
		queue := list.New()
		queue.PushBack(&skylab_test_runner.TestRequest{
			DynamicDeps: []*skylab_test_runner.DynamicDep{
				{
					Key:   "key",
					Value: "value",
				},
			},
		})
		var testRequest *skylab_test_runner.TestRequest
		err := common_commands.Instantiate_PopFromQueue(queue, func(element any) {
			testRequest = element.(*skylab_test_runner.TestRequest)
		})
		So(err, ShouldBeNil)
		So(testRequest, ShouldNotBeNil)
		So(testRequest.DynamicDeps, ShouldHaveLength, 1)
		So(testRequest.DynamicDeps[0].Key, ShouldEqual, "key")
	})
}
