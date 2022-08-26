// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/auth/identity"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/auth/authtest"

	ufspb "infra/unifiedfleet/api/v1/models"
	api "infra/unifiedfleet/api/v1/rpc"
	"infra/unifiedfleet/app/model/configuration"
)

const (
	LAUNCHED_BOARD                = "launched_board"
	UNLAUNCHED_BOARD              = "unlaunched_board"
	LAUNCHED_BOARD_PRIVATE_MODELS = "launched_board_private_models"
	VALID_IMAGE_EVE               = "eve-public/R105-14988.0.0"
)

func TestIsValidPublicChromiumTest(t *testing.T) {
	t.Parallel()
	ctx := auth.WithState(testingContext(), &authtest.FakeState{
		Identity:       "user:abc@def.com",
		IdentityGroups: []string{"public-chromium-in-chromeos-builders"},
	})
	configuration.AddPublicBoardModelData(ctx, "eve", []string{"eve"}, false)
	Convey("Is Valid Public Chromium Test", t, func() {
		Convey("happy path", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Board:    "eve",
				Model:    "eve",
				Image:    VALID_IMAGE_EVE,
			}

			err := IsValidTest(ctx, req)

			So(err, ShouldBeNil)
		})
		Convey("Private test name and public auth group member", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "private",
				Board:    "eve",
				Model:    "eve",
				Image:    VALID_IMAGE_EVE,
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
				Image:    VALID_IMAGE_EVE,
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
				Image:    VALID_IMAGE_EVE,
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
				Image:    VALID_IMAGE_EVE,
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
				Image:    VALID_IMAGE_EVE,
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
				Image:    "chromiumos-image-archive/eve-public/LATEST-14695.113.3",
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
				Image:    VALID_IMAGE_EVE,
			}

			err := IsValidTest(ctx, req)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Test name cannot be empty")
		})
		Convey("Missing Board", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Model:    "eve",
				Image:    VALID_IMAGE_EVE,
			}

			err := IsValidTest(ctx, req)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Board cannot be empty")
		})
		Convey("Missing Models - returns error if board has private models", func() {
			configuration.AddPublicBoardModelData(ctx, "fakePrivateBoard", []string{"fakeModelLaunched"}, true)
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Board:    "fakePrivateBoard",
				Image:    VALID_IMAGE_EVE,
			}

			err := IsValidTest(ctx, req)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Model cannot be empty as the specified board has unlaunched models")
		})
		Convey("Public Model and Public Board With Private Model - succeeds", func() {
			configuration.AddPublicBoardModelData(ctx, "fakePrivateBoard", []string{"fakeModelLaunched"}, true)
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Board:    "fakePrivateBoard",
				Model:    "fakeModelLaunched",
				Image:    VALID_IMAGE_EVE,
			}

			err := IsValidTest(ctx, req)

			So(err, ShouldBeNil)
		})
		Convey("Private Model and Public Board With Public Models - returns error", func() {
			configuration.AddPublicBoardModelData(ctx, "fakePrivateBoard", []string{"fakeModelLaunched"}, true)
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Board:    "fakePrivateBoard",
				Model:    "fakeModelUnLaunched",
				Image:    VALID_IMAGE_EVE,
			}

			err := IsValidTest(ctx, req)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "private model")
		})
		Convey("Missing Models - ok for public boards with only public models", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Board:    "eve",
				Image:    VALID_IMAGE_EVE,
			}

			err := IsValidTest(ctx, req)

			So(err, ShouldBeNil)
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

