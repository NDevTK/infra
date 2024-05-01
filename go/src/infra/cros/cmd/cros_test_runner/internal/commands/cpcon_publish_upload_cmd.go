// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"
	"os"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"

	"infra/cros/cmd/common_lib/common"
	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/cros_test_runner/data"
)

// CpconPublishUploadCmd represents cpcon publish upload cmd.
type CpconPublishUploadCmd struct {
	*interfaces.SingleCmdByExecutor

	// Deps
	CpconPublishSrcDir string
	CpconJobName       string
	GcsURL             string
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *CpconPublishUploadCmd) ExtractDependencies(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

	var err error
	switch sk := ski.(type) {
	case *data.HwTestStateKeeper:
		err = cmd.extractDepsFromHwTestStateKeeper(ctx, sk)
	case *data.LocalTestStateKeeper:
		err = cmd.extractDepsFromHwTestStateKeeper(ctx, &sk.HwTestStateKeeper)

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
		logging.Warningf(ctx, "SWARMING_TASK_ID not set, setting to default %s", "local_run")
		swarmingTaskId = "local_run"
	}
	formattedSwarmingTaskId := common.FormatSwarmingTaskId(swarmingTaskId)
	jobName := fmt.Sprintf("swarming-%s0", formattedSwarmingTaskId)
	cmd.CpconJobName = jobName

	if sk.CpconPublishSrcDir == "" {
		return fmt.Errorf("Cmd %q missing dependency: CpconPublishSrcDir", cmd.GetCommandType())
	}
	cmd.CpconPublishSrcDir = sk.CpconPublishSrcDir

	if sk.GcsURL == "" {
		return fmt.Errorf("Cmd %q missing dependency: GcsURL", cmd.GetCommandType())
	}
	cmd.GcsURL = sk.GcsURL

	return nil
}

func NewCpconPublishUploadCmd(executor interfaces.ExecutorInterface) *CpconPublishUploadCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(CpconPublishUploadCmdType, executor)
	cmd := &CpconPublishUploadCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
