// Copyright 2023 The Chromium Authors
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

// CpconPublishServiceStartCmd represents Cpcon publish service cmd.
type CpconPublishServiceStartCmd struct {
	*interfaces.SingleCmdByExecutor

	// Deps
	CpconPublishSrcDir string
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *CpconPublishServiceStartCmd) ExtractDependencies(
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

func (cmd *CpconPublishServiceStartCmd) extractDepsFromHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	if sk.CpconPublishSrcDir == "" {
		return fmt.Errorf("Cmd %q missing dependency: CpconPublishSrcDir", cmd.GetCommandType())
	}

	cmd.CpconPublishSrcDir = sk.CpconPublishSrcDir

	return nil
}

func NewCpconPublishServiceStartCmd(executor interfaces.ExecutorInterface) *CpconPublishServiceStartCmd {
	singleCmdExec := interfaces.NewSingleCmdByExecutor(CpconPublishStartCmdType, executor)
	cmd := &CpconPublishServiceStartCmd{SingleCmdByExecutor: singleCmdExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
