// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ufs

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc"

	models "infra/unifiedfleet/api/v1/models"
	lab "infra/unifiedfleet/api/v1/models/chromeos/lab"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
)

// TestGetPools tests that GetPools passes an appropriately annotated name to the
func TestGetPools(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	c := &fakeGetPoolsClient{}
	t.Run("check pools", func(t *testing.T) {
		// t.Parallel() -- test interacts with fake client statefully.
		expected := []string{"aaaa"}
		actual, err := GetPools(ctx, c, "a")
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("unexpected diff (-want +got): %s", diff)
		}
	})
	t.Run("check name used", func(t *testing.T) {
		// t.Parallel() -- test interacts with fake client statefully.
		expected := []string{"machineLSEs/a"}
		actual := c.names
		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("unexpected diff (-want +got): %s", diff)
		}
	})
}

// FakeMachine is a fake DUT with pool "aaaa".
var fakeMachine = &models.MachineLSE{
	Lse: &models.MachineLSE_ChromeosMachineLse{
		ChromeosMachineLse: &models.ChromeOSMachineLSE{
			ChromeosLse: &models.ChromeOSMachineLSE_DeviceLse{
				DeviceLse: &models.ChromeOSDeviceLSE{
					Device: &models.ChromeOSDeviceLSE_Dut{
						Dut: &lab.DeviceUnderTest{
							Pools: []string{"aaaa"},
						},
					},
				},
			},
		},
	},
}

// FakeGetPoolsClient mimics a UFS client and records what it was asked to look up.
type fakeGetPoolsClient struct {
	names []string
}

// GetMachineLSE always returns a fake machine.
func (f *fakeGetPoolsClient) GetMachineLSE(ctx context.Context, in *ufsAPI.GetMachineLSERequest, opts ...grpc.CallOption) (*models.MachineLSE, error) {
	f.names = append(f.names, in.GetName())
	return fakeMachine, nil
}
