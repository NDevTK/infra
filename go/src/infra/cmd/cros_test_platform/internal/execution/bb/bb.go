// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package bb implements a skylab.Client using calls to BuildBucket.
package bb

import (
	"context"
	"infra/cmd/cros_test_platform/internal/execution/skylab"
	"infra/cmd/cros_test_platform/internal/execution/swarming"
	"infra/libs/skylab/request"

	"go.chromium.org/chromiumos/infra/proto/go/test_platform/config"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
)

type bbSkylabClient struct {
	swarmingClient swarming.Client
}

// NewSkylabClient creates a new skylab.Client.
func NewSkylabClient(ctx context.Context, cfg *config.Config) (skylab.Client, error) {
	sc, err := swarming.NewClient(ctx, cfg.SkylabSwarming)
	if err != nil {
		return nil, errors.Annotate(err, "create Skylab client").Err()
	}
	return &bbSkylabClient{
		swarmingClient: sc,
	}, nil
}

// ValidateArgs checks whether this test has dependencies satisfied by
// at least one Skylab bot.
func (c *bbSkylabClient) ValidateArgs(ctx context.Context, args *request.Args) (bool, error) {
	dims, err := args.StaticDimensions()
	if err != nil {
		return false, errors.Annotate(err, "validate dependencies").Err()
	}
	exists, err := c.swarmingClient.BotExists(ctx, dims)
	if err != nil {
		return false, errors.Annotate(err, "validate dependencies").Err()
	}
	if !exists {
		logging.Warningf(ctx, "Dependency validation failed for %s: no bot exists with dimensions %+v.", args.TestRunnerRequest.GetTest().GetAutotest().GetName(), dims)
	}
	return exists, nil
}

// LaunchTask stub.
func (c *bbSkylabClient) LaunchTask(ctx context.Context, args *request.Args) (skylab.TaskReference, error) {
	panic("Not yet implemented.")
}

// FetchResults stub.
func (c *bbSkylabClient) FetchResults(ctx context.Context, t skylab.TaskReference) (*skylab.FetchResultsResponse, error) {
	panic("Not yet implemented.")
}

// URL stub.
func (c *bbSkylabClient) URL(t skylab.TaskReference) string {
	panic("Not yet implemented.")
}

// SwarmingTaskID stub.
func (c *bbSkylabClient) SwarmingTaskID(t skylab.TaskReference) string {
	panic("Not yet implemented.")
}
