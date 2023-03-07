// Copyright 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package run

import (
	"context"
	"testing"

	"infra/cmd/crosfleet/internal/buildbucket"

	"google.golang.org/grpc"

	models "infra/unifiedfleet/api/v1/models"
	ufsapi "infra/unifiedfleet/api/v1/rpc"
)

type fakeUFSClient struct{}

// CheckFleetTestsPolicy checks the fleet test policy for the given test parameters.
func (c fakeUFSClient) CheckFleetTestsPolicy(ctx context.Context, req *ufsapi.CheckFleetTestsPolicyRequest, opt ...grpc.CallOption) (*ufsapi.CheckFleetTestsPolicyResponse, error) {
	return &ufsapi.CheckFleetTestsPolicyResponse{
		IsTestValid: true,
		TestStatus: &ufsapi.TestStatus{
			Code: ufsapi.TestStatus_OK,
		},
	}, nil
}

// GetMachineLSE gets information about a DUT.
func (c fakeUFSClient) GetMachineLSE(ctx context.Context, req *ufsapi.GetMachineLSERequest, opt ...grpc.CallOption) (*models.MachineLSE, error) {
	return nil, nil
}

// GetMachine retrieves the details of the machine.
func (c *fakeUFSClient) GetMachine(ctx context.Context, req *ufsapi.GetMachineRequest, opt ...grpc.CallOption) (*models.Machine, error) {
	return nil, nil
}

func TestSuiteNoModels(t *testing.T) {
	r := suiteRun{
		testCommonFlags: testCommonFlags{
			exitEarly: true,
			repeats:   1,
			priority:  DefaultSwarmingPriority,
			release:   "R112-15357.0.0",
			board:     "drallion",
			pool:      "DUT_POOL_QUOTA",
		},
	}
	ctx := context.Background()

	ufs := &fakeUFSClient{}
	bb := &buildbucket.FakeClient{
		Client: buildbucket.FakeBuildClient{
			ExpectedSchedule: buildbucket.ScheduleParams{
				BuilderName: "cros_test_platform",
				Tags: map[string]string{
					"crosfleet-tool": "suite",
					"label-board":    "drallion",
					"label-image":    "drallion-release/R112-15357.0.0",
					"label-pool":     "DUT_POOL_QUOTA",
					"user_agent":     "crosfleet",
				},
			},
		},
	}
	err := r.innerRun(nil, []string{"bvt-installer"}, ctx, bb, ufs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
