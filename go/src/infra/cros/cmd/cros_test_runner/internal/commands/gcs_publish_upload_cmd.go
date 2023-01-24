// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"

	"go.chromium.org/luci/common/errors"
)

// GcsPublishUploadCmd represents gcs publish upload cmd.
type GcsPublishUploadCmd struct {
	*interfaces.SingleCmdByExecutor

	// Deps
	GcsUrl string
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *GcsPublishUploadCmd) ExtractDependencies(ctx context.Context, ski interfaces.StateKeeperInterface) error {
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

func (cmd *GcsPublishUploadCmd) extractDepsFromHwTestStateKeeper(ctx context.Context, sk *data.HwTestStateKeeper) error {
	if sk.GcsUrl == "" {
		return fmt.Errorf("Cmd %q missing dependency: GcsUrl", cmd.GetCommandType())
	}

	cmd.GcsUrl = sk.GcsUrl
	return nil
}

func NewGcsPublishUploadCmd(executor interfaces.ExecutorInterface) *GcsPublishUploadCmd {
	cmd := &GcsPublishUploadCmd{SingleCmdByExecutor: interfaces.NewSingleCmdByExecutor(GcsPublishUploadCmdType, executor)}
	cmd.ConcreteCmd = cmd
	return cmd
}
