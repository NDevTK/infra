// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package executors

import (
	"context"
	"fmt"

	"go.chromium.org/chromiumos/config/go/test/api"
	testapi "go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"
	"infra/cros/cmd/cros_test_runner/common"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
)

// CacheServerExecutor implements the execution of the step defined in
// commands.CacheServerStartCmd.
type CacheServerExecutor struct {
	*interfaces.AbstractExecutor

	// Dependencies for Injection
	Container interfaces.ContainerInterface
}

func NewCacheServerExecutor(container interfaces.ContainerInterface) *CacheServerExecutor {
	absExec := interfaces.NewAbstractExecutor(CacheServerExecutorType)
	return &CacheServerExecutor{AbstractExecutor: absExec, Container: container}
}

func (ex *CacheServerExecutor) ExecuteCommand(
	ctx context.Context,
	cmdInterface interfaces.CommandInterface) error {

	switch cmd := cmdInterface.(type) {
	case *commands.CacheServerStartCmd:
		return ex.vmCacheServerStartCommandExecution(ctx, cmd)
	default:
		return fmt.Errorf("Command type %s is not supported by %s executor type!", cmd.GetCommandType(), ex.GetExecutorType())
	}
}

// vmCacheServerStartCommandExecution executes the "VM cache server start" step.
func (ex *CacheServerExecutor) vmCacheServerStartCommandExecution(
	ctx context.Context,
	cmd *commands.CacheServerStartCmd) error {

	var err error
	step, ctx := build.StartStep(ctx, "VM cache server start")
	defer func() { step.End(err) }()

	logging.Infof(ctx, "Starting cache server.")

	csTemplate := &testapi.CacheServerTemplate{}
	template := &api.Template{
		Container: &api.Template_CacheServer{
			CacheServer: csTemplate,
		},
	}

	// Process container.
	serverAddress, err := ex.Container.ProcessContainer(ctx, template)
	if err != nil {
		return errors.Annotate(err, "error processing container: ").Err()
	}

	// Process dut server address.
	cacheServerAddress, err := common.GetIpEndpoint(serverAddress)
	if err != nil {
		return errors.Annotate(err, "error while creating ip endpoint from server address: ").Err()
	}

	logging.Infof(ctx, "Cacheserver started at address: %v", cacheServerAddress)
	cmd.CacheServerAddress = cacheServerAddress
	return err
}
