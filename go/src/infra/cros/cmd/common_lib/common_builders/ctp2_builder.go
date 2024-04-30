// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common_builders

import (
	"context"

	"golang.org/x/exp/maps"

	testapi "go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform"

	"infra/cros/cmd/common_lib/interfaces"
)

type ManifestFetcher func(ctx context.Context, s string) (string, error)

type CTPV2FromV1 struct {
	interfaces.CTPv2Builder

	ctx             context.Context
	v2              *testapi.CTPv2Request
	v1              []*test_platform.Request
	manifestFetcher ManifestFetcher
}

func NewCTPV2FromV1(ctx context.Context, v1 map[string]*test_platform.Request) *CTPV2FromV1 {
	return &CTPV2FromV1{
		v1: maps.Values(v1),
		v2: &testapi.CTPv2Request{
			Requests: []*testapi.CTPRequest{},
		},
		ctx:             ctx,
		manifestFetcher: GetBuilderManifestFromContainer,
	}
}

func NewCTPV2FromV1WithCustomManifestFetcher(ctx context.Context, v1 map[string]*test_platform.Request, manifestFetcher ManifestFetcher) *CTPV2FromV1 {
	if manifestFetcher == nil {
		manifestFetcher = GetBuilderManifestFromContainer
	}
	return &CTPV2FromV1{
		v1: maps.Values(v1),
		v2: &testapi.CTPv2Request{
			Requests: []*testapi.CTPRequest{},
		},
		ctx:             ctx,
		manifestFetcher: manifestFetcher,
	}
}

func (builder *CTPV2FromV1) BuildRequest() *testapi.CTPv2Request {
	for _, v1Request := range builder.v1 {
		builder.v2.Requests = append(builder.v2.Requests, buildCTPRequest(v1Request))
	}

	builder.v2.Requests = GroupV2Requests(builder.ctx, builder.v2.Requests, builder.manifestFetcher)
	return builder.v2
}
