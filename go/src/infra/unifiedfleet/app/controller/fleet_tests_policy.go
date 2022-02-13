// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server/auth"
	"google.golang.org/grpc/codes"

	api "infra/unifiedfleet/api/v1/rpc"
	"infra/unifiedfleet/app/util"

	grpcStatus "google.golang.org/grpc/status"
)

const (
	// LUCI Auth group which is used to verify if a service account has permissions to run public Chromium tests in ChromeOS lab
	PublicUsersToChromeOSAuthGroup = "public-chromium-in-chromeos-builders"
)

func IsValidTest(ctx context.Context, req *api.CheckFleetTestsPolicyRequest) error {
	logging.Infof(ctx, "Request to check from crosfleet: %s", req)
	isValidPublicGroupMember, err := isPublicGroupMember(ctx)
	logging.Infof(ctx, "Service account being validated: %s", auth.CurrentIdentity(ctx).Email())
	if err != nil {
		// Ignoring error for now till we validate the service account membership check is correct
		logging.Errorf(ctx, "Request to check public chrome auth group membership failed: %s", err)
		return nil
		// return err
	}

	if !isValidPublicGroupMember {
		return nil
	}

	// Validate if the board and model are public
	if req.Board == "" {
		return grpcStatus.Errorf(codes.InvalidArgument, "Invalid input - Board cannot be empty for public tests.")
	}
	if !contains(getValidPublicBoards(), req.Board) {
		return grpcStatus.Errorf(codes.InvalidArgument, util.InvalidBoard, req.Board)
	}
	if req.Model == "" {
		return grpcStatus.Errorf(codes.InvalidArgument, "Invalid input - Model cannot be empty for public tests.")
	}
	if !contains(getValidPublictModels(), req.Model) {
		return grpcStatus.Errorf(codes.InvalidArgument, util.InvalidModel, req.Model)
	}

	// Validate Test Name
	if req.TestName == "" {
		return grpcStatus.Errorf(codes.InvalidArgument, "Invalid input - Test name cannot be empty for public tests.")
	}
	if !contains(getValidPublicTestNames(), req.TestName) {
		return grpcStatus.Errorf(codes.InvalidArgument, util.InvalidTest, req.TestName)
	}

	// Validate Image
	if req.Image == "" {
		return grpcStatus.Errorf(codes.InvalidArgument, "Invalid input - Image cannot be empty for public tests.")
	}
	if !contains(getValidPublicImages(), req.Image) {
		return grpcStatus.Errorf(codes.InvalidArgument, util.InvalidImage, req.Image)
	}

	return nil
}

func isPublicGroupMember(ctx context.Context) (bool, error) {
	// isPublicGroupMember, err := auth.IsMember(ctx, PublicUsersToChromeOSAuthGroup)
	// if err != nil {
	// 	logging.Errorf(ctx, "Check group %q membership failed while verifying if the test is tiggered by public users: %s", PublicUsersToChromeOSAuthGroup, err.Error())
	// 	return false, status.Errorf(codes.Internal, "can't check access group membership: %s", err)
	// }
	// return isPublicGroupMember, nil
	return true, nil
}

func getValidPublicTestNames() []string {
	return []string{"tast.lacros"}
}

func getValidPublicBoards() []string {
	return []string{"eve", "kevin"}
}

func getValidPublictModels() []string {
	return []string{"eve", "kevin"}
}

func getValidPublicImages() []string {
	return []string{"R100-14495.0.0-rc1"}
}

func contains(listItems []string, name string) bool {
	for _, item := range listItems {
		if name == item {
			return true
		}
	}
	return false
}
