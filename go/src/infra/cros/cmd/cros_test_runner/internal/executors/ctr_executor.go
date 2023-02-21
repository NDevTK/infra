// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package executors

import (
	"context"
	"fmt"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"

	testapi "go.chromium.org/chromiumos/config/go/test/api"

	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
	"infra/cros/cmd/cros_test_runner/internal/tools/crostoolrunner"
)

// CrosProvisionExecutor represents executor
// for all crostoolrunner (ctr) related commands.
type CtrExecutor struct {
	*interfaces.AbstractExecutor

	Ctr                        *crostoolrunner.CrosToolRunner
	CrosProvisionServiceClient testapi.GenericProvisionServiceClient
}

func NewCtrExecutor(ctr *crostoolrunner.CrosToolRunner) *CtrExecutor {
	absExec := interfaces.NewAbstractExecutor(CtrExecutorType)
	return &CtrExecutor{AbstractExecutor: absExec, Ctr: ctr}
}

func (ex *CtrExecutor) ExecuteCommand(ctx context.Context, cmdInterface interfaces.CommandInterface) error {
	switch cmd := cmdInterface.(type) {
	case *commands.CtrServiceStartAsyncCmd:
		return ex.startAsyncCommandExecution(ctx, cmd)
	case *commands.GcloudAuthCmd:
		return ex.gcloudAuthCommandExecution(ctx, cmd)
	case *commands.CtrServiceStopCmd:
		ex.stopCommandExecution(ctx, cmd)

	default:
		return fmt.Errorf(
			"Command type %s, %T, %v is not supported by %s executor type!",
			cmd.GetCommandType(),
			cmdInterface,
			cmdInterface,
			ex.GetExecutorType())
	}

	return nil
}

// startAsyncCommandExecution executes the start ctr server
// asynchronously command.
func (ex *CtrExecutor) startAsyncCommandExecution(
	ctx context.Context,
	cmd *commands.CtrServiceStartAsyncCmd) error {

	var err error
	step, ctx := build.StartStep(ctx, "Ctr service start")
	defer func() { step.End(err) }()

	err = ex.StartAsync(ctx)
	if err != nil {
		return errors.Annotate(err, "Start ctr cmd err: ").Err()
	}

	return err
}

// stopCommandExecution executes stop ctr server command.
func (ex *CtrExecutor) stopCommandExecution(
	ctx context.Context,
	cmd *commands.CtrServiceStopCmd) error {

	var err error
	step, ctx := build.StartStep(ctx, "Ctr service stop")
	defer func() { step.End(err) }()

	err = ex.Stop(ctx)
	if err != nil {
		return errors.Annotate(err, "Stop ctr cmd err: ").Err()
	}

	return err
}

// gcloudAuthCommandExecution executes the gcloud registry auth command.
func (ex *CtrExecutor) gcloudAuthCommandExecution(
	ctx context.Context,
	cmd *commands.GcloudAuthCmd) error {

	var err error
	step, ctx := build.StartStep(ctx, "Gcloud Auth")
	defer func() { step.End(err) }()

	err = ex.GcloudAuth(ctx, cmd.DockerKeyFileLocation)
	if err != nil {
		return errors.Annotate(err, "gcloud auth cmd err: ").Err()
	}

	return err
}

// StartAsync starts the ctr server asynchronously.
func (ex *CtrExecutor) StartAsync(ctx context.Context) error {
	// Initialize ctr
	err := ex.Ctr.Initialize(ctx)
	if err != nil {
		logging.Infof(ctx, fmt.Sprintf("cros-tool-runner initialization error: %s", err.Error()))
		return errors.Annotate(err, "cros-tool-runner initialization error: ").Err()
	} else {
		logging.Infof(ctx, "CTR initialization succeeded!")
	}

	// Start CTR Server async
	err = ex.Ctr.StartCTRServerAsync(ctx)
	if err != nil {
		logging.Infof(ctx, "error during starting ctr server: %s", err.Error())
		return errors.Annotate(err, "error during starting ctr server: ").Err()
	}

	// Retrieve server address from metadata
	serverAddress, err := ex.Ctr.GetServerAddressFromServiceMetadata(ctx)
	if err != nil {
		return errors.Annotate(err, "cros-tool-runner retrieve server address error: ").Err()
	}

	// Connect to server
	_, err = ex.Ctr.ConnectToCTRServer(ctx, serverAddress)
	if err != nil {
		return errors.Annotate(err, "cros-tool-runner connect to server error: ").Err()
	}

	return nil
}

// GcloudAuth does the gcloud auth through ctr.
func (ex *CtrExecutor) GcloudAuth(ctx context.Context, dockerKeyFileLocation string) error {
	_, err := ex.Ctr.GcloudAuth(ctx, dockerKeyFileLocation)
	if err != nil {
		return errors.Annotate(err, "error during gcloud cmd: ").Err()
	}

	return nil
}

// Stop stops the ctr server.
func (ex *CtrExecutor) Stop(ctx context.Context) error {
	err := ex.Ctr.StopCTRServer(ctx)
	if err != nil {
		logging.Infof(ctx, "error during stopping ctr server: %s", err.Error())
		return errors.Annotate(err, "error during stopping ctr server: ").Err()
	}

	return nil

}
