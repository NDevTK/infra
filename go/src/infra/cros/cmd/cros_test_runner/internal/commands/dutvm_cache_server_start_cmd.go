// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"

	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/common/errors"

	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
)

// DutVmCacheServerStartCmd defines the I/O to start a cache server on GCE.
type DutVmCacheServerStartCmd struct {
	*interfaces.SingleCmdByExecutor
	// Deps
	DutTopology *labapi.DutTopology

	// Updates
	CacheServerAddress *labapi.IpEndpoint
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *DutVmCacheServerStartCmd) ExtractDependencies(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

	var err error
	switch sk := ski.(type) {
	case *data.HwTestStateKeeper:
		err = cmd.extractDepsFromHwTestStateKeeper(ctx, sk)

	default:
		return fmt.Errorf("stateKeeper '%T' is not supported by cmd type %s", sk, cmd.GetCommandType())
	}

	if err != nil {
		return errors.Annotate(err, "error during extracting dependencies for command %s: ", cmd.GetCommandType()).Err()
	}

	return nil
}

// extractDepsFromHwTestStateKeeper validates DutTopology and keeps a local reference.
func (cmd *DutVmCacheServerStartCmd) extractDepsFromHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	if sk.DutTopology == nil {
		return fmt.Errorf("cmd %q missing dependency: DutTopology", cmd.GetCommandType())
	}
	if len(sk.DutTopology.GetDuts()) == 0 {
		return fmt.Errorf("cmd %q missing dependency: PrimaryDut", cmd.GetCommandType())
	}
	primaryDut := sk.DutTopology.GetDuts()[0]

	if primaryDut.GetChromeos().GetSsh() == nil {
		return fmt.Errorf("cmd %q missing dependency: DutSshAddress", cmd.GetCommandType())
	}

	if primaryDut.GetCacheServer().GetAddress() != nil {
		return fmt.Errorf("cmd %q incorrect dependency: CacheServerAddress should not be populated", cmd.GetCommandType())
	}

	cmd.DutTopology = sk.DutTopology

	return nil
}

// UpdateStateKeeper updates the state keeper with info from the cmd.
func (cmd *DutVmCacheServerStartCmd) UpdateStateKeeper(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

	var err error
	switch sk := ski.(type) {
	case *data.HwTestStateKeeper:
		err = cmd.updateVmTestStateKeeper(ctx, sk)
	}

	if err != nil {
		return errors.Annotate(err, "error during updating for command %s: ", cmd.GetCommandType()).Err()
	}

	return nil
}

// updateVmTestStateKeeper decorates DutTopology with CacheServer populated.
func (cmd *DutVmCacheServerStartCmd) updateVmTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	if cmd.CacheServerAddress != nil {
		cacheServer := &labapi.CacheServer{
			Address: cmd.CacheServerAddress,
		}
		for _, dut := range sk.DutTopology.GetDuts() {
			dut.CacheServer = cacheServer
		}
	}

	return nil
}

func NewDutVmCacheServerStartCmd(executor interfaces.ExecutorInterface) *DutVmCacheServerStartCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(DutVmCacheServerStartCmdType, executor)
	cmd := &DutVmCacheServerStartCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
