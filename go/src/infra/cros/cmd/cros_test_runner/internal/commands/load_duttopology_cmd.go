// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"

	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/common/errors"
)

// LoadDutTopologyCmd represents load dut toplogy cmd.
type LoadDutTopologyCmd struct {
	*interfaces.SingleCmdByExecutor

	// Deps
	HostName string

	// Updates
	DutTopology *labapi.DutTopology
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *LoadDutTopologyCmd) ExtractDependencies(
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

// UpdateStateKeeper updates the state keeper with info from the cmd.
func (cmd *LoadDutTopologyCmd) UpdateStateKeeper(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

	var err error
	switch sk := ski.(type) {
	case *data.HwTestStateKeeper:
		err = cmd.updateHwTestStateKeeper(ctx, sk)
	}

	if err != nil {
		return errors.Annotate(err, "error during updating for command %s: ", cmd.GetCommandType()).Err()
	}

	return nil
}

func (cmd *LoadDutTopologyCmd) extractDepsFromHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	if sk.HostName == "" {
		return fmt.Errorf("Cmd %q missing dependency: HostName", cmd.GetCommandType())
	}
	cmd.HostName = sk.HostName

	return nil
}

func (cmd *LoadDutTopologyCmd) updateHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	if cmd.DutTopology != nil {
		sk.DutTopology = cmd.DutTopology
	}

	return nil
}

func NewLoadDutTopologyCmd(executor interfaces.ExecutorInterface) *LoadDutTopologyCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(LoadDutTopologyCmdType, executor)
	cmd := &LoadDutTopologyCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
