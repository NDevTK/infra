// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cache

import (
	"context"
	"fmt"
	"testing"

	ufspb "infra/unifiedfleet/api/v1/models"
	ufsapi "infra/unifiedfleet/api/v1/rpc"

	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

func TestFindCacheServer_single(t *testing.T) {
	t.Parallel()

	c := &fakeClient{
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

	locator := NewLocator()
	got, err := locator.FindCacheServer("200.200.200.201", c)
	if err != nil {
		t.Fatal(err)
	}
	want := &labapi.IpEndpoint{
		Address: "200.200.200.208",
		Port:    55,
	}
	if !proto.Equal(want, got) {
		t.Errorf("FindCacheServer() mismatch (-want +got):\n%s\n%s", want, got)
	}
}

func TestFindCacheServer_zone(t *testing.T) {
	t.Parallel()

	c := &fakeClient{
		CachingServices: &ufsapi.ListCachingServicesResponse{
			CachingServices: []*ufspb.CachingService{
				{
					Name:  "cachingservice/200.200.200.208",
					Port:  55,
					Zones: []ufspb.Zone{ufspb.Zone_ZONE_CHROMEOS2},
					State: ufspb.State_STATE_SERVING,
				},
				{
					Name:  "cachingservice/100.100.100.108",
					Port:  55,
					Zones: []ufspb.Zone{ufspb.Zone_ZONE_SFO36_OS},
					State: ufspb.State_STATE_SERVING,
				},
			},
		},
		MachineLSEs: map[string]*ufspb.MachineLSE{
			"machineLSEs/SU-name": {
				Zone: "ZONE_SFO36_OS",
			},
		},
	}

	locator := NewLocator()
	got, err := locator.FindCacheServer("SU-name", c)
	if err != nil {
		t.Fatal(err)
	}
	want := &labapi.IpEndpoint{
		Address: "100.100.100.108",
		Port:    55,
	}
	if !proto.Equal(want, got) {
		t.Errorf("FindCacheServer() mismatch (-want +got):\n%s\n%s", want, got)
	}
}

func TestParseAddress_happy(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		addr string
		want address
	}{
		"scheme and server": {
			addr: "http://100.100.100.108",
			want: address{Ip: "100.100.100.108", Port: 80},
		},
		"server and port": {
			addr: "server:8000",
			want: address{Ip: "server", Port: 8000},
		},
		"scheme server and port": {
			addr: "http://server:8001",
			want: address{Ip: "server", Port: 8001},
		},
	}
	for n, tc := range cases {
		tc := tc
		t.Run(n, func(t *testing.T) {
			got, err := parseAddress(tc.addr)
			if err != nil {
				t.Errorf("parseAddress(%q) err %v, want %v", tc, err, nil)
			}
			if *got != tc.want {
				t.Errorf("parseAddress(%q) got %q, want %q", tc, got, tc.want)
			}
		})
	}
}

func TestParseAddress_errors(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"empty string":            "",
		"emtty string 2":          "    ",
		"neither scheme nor port": "server",
		"port is not a number":    "server:port",
		"extra part":              "server:port:more",
	}
	for n, tc := range cases {
		tc := tc
		t.Run(n, func(t *testing.T) {
			t.Parallel()
			_, err := parseAddress(tc)
			if err == nil {
				t.Errorf("parseAddress(%q) error nil, want not nil", tc)
			}
		})
	}
}

type fakeClient struct {
	ufsapi.FleetClient
	CachingServices *ufsapi.ListCachingServicesResponse
	MachineLSEs     map[string]*ufspb.MachineLSE
	Machines        map[string]*ufspb.Machine
}

func (s fakeClient) ListCachingServices(context.Context, *ufsapi.ListCachingServicesRequest, ...grpc.CallOption) (*ufsapi.ListCachingServicesResponse, error) {
	return proto.Clone(s.CachingServices).(*ufsapi.ListCachingServicesResponse), nil
}

func (s fakeClient) GetMachineLSE(_ context.Context, req *ufsapi.GetMachineLSERequest, o ...grpc.CallOption) (*ufspb.MachineLSE, error) {
	if e, ok := s.MachineLSEs[req.GetName()]; ok {
		return e, nil
	}
	return nil, fmt.Errorf("zone for %q not found", req.GetName())
}

func (s fakeClient) GetMachine(_ context.Context, req *ufsapi.GetMachineRequest, o ...grpc.CallOption) (*ufspb.Machine, error) {
	if e, ok := s.Machines[req.GetName()]; ok {
		return e, nil
	}
	return nil, fmt.Errorf("zone for %q not found", req.GetName())

}
