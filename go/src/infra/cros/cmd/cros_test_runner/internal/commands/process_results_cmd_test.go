// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands_test

import (
	"context"
	"infra/cros/cmd/cros_test_runner/data"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
)

func TestProcessResultsCmdDeps_UnsupportedSK(t *testing.T) {
	t.Parallel()
	Convey("Unsupported state keeper", t, func() {
		ctx := context.Background()
		sk := &UnsupportedStateKeeper{}
		cmd := commands.NewProcessResultsCmd()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldNotBeNil)
	})
}

func TestProcessResultsCmdDeps_MissingDeps(t *testing.T) {
	t.Parallel()
	Convey("Cmd missing deps", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{CftTestRequest: nil}
		cmd := commands.NewProcessResultsCmd()
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)
	})
}

func TestProcessResultsCmdDeps_UpdateSK(t *testing.T) {
	t.Parallel()
	Convey("Update SK", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{CftTestRequest: nil}
		cmd := commands.NewProcessResultsCmd()
		err := cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
	})
}

func TestProcessResultsCmdDeps_Execute(t *testing.T) {
	t.Parallel()
	Convey("BuildInputValidationCmd execute with passing values", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{
			CftTestRequest: &skylab_test_runner.CFTTestRequest{ParentBuildId: 12345678},
			GcsUrl:         "some/url",
			TesthausUrl:    "some/url",
			ProvisionResponses: map[string][]*api.InstallResponse{
				"primaryDevice": {
					{Status: api.InstallResponse_STATUS_SUCCESS},
				},
			},
			TestResponses: &api.CrosTestResponse{
				TestCaseResults: []*api.TestCaseResult{
					{
						TestCaseId: &api.TestCase_Id{Value: "testId"},
						Verdict:    &api.TestCaseResult_Pass_{},
					},
				},
			},
		}
		cmd := commands.NewProcessResultsCmd()

		// Extract deps first
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)

		// Execute cmd
		err = cmd.Execute(ctx)
		So(err, ShouldBeNil)

		// Update SK
		err = cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
		So(sk.SkylabResult, ShouldNotBeNil)
	})

	Convey("BuildInputValidationCmd execute with missing provision resp", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{
			CftTestRequest: &skylab_test_runner.CFTTestRequest{ParentBuildId: 12345678},
			GcsUrl:         "some/url",
			TesthausUrl:    "some/url",
			TestResponses: &api.CrosTestResponse{
				TestCaseResults: []*api.TestCaseResult{
					{
						TestCaseId: &api.TestCase_Id{Value: "testId"},
						Verdict:    &api.TestCaseResult_Pass_{},
					},
				},
			},
		}
		cmd := commands.NewProcessResultsCmd()

		// Extract deps first
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)

		// Execute cmd
		err = cmd.Execute(ctx)
		So(err, ShouldBeNil)

		// Update SK
		err = cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
		So(sk.SkylabResult, ShouldNotBeNil)
	})

	Convey("BuildInputValidationCmd execute with missing test results", t, func() {
		ctx := context.Background()
		sk := &data.HwTestStateKeeper{
			CftTestRequest: &skylab_test_runner.CFTTestRequest{
				ParentBuildId: 12345678,
				TestSuites: []*api.TestSuite{
					{
						Name: "suite",
						Spec: &api.TestSuite_TestCaseIds{
							TestCaseIds: &api.TestCaseIdList{
								TestCaseIds: []*api.TestCase_Id{
									{
										Value: "test1",
									},
								},
							},
						},
					},
				},
			},
			GcsUrl:      "some/url",
			TesthausUrl: "some/url",
			ProvisionResponses: map[string][]*api.InstallResponse{
				"primaryDevice": {
					{Status: api.InstallResponse_STATUS_SUCCESS},
				},
			},
		}
		cmd := commands.NewProcessResultsCmd()

		// Extract deps first
		err := cmd.ExtractDependencies(ctx, sk)
		So(err, ShouldBeNil)

		// Execute cmd
		err = cmd.Execute(ctx)
		So(err, ShouldBeNil)

		// Update SK
		err = cmd.UpdateStateKeeper(ctx, sk)
		So(err, ShouldBeNil)
		So(sk.SkylabResult, ShouldNotBeNil)
	})
}
