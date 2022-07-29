// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ufs

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc"

	models "infra/unifiedfleet/api/v1/models"
	lab "infra/unifiedfleet/api/v1/models/chromeos/lab"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
)

// TestNewClient tests that NewClient responds in appropriate ways
// to ill-formed arguments. Not a deep test.
func TestNewClient(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	hc := http.DefaultClient

	_, err := NewClient(ctx, hc, "")
	if err == nil {
		t.Errorf("expected error to not be nil")
	}
}

// TestGetPools tests that GetPools passes an appropriately annotated name to the
func TestGetPools(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	c := &fakeGetPoolsClient{}
	expectedPools := []string{"aaaa"}
	actualPools, err := GetPools(ctx, c, "a")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if diff := cmp.Diff(expectedPools, actualPools); diff != "" {
		t.Errorf("unexpected diff (-want +got): %s", diff)
	}
	expectedName := map[string]bool{"a": true}
	actualName := c.names
	if diff := cmp.Diff(expectedName, actualName); diff != "" {
		t.Errorf("unexpected diff (-want +got): %s", diff)
	}
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
	names map[string]bool
}

// GetMachineLSE always returns a fake machine.
func (f *fakeGetPoolsClient) GetMachineLSE(ctx context.Context, in *ufsAPI.GetMachineLSERequest, opts ...grpc.CallOption) (*models.MachineLSE, error) {
	if f.names == nil {
		f.names = map[string]bool{}
	}
	f.names[in.GetName()] = true
	return fakeMachine, nil
}

// GetMachineLSE always returns a fake machine.
func (f *fakeGetPoolsClient) GetChromeOSDeviceData(ctx context.Context, in *ufsAPI.GetChromeOSDeviceDataRequest, opts ...grpc.CallOption) (*models.ChromeOSDeviceData, error) {
	if f.names == nil {
		f.names = map[string]bool{}
	}
	f.names[in.GetHostname()] = true
	return &models.ChromeOSDeviceData{
		LabConfig: fakeMachine,
	}, nil
}

// GetDeviceData always returns a fake host.
func (f *fakeGetPoolsClient) GetDeviceData(ctx context.Context, in *ufsAPI.GetDeviceDataRequest, opts ...grpc.CallOption) (*ufsAPI.GetDeviceDataResponse, error) {
	if f.names == nil {
		f.names = map[string]bool{}
	}
	f.names[in.GetHostname()] = true
	return &ufsAPI.GetDeviceDataResponse{
		Resource: &ufsAPI.GetDeviceDataResponse_ChromeOsDeviceData{
			ChromeOsDeviceData: &models.ChromeOSDeviceData{
				LabConfig: fakeMachine,
			},
		},
	}, nil
}
