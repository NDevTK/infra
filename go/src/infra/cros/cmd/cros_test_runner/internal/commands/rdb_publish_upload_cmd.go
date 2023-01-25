// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"
	"infra/cros/cmd/cros_test_runner/common"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"

	artifactpb "go.chromium.org/chromiumos/config/go/test/artifact"
	"go.chromium.org/luci/common/errors"
)

// RdbPublishUploadCmd represents rdb publish upload cmd.
type RdbPublishUploadCmd struct {
	*interfaces.SingleCmdByExecutor

	// Deps
	CurrentInvocationId string
	TestResultForRdb    *artifactpb.TestResult
	StainlessUrl        string
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *RdbPublishUploadCmd) ExtractDependencies(
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

func (cmd *RdbPublishUploadCmd) extractDepsFromHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	if sk.CurrentInvocationId == "" {
		return fmt.Errorf("Cmd %q missing dependency: CurrentInvocationId", cmd.GetCommandType())
	}
	if sk.StainlessUrl == "" {
		return fmt.Errorf("Cmd %q missing dependency: StainlessUrl", cmd.GetCommandType())
	}
	testResult, err := cmd.constructTestResultFromStateKeeper(ctx, sk)
	if err != nil {
		return errors.Annotate(err, fmt.Sprintf("Cmd %q missing dependency: TestResultForRdb", cmd.GetCommandType())).Err()
	}

	cmd.CurrentInvocationId = sk.CurrentInvocationId
	cmd.StainlessUrl = sk.StainlessUrl
	cmd.TestResultForRdb = testResult

	return nil
}

func (cmd *RdbPublishUploadCmd) constructTestResultFromStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) (*artifactpb.TestResult, error) {

	// TODO (azrahman): construct test result here with mandatory information. Temporarily using mocked data for testing.
	return common.GetMockedTestResultProto(), nil
}

func NewRdbPublishUploadCmd(executor interfaces.ExecutorInterface) *RdbPublishUploadCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(RdbPublishUploadCmdType, executor)
	cmd := &RdbPublishUploadCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
