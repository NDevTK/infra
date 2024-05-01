// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"

	"go.chromium.org/luci/common/errors"

	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/cros_test_runner/data"
)

// GcsPublishUploadCmd represents gcs publish upload cmd.
type GcsPublishUploadCmd struct {
	*interfaces.SingleCmdByExecutor

	// Deps
	GcsURL string
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *GcsPublishUploadCmd) ExtractDependencies(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

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

func (cmd *GcsPublishUploadCmd) extractDepsFromHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	if sk.GcsURL == "" {
		return fmt.Errorf("Cmd %q missing dependency: GcsURL", cmd.GetCommandType())
	}

	cmd.GcsURL = sk.GcsURL
	return nil
}

func NewGcsPublishUploadCmd(executor interfaces.ExecutorInterface) *GcsPublishUploadCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(GcsPublishUploadCmdType, executor)
	cmd := &GcsPublishUploadCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
