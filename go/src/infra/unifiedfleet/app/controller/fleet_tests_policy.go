// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"fmt"
	"time"

	"go.chromium.org/luci/auth/identity"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server/auth"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"

	ufspb "infra/unifiedfleet/api/v1/models"
	api "infra/unifiedfleet/api/v1/rpc"
	"infra/unifiedfleet/app/model/configuration"
)

const (
	// LUCI Auth group which is used to verify if a service account has permissions to run public Chromium tests in ChromeOS lab
	PublicUsersToChromeOSAuthGroup = "public-chromium-in-chromeos-builders"

	// Date Format to parse launch date for Device info read from DLM
	DateFormat = "2006-01-02"
)

// InvalidBoardError is the error raised when a private board is specified for a public test
type InvalidBoardError struct {
	Board string
}

func (e *InvalidBoardError) Error() string {
	return fmt.Sprintf("Cannnot run public tests on a private board : %s", e.Board)
}

// InvalidModelError is the error raised when a private model is specified for a public test
type InvalidModelError struct {
	Model string
}

func (e *InvalidModelError) Error() string {
	return fmt.Sprintf("Cannot run public tests on a private model : %s", e.Model)
}

// InvalidImageError is the error raised when an invalid image is specified for a public test
type InvalidImageError struct {
	Image string
}

func (e *InvalidImageError) Error() string {
	return fmt.Sprintf("Cannot run public tests on an image which is not allowlisted : %s", e.Image)
}

// InvalidTestError is the error raised when an invalid image is specified for a public test
type InvalidTestError struct {
	TestName string
}

func (e *InvalidTestError) Error() string {
	return fmt.Sprintf("Public user cannnot run the not allowlisted test : %s", e.TestName)
}

func IsValidTest(ctx context.Context, req *api.CheckFleetTestsPolicyRequest) error {
	isMemberInPublicGroup, err := isPublicGroupMember(ctx, req)
	if err != nil {
		// Ignoring error for now till we validate the service account membership check is correct
		logging.Errorf(ctx, "Request to check public chrome auth group membership failed: %s", err)
		return nil
	}

	if !isMemberInPublicGroup {
		return nil
	}

	// Validate if the board and model are public
	if err := validatePublicBoardModel(ctx, req.Board, req.Model); err != nil {
		return err
	}

	// Validate Test Name
	if req.TestName == "" {
		return grpcStatus.Errorf(codes.InvalidArgument, "Invalid input - Test name cannot be empty for public tests.")
	}
	if !contains(getValidPublicTestNames(), req.TestName) {
		return &InvalidTestError{TestName: req.TestName}
	}

	// Validate Image
	if req.Image == "" {
		return grpcStatus.Errorf(codes.InvalidArgument, "Invalid input - Image cannot be empty for public tests.")
	}
	if !contains(getValidPublicImages(), req.Image) {
		return &InvalidImageError{Image: req.Image}
	}

	return nil
}

func ImportPublicBoardsAndModels(ctx context.Context, goldenEyeDevices *ufspb.GoldenEyeDevices) error {
	boardPublicModelMap := make(map[string][]string)
	boardHasPrivateModelMap := make(map[string]bool)
	for _, device := range goldenEyeDevices.Devices {
		if device.LaunchDate == "" {
			continue
		}
		launchDate, err := time.Parse(DateFormat, device.LaunchDate)
		if err != nil {
			// Ignore and process the rest of the data
			logging.Infof(ctx, "Failed to parse Launch Date from Golden Eye Device data %s", device.LaunchDate)
			continue
		}
		if launchDate.Before(time.Now()) {
			// Already launched board and model, can be added to allowed list
			for _, board := range device.Boards {
				logging.Infof(ctx, "Launched Board from Golden Eye Device data %s", board.PublicCodename)
				for _, model := range board.Models {
					boardPublicModelMap[board.GetPublicCodename()] = append(boardPublicModelMap[board.GetPublicCodename()], model.Name)
				}
			}
		} else {
			// Flag the board for private model(s)
			for _, board := range device.Boards {
				boardHasPrivateModelMap[board.GetPublicCodename()] = true
			}
		}
	}
	for board, models := range boardPublicModelMap {
		configuration.AddPublicBoardModelData(ctx, board, models, boardHasPrivateModelMap[board])
	}
	return nil
}

func isPublicGroupMember(ctx context.Context, req *api.CheckFleetTestsPolicyRequest) (bool, error) {
	var ident identity.Identity
	var err error
	if req.GetTestServiceAccount() != "" {
		ident, err = identity.MakeIdentity(req.GetTestServiceAccount())
		if err != nil {
			logging.WithError(err).Errorf(ctx, "Failed to create identity for %q.", req.GetTestServiceAccount())
			return false, nil
		}
	} else {
		ident = auth.CurrentIdentity(ctx)
	}

	logging.Infof(ctx, "CheckFleetTestsPolicyRequest: %s", req)
	logging.Infof(ctx, "Service account being validated: %s", ident.Email())

	state := auth.GetState(ctx)
	if state == nil {
		logging.Errorf(ctx, "Failed to check auth, no State in context.")
		return false, nil
	}
	authDB := state.DB()
	if authDB == nil {
		logging.Errorf(ctx, "Failed to check auth, nil auth DB in State.")
		return false, nil
	}

	isMemberInPublicGroup, err := authDB.IsMember(ctx, ident, []string{PublicUsersToChromeOSAuthGroup})
	if err != nil {
		// Ignoring error for now till we validate the service account membership check is correct
		logging.Errorf(ctx, "Request to check public chrome auth group membership failed: %s", err)
		return false, nil
	}
	return isMemberInPublicGroup, nil
}

func getValidPublicTestNames() []string {
	return []string{"tast.lacros"}
}

func validatePublicBoardModel(ctx context.Context, board string, model string) error {
	if board == "" {
		return grpcStatus.Errorf(codes.InvalidArgument, "Invalid input - Board cannot be empty for public tests.")
	}

	publicBoardEntity, err := configuration.GetPublicBoardModelData(ctx, board)
	if err != nil {
		return &InvalidBoardError{Board: board}
	}
	if model == "" {
		if publicBoardEntity.BoardHasPrivateModels {
			return grpcStatus.Errorf(codes.InvalidArgument, "Invalid input - Model cannot be empty as the specified board has unlaunched models.")
		} else {
			return nil
		}
	}
	for _, m := range publicBoardEntity.Models {
		if m == model {
			return nil
		}
	}
	return &InvalidModelError{Model: model}
}

func getValidPublicImages() []string {
	return []string{"chromiumos-image-archive/eve-public/R105-14988.0.0", "chromiumos-image-archive/octopus-public/R105-14988.0.0"}
}

func contains(listItems []string, name string) bool {
	for _, item := range listItems {
		if name == item {
			return true
		}
	}
	return false
}
