// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/auth/authtest"

	api "infra/unifiedfleet/api/v1/rpc"
	"infra/unifiedfleet/app/model/configuration"
)

func TestGetPublicChromiumTestStatus(t *testing.T) {
	t.Parallel()
	ctx := auth.WithState(testingContext(), &authtest.FakeState{
		Identity:       "user:abc@def.com",
		IdentityGroups: []string{"public-chromium-in-chromeos-builders"},
	})
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()
	configuration.AddPublicBoardModelData(ctx, "eve", []string{"eve"}, false)
	Convey("Check Fleet Policy For Tests", t, func() {
		Convey("happy path", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName:  "tast.lacros",
				Board:     "eve",
				Model:     "eve",
				Image:     "eve-public/R105-14988.0.0",
				QsAccount: "chromium",
			}

			res, err := tf.Fleet.CheckFleetTestsPolicy(ctx, req)
			So(err, ShouldBeNil)
			So(res.TestStatus.Code, ShouldEqual, api.TestStatus_OK)
		})
		Convey("Private board", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Board:    "private",
				Model:    "eve",
				Image:    "R100-14495.0.0-rc1",
			}

			res, err := tf.Fleet.CheckFleetTestsPolicy(ctx, req)
			So(err, ShouldBeNil)
			So(res.TestStatus.Code, ShouldEqual, api.TestStatus_NOT_A_PUBLIC_BOARD)
		})
		Convey("Private model", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Board:    "eve",
				Model:    "private",
				Image:    "R100-14495.0.0-rc1",
			}

			res, err := tf.Fleet.CheckFleetTestsPolicy(ctx, req)
			So(err, ShouldBeNil)
			So(res.TestStatus.Code, ShouldEqual, api.TestStatus_NOT_A_PUBLIC_MODEL)
		})
		Convey("Non allowlisted image", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Board:    "eve",
				Model:    "eve",
				Image:    "invalid",
			}

			res, err := tf.Fleet.CheckFleetTestsPolicy(ctx, req)
			So(err, ShouldBeNil)
			So(res.TestStatus.Code, ShouldEqual, api.TestStatus_NOT_A_PUBLIC_IMAGE)
		})
		Convey("Private test name and public auth group member", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "private",
				Board:    "eve",
				Model:    "eve",
				Image:    "R100-14495.0.0-rc1",
			}

			res, err := tf.Fleet.CheckFleetTestsPolicy(ctx, req)
			So(err, ShouldBeNil)
			So(res.TestStatus.Code, ShouldEqual, api.TestStatus_NOT_A_PUBLIC_TEST)
		})
		Convey("Public test name and not a public auth group member", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Board:    "eve",
				Model:    "eve",
				Image:    "R100-14495.0.0-rc1",
			}
			ctx := auth.WithState(testingContext(), &authtest.FakeState{
				Identity: "user:abc@def.com",
			})

			res, err := tf.Fleet.CheckFleetTestsPolicy(ctx, req)
			So(err, ShouldBeNil)
			So(res.TestStatus.Code, ShouldEqual, api.TestStatus_OK)
		})
		Convey("Private test name and not a public auth group member", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "private",
				Board:    "eve",
				Model:    "eve",
				Image:    "R100-14495.0.0-rc1",
			}
			ctx := auth.WithState(testingContext(), &authtest.FakeState{
				Identity: "user:abc@def.com",
			})

			res, err := tf.Fleet.CheckFleetTestsPolicy(ctx, req)
			So(err, ShouldBeNil)
			So(res.TestStatus.Code, ShouldEqual, api.TestStatus_OK)
		})
		Convey("Missing Test names", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "",
				Board:    "eve",
				Model:    "eve",
				Image:    "R100-14495.0.0-rc1",
			}

			_, err := tf.Fleet.CheckFleetTestsPolicy(ctx, req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Test name cannot be empty")
		})
		Convey("Missing Board", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Model:    "eve",
				Image:    "R100-14495.0.0-rc1",
			}

			_, err := tf.Fleet.CheckFleetTestsPolicy(ctx, req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Board cannot be empty")
		})
		Convey("Missing Models", func() {
			configuration.AddPublicBoardModelData(ctx, "fakeBoard", []string{"fakeModel"}, true)
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Board:    "fakeBoard",
				Image:    "R100-14495.0.0-rc1",
			}

			_, err := tf.Fleet.CheckFleetTestsPolicy(ctx, req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Model cannot be empty as the specified board has unlaunched models")
		})
		Convey("Missing Models - succeeds for boards with only public models", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Board:    "eve",
				Image:    "R100-14495.0.0-rc1",
			}

			_, err := tf.Fleet.CheckFleetTestsPolicy(ctx, req)
			So(err, ShouldBeNil)
		})
		Convey("Missing Image", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Board:    "eve",
				Model:    "eve",
			}

			_, err := tf.Fleet.CheckFleetTestsPolicy(ctx, req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Image cannot be empty")
		})
		Convey("Invalid QsAccount", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName:  "tast.lacros",
				Board:     "eve",
				Model:     "eve",
				Image:     "eve-public/R105-14988.0.0",
				QsAccount: "invalid",
			}

			res, err := tf.Fleet.CheckFleetTestsPolicy(ctx, req)
			So(err, ShouldBeNil)
			So(res.TestStatus.Code, ShouldEqual, api.TestStatus_INVALID_QS_ACCOUNT)
		})
	})
}
