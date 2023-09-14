// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package bb implements a BuildBucket.Client using calls to BuildBucket.
package bb

import (
	"context"
	"net/http"
	"time"

	"go.chromium.org/chromiumos/infra/proto/go/test_platform/result_flow"
	pb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/retry"
	"go.chromium.org/luci/common/retry/transient"
	"go.chromium.org/luci/grpc/prpc"
	"google.golang.org/genproto/protobuf/field_mask"
)

// Client defines an interface used to interact with Buildbucket.
type Client interface {
	GetTargetBuilds(context.Context, []int64) ([]*pb.Build, error)
}

type resultFlowBBClient struct {
	client        pb.BuildsClient
	requestFields []string
	br            *pb.BatchResponse
}

// NewClient creates a new Client to interact with BuildBucket.
func NewClient(ctx context.Context, conf *result_flow.BuildbucketConfig, fields []string, h *http.Client) (Client, error) {
	b := pb.NewBuildsPRPCClient(&prpc.Client{
		C:    h,
		Host: conf.Host,
	})

	return &resultFlowBBClient{
		client:        b,
		requestFields: fields,
	}, nil
}

// GetTargetBuilds gets the pb.Build object from BuildBucket API via a batch request.
func (c *resultFlowBBClient) GetTargetBuilds(ctx context.Context, bIDs []int64) ([]*pb.Build, error) {
	if len(bIDs) == 0 {
		return nil, nil
	}
	// Retry the batch buildbucket call upon errors.
	err := retry.Retry(
		ctx,
		transient.Only(retry.Default),
		func() error {
			var berr error
			c.br, berr = c.client.Batch(ctx, batchRequest(c.requestFields, bIDs))
			if berr != nil {
				// All requests in the batch failed. Assume they are all transient errors.
				return transient.Tag.Apply(berr)
			}
			return nil
		},
		func(err error, d time.Duration) {
			logging.Warningf(
				ctx,
				"Transient error calling Buildbucket: %v. Retrying... with delay %s",
				err,
				d.String(),
			)
		})
	if err != nil {
		return nil, err
	}

	var builds []*pb.Build
	for _, r := range c.br.Responses {
		if _, ok := r.Response.(*pb.BatchResponse_Response_Error); ok {
			logging.Errorf(ctx, "failed to read a single build, err: %v", r)
			continue
		}
		res := r.Response.(*pb.BatchResponse_Response_GetBuild)
		builds = append(builds, res.GetBuild)
	}
	return builds, nil
}

func batchRequest(f []string, bIDs []int64) *pb.BatchRequest {
	var r []*pb.BatchRequest_Request
	for _, bID := range bIDs {
		r = append(r, &pb.BatchRequest_Request{
			Request: &pb.BatchRequest_Request_GetBuild{
				GetBuild: &pb.GetBuildRequest{
					Id:     bID,
					Fields: &field_mask.FieldMask{Paths: f},
				},
			},
		})
	}
	return &pb.BatchRequest{Requests: r}
}
