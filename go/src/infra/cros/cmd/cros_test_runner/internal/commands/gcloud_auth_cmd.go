// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"

	"go.chromium.org/luci/common/errors"

	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
)

// GcloudAuthCmd represents gcloud auth cmd.
type GcloudAuthCmd struct {
	*interfaces.SingleCmdByExecutor

	// Deps
	DockerKeyFileLocation string
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *GcloudAuthCmd) ExtractDependencies(ctx context.Context, ski interfaces.StateKeeperInterface) error {
	var err error
	switch sk := ski.(type) {
	case *data.HwTestStateKeeper:
		err = cmd.extractDepsFromHwTestStateKeeper(ctx, sk)
	default:
		return fmt.Errorf("StateKeeper '%T' is not supported by cmd type %s.", sk, cmd.GetCommandType())
	}

	if err != nil {
		return errors.Annotate(err, "error during extracting dependencies for command %s: ", cmd.GetCommandType()).Err()
	}

	return nil
}

func (cmd *GcloudAuthCmd) extractDepsFromHwTestStateKeeper(ctx context.Context, sk *data.HwTestStateKeeper) error {
	if sk.DockerKeyFileLocation == "" {
		return fmt.Errorf("Cmd %q missing dependency: DockerKeyFileLocation", cmd.GetCommandType())
	}
	cmd.DockerKeyFileLocation = sk.DockerKeyFileLocation

	return nil
}

func NewGcloudAuthCmd(executor interfaces.ExecutorInterface) *GcloudAuthCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(GcloudAuthCmdType, executor)
	cmd := &GcloudAuthCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
