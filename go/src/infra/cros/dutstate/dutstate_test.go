// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dutstate

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	ufsProto "infra/unifiedfleet/api/v1/models"
	ufslab "infra/unifiedfleet/api/v1/models/chromeos/lab"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	"infra/unifiedfleet/app/util"
	ufsUtil "infra/unifiedfleet/app/util"
)

type FakeUFSClient struct {
	getStateMap    map[string]ufsProto.State
	getStateErr    error
	updateStateMap map[string]ufsProto.State
	updateStateErr error
}

func TestReadState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("Read state from USF", t, func() {
		c := &FakeUFSClient{
			getStateMap: map[string]ufsProto.State{
				"os:machineLSEs/host1":         ufsProto.State_STATE_REPAIR_FAILED,
				"os:machineLSEs/host2":         ufsProto.State_STATE_DEPLOYED_TESTING,
				"os-partner:machineLSEs/host1": ufsProto.State_STATE_DEPLOYED_PRE_SERVING,
			},
		}
		// no namespace - should default to `os`
		r := Read(ctx, c, "host1")
		So(r.State, ShouldEqual, "repair_failed")
		So(r.Time, ShouldNotEqual, 0)

		r = Read(ctx, c, "host2")
		So(r.State, ShouldEqual, "manual_repair")
		So(r.Time, ShouldNotEqual, 0)

		r = Read(ctx, c, "not_found")
		So(r.State, ShouldEqual, "unknown")
		So(r.Time, ShouldEqual, 0)

		r = Read(ctx, c, "fail")
		So(r.State, ShouldEqual, "unknown")
		So(r.Time, ShouldEqual, 0)

		// explicitly set os context, should give the same results
		osCtx := ctxWithNamespace(ufsUtil.OSNamespace)
		r = Read(osCtx, c, "host1")
		So(r.State, ShouldEqual, "repair_failed")
		So(r.Time, ShouldNotEqual, 0)

		// explicitly set partner context, should fetch a different DUT
		partnerCtx := ctxWithNamespace(ufsUtil.OSPartnerNamespace)
		r = Read(partnerCtx, c, "host1")
		So(r.State, ShouldEqual, "needs_deploy")
		So(r.Time, ShouldNotEqual, 0)
	})
}

func TestUpdateState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("Read state from USF", t, func() {
		c := &FakeUFSClient{
			updateStateMap: map[string]ufsProto.State{},
		}

		// dont explicitly set context
		Convey("set repair_failed and expect REPAIR_FAILED", func() {
			e := Update(ctx, c, "host1", "repair_failed")
			So(e, ShouldBeNil)
			So(c.updateStateMap, ShouldHaveLength, 1)
			So(c.updateStateMap["os:machineLSEs/host1"], ShouldEqual, ufsProto.State_STATE_REPAIR_FAILED)
		})

		Convey("set manual_repair and expect DEPLOYED_TESTING", func() {
			e := Update(ctx, c, "host2", "manual_repair")
			So(e, ShouldBeNil)
			So(c.updateStateMap, ShouldHaveLength, 1)
			So(c.updateStateMap["os:machineLSEs/host2"], ShouldEqual, ufsProto.State_STATE_DEPLOYED_TESTING)
		})

		Convey("set incorrect state and expect UNSPECIFIED for UFS", func() {
			e := Update(ctx, c, "host2", "wrong_state")
			So(e, ShouldBeNil)
			So(c.updateStateMap, ShouldHaveLength, 1)
			So(c.updateStateMap["os:machineLSEs/host2"], ShouldEqual, ufsProto.State_STATE_UNSPECIFIED)
		})

		// explicitly set os context and expect same result as default
		Convey("set repair_failed and expect REPAIR_FAILED in os namespace", func() {
			osCtx := ctxWithNamespace(ufsUtil.OSNamespace)
			e := Update(osCtx, c, "host1", "repair_failed")
			So(e, ShouldBeNil)
			So(c.updateStateMap, ShouldHaveLength, 1)
			So(c.updateStateMap["os:machineLSEs/host1"], ShouldEqual, ufsProto.State_STATE_REPAIR_FAILED)
		})

		// update DUT in separate namespace, should touch a different machine
		Convey("set state in separate namespace", func() {
			partnerCtx := ctxWithNamespace(ufsUtil.OSPartnerNamespace)
			e := Update(partnerCtx, c, "host1", "manual_repair")
			So(e, ShouldBeNil)
			So(c.updateStateMap, ShouldHaveLength, 1)
			So(c.updateStateMap["os-partner:machineLSEs/host1"], ShouldEqual, ufsProto.State_STATE_DEPLOYED_TESTING)
		})
	})
}

