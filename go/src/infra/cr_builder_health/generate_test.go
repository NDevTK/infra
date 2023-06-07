// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"testing"

	buildbucketpb "go.chromium.org/luci/buildbucket/proto"

	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"

	. "github.com/smartystreets/goconvey/convey"
)

type bbClientMock struct {
	setHealthCalls int
}

var existantBuilder = "existant-builder"
var nonExistantBuilder = "non-existant-builder"

func (c *bbClientMock) SetBuilderHealth(ctx context.Context, in *buildbucketpb.SetBuilderHealthRequest, opts ...grpc.CallOption) (*buildbucketpb.SetBuilderHealthResponse, error) {
	c.setHealthCalls += 1
	result := &buildbucketpb.SetBuilderHealthResponse{
		Responses: []*buildbucketpb.SetBuilderHealthResponse_Response{},
	}

	for _, req := range in.Health {
		if req.Id.Builder == existantBuilder {
			result.Responses = append(result.Responses, &buildbucketpb.SetBuilderHealthResponse_Response{
				Response: &buildbucketpb.SetBuilderHealthResponse_Response_Result{},
			})
		} else {
			result.Responses = append(result.Responses, &buildbucketpb.SetBuilderHealthResponse_Response{
				Response: &buildbucketpb.SetBuilderHealthResponse_Response_Error{
					Error: &status.Status{
						Code:    400,
						Message: "Invalid builder name",
					},
				},
			})
		}
	}

	return result, nil
}

func TestGenerate(t *testing.T) {
	t.Parallel()

	Convey("RPC Buildbucket is called ok", t, func() {
		ctx := context.Background()
		client := &bbClientMock{}
		rows := []Row{{
			Bucket:         "bucket",
			Builder:        existantBuilder,
			BuildMinsP50:   59,
			BuildMinsP95:   119,
			PendingMinsP50: 59,
			PendingMinsP95: 119,
			FailRate:       0.05,
			InfraFailRate:  0,
		}}
		err := rpcBuildbucket(ctx, rows, client)
		So(client.setHealthCalls, ShouldEqual, 1)
		So(ctx.Err(), ShouldBeNil)
		So(err, ShouldBeNil)
	})

	Convey("RPC Buildbucket is called error", t, func() {
		ctx := context.Background()
		client := &bbClientMock{}
		rows := []Row{
			{
				Bucket:         "bucket",
				Builder:        nonExistantBuilder,
				BuildMinsP50:   59,
				BuildMinsP95:   119,
				PendingMinsP50: 59,
				PendingMinsP95: 119,
				FailRate:       0.05,
				InfraFailRate:  0,
			},
			{
				Bucket:         "bucket",
				Builder:        existantBuilder,
				BuildMinsP50:   59,
				BuildMinsP95:   119,
				PendingMinsP50: 59,
				PendingMinsP95: 119,
				FailRate:       0.05,
				InfraFailRate:  0,
			},
		}
		err := rpcBuildbucket(ctx, rows, client)
		So(err.Error(), ShouldContainSubstring, "1")
		So(client.setHealthCalls, ShouldEqual, 1)
		So(ctx.Err(), ShouldBeNil)
	})
}
