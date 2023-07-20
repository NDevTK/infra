// Copyright 2023 The Chromium Authors
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
	"google.golang.org/grpc"

	"infra/cros/cmd/common_lib/common"
	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/ctpv2/internal/commands"
)

// CrosTestFinderExecutor represents executor for all cros-test-finder related commands.
type FilterExecutor struct {
	*interfaces.AbstractExecutor

	Container           interfaces.ContainerInterface
	FilterServiceClient testapi.GenericFilterServiceClient
	ServerAddress       string
}

func NewFilterExecutor(container interfaces.ContainerInterface) *FilterExecutor {
	absExec := interfaces.NewAbstractExecutor(FilterExecutorType)
	return &FilterExecutor{AbstractExecutor: absExec, Container: container}
}

func (ex *FilterExecutor) ExecuteCommand(
	ctx context.Context,
	cmdInterface interfaces.CommandInterface) error {

	switch cmd := cmdInterface.(type) {
	case *commands.FilterStartCmd:
		return ex.filterContainerStartCommandExecution(ctx, cmd)
	case *commands.FilterExecutionCmd:
		return ex.filterExecutionCommandExecution(ctx, cmd)
	default:
		return fmt.Errorf(
			"Command type %s is not supported by %s executor type!",
			cmd.GetCommandType(),
			ex.GetExecutorType())
	}
}

// testStartCommandExecution executes the test server start command.
func (ex *FilterExecutor) filterContainerStartCommandExecution(
	ctx context.Context,
	cmd *commands.FilterStartCmd) error {

	var err error
	step, ctx := build.StartStep(ctx, "Test Finder service start")
	defer func() { step.End(err) }()

	err = ex.StartContainerService(ctx)
	logErr := common.WriteContainerLogToStepLog(ctx, ex.Container, step, "cros-test-finder log")
	if err != nil {
		return errors.Annotate(err, "Start test finder service cmd err: ").Err()
	}
	if logErr != nil {
		logging.Infof(ctx, "error during writing cros-test-finder log contents: %s", err)
	}

	return err
}

// testExecutionCommandExecution executes the test execution command.
func (ex *FilterExecutor) filterExecutionCommandExecution(
	ctx context.Context,
	cmd *commands.FilterExecutionCmd) error {

	var err error
	step, ctx := build.StartStep(ctx, "Test Finder execution")
	defer func() { step.End(err) }()

	testReq := &testapi.InternalTestplan{}

	common.WriteProtoToStepLog(ctx, step, testReq, "test finder request")

	testResp, err := ex.ExecuteFilter(ctx, testReq)

	if err != nil {
		return errors.Annotate(err, "Tests execution cmd err: ").Err()
	}
	//cmd.TestSuites = testResp.TestSuites

	common.WriteProtoToStepLog(ctx, step, testResp, "test finder response")

	cmd.OutputTestPlan = testResp

	return err
}

// Start starts the cros-test-finder server.
func (ex *FilterExecutor) StartContainerService(ctx context.Context) error {
	// TODO: Get this info from input/deps. Make it work for generic container.
	template := &api.Template{
		Container: &api.Template_CrosTestFinder{
			CrosTestFinder: &testapi.CrosTestFinderTemplate{},
		},
	}

	// Process container.
	serverAddress, err := ex.Container.ProcessContainer(ctx, template)
	if err != nil {
		return errors.Annotate(err, "error processing container: ").Err()
	}

	ex.ServerAddress = serverAddress

	// Connect with the service.
	conn, err := common.ConnectWithService(ctx, serverAddress)
	if err != nil {
		logging.Infof(
			ctx,
			"error during connecting with cros-test-finder server at %s: %s",
			serverAddress,
			err.Error())
		return err
	}
	logging.Infof(ctx, "Connected with cros-test-finder service.")

	// Create new client.
	filterServiceClient := api.NewGenericFilterServiceClient(conn)
	if filterServiceClient == nil {
		return fmt.Errorf("filterServiceClient is nil")
	}

	ex.FilterServiceClient = filterServiceClient

	return nil
}

// ExecuteTests invokes the run tests endpoint of cros-test.
func (ex *FilterExecutor) ExecuteFilter(
	ctx context.Context,
	filterReq *testapi.InternalTestplan) (*testapi.InternalTestplan, error) {
	if filterReq == nil {
		return nil, fmt.Errorf("Cannot find tests for nil test finder request.")
	}
	if ex.FilterServiceClient == nil {
		return nil, fmt.Errorf("FilterServiceClient is nil in CrosTestFinderExecutor")
	}
	findTestResp, err := ex.FilterServiceClient.Execute(ctx, filterReq, grpc.EmptyCallOption{})
	if err != nil {
		return nil, errors.Annotate(err, "test execution failure: ").Err()
	}

	return findTestResp, nil
}