func TestConvertToUFSState(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		in  State
		out ufsProto.State
	}{
		{
			State("ready"),
			ufsProto.State_STATE_SERVING,
		},
		{
			State("repair_failed"),
			ufsProto.State_STATE_REPAIR_FAILED,
		},
		{
			State("Ready "),
			ufsProto.State_STATE_UNSPECIFIED,
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(string(tc.in), func(t *testing.T) {
			t.Parallel()
			got := ConvertToUFSState(tc.in)
			if diff := cmp.Diff(tc.out, got); diff != "" {
				t.Errorf("output mismatch (-want +got): %s\n", diff)
			}
		})
	}
}

func TestConvertFromUFSState(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		in  ufsProto.State
		out State
	}{
		{
			ufsProto.State_STATE_SERVING,
			State("ready"),
		},
		{
			ufsProto.State_STATE_DEPLOYED_PRE_SERVING,
			State("needs_deploy"),
		},
		{
			ufsProto.State_STATE_UNSPECIFIED,
			State("unknown"),
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.in.String(), func(t *testing.T) {
			t.Parallel()
			got := ConvertFromUFSState(tc.in)
			if diff := cmp.Diff(tc.out, got); diff != "" {
				t.Errorf("output mismatch (-want +got): %s\n", diff)
			}
		})
	}
}

func TestStateString(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		in  State
		out string
	}{
		{
			Ready,
			"ready",
		},
		{
			NeedsRepair,
			"needs_repair",
		},
		{
			NeedsReset,
			"needs_reset",
		},
		{
			Reserved,
			"reserved",
		},
		{
			State("Some custom"),
			"Some custom",
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.in.String(), func(t *testing.T) {
			t.Parallel()
			got := tc.in.String()
			if diff := cmp.Diff(tc.out, got); diff != "" {
				t.Errorf("output mismatch (-want +got): %s\n", diff)
			}
		})
	}
}

func (c *FakeUFSClient) GetMachineLSE(ctx context.Context, req *ufsAPI.GetMachineLSERequest, opts ...grpc.CallOption) (*ufsProto.MachineLSE, error) {
	// we can use ns_and_name to look for DUTs (instead of just name) so that
	// we can test the client has the right context set during tests
	ns, err := fetchNamespaceFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if c.getStateErr == nil {
		ns_and_name := fmt.Sprintf("%s:%s", ns, req.GetName())
		if ns_and_name == "os:machineLSEs/fail" {
			return nil, status.Error(codes.Unknown, "Somthing else")
		}
		if ns_and_name == "os:machineLSEs/not_found" {
			return nil, status.Error(codes.NotFound, "not_found")
		}
		if ns_and_name == "os:machineLSEs/host1" {
			return &ufsProto.MachineLSE{
				Name:          req.GetName(),
				ResourceState: c.getStateMap[ns_and_name],
				UpdateTime:    timestamppb.Now(),
				Lse: &ufsProto.MachineLSE_ChromeosMachineLse{
					ChromeosMachineLse: &ufsProto.ChromeOSMachineLSE{
						ChromeosLse: &ufsProto.ChromeOSMachineLSE_DeviceLse{
							DeviceLse: &ufsProto.ChromeOSDeviceLSE{
								Device: &ufsProto.ChromeOSDeviceLSE_Dut{
									Dut: &ufslab.DeviceUnderTest{},
								},
							},
						},
					},
				},
				Machines: []string{"1"},
			}, nil
		}
		if ns_and_name == "os:machineLSEs/host2" {
			return &ufsProto.MachineLSE{
				Name:          req.GetName(),
				ResourceState: c.getStateMap[ns_and_name],
				UpdateTime:    timestamppb.Now(),
				Lse: &ufsProto.MachineLSE_ChromeosMachineLse{
					ChromeosMachineLse: &ufsProto.ChromeOSMachineLSE{
						ChromeosLse: &ufsProto.ChromeOSMachineLSE_DeviceLse{
							DeviceLse: &ufsProto.ChromeOSDeviceLSE{
								Device: &ufsProto.ChromeOSDeviceLSE_Labstation{
									Labstation: &ufslab.Labstation{},
								},
							},
						},
					},
				},
				Machines: []string{"2"},
			}, nil
		}
		// note this is the same hostname as above case, with a different namespace
		if ns_and_name == "os-partner:machineLSEs/host1" {
			return &ufsProto.MachineLSE{
				Name:          req.GetName(),
				ResourceState: c.getStateMap[ns_and_name],
				UpdateTime:    timestamppb.Now(),
				Lse: &ufsProto.MachineLSE_ChromeosMachineLse{
					ChromeosMachineLse: &ufsProto.ChromeOSMachineLSE{
						ChromeosLse: &ufsProto.ChromeOSMachineLSE_DeviceLse{
							DeviceLse: &ufsProto.ChromeOSDeviceLSE{
								Device: &ufsProto.ChromeOSDeviceLSE_Labstation{
									Labstation: &ufslab.Labstation{},
								},
							},
						},
					},
				},
				Machines: []string{"p1"},
			}, nil
		}
	}
	return nil, c.getStateErr
}

