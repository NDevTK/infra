// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/auth/authtest"

	ufspb "infra/unifiedfleet/api/v1/models"
	api "infra/unifiedfleet/api/v1/rpc"
	"infra/unifiedfleet/app/model/configuration"
)

func TestIsValidPublicChromiumTest(t *testing.T) {
	t.Parallel()
	ctx := auth.WithState(testingContext(), &authtest.FakeState{
		Identity:       "user:abc@def.com",
		IdentityGroups: []string{"public-chromium-in-chromeos-builders"},
	})
	Convey("Is Valid Public Chromium Test", t, func() {
		Convey("happy path", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Board:    "eve",
				Model:    "eve",
				Image:    "eve-full/R100-14495.0.0-rc1",
			}

			err := IsValidTest(ctx, req)

			So(err, ShouldBeNil)
		})
		Convey("Private test name and public auth group member", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "private",
				Board:    "eve",
				Model:    "eve",
				Image:    "eve-full/R100-14495.0.0-rc1",
			}

			err := IsValidTest(ctx, req)
			err, ok := err.(*InvalidTestError)

			So(err, ShouldNotBeNil)
			So(ok, ShouldBeTrue)
		})
		Convey("Public test name and not a public auth group member", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Board:    "eve",
				Model:    "eve",
				Image:    "eve-full/R100-14495.0.0-rc1",
			}
			newCtx := auth.WithState(ctx, &authtest.FakeState{
				Identity: "user:abc@def.com",
			})

			err := IsValidTest(newCtx, req)

			So(err, ShouldBeNil)
		})
		Convey("Private test name and not a public auth group member", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "private",
				Board:    "eve",
				Model:    "eve",
				Image:    "eve-full/R100-14495.0.0-rc1",
			}
			ctx := auth.WithState(ctx, &authtest.FakeState{
				Identity: "user:abc@def.com",
			})

			err := IsValidTest(ctx, req)

			So(err, ShouldBeNil)
		})
		Convey("Public test and private board", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Board:    "private",
				Model:    "eve",
				Image:    "eve-full/R100-14495.0.0-rc1",
			}

			err := IsValidTest(ctx, req)
			err, ok := err.(*InvalidBoardError)

			So(err, ShouldNotBeNil)
			So(ok, ShouldBeTrue)
		})
		Convey("Public test and private model", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Board:    "eve",
				Model:    "private",
				Image:    "eve-full/R100-14495.0.0-rc1",
			}

			err := IsValidTest(ctx, req)
			err, ok := err.(*InvalidModelError)

			So(err, ShouldNotBeNil)
			So(ok, ShouldBeTrue)
		})
		Convey("Public test and incorrect image", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Board:    "eve",
				Model:    "eve",
				Image:    "eve-full/R100-14495.0.0-rc1-abcd",
			}

			err := IsValidTest(ctx, req)
			err, ok := err.(*InvalidImageError)

			So(err, ShouldNotBeNil)
			So(ok, ShouldBeTrue)
		})
		Convey("Missing Test names", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "",
				Board:    "eve",
				Model:    "eve",
				Image:    "eve-full/R100-14495.0.0-rc1",
			}

			err := IsValidTest(ctx, req)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Test name cannot be empty")
		})
		Convey("Missing Board", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Model:    "eve",
				Image:    "eve-full/R100-14495.0.0-rc1",
			}

			err := IsValidTest(ctx, req)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Board cannot be empty")
		})
		Convey("Missing Models", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Board:    "eve",
				Image:    "eve-full/R100-14495.0.0-rc1",
			}

			err := IsValidTest(ctx, req)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Model cannot be empty")
		})
		Convey("Missing Image", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Board:    "eve",
				Model:    "eve",
			}

			err := IsValidTest(ctx, req)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Image cannot be empty")
		})
	})
}

func TestImportPublicBoardsAndModels(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	Convey("Import Public Boards and Models", t, func() {
		Convey("Happy Path", func() {
			mockDevice := mockDevicesLaunched()
			err := ImportPublicBoardsAndModels(ctx, mockDevice)
			So(err, ShouldBeNil)
		})
		Convey("Happy Path Check DataStore", func() {
			mockDevice := mockDevicesLaunched()
			err := ImportPublicBoardsAndModels(ctx, mockDevice)
			entity, err := configuration.GetPublicBoardModelData(ctx, mockDevice.Devices[0].Boards[0].PublicCodename)
			So(err, ShouldBeNil)
			So(entity.Board, ShouldEqual, mockDevice.Devices[0].Boards[0].PublicCodename)
			So(entity.Models, ShouldResemble, getModelNamesFromMockBoard(mockDevice.Devices[0].Boards[0]))
		})
		Convey("Empty Input", func() {
			mockDevice := &ufspb.GoldenEyeDevices{}
			err := ImportPublicBoardsAndModels(ctx, mockDevice)
			entity, err := configuration.GetPublicBoardModelData(ctx, "test")
			So(err, ShouldNotBeNil)
			So(entity, ShouldBeNil)
		})
		Convey("Unlaunched Devices", func() {
			mockDevice := mockDevicesUnlaunched()
			err := ImportPublicBoardsAndModels(ctx, mockDevice)
			So(err, ShouldBeNil)
		})
		Convey("Unlaunched Devices not saved to DataStore", func() {
			mockDevice := mockDevicesUnlaunched()
			err := ImportPublicBoardsAndModels(ctx, mockDevice)
			So(err, ShouldBeNil)
			entity, err := configuration.GetPublicBoardModelData(ctx, mockDevice.Devices[0].Boards[0].PublicCodename)
			So(err, ShouldNotBeNil)
			So(entity, ShouldBeNil)
		})
	})
}

func mockDevicesLaunched() *ufspb.GoldenEyeDevices {
	return &ufspb.GoldenEyeDevices{
		Devices: []*ufspb.GoldenEyeDevice{
			{
				LaunchDate: "2022-05-01",
				Boards: []*ufspb.Board{
					{
						PublicCodename: "board1",
						Models: []*ufspb.Model{
							{Name: "model1"},
							{Name: "model2"},
						},
					},
				},
			},
		},
	}
}

func mockDevicesUnlaunched() *ufspb.GoldenEyeDevices {
	return &ufspb.GoldenEyeDevices{
		Devices: []*ufspb.GoldenEyeDevice{
			{
				LaunchDate: "2023-05-01",
				Boards: []*ufspb.Board{
					{
						PublicCodename: "boardNew",
						Models: []*ufspb.Model{
							{Name: "modelNew1"},
							{Name: "modelNew2"},
						},
					},
				},
			},
		},
	}
}

func getModelNamesFromMockBoard(board *ufspb.Board) (models []string) {
	for _, model := range board.Models {
		models = append(models, model.Name)
	}
	return models
}
