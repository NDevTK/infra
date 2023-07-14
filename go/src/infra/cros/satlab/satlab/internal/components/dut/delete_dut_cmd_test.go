// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dut

import (
	"context"
	ufsModels "infra/unifiedfleet/api/v1/models"
	ufsApi "infra/unifiedfleet/api/v1/rpc"
	ufspb "infra/unifiedfleet/api/v1/rpc"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type mockDeleteClient struct {
	getMachineLSECalls    []*ufspb.GetMachineLSERequest
	deleteMachineLSECalls []*ufspb.DeleteMachineLSERequest
}

func (c *mockDeleteClient) DeleteMachineLSE(ctx context.Context, req *ufsApi.DeleteMachineLSERequest, ops ...grpc.CallOption) (*emptypb.Empty, error) {
	c.deleteMachineLSECalls = append(c.deleteMachineLSECalls, req)
	return &emptypb.Empty{}, nil
}

func (c *mockDeleteClient) GetMachineLSE(ctx context.Context, req *ufsApi.GetMachineLSERequest, opts ...grpc.CallOption) (*ufsModels.MachineLSE, error) {
	c.getMachineLSECalls = append(c.getMachineLSECalls, req)
	return &ufsModels.MachineLSE{Name: req.Name}, nil
}

func Test_innerRunWithClients(t *testing.T) {
	tests := []struct {
		name            string
		cmd             *deleteDUT
		dutNames        []string
		wantGetCalls    []*ufsApi.GetMachineLSERequest
		wantDeleteCalls []*ufsApi.DeleteMachineLSERequest
	}{
		{
			name:     "delete calls ufs for duts passed in",
			cmd:      &deleteDUT{},
			dutNames: []string{"dut1", "dut2"},
			wantGetCalls: []*ufsApi.GetMachineLSERequest{
				{Name: "machineLSEs/dut1"},
				{Name: "machineLSEs/dut2"},
			},
			wantDeleteCalls: []*ufsApi.DeleteMachineLSERequest{
				{Name: "machineLSEs/dut1"},
				{Name: "machineLSEs/dut2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ufs := &mockDeleteClient{}

			if err := innerRunWithClients(context.Background(), tt.cmd, tt.dutNames, ufs); err != nil {
				t.Errorf("innerRunWithClients() error = %v", err)
			}

			if diff := cmp.Diff(tt.wantGetCalls, ufs.getMachineLSECalls, cmpopts.IgnoreUnexported(ufsApi.GetMachineLSERequest{})); diff != "" {
				t.Errorf("unexpected diff in get calls: %s", diff)
			}

			if diff := cmp.Diff(tt.wantDeleteCalls, ufs.deleteMachineLSECalls, cmpopts.IgnoreUnexported(ufsApi.DeleteMachineLSERequest{})); diff != "" {
				t.Errorf("unexpected diff in delete calls: %s", diff)
			}
		})
	}
}
