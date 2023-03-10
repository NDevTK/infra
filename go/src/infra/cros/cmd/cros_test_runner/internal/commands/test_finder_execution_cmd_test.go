// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands_test

import (
	"context"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/containers"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/executors"
	"infra/cros/cmd/cros_test_runner/internal/tools/crostoolrunner"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
)

func TestTestFinderExecutionCmd_UnsupportedSK(t *testing.T) {
	t.Parallel()
	Convey("Unsupported state keeper", t, func() {
		ctx := context.Background()
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosTestFinderTemplatedContainer("container/image/path", ctr)
		exec := executors.NewCrosTestFinderExecutor(cont)
		cmd := commands.NewTestFinderExecutionCmd(exec)
		sk := &UnsupportedStateKeeper{}
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestTestFinderExecutionCmd_MissingDeps(t *testing.T) {
	t.Parallel()
	Convey("Cmd missing deps", t, func() {
		ctx := context.Background()
		sk := &data.LocalTestStateKeeper{}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosTestFinderTemplatedContainer("container/image/path", ctr)
		exec := executors.NewCrosTestFinderExecutor(cont)
		cmd := commands.NewTestFinderExecutionCmd(exec)
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)
	})
}

func TestTestFinderExecutionCmd_UpdateSK(t *testing.T) {
	t.Parallel()
	Convey("Cmd with no updates", t, func() {
		ctx := context.Background()
		sk := &data.LocalTestStateKeeper{}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosTestFinderTemplatedContainer("container/image/path", ctr)
		exec := executors.NewCrosTestFinderExecutor(cont)
		cmd := commands.NewTestFinderExecutionCmd(exec)
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
	})
}

func TestTestFinderExecutionCmd_ExtractDepsSuccess(t *testing.T) {
	t.Parallel()

	Convey("TestFinderExecutionCmd extract deps", t, func() {
		ctx := context.Background()
		sk := &data.LocalTestStateKeeper{
			Tests:       []string{"test1"},
			Tags:        []string{"group:test"},
			TagsExclude: []string{"group:nottest"},
		}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosTestFinderTemplatedContainer("container/image/path", ctr)
		exec := executors.NewCrosTestFinderExecutor(cont)
		cmd := commands.NewTestFinderExecutionCmd(exec)

		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)
		So(cmd.Tests, ShouldResemble, sk.Tests)
		So(cmd.Tags, ShouldResemble, sk.Tags)
		So(cmd.TagsExclude, ShouldResemble, sk.TagsExclude)
	})
}

func TestTestFinderExecutionCmd_UpdateSKSuccess(t *testing.T) {
	t.Parallel()
	Convey("TestFinderExecutionCmd update SK", t, func() {
		ctx := context.Background()
		sk := &data.LocalTestStateKeeper{HwTestStateKeeper: data.HwTestStateKeeper{CftTestRequest: &skylab_test_runner.CFTTestRequest{}}}
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		cont := containers.NewCrosTestFinderTemplatedContainer("container/image/path", ctr)
		exec := executors.NewCrosTestFinderExecutor(cont)
		cmd := commands.NewTestFinderExecutionCmd(exec)

		cmd.TestSuites = []*api.TestSuite{
			{
				Spec: &api.TestSuite_TestCases{
					TestCases: &api.TestCaseList{
						TestCases: []*api.TestCase{
							{
								Id: &api.TestCase_Id{
									Value: "test1",
								},
							},
							{
								Id: &api.TestCase_Id{
									Value: "test2",
								},
							},
						},
					},
				},
			},
		}

		expectedResponse := []*api.TestSuite{
			{
				Spec: &api.TestSuite_TestCaseIds{
					TestCaseIds: &api.TestCaseIdList{
						TestCaseIds: []*api.TestCase_Id{
							{
								Value: "test1",
							},
							{
								Value: "test2",
							},
						},
					},
				},
			},
		}

		// Update SK
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
		So(sk.CftTestRequest.TestSuites, ShouldResemble, expectedResponse)
	})
}
