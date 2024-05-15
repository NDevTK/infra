// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package buildbucket implements the interface required to schedule builder
// requests on the LUCI BuildBucket architecture.
package buildbucket

import (
	"context"

	"go.chromium.org/luci/auth/client/authcli"
	bb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/grpc/prpc"

	"infra/cmdsupport/cmdlib"
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
}

// client implements the Scheduler interface.
type client struct {
	ctx               context.Context
	buildBucketClient bb.BuildsClient
	isProd            bool
	dryRun            bool
}

// InitScheduler returns an operable Scheduler interface.
func InitScheduler(ctx context.Context, authOpts *authcli.Flags, isProd, dryRun bool) (Scheduler, error) {
	// Build the underlying HTTP client with the proper Auth Scoping.
	httpClient, err := cmdlib.NewHTTPClient(ctx, authOpts)
	if err != nil {
		return nil, err
	}

	// Form the custom prpc client that LUCI requires.
	prpcClient := &prpc.Client{
		C:    httpClient,
		Host: buildBucketHost,
	}

	return &client{
		ctx:               ctx,
		buildBucketClient: bb.NewBuildsClient(prpcClient),
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
