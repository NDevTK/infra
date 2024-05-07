// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package clients

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	ufspb "infra/unifiedfleet/api/v1/models"
	chromeosLab "infra/unifiedfleet/api/v1/models/chromeos/lab"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
)

type contextKey string

// MockUFSClientKey is used for testing.
var MockUFSClientKey contextKey = "used in tests only for setting the mock UFSClient"

// UFSClient is UFS API	wrapper for BotsRegulator specific usage.
// It is used for mocking and testing.
type UFSClient interface {
	ListMachineLSEs(ctx context.Context, in *ufsAPI.ListMachineLSEsRequest, opts ...grpc.CallOption) (*ufsAPI.ListMachineLSEsResponse, error)
	ListMachines(ctx context.Context, in *ufsAPI.ListMachinesRequest, opts ...grpc.CallOption) (*ufsAPI.ListMachinesResponse, error)
	ListSchedulingUnits(ctx context.Context, in *ufsAPI.ListSchedulingUnitsRequest, opts ...grpc.CallOption) (*ufsAPI.ListSchedulingUnitsResponse, error)
	UpdateMachineLSE(ctx context.Context, in *ufsAPI.UpdateMachineLSERequest, opts ...grpc.CallOption) (*ufspb.MachineLSE, error)
}

func NewUFSClient(ctx context.Context, host string) (UFSClient, error) {
	if mockClient, ok := ctx.Value(MockUFSClientKey).(UFSClient); ok {
		return mockClient, nil
	}
	pc, err := rawPRPCClient(ctx, host)
	if err != nil {
		return nil, err
	}
	ic := ufsAPI.NewFleetPRPCClient(pc)
	return &ufsService{
		client: ic,
	}, nil
}

// ufsService is used in non-test environments.
type ufsService struct {
	client ufsAPI.FleetClient
}

func (u *ufsService) ListMachineLSEs(ctx context.Context, in *ufsAPI.ListMachineLSEsRequest, opts ...grpc.CallOption) (*ufsAPI.ListMachineLSEsResponse, error) {
	return u.client.ListMachineLSEs(ctx, in, opts...)
}

func (u *ufsService) ListMachines(ctx context.Context, in *ufsAPI.ListMachinesRequest, opts ...grpc.CallOption) (*ufsAPI.ListMachinesResponse, error) {
	return u.client.ListMachines(ctx, in, opts...)
}

func (u *ufsService) ListSchedulingUnits(ctx context.Context, in *ufsAPI.ListSchedulingUnitsRequest, opts ...grpc.CallOption) (*ufsAPI.ListSchedulingUnitsResponse, error) {
	return u.client.ListSchedulingUnits(ctx, in, opts...)
}

func (u *ufsService) UpdateMachineLSE(ctx context.Context, in *ufsAPI.UpdateMachineLSERequest, opts ...grpc.CallOption) (*ufspb.MachineLSE, error) {
	return u.client.UpdateMachineLSE(ctx, in, opts...)
}

// SetUFSNamespace is a helper function to set UFS namespace in context.
func SetUFSNamespace(ctx context.Context, namespace string) context.Context {
	md := metadata.Pairs("namespace", namespace)
	return metadata.NewOutgoingContext(ctx, md)
}

// InitializeUpdateDUTRequest return a new initialized UpdateMachineLSERequest.
func InitializeUpdateDUTRequest(hostname, hive string) *ufsAPI.UpdateMachineLSERequest {
	// An empty machineLSE is enough. UFS will fetch the correct lse from the machinelse.name.
	// 3679c23a3c07de90bc8d4241ea77416cf3dcda45:infra/go/src/infra/unifiedfleet/app/controller/dut.go;l=160
	// Shivas for ref: 132b2fe1a670c91e9eaad45b3cb0d04601ad0ce3:go/src/infra/cmd/shivas/internal/ufs/subcmds/dut/update_dut_batch.go;l=255
	lse := &ufspb.MachineLSE{
		Lse: &ufspb.MachineLSE_ChromeosMachineLse{
			ChromeosMachineLse: &ufspb.ChromeOSMachineLSE{
				ChromeosLse: &ufspb.ChromeOSMachineLSE_DeviceLse{
					DeviceLse: &ufspb.ChromeOSDeviceLSE{
						Device: &ufspb.ChromeOSDeviceLSE_Dut{
							Dut: &chromeosLab.DeviceUnderTest{
								Peripherals: &chromeosLab.Peripherals{
									Chameleon:     &chromeosLab.Chameleon{},
									Servo:         &chromeosLab.Servo{},
									Rpm:           &chromeosLab.OSRPM{},
									Audio:         &chromeosLab.Audio{},
									Wifi:          &chromeosLab.Wifi{},
									Touch:         &chromeosLab.Touch{},
									CameraboxInfo: &chromeosLab.Camerabox{},
									Dolos:         &chromeosLab.Dolos{},
								},
							},
						},
					},
				},
			},
		},
	}
	lse.Name = ufsUtil.AddPrefix(ufsUtil.MachineLSECollection, hostname)
	lse.Hostname = hostname
	lse.GetChromeosMachineLse().GetDeviceLse().GetDut().Hostname = hostname
	lse.GetChromeosMachineLse().GetDeviceLse().GetDut().Hive = hive
	req := &ufsAPI.UpdateMachineLSERequest{
		MachineLSE: lse,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{"dut.hive"},
		},
	}
	return req
}
