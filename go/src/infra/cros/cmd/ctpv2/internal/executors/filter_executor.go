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
	ctpv2_data "infra/cros/cmd/ctpv2/data"
	"infra/cros/cmd/ctpv2/internal/commands"
)

// FilterExecutor represents executor for all filter related commands.
type FilterExecutor struct {
	*interfaces.AbstractExecutor

	FilterServiceClient testapi.GenericFilterServiceClient
	ContainerInfo       *ctpv2_data.ContainerInfo
}

func NewFilterExecutor() *FilterExecutor {
	absExec := interfaces.NewAbstractExecutor(FilterExecutorType)
	return &FilterExecutor{AbstractExecutor: absExec}
}

func (ex *FilterExecutor) ExecuteCommand(
	ctx context.Context,
	cmdInterface interfaces.CommandInterface) error {

	switch cmd := cmdInterface.(type) {
	case *commands.FilterExecutionCmd:
		return ex.filterExecutionCommandExecution(ctx, cmd)
	default:
		return fmt.Errorf(
			"Command type %s is not supported by %s executor type!",
			cmd.GetCommandType(),
			ex.GetExecutorType())
	}
}

// filterExecutionCommandExecution executes filter execution command.
func (ex *FilterExecutor) filterExecutionCommandExecution(
	ctx context.Context,
	cmd *commands.FilterExecutionCmd) error {

	ex.ContainerInfo = cmd.ContainerInfo

	var err error
	step, ctx := build.StartStep(ctx, fmt.Sprintf("Filter execution: %s", ex.ContainerInfo.GetKey()))
	defer func() { step.End(err) }()

	common.WriteProtoToStepLog(ctx, step, cmd.InputTestPlan, "filter request")

	fitlerResp, err := ex.ExecuteFilter(ctx, cmd.InputTestPlan)
	if err != nil {
		return errors.Annotate(err, "Filter execution cmd err: ").Err()
	}

	common.WriteProtoToStepLog(ctx, step, fitlerResp, "filter response")
	cmd.OutputTestPlan = fitlerResp

	return err
}

// ExecuteTests invokes the run tests endpoint of cros-test.
func (ex *FilterExecutor) ExecuteFilter(
	ctx context.Context,
	filterReq *testapi.InternalTestplan) (*testapi.InternalTestplan, error) {
	if filterReq == nil {
		return nil, fmt.Errorf("Cannot execute filter for nil filter request.")
	}
	if ex.ContainerInfo == nil {
		return nil, fmt.Errorf("Cannot execute filter with nil container info.")
	}
	if ex.ContainerInfo.ServiceEndpoint == nil {
		return nil, fmt.Errorf("Cannot execute filter for nil service endpoint.")
	}

	filterEndpointStr, err := ex.ContainerInfo.GetEndpointString()
	if err != nil {
		return nil, errors.Annotate(err, "error while getting filter endpoint str: ").Err()
	}

	// Connect with the filter service.
	conn, err := common.ConnectWithService(ctx, filterEndpointStr)
	if err != nil {
		logging.Infof(
			ctx,
			"error during connecting with filter server at %s: %s",
			filterEndpointStr,
			err.Error())
		return nil, err
	}
	logging.Infof(ctx, "Connected with filter service.")

	// Create new client.
	filterServiceClient := api.NewGenericFilterServiceClient(conn)
	if filterServiceClient == nil {
		return nil, fmt.Errorf("filterServiceClient is nil")
	}

	// Call filter grpc endpoint
	findTestResp, err := filterServiceClient.Execute(ctx, filterReq, grpc.EmptyCallOption{})
	if err != nil {
		return nil, errors.Annotate(err, "filter grpc execution failure: ").Err()
	}

	return findTestResp, nil
}
