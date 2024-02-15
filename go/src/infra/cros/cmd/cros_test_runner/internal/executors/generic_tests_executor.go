// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package executors

import (
	"context"
	"fmt"

	"go.chromium.org/chromiumos/config/go/test/api"
	testapi "go.chromium.org/chromiumos/config/go/test/api"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"
	"google.golang.org/grpc"

	"infra/cros/cmd/common_lib/common"
	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/cros_test_runner/internal/commands"
)

// GenericTestsExecutor represents executor for all test execution related commands.
type GenericTestsExecutor struct {
	*interfaces.AbstractExecutor
}

func NewGenericTestsExecutor() *GenericTestsExecutor {
	absExec := interfaces.NewAbstractExecutor(GenericTestsExecutorType)
	return &GenericTestsExecutor{AbstractExecutor: absExec}
}

func (ex *GenericTestsExecutor) ExecuteCommand(
	ctx context.Context,
	cmdInterface interfaces.CommandInterface) error {

	switch cmd := cmdInterface.(type) {
	case *commands.GenericTestsCmd:
		return ex.genericTestsHandler(ctx, cmd)
	default:
		return fmt.Errorf(
			"Command type %s is not supported by %s executor type!",
			cmd.GetCommandType(),
			ex.GetExecutorType())
	}
}

// genericTestsHandler handles incoming TestRequests.
func (ex *GenericTestsExecutor) genericTestsHandler(
	ctx context.Context,
	cmd *commands.GenericTestsCmd) (err error) {
	stepName := "Test Execution service"
	if cmd.Identifier != "" {
		stepName = fmt.Sprintf("%s: %s", stepName, cmd.Identifier)
	}
	step, ctx := build.StartStep(ctx, stepName)
	defer func() { step.End(err) }()

	common.WriteProtoToStepLog(ctx, step, cmd.TestRequest, "test service request")

	client, err := ex.ConnectToService(ctx, cmd.TestRequest.GetServiceAddress())
	if err != nil {
		err = fmt.Errorf("error connecting to test execution service, %s", err)
		return
	}

	resp, err := ex.RunTests(ctx, client, cmd.TestRequest.TestRequest)
	if err != nil {
		err = errors.Annotate(err, "Tests execution cmd err: ").Err()
		logging.Infof(ctx, "%s", err)
	}

	cmd.TestResponses = resp

	common.WriteProtoToStepLog(ctx, step, resp, "test response")

	return
}

// ConnectToService connects to the ExecutionServiceClient attached to the server address.
func (ex *GenericTestsExecutor) ConnectToService(
	ctx context.Context,
	endpoint *labapi.IpEndpoint) (api.ExecutionServiceClient, error) {
	var err error
	step, ctx := build.StartStep(ctx, "Establish Connection")
	defer func() { step.End(err) }()

	// Connect with the service.
	address := common.GetServerAddress(endpoint)
	conn, err := common.ConnectWithService(ctx, address)
	if err != nil {
		logging.Infof(
			ctx,
			"error during connecting with test execution server at %s: %s",
			address,
			err.Error())
		return nil, err
	}
	logging.Infof(ctx, "Connected with test execution service.")

	// Create new client.
	client := api.NewExecutionServiceClient(conn)
	if client == nil {
		err = fmt.Errorf("ExecutionServiceClient is nil")
		return nil, err
	}

	return client, err
}

// RunTests invokces the RunTests endpoint of the ExecutionServiceClient
func (ex *GenericTestsExecutor) RunTests(
	ctx context.Context,
	client api.ExecutionServiceClient,
	req *api.CrosTestRequest,
) (resp *testapi.CrosTestResponse, err error) {
	step, ctx := build.StartStep(ctx, "Run Tests")
	defer func() { step.End(err) }()

	if req == nil {
		err = fmt.Errorf("CrosTestRequest is nil")
		return
	}

	if client == nil {
		err = fmt.Errorf("ExecutionServiceClient is nil")
		return
	}

	common.WriteProtoToStepLog(ctx, step, req, "cros test request")
	runTestsOp, err := client.RunTests(ctx, req, grpc.EmptyCallOption{})
	if err != nil {
		err = errors.Annotate(err, "run tests failure: ").Err()
		return
	}

	// TODO: Fix this to use common.WaitLro
	opResp, err := common.ProcessDoneLro(ctx, runTestsOp)
	if err != nil {
		err = errors.Annotate(err, "run tests lro failure: ").Err()
		return
	}

	resp = &testapi.CrosTestResponse{}
	if err = opResp.UnmarshalTo(resp); err != nil {
		err = errors.Annotate(err, "run tests lro response unmarshalling failed: ").Err()
		return
	}

	return
}