func TestIsPublicGroupMember(t *testing.T) {
	t.Parallel()
	Convey("Is Public Group Member", t, func() {
		Convey("happy path", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Board:    "eve",
				Model:    "eve",
				Image:    VALID_IMAGE_EVE,
			}
			ctx := auth.WithState(testingContext(), &authtest.FakeState{
				Identity:       "user:abc@def.com",
				IdentityGroups: []string{"public-chromium-in-chromeos-builders"},
			})

			publicGroupMember, err := isPublicGroupMember(ctx, req)

			So(err, ShouldBeNil)
			So(publicGroupMember, ShouldBeTrue)
		})
		Convey("happy path - request with test service account", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName:           "tast.lacros",
				Board:              "eve",
				Model:              "eve",
				Image:              VALID_IMAGE_EVE,
				TestServiceAccount: "user:abc@def.com",
			}
			ctx := auth.WithState(testingContext(), &authtest.FakeState{
				Identity: identity.AnonymousIdentity,
				FakeDB: authtest.NewFakeDB(
					authtest.MockMembership("user:abc@def.com", PublicUsersToChromeOSAuthGroup),
				),
			})

			publicGroupMember, err := isPublicGroupMember(ctx, req)

			So(err, ShouldBeNil)
			So(publicGroupMember, ShouldBeTrue)
		})
		Convey("Test service account not a public auth group member - Returns false", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName:           "tast.lacros",
				Board:              "eve",
				Model:              "eve",
				Image:              VALID_IMAGE_EVE,
				TestServiceAccount: "abc@def.com",
			}
			ctx := auth.WithState(testingContext(), &authtest.FakeState{
				Identity: identity.AnonymousIdentity,
			})

			publicGroupMember, err := isPublicGroupMember(ctx, req)

			So(err, ShouldBeNil)
			So(publicGroupMember, ShouldBeFalse)
		})
		Convey("No Test service account and empty context - Returns false", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Board:    "eve",
				Model:    "eve",
				Image:    VALID_IMAGE_EVE,
			}
			ctx := auth.WithState(testingContext(), &authtest.FakeState{
				Identity: identity.AnonymousIdentity,
			})

			publicGroupMember, err := isPublicGroupMember(ctx, req)

			So(err, ShouldBeNil)
			So(publicGroupMember, ShouldBeFalse)
		})
		Convey("Not a public group member", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Board:    "eve",
				Model:    "eve",
				Image:    VALID_IMAGE_EVE,
			}
			ctx := auth.WithState(testingContext(), &authtest.FakeState{
				Identity: "user:abc@def.com",
			})

			publicGroupMember, err := isPublicGroupMember(ctx, req)

			So(err, ShouldBeNil)
			So(publicGroupMember, ShouldBeFalse)
		})
		Convey("Nil State - Returns false", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Board:    "eve",
				Model:    "eve",
				Image:    VALID_IMAGE_EVE,
			}
			ctx := auth.WithState(testingContext(), nil)

			publicGroupMember, err := isPublicGroupMember(ctx, req)

			So(err, ShouldBeNil)
			So(publicGroupMember, ShouldBeFalse)
		})
		Convey("Nil State DB - Returns false", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Board:    "eve",
				Model:    "eve",
				Image:    VALID_IMAGE_EVE,
			}
			ctx := auth.WithState(testingContext(), &authtest.FakeState{
				FakeDB: nil,
			})
			publicGroupMember, err := isPublicGroupMember(ctx, req)

			So(err, ShouldBeNil)
			So(publicGroupMember, ShouldBeFalse)
		})
		Convey("Anonymous Identity - Returns true", func() {
			req := &api.CheckFleetTestsPolicyRequest{
				TestName: "tast.lacros",
				Board:    "eve",
				Model:    "eve",
				Image:    VALID_IMAGE_EVE,
			}
			ctx := auth.WithState(testingContext(), &authtest.FakeState{
				IdentityGroups: []string{"public-chromium-in-chromeos-builders"},
			})
			publicGroupMember, err := isPublicGroupMember(ctx, req)

			So(err, ShouldBeNil)
			So(publicGroupMember, ShouldBeTrue)
		})
	})
}

