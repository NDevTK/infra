// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"
	"os"

	"infra/cros/cmd/cros_test_runner/common"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"

	"go.chromium.org/luci/common/errors"
)

// CpconPublishUploadCmd represents cpcon publish upload cmd.
type CpconPublishUploadCmd struct {
	*interfaces.SingleCmdByExecutor

	// Deps
	CpconJobName string
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *CpconPublishUploadCmd) ExtractDependencies(
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

func (cmd *CpconPublishUploadCmd) extractDepsFromHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	swarmingTaskId := os.Getenv("SWARMING_TASK_ID")
	if swarmingTaskId == "" {
		return fmt.Errorf("Cmd %q missing dependency: SWARMING_TASK_ID in env to construct CpconJobName", cmd.GetCommandType())
	}
	formattedSwarmingTaskId := common.FormatSwarmingTaskId(swarmingTaskId)
	jobName := fmt.Sprintf("swarming-%s0", formattedSwarmingTaskId)
	cmd.CpconJobName = jobName

	return nil
}

func NewCpconPublishUploadCmd(executor interfaces.ExecutorInterface) *CpconPublishUploadCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(CpconPublishUploadCmdType, executor)
	cmd := &CpconPublishUploadCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
