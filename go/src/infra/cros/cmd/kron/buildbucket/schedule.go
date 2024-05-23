// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package buildbucket implements the interface required to schedule builder
// requests on the LUCI BuildBucket architecture.
package buildbucket

import (
	"context"
	"fmt"

	"go.chromium.org/luci/auth/client/authcli"
	bb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/grpc/prpc"

	"infra/cmdsupport/cmdlib"
	"infra/cros/cmd/kron/common"
)

// buildBucketHost is the URL host for the BuildBucket API.
const buildBucketHost = "cr-buildbucket.appspot.com"

var (
	CtpBuilderIDProd = bb.BuilderID{
		Project: "chromeos",
		Bucket:  "testplatform",
		Builder: "cros_test_platform",
	}
	CtpBuilderIDStaging = bb.BuilderID{
		Project: "chromeos",
		Bucket:  "testplatform",
		Builder: "cros_test_platform-dev",
	}
)

// Scheduler interface type describes the BB API functionality connection.
type Scheduler interface {
	Schedule(request *bb.ScheduleBuildRequest) (*bb.Build, error)
	GetBuildStatus(buildID int64) (*bb.Build, error)
}

// client implements the Scheduler interface.
type client struct {
	ctx               context.Context
	buildBucketClient bb.BuildsClient
	isProd            bool
	dryRun            bool
}

// NewBBClient returns a bb client
func NewBBClient(ctx context.Context, authOpts *authcli.Flags) (bb.BuildsClient, error) {
	httpClient, err := cmdlib.NewHTTPClient(context.Background(), authOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create BuildBucket client: %w", err)
	}
	pClient := &prpc.Client{
		C:    httpClient,
		Host: buildBucketHost,
	}
	return bb.NewBuildsClient(pClient), nil
}

// InitScheduler returns an operable Scheduler interface.
func InitScheduler(ctx context.Context, authOpts *authcli.Flags, isProd, dryRun bool) (Scheduler, error) {
	// Build the underlying HTTP client with the proper Auth Scoping.
	bbclient, err := NewBBClient(ctx, authOpts)
	if err != nil {
		return nil, err
	}

	return &client{
		ctx:               ctx,
		buildBucketClient: bbclient,
		isProd:            isProd,
		dryRun:            dryRun,
	}, nil
}

// Schedule takes in a ScheduleBuildRequest and schedules it via the BuildBucket
// API.
func (c *client) Schedule(request *bb.ScheduleBuildRequest) (*bb.Build, error) {
	build, err := c.buildBucketClient.ScheduleBuild(c.ctx, request)
	if err != nil {
		return nil, err
	}

	return build, nil
}

// GetBuildStatus returns the status of a build.
func (c *client) GetBuildStatus(buildID int64) (*bb.Build, error) {
	ctx := context.Background()
	statusReq := &bb.GetBuildStatusRequest{
		Id: buildID,
	}
	build, err := c.buildBucketClient.GetBuildStatus(ctx, statusReq)
	if err != nil {
		common.Stdout.Printf("Failed to fetch build status for build id :%d", buildID)
		return nil, err
	}
	return build, nil
}