func (c *FakeUFSClient) UpdateMachineLSE(ctx context.Context, req *ufsAPI.UpdateMachineLSERequest, opts ...grpc.CallOption) (*ufsProto.MachineLSE, error) {
	// we can use ns_and_name to look for DUTs (instead of just name) so that
	// we can test the client has the right context set during tests
	ns, err := fetchNamespaceFromContext(ctx)
	if err != nil {
		return nil, err
	}
	ns_and_name := fmt.Sprintf("%s:%s", ns, req.GetMachineLSE().GetName())

	if c.updateStateErr == nil {
		c.updateStateMap[ns_and_name] = req.GetMachineLSE().GetResourceState()
		if ns_and_name == "os:machineLSEs/host1" {
			return &ufsProto.MachineLSE{
				Name:          req.GetMachineLSE().GetName(),
				ResourceState: req.GetMachineLSE().GetResourceState(),
				UpdateTime:    timestamppb.Now(),
				Lse: &ufsProto.MachineLSE_ChromeosMachineLse{
					ChromeosMachineLse: &ufsProto.ChromeOSMachineLSE{
						ChromeosLse: &ufsProto.ChromeOSMachineLSE_DeviceLse{
							DeviceLse: &ufsProto.ChromeOSDeviceLSE{
								Device: &ufsProto.ChromeOSDeviceLSE_Dut{
									Dut: &ufslab.DeviceUnderTest{},
								},
							},
						},
					},
				},
			}, nil
		}
		if ns_and_name == "os:machineLSEs/host2" {
			return &ufsProto.MachineLSE{
				Name:          req.GetMachineLSE().GetName(),
				ResourceState: req.GetMachineLSE().GetResourceState(),
				UpdateTime:    timestamppb.Now(),
				Lse: &ufsProto.MachineLSE_ChromeosMachineLse{
					ChromeosMachineLse: &ufsProto.ChromeOSMachineLSE{
						ChromeosLse: &ufsProto.ChromeOSMachineLSE_DeviceLse{
							DeviceLse: &ufsProto.ChromeOSDeviceLSE{
								Device: &ufsProto.ChromeOSDeviceLSE_Labstation{
									Labstation: &ufslab.Labstation{},
								},
							},
						},
					},
				},
			}, nil
		}
		if ns_and_name == "os-partner:machineLSEs/host1" {
			return &ufsProto.MachineLSE{
				Name:          req.GetMachineLSE().GetName(),
				ResourceState: req.GetMachineLSE().GetResourceState(),
				UpdateTime:    timestamppb.Now(),
				Lse: &ufsProto.MachineLSE_ChromeosMachineLse{
					ChromeosMachineLse: &ufsProto.ChromeOSMachineLSE{
						ChromeosLse: &ufsProto.ChromeOSMachineLSE_DeviceLse{
							DeviceLse: &ufsProto.ChromeOSDeviceLSE{
								Device: &ufsProto.ChromeOSDeviceLSE_Dut{
									Dut: &ufslab.DeviceUnderTest{},
								},
							},
						},
					},
				},
			}, nil
		}
	}
	return nil, c.updateStateErr
}

func fetchNamespaceFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unknown, "failed getting metadata")
	}
	namespace, ok := md[util.Namespace]
	if !ok {
		return "", status.Error(codes.Unknown, "no namespace in metadata")
	}

	return namespace[0], nil
}

func ctxWithNamespace(ns string) context.Context {
	ctx := context.Background()
	newMetadata := metadata.Pairs(ufsUtil.Namespace, ns)
	return metadata.NewOutgoingContext(ctx, newMetadata)
}
