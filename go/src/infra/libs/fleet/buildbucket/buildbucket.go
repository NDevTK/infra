// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package `buildbucket` contains all the necessary code to schedule a CTP
// build for running a test using buildbucket APIs.
package buildbucket

import (
	"context"
	"net/http"

	bb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/grpc/prpc"
	"go.chromium.org/luci/server/auth"
)

// buildBucketHost is the URL host for the BuildBucket API.
const buildBucketHost = "cr-buildbucket.appspot.com"

// BuildbucketClient interface provides subset of Buildbucket methods relevant to Fleet use cases
type BuildbucketClient interface {
	ScheduleCTPBuild(ctx context.Context) (*bb.Build, error)
}

// client wraps the buildbucket client.
type client struct {
	buildBucketClient bb.BuildsClient
}

// NewClient creates a client to communicate with Buildbucket.
func NewClient(ctx context.Context) (*client, error) {
	bbClient, err := newBuildsClient(ctx, buildBucketHost)
	if err != nil {
		return nil, err
	}

	return &client{
		buildBucketClient: bbClient,
	}, nil
}

func newBuildsClient(ctx context.Context, host string) (bb.BuildsClient, error) {
	t, err := auth.GetRPCTransport(ctx, auth.AsSelf)
	if err != nil {
		return nil, err
	}
	return bb.NewBuildsPRPCClient(
		&prpc.Client{
			C:       &http.Client{Transport: t},
			Host:    host,
			Options: prpc.DefaultOptions(),
		}), nil
}
