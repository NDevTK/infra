// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package bb defines an interface for interacting with buildbucket.
package bb

import (
	"context"

	"infra/cmd/cros_test_platform/internal/site"

	"github.com/golang/protobuf/jsonpb"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"go.chromium.org/luci/auth"
	buildbucket_pb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/grpc/prpc"
	"google.golang.org/genproto/protobuf/field_mask"
)

// Client defines an interface used to interact with a buildbucket service.
type Client interface {
	ScheduleBuild(ctx context.Context, in *buildbucket_pb.ScheduleBuildRequest) (*Build, error)
	GetBuild(ctx context.Context, ID int64) (*Build, error)
}

// Client defines an interface used to interact with a buildbucket service.
type client struct {
	client buildbucket_pb.BuildsClient
}

// Build contains selected information from a buildbucket Build.
type Build struct {
	ID     int64
	Status buildbucket_pb.Status
	Result *skylab_test_runner.Result
}

// NewClient returns a new client to interact with buildbucket builds.
func NewClient(ctx context.Context, bbHost string) (Client, error) {
	hClient, err := auth.NewAuthenticator(ctx, auth.SilentLogin, site.DefaultAuthOptions).Client()
	if err != nil {
		return nil, errors.Annotate(err, "new BB client").Err()
	}

	return &client{
		buildbucket_pb.NewBuildsPRPCClient(&prpc.Client{
			C:    hClient,
			Host: bbHost,
		}),
	}, nil
}

func (c *client) ScheduleBuild(ctx context.Context, req *buildbucket_pb.ScheduleBuildRequest) (*Build, error) {
	// TODO(crbug.com/1038378): add retries.
	b, err := c.client.ScheduleBuild(ctx, req)
	if err != nil {
		return nil, errors.Annotate(err, "Schedule BB build").Err()
	}
	return &Build{
		ID:     b.Id,
		Status: b.Status,
	}, nil
}

// getBuildFields is the list of buildbucket fields that are needed.
var getBuildFields = []string{
	"id",
	// Build details are parsed from the build's output properties.
	"output.properties",
	// Build status is used to determine whether the build is complete.
	"status",
}

func (c *client) GetBuild(ctx context.Context, ID int64) (*Build, error) {
	req := &buildbucket_pb.GetBuildRequest{
		Id:     ID,
		Fields: &field_mask.FieldMask{Paths: getBuildFields},
	}
	rawBuild, err := c.client.GetBuild(ctx, req)
	if err != nil {
		return nil, errors.Annotate(err, "get build").Err()
	}
	result, err := extractResult(rawBuild)
	if err != nil {
		return nil, errors.Annotate(err, "get build").Err()
	}

	return &Build{
		ID:     rawBuild.Id,
		Status: rawBuild.Status,
		Result: result,
	}, nil
}

func extractResult(rawBuild *buildbucket_pb.Build) (*skylab_test_runner.Result, error) {
	op := rawBuild.GetOutput().GetProperties().GetFields()
	if op == nil {
		return nil, nil
	}
	rawResult, ok := op["result"]
	if !ok {
		return nil, nil
	}

	m := jsonpb.Marshaler{}
	json, err := m.MarshalToString(rawResult)
	if err != nil {
		return nil, errors.Annotate(err, "get test_runner result").Err()
	}
	result := &skylab_test_runner.Result{}
	if err := jsonpb.UnmarshalString(json, result); err != nil {
		return nil, errors.Annotate(err, "get test_runner result").Err()
	}
	return result, nil
}
