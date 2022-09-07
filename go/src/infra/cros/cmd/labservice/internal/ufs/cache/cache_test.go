// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cache

import (
	"context"
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

type fakeClient struct {
	ufsapi.FleetClient
	CachingServices *ufsapi.ListCachingServicesResponse
}

func (s fakeClient) ListCachingServices(context.Context, *ufsapi.ListCachingServicesRequest, ...grpc.CallOption) (*ufsapi.ListCachingServicesResponse, error) {
	return proto.Clone(s.CachingServices).(*ufsapi.ListCachingServicesResponse), nil
}