func TestImportPublicBoardsAndModels(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	Convey("Import Public Boards and Models", t, func() {
		Convey("Happy Path", func() {
			mockDevice := mockDevices()

			err := ImportPublicBoardsAndModels(ctx, mockDevice)

			So(err, ShouldBeNil)
		})
		Convey("Happy Path Check DataStore", func() {
			mockDevice := mockDevices()

			dataerr := ImportPublicBoardsAndModels(ctx, mockDevice)
			entity, err := configuration.GetPublicBoardModelData(ctx, LAUNCHED_BOARD)

			So(dataerr, ShouldBeNil)
			So(err, ShouldBeNil)
			So(entity.Board, ShouldEqual, LAUNCHED_BOARD)
			So(len(entity.Models), ShouldEqual, 4)
			So(entity.BoardHasPrivateModels, ShouldBeFalse)
		})
		Convey("Empty Input", func() {
			mockDevice := &ufspb.GoldenEyeDevices{}

			dataerr := ImportPublicBoardsAndModels(ctx, mockDevice)
			entity, err := configuration.GetPublicBoardModelData(ctx, "test")

			So(dataerr, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(entity, ShouldBeNil)
		})
		Convey("Unlaunched Devices not saved to DataStore", func() {
			mockDevice := mockDevices()

			dataerr := ImportPublicBoardsAndModels(ctx, mockDevice)
			entity, err := configuration.GetPublicBoardModelData(ctx, UNLAUNCHED_BOARD)

			So(dataerr, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(entity, ShouldBeNil)
		})
		Convey("Unlaunched Models not saved to DataStore", func() {
			mockDevice := mockDevices()

			dataerr := ImportPublicBoardsAndModels(ctx, mockDevice)
			entity, err := configuration.GetPublicBoardModelData(ctx, LAUNCHED_BOARD_PRIVATE_MODELS)

			So(dataerr, ShouldBeNil)
			So(err, ShouldBeNil)
			So(entity, ShouldNotBeNil)
			So(len(entity.Models), ShouldEqual, 2)
			So(entity.BoardHasPrivateModels, ShouldBeTrue)
		})
	})
}

func TestValidatePublicBoardModel(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	configuration.AddPublicBoardModelData(ctx, "eve", []string{"eve"}, false)
	Convey("Validate Board and Model", t, func() {
		Convey("Happy Path", func() {
			err := validatePublicBoardModel(ctx, "eve", "eve")
			So(err, ShouldBeNil)
		})
		Convey("Private Board", func() {
			err := validatePublicBoardModel(ctx, "board", "eve")
			err, ok := err.(*InvalidBoardError)

			So(err, ShouldNotBeNil)
			So(ok, ShouldBeTrue)
		})
		Convey("Private Model", func() {
			err := validatePublicBoardModel(ctx, "eve", "model")
			err, ok := err.(*InvalidModelError)

			So(err, ShouldNotBeNil)
			So(ok, ShouldBeTrue)
		})
	})
}

func mockDevices() *ufspb.GoldenEyeDevices {
	return &ufspb.GoldenEyeDevices{
		Devices: []*ufspb.GoldenEyeDevice{
			{
				LaunchDate: "2022-05-01",
				Boards: []*ufspb.Board{
					{
						PublicCodename: LAUNCHED_BOARD,
						Models: []*ufspb.Model{
							{Name: "model1"},
							{Name: "model2"},
						},
					},
				},
			},
			{
				LaunchDate: "2022-05-01",
				Boards: []*ufspb.Board{
					{
						PublicCodename: LAUNCHED_BOARD_PRIVATE_MODELS,
						Models: []*ufspb.Model{
							{Name: "modelOld1"},
							{Name: "modelOld2"},
						},
					},
				},
			},
			{
				LaunchDate: "2021-01-01",
				Boards: []*ufspb.Board{
					{
						PublicCodename: LAUNCHED_BOARD,
						Models: []*ufspb.Model{
							{Name: "model3"},
							{Name: "model4"},
						},
					},
				},
			},
			{
				LaunchDate: time.Now().Add(time.Hour * 10000).Format(DateFormat),
				Boards: []*ufspb.Board{
					{
						PublicCodename: LAUNCHED_BOARD_PRIVATE_MODELS,
						Models: []*ufspb.Model{
							{Name: "modelNew1"},
							{Name: "modelNew2"},
						},
					},
				},
			},
			{
				LaunchDate: time.Now().Add(time.Hour * 10000).Format(DateFormat),
				Boards: []*ufspb.Board{
					{
						PublicCodename: UNLAUNCHED_BOARD,
						Models: []*ufspb.Model{
							{Name: "model2New1"},
							{Name: "model2New2"},
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
