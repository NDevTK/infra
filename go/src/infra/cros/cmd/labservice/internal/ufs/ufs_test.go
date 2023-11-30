// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ufs

import (
	"context"
	"net"
	"testing"

	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"

	"infra/cros/cmd/labservice/internal/ufs/cache"
	ufspb "infra/unifiedfleet/api/v1/models"
	lab "infra/unifiedfleet/api/v1/models/chromeos/lab"
	manufacturing "infra/unifiedfleet/api/v1/models/chromeos/manufacturing"
	ufsapi "infra/unifiedfleet/api/v1/rpc"
)

func TestGetChromeOsDutTopology_single(t *testing.T) {
	t.Parallel()
	ctx, cf := context.WithCancel(context.Background())
	defer cf()
	s := &fakeServer{
		ChromeOSDeviceData: &ufspb.ChromeOSDeviceData{
			LabConfig: &ufspb.MachineLSE{
				Hostname: "200.200.200.200",
				Lse: &ufspb.MachineLSE_ChromeosMachineLse{
					ChromeosMachineLse: &ufspb.ChromeOSMachineLSE{
						ChromeosLse: &ufspb.ChromeOSMachineLSE_DeviceLse{
							DeviceLse: &ufspb.ChromeOSDeviceLSE{
								Device: &ufspb.ChromeOSDeviceLSE_Dut{
									Dut: &lab.DeviceUnderTest{
										Peripherals: &lab.Peripherals{
											Audio: &lab.Audio{
												AudioBox: true,
												Atrus:    true,
											},
											Chameleon: &lab.Chameleon{
												AudioBoard:           true,
												ChameleonPeripherals: []lab.ChameleonType{lab.ChameleonType_CHAMELEON_TYPE_DP, lab.ChameleonType_CHAMELEON_TYPE_V3},
											},
											Servo: &lab.Servo{
												ServoHostname: "servo_host",
												ServoPort:     33,
											},
											Wifi: &lab.Wifi{
												Wificell:    true,
												AntennaConn: lab.Wifi_CONN_CONDUCTIVE,
											},
											Touch: &lab.Touch{
												Mimo: true,
											},
											Camerabox: true,
											CameraboxInfo: &lab.Camerabox{
												Facing: lab.Camerabox_FACING_FRONT,
											},
											Cable: []*lab.Cable{
												{
													Type: lab.CableType_CABLE_AUDIOJACK,
												},
											},
											BluetoothPeers: []*lab.BluetoothPeer{
												{
													Device: &lab.BluetoothPeer_RaspberryPi{
														RaspberryPi: &lab.RaspberryPi{
															Hostname: "test-btp1",
															State:    lab.PeripheralState_WORKING,
														},
													},
												},
												{
													Device: &lab.BluetoothPeer_RaspberryPi{
														RaspberryPi: &lab.RaspberryPi{
															Hostname: "test-btp2",
															State:    lab.PeripheralState_BROKEN,
														},
													},
												},
												{
													Device: &lab.BluetoothPeer_RaspberryPi{
														RaspberryPi: &lab.RaspberryPi{
															Hostname: "test-btp3",
															State:    lab.PeripheralState_UNKNOWN,
														},
													},
												},
											},
										},
										Siminfo: []*lab.SIMInfo{
											{
												SlotId:   1,
												Type:     lab.SIMType_SIM_DIGITAL,
												Eid:      "eid1",
												TestEsim: true,
												ProfileInfo: []*lab.SIMProfileInfo{
													{
														Iccid:       "iccid1",
														SimPin:      "pin1",
														SimPuk:      "puk1",
														CarrierName: lab.NetworkProvider_NETWORK_ATT,
														OwnNumber:   "number1",
													},
													{
														Iccid:       "iccid2",
														SimPin:      "pin2",
														SimPuk:      "puk2",
														CarrierName: lab.NetworkProvider_NETWORK_TMOBILE,
														OwnNumber:   "number2",
													},
												},
											},
											{
												SlotId:   2,
												Type:     lab.SIMType_SIM_PHYSICAL,
												Eid:      "eid2",
												TestEsim: false,
												ProfileInfo: []*lab.SIMProfileInfo{
													{
														Iccid:       "iccid3",
														SimPin:      "pin3",
														SimPuk:      "puk3",
														CarrierName: lab.NetworkProvider_NETWORK_VERIZON,
														OwnNumber:   "number3",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Machine: &ufspb.Machine{
				Name: "mary",
				Device: &ufspb.Machine_ChromeosMachine{
					ChromeosMachine: &ufspb.ChromeOSMachine{
						BuildTarget: "build-target",
						Model:       "model",
					},
				},
			},
			ManufacturingConfig: &manufacturing.ManufacturingConfig{
				HwidComponent: []string{
					"fake-component1",
					"fake-component2",
				},
			},
			HwidData: &ufspb.HwidData{
				Sku:  "fake-sku",
				Hwid: "fake-hwid",
				DutLabel: &ufspb.DutLabel{
					Labels: []*ufspb.DutLabel_Label{
						{
							Name:  "phase",
							Value: "EVT-Maple",
						},
					},
				},
			},
			DutState: &lab.DutState{
				Chameleon: lab.PeripheralState_NOT_APPLICABLE,
			},
		},
		CachingServices: &ufsapi.ListCachingServicesResponse{
			CachingServices: []*ufspb.CachingService{
				{
					Name:           "cachingservice/200.200.200.208",
					Port:           55,
					ServingSubnets: []string{"200.200.200.200/24"},
					State:          ufspb.State_STATE_SERVING,
				},
			},
		},
	}
	cl := cache.NewLocator()
	c := newFakeClient(ctx, t, s)
	inventory := NewInventory(c, cl)
	got, err := inventory.GetDutTopology(ctx, "alice")
	if err != nil {
		t.Fatal(err)
	}
	want := &labapi.DutTopology{
		Id: &labapi.DutTopology_Id{Value: "alice"},
		Duts: []*labapi.Dut{
			{
				Id: &labapi.Dut_Id{Value: "200.200.200.200"},
				DutType: &labapi.Dut_Chromeos{
					Chromeos: &labapi.Dut_ChromeOS{
						Audio: &labapi.Audio{
							AudioBox: true,
							Atrus:    true,
						},
						Chameleon: &labapi.Chameleon{
							AudioBoard:  true,
							Peripherals: []labapi.Chameleon_Peripheral{labapi.Chameleon_DP},
							Types:       []labapi.Chameleon_Type{labapi.Chameleon_V3},
							State:       labapi.PeripheralState_NOT_APPLICABLE,
						},
						Servo: &labapi.Servo{
							Present: true,
							ServodAddress: &labapi.IpEndpoint{
								Address: "servo_host",
								Port:    33,
							},
						},
						Ssh: &labapi.IpEndpoint{
							Address: "200.200.200.200",
							Port:    22,
						},
						Wifi: &labapi.Wifi{
							Environment: labapi.Wifi_WIFI_CELL,
							Antenna: &labapi.WifiAntenna{
								Connection: labapi.WifiAntenna_CONDUCTIVE,
							},
						},
						Touch: &labapi.Touch{
							Mimo: true,
						},
						Camerabox: &labapi.Camerabox{
							Facing: labapi.Camerabox_FRONT,
						},
						Cables: []*labapi.Cable{
							{
								Type: labapi.Cable_AUDIOJACK,
							},
						},
						DutModel: &labapi.DutModel{
							BuildTarget: "build-target",
							ModelName:   "model",
						},
						HwidComponent: []string{
							"fake-component1",
							"fake-component2",
						},
						BluetoothPeers: []*labapi.BluetoothPeer{
							{
								Hostname: "test-btp1",
								State:    labapi.PeripheralState_WORKING,
							},
							{
								Hostname: "test-btp2",
								State:    labapi.PeripheralState_BROKEN,
							},
							{
								Hostname: "test-btp3",
								State:    labapi.PeripheralState_PERIPHERAL_STATE_UNSPECIFIED,
							},
						},
						Sku:   "fake-sku",
						Hwid:  "fake-hwid",
						Phase: labapi.Phase_EVT_MAPLE,
						SimInfos: []*labapi.SIMInfo{
							{
								SlotId:   1,
								Type:     labapi.SIMType_SIM_DIGITAL,
								Eid:      "eid1",
								TestEsim: true,
								ProfileInfo: []*labapi.SIMProfileInfo{
									{
										Iccid:       "iccid1",
										SimPin:      "pin1",
										SimPuk:      "puk1",
										CarrierName: labapi.NetworkProvider_NETWORK_ATT,
										OwnNumber:   "number1",
									},
									{
										Iccid:       "iccid2",
										SimPin:      "pin2",
										SimPuk:      "puk2",
										CarrierName: labapi.NetworkProvider_NETWORK_TMOBILE,
										OwnNumber:   "number2",
									},
								},
							},
							{
								SlotId:   2,
								Type:     labapi.SIMType_SIM_PHYSICAL,
								Eid:      "eid2",
								TestEsim: false,
								ProfileInfo: []*labapi.SIMProfileInfo{
									{
										Iccid:       "iccid3",
										SimPin:      "pin3",
										SimPuk:      "puk3",
										CarrierName: labapi.NetworkProvider_NETWORK_VERIZON,
										OwnNumber:   "number3",
									},
								},
							},
						},
					},
				},
				CacheServer: &labapi.CacheServer{
					Address: &labapi.IpEndpoint{
						Address: "200.200.200.208",
						Port:    55,
					},
				},
			},
		},
	}
	if !proto.Equal(want, got) {
		t.Errorf("GetDutTopology() mismatch (-want +got):\n%s\n%s", want, got)
	}
}

func TestGetAndroidDutTopology_single(t *testing.T) {
	t.Parallel()
	ctx, cf := context.WithCancel(context.Background())
	defer cf()
	hostname := "dummy_hostname"
	associatedHostname := "dummy_associated_hostname"
	serialNumber := "1234567890"
	buildTarget := "dummy_build_target"
	model := "dummy_model"
	dutTopologyId := "dummy_android_dut_topology_id"
	s := &fakeServer{
		AttachedDeviceData: &ufsapi.AttachedDeviceData{
			LabConfig: &ufspb.MachineLSE{
				Hostname: hostname,
				Lse: &ufspb.MachineLSE_AttachedDeviceLse{
					AttachedDeviceLse: &ufspb.AttachedDeviceLSE{
						OsVersion: &ufspb.OSVersion{
							Value:       "dummy_value",
							Description: "dummy_description",
							Image:       "dummy_image",
						},
						AssociatedHostname: associatedHostname,
					},
				},
			},
			Machine: &ufspb.Machine{
				SerialNumber: serialNumber,
				Device: &ufspb.Machine_AttachedDevice{
					AttachedDevice: &ufspb.AttachedDevice{
						BuildTarget: buildTarget,
						Model:       model,
					},
				},
			},
		},
	}
	c := newFakeClient(ctx, t, s)
	inventory := NewInventory(c, cache.NewLocator())
	got, err := inventory.GetDutTopology(ctx, dutTopologyId)
	if err != nil {
		t.Fatal(err)
	}
	want := &labapi.DutTopology{
		Id: &labapi.DutTopology_Id{Value: dutTopologyId},
		Duts: []*labapi.Dut{
			{
				Id: &labapi.Dut_Id{Value: hostname},
				DutType: &labapi.Dut_Android_{
					Android: &labapi.Dut_Android{
						AssociatedHostname: &labapi.IpEndpoint{
							Address: associatedHostname,
						},
						Name:         hostname,
						SerialNumber: serialNumber,
						DutModel: &labapi.DutModel{
							BuildTarget: buildTarget,
							ModelName:   model,
						},
					},
				},
			},
		},
	}
	if !proto.Equal(want, got) {
		t.Errorf("GetDutTopology() mismatch (-want +got):\n%s\n%s", want, got)
	}
}

func TestGetChromeOsDevboardTopology_single(t *testing.T) {
	t.Parallel()
	ctx, cf := context.WithCancel(context.Background())
	defer cf()
	s := &fakeServer{
		ChromeOSDeviceData: &ufspb.ChromeOSDeviceData{
			LabConfig: &ufspb.MachineLSE{
				Hostname: "200.200.200.200",
				Lse: &ufspb.MachineLSE_ChromeosMachineLse{
					ChromeosMachineLse: &ufspb.ChromeOSMachineLSE{
						ChromeosLse: &ufspb.ChromeOSMachineLSE_DeviceLse{
							DeviceLse: &ufspb.ChromeOSDeviceLSE{
								Device: &ufspb.ChromeOSDeviceLSE_Devboard{
									Devboard: &lab.Devboard{
										Pools: []string{"test-pool"},
										Servo: &lab.Servo{
											ServoHostname: "servo-host",
											ServoPort:     33,
										},
									},
								},
							},
						},
					},
				},
			},
			Machine: &ufspb.Machine{
				Name: "mary",
				Device: &ufspb.Machine_Devboard{
					Devboard: &ufspb.Devboard{
						Board: &ufspb.Devboard_Andreiboard{
							Andreiboard: &ufspb.Andreiboard{
								UltradebugSerial: "fake-serial",
							},
						},
					},
				},
			},
			ManufacturingConfig: &manufacturing.ManufacturingConfig{
				HwidComponent: []string{
					"fake-component1",
					"fake-component2",
				},
			},
			HwidData: &ufspb.HwidData{
				Sku:  "fake-sku",
				Hwid: "fake-hwid",
				DutLabel: &ufspb.DutLabel{
					Labels: []*ufspb.DutLabel_Label{
						{
							Name:  "phase",
							Value: "EVT-Maple",
						},
					},
				},
			},
			DutState: &lab.DutState{
				Chameleon: lab.PeripheralState_NOT_APPLICABLE,
			},
		},
		CachingServices: &ufsapi.ListCachingServicesResponse{
			CachingServices: []*ufspb.CachingService{
				{
					Name:           "cachingservice/200.200.200.208",
					Port:           55,
					ServingSubnets: []string{"200.200.200.200/24"},
					State:          ufspb.State_STATE_SERVING,
				},
			},
		},
	}
	cl := cache.NewLocator()
	c := newFakeClient(ctx, t, s)
	inventory := NewInventory(c, cl)
	got, err := inventory.GetDutTopology(ctx, "alice")
	if err != nil {
		t.Fatal(err)
	}
	want := &labapi.DutTopology{
		Id: &labapi.DutTopology_Id{Value: "alice"},
		Duts: []*labapi.Dut{
			{
				Id: &labapi.Dut_Id{Value: "200.200.200.200"},
				DutType: &labapi.Dut_Devboard_{
					Devboard: &labapi.Dut_Devboard{
						BoardType:        "andreiboard",
						UltradebugSerial: "fake-serial",
						Servo: &labapi.Servo{
							Present: true,
							ServodAddress: &labapi.IpEndpoint{
								Address: "servo-host",
								Port:    33,
							},
						},
					},
				},
				CacheServer: &labapi.CacheServer{
					Address: &labapi.IpEndpoint{
						Address: "200.200.200.208",
						Port:    55,
					},
				},
			},
		},
	}
	if !proto.Equal(want, got) {
		t.Errorf("GetDutTopology() mismatch (-want +got):\n%s\n%s", want, got)
	}
}

type fakeServer struct {
	ufsapi.UnimplementedFleetServer
	ChromeOSDeviceData *ufspb.ChromeOSDeviceData
	AttachedDeviceData *ufsapi.AttachedDeviceData
	CachingServices    *ufsapi.ListCachingServicesResponse
}

func (s *fakeServer) GetDeviceData(ctx context.Context, in *ufsapi.GetDeviceDataRequest) (*ufsapi.GetDeviceDataResponse, error) {
	if s.ChromeOSDeviceData != nil {
		return &ufsapi.GetDeviceDataResponse{
			Resource: &ufsapi.GetDeviceDataResponse_ChromeOsDeviceData{
				ChromeOsDeviceData: proto.Clone(s.ChromeOSDeviceData).(*ufspb.ChromeOSDeviceData),
			},
			ResourceType: ufsapi.GetDeviceDataResponse_RESOURCE_TYPE_CHROMEOS_DEVICE,
		}, nil
	}
	return &ufsapi.GetDeviceDataResponse{
		Resource: &ufsapi.GetDeviceDataResponse_AttachedDeviceData{
			AttachedDeviceData: proto.Clone(s.AttachedDeviceData).(*ufsapi.AttachedDeviceData),
		},
		ResourceType: ufsapi.GetDeviceDataResponse_RESOURCE_TYPE_ATTACHED_DEVICE,
	}, nil
}

func (s *fakeServer) ListCachingServices(ctx context.Context, in *ufsapi.ListCachingServicesRequest) (*ufsapi.ListCachingServicesResponse, error) {
	return proto.Clone(s.CachingServices).(*ufsapi.ListCachingServicesResponse), nil
}

// Make a fake client for testing.
// Cancel the context to clean up the fake server and client.
func newFakeClient(ctx context.Context, t *testing.T, s ufsapi.FleetServer) ufsapi.FleetClient {
	gs := grpc.NewServer()
	ufsapi.RegisterFleetServer(gs, s)
	l := bufconn.Listen(4096)
	go gs.Serve(l)
	go func() {
		<-ctx.Done()
		// This also closes the listener.
		gs.Stop()
	}()
	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(),
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return l.Dial() }))
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		<-ctx.Done()
		conn.Close()
	}()
	return ufsapi.NewFleetClient(conn)
}
