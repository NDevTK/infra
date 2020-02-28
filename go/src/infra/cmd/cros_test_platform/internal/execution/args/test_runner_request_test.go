// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package args contains the logic for assembling all data required for
// creating an individual task request.
package args

import (
	"context"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"

	. "github.com/smartystreets/goconvey/convey"

	build_api "go.chromium.org/chromiumos/infra/proto/go/chromite/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/steps"
)

var noDeadline time.Time

func TestProvisionableLabels(t *testing.T) {
	Convey("Given a test that specifies software dependencies", t, func() {
		ctx := context.Background()
		params := &test_platform.Request_Params{
			SoftwareDependencies: []*test_platform.Request_Params_SoftwareDependency{
				{
					Dep: &test_platform.Request_Params_SoftwareDependency_ChromeosBuild{ChromeosBuild: "foo-build"},
				},
				{
					Dep: &test_platform.Request_Params_SoftwareDependency_RoFirmwareBuild{RoFirmwareBuild: "foo-ro-firmware"},
				},
				{
					Dep: &test_platform.Request_Params_SoftwareDependency_RwFirmwareBuild{RwFirmwareBuild: "foo-rw-firmware"},
				},
			},
		}

		Convey("when generating a test runner request", func() {
			g := NewGenerator(basicInvocation(), params, nil, "", noDeadline)
			got, err := g.testRunnerRequest(ctx)
			So(err, ShouldBeNil)
			Convey("the provisionable labels match the software dependencies", func() {
				So(got.Prejob, ShouldNotBeNil)
				So(got.Prejob.ProvisionableLabels, ShouldNotBeNil)
				So(got.Prejob.ProvisionableLabels["cros-version"], ShouldEqual, "foo-build")
				So(got.Prejob.ProvisionableLabels["fwro-version"], ShouldEqual, "foo-ro-firmware")
				So(got.Prejob.ProvisionableLabels["fwrw-version"], ShouldEqual, "foo-rw-firmware")
			})
		})
	})
}

func TestTestEnvironment(t *testing.T) {
	Convey("Given a", t, func() {
		ctx := context.Background()

		cases := []struct {
			description          string
			environment          build_api.AutotestTest_ExecutionEnvironment
			expectedIsClientTest bool
		}{
			{
				description:          "client test",
				environment:          build_api.AutotestTest_EXECUTION_ENVIRONMENT_CLIENT,
				expectedIsClientTest: true,
			},
			{
				description:          "server test",
				environment:          build_api.AutotestTest_EXECUTION_ENVIRONMENT_SERVER,
				expectedIsClientTest: false,
			},
		}
		for _, c := range cases {
			Convey(c.description, func() {
				invocation := &steps.EnumerationResponse_AutotestInvocation{
					Test: &build_api.AutotestTest{
						Name:                 "foo-test",
						ExecutionEnvironment: c.environment,
					},
				}
				Convey("when generating a test runner request", func() {
					g := NewGenerator(invocation, &test_platform.Request_Params{}, nil, "", noDeadline)
					got, err := g.testRunnerRequest(ctx)
					So(err, ShouldBeNil)
					Convey("the test field is populated correctly.", func() {
						So(got.Test, ShouldNotBeNil)
						So(got.Test.GetAutotest(), ShouldNotBeNil)
						So(got.Test.GetAutotest().Name, ShouldEqual, "foo-test")
						So(got.Test.GetAutotest().IsClientTest, ShouldEqual, c.expectedIsClientTest)
						So(got.Test.GetAutotest().DisplayName, ShouldEqual, "foo-test")
					})
				})
			})
		}
	})
}

func TestTestArgs(t *testing.T) {
	Convey("Given a request that specifies test args", t, func() {
		ctx := context.Background()
		invocation := &steps.EnumerationResponse_AutotestInvocation{
			Test: &build_api.AutotestTest{
				ExecutionEnvironment: build_api.AutotestTest_EXECUTION_ENVIRONMENT_CLIENT,
			},
			TestArgs: "foo=bar baz=qux",
		}
		Convey("when generating a test runner request", func() {
			g := NewGenerator(invocation, &test_platform.Request_Params{}, nil, "", noDeadline)
			got, err := g.testRunnerRequest(ctx)
			So(err, ShouldBeNil)
			Convey("the test args are propagated correctly.", func() {
				So(got.Test, ShouldNotBeNil)
				So(got.Test.GetAutotest(), ShouldNotBeNil)
				So(got.Test.GetAutotest().TestArgs, ShouldEqual, "foo=bar baz=qux")
			})
		})
	})
}

