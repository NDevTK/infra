// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cron

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	gcepAPI "go.chromium.org/luci/gce/api/config/v1"

	"infra/cros/botsregulator/internal/clients"
	"infra/cros/botsregulator/internal/regulator"
	ufspb "infra/unifiedfleet/api/v1/models"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
)

func TestRegulate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockUFS := clients.NewMockUFSClient(mockCtrl)
	mockGCEP := clients.NewMockGCEPClient(mockCtrl)
	ctx := context.Background()
	ctx = context.WithValue(ctx, clients.MockGCEPClientKey, mockGCEP)
	ctx = context.WithValue(ctx, clients.MockUFSClientKey, mockUFS)

	opts := &regulator.RegulatorOptions{
		BPI:       "bpi.endpoint",
		UFS:       "ufs.enpoint",
		Hive:      "cloudbots",
		CfID:      "cloudbots-dev",
		Namespace: "os",
	}

	ctxWithNS := clients.SetUFSNamespace(ctx, "os")
	gomock.InOrder(
		mockUFS.EXPECT().ListMachineLSEs(ctxWithNS, &ufsAPI.ListMachineLSEsRequest{
			Filter:   "hive=cloudbots",
			KeysOnly: true,
			PageSize: 1000,
		}).Return(&ufsAPI.ListMachineLSEsResponse{
			MachineLSEs: []*ufspb.MachineLSE{
				{Name: "machineLSEs/dut-1"},
				{Name: "machineLSEs/dut-2"},
				{Name: "machineLSEs/dut-3"},
				{Name: "machineLSEs/dut-4"},
			}}, nil),
		mockUFS.EXPECT().ListSchedulingUnits(ctxWithNS, &ufsAPI.ListSchedulingUnitsRequest{
			PageSize: 1000,
		}).Return(&ufsAPI.ListSchedulingUnitsResponse{
			SchedulingUnits: []*ufspb.SchedulingUnit{
				{Name: "schedulingunits/su-1", MachineLSEs: []string{"dut-1"}},
				{Name: "schedulingunits/su-2", MachineLSEs: []string{"dut-2", "dut-3"}},
				{Name: "schedulingunits/su-3", MachineLSEs: []string{"dut-8", "dut-9"}},
			}}, nil),
		mockGCEP.EXPECT().Get(ctx, &gcepAPI.GetRequest{
			Id: "cloudbots-dev",
		}).Return(&gcepAPI.Config{
			Prefix: "cloudbots-dev",
		}, nil),
		mockGCEP.EXPECT().Update(ctx, &gcepAPI.UpdateRequest{
			Id: "cloudbots-dev",
			Config: &gcepAPI.Config{
				Prefix: "cloudbots-dev",
				Duts: map[string]*emptypb.Empty{
					"su-1":  {},
					"su-2":  {},
					"dut-4": {},
				},
			},
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{"config.duts"},
			},
		}),
	)

	// Fake Cloud Run environment.
	t.Setenv("K_SERVICE", "bots-regulator-test")

	err := Regulate(ctx, opts)
	if err != nil {
		t.Fatalf("should not error: %v", err)
	}

}