func TestKeyvals(t *testing.T) {
	Convey("Given a request that specifies test args", t, func() {
		ctx := context.Background()
		invocation := &steps.EnumerationResponse_AutotestInvocation{
			Test: &build_api.AutotestTest{
				ExecutionEnvironment: build_api.AutotestTest_EXECUTION_ENVIRONMENT_CLIENT,
			},
			ResultKeyvals: map[string]string{
				"test-level-key": "test-value",
				"ambiguous-key":  "test-value",
			},
			DisplayName: "fancy-name",
		}
		params := &test_platform.Request_Params{
			Decorations: &test_platform.Request_Params_Decorations{
				AutotestKeyvals: map[string]string{
					"request-level-key": "request-value",
					"ambiguous-key":     "request-value",
				},
			},
			SoftwareDependencies: []*test_platform.Request_Params_SoftwareDependency{
				{
					Dep: &test_platform.Request_Params_SoftwareDependency_ChromeosBuild{ChromeosBuild: "foo-build"},
				},
			},
		}
		want := map[string]string{
			"ambiguous-key":     "request-value",
			"build":             "foo-build",
			"label":             "fancy-name",
			"parent_job_id":     "foo-ID",
			"request-level-key": "request-value",
			"test-level-key":    "test-value",
		}
		Convey("when generating a test runner request", func() {
			g := NewGenerator(invocation, params, nil, "foo-ID", noDeadline)
			got, err := g.testRunnerRequest(ctx)
			So(err, ShouldBeNil)
			Convey("the test args are propagated correctly.", func() {
				So(got.Test, ShouldNotBeNil)
				So(got.Test.GetAutotest(), ShouldNotBeNil)
				So(got.Test.GetAutotest().Keyvals, ShouldResemble, want)
			})
		})
	})
}

func TestConstructedDisplayName(t *testing.T) {
	Convey("Given a request does not specify a display name", t, func() {
		ctx := context.Background()
		invocation := &steps.EnumerationResponse_AutotestInvocation{
			Test: &build_api.AutotestTest{
				Name:                 "foo-name",
				ExecutionEnvironment: build_api.AutotestTest_EXECUTION_ENVIRONMENT_CLIENT,
			},
		}
		params := &test_platform.Request_Params{
			SoftwareDependencies: []*test_platform.Request_Params_SoftwareDependency{
				{
					Dep: &test_platform.Request_Params_SoftwareDependency_ChromeosBuild{ChromeosBuild: "foo-build"},
				},
			},
			Decorations: &test_platform.Request_Params_Decorations{
				AutotestKeyvals: map[string]string{
					"suite": "foo-suite",
				},
			},
		}
		want := "foo-build/foo-suite/foo-name"
		Convey("when generating a test runner request", func() {
			g := NewGenerator(invocation, params, nil, "", noDeadline)
			got, err := g.testRunnerRequest(ctx)
			So(err, ShouldBeNil)
			Convey("the display name is generated correctly.", func() {
				So(got.Test, ShouldNotBeNil)
				So(got.Test.GetAutotest(), ShouldNotBeNil)
				So(got.Test.GetAutotest().DisplayName, ShouldEqual, want)
				So(got.Test.GetAutotest().Keyvals, ShouldNotBeNil)
				So(got.Test.GetAutotest().Keyvals["label"], ShouldEqual, want)
			})
		})
	})
}

func TestDeadline(t *testing.T) {
	Convey("Given a request that specifies a deadline", t, func() {
		ctx := context.Background()
		ts, _ := time.Parse(time.RFC3339, "2020-02-27T12:47:42Z")
		Convey("when generating a test runner request", func() {
			g := NewGenerator(basicInvocation(), &test_platform.Request_Params{}, nil, "", ts)
			got, err := g.testRunnerRequest(ctx)
			So(err, ShouldBeNil)
			Convey("the deadline is set correctly.", func() {
				So(ptypes.TimestampString(got.Deadline), ShouldEqual, "2020-02-27T12:47:42Z")
			})
		})
	})
}

func TestNoDeadline(t *testing.T) {
	Convey("Given a request that does not specify a deadline", t, func() {
		ctx := context.Background()
		Convey("when generating a test runner request", func() {
			g := NewGenerator(basicInvocation(), &test_platform.Request_Params{}, nil, "", noDeadline)
			got, err := g.testRunnerRequest(ctx)
			So(err, ShouldBeNil)
			Convey("the deadline should not be set.", func() {
				So(got.Deadline, ShouldBeNil)
			})
		})
	})
}

func basicInvocation() *steps.EnumerationResponse_AutotestInvocation {
	return &steps.EnumerationResponse_AutotestInvocation{
		Test: &build_api.AutotestTest{
			ExecutionEnvironment: build_api.AutotestTest_EXECUTION_ENVIRONMENT_CLIENT,
		},
	}
}
