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

// GenericPublishExecutor represents executor for all publish related commands.
type GenericPublishExecutor struct {
	*interfaces.AbstractExecutor
}

func NewGenericPublishExecutor() *GenericPublishExecutor {
	absExec := interfaces.NewAbstractExecutor(GenericPublishExecutorType)
	return &GenericPublishExecutor{AbstractExecutor: absExec}
}

func (ex *GenericPublishExecutor) ExecuteCommand(
	ctx context.Context,
	cmdInterface interfaces.CommandInterface) error {

	switch cmd := cmdInterface.(type) {
	case *commands.GenericPublishCmd:
		return ex.genericPublishHandler(ctx, cmd)
	default:
		return fmt.Errorf(
			"Command type %s is not supported by %s executor type!",
			cmd.GetCommandType(),
			ex.GetExecutorType())
	}
}

// genericPublishHandler handles incoming PublishRequests.
func (ex *GenericPublishExecutor) genericPublishHandler(
	ctx context.Context,
	cmd *commands.GenericPublishCmd) (err error) {
	stepName := "Publish service"
	if cmd.Identifier != "" {
		stepName = fmt.Sprintf("%s: %s", stepName, cmd.Identifier)
	}
	step, ctx := build.StartStep(ctx, stepName)
	defer func() { step.End(err) }()

	common.WriteProtoToStepLog(ctx, step, cmd.PublishRequest, "publish service request")

	client, err := ex.ConnectToService(ctx, cmd.PublishRequest.GetServiceAddress())
	if err != nil {
		err = fmt.Errorf("error connecting to publish service, %s", err)
		return
	}

	resp, err := ex.Publish(ctx, client, cmd.PublishRequest.PublishRequest)
	if err != nil {
		err = errors.Annotate(err, "Publish cmd err: ").Err()
	}
	cmd.PublishResp = resp

	common.WriteProtoToStepLog(ctx, step, resp, "publish response")

	return
}

// ConnectToService connects to the GenericPublishServiceClient attached to the server address.
func (ex *GenericPublishExecutor) ConnectToService(
	ctx context.Context,
	endpoint *labapi.IpEndpoint) (api.GenericPublishServiceClient, error) {
	var err error
	step, ctx := build.StartStep(ctx, "Establish Connection")
	defer func() { step.End(err) }()

	// Connect with the service.
	address := common.GetServerAddress(endpoint)
	conn, err := common.ConnectWithService(ctx, address)
	if err != nil {
		logging.Infof(
			ctx,
			"error during connecting with publish server at %s: %s",
			address,
			err.Error())
		return nil, err
	}
	logging.Infof(ctx, "Connected with publish service.")

	// Create new client.
	client := api.NewGenericPublishServiceClient(conn)
	if client == nil {
		err = fmt.Errorf("GenericPublishServiceClient is nil")
		return nil, err
	}

	return client, err
}

// Publish invokces the Publish endpoint of the GenericPublishServiceClient
func (ex *GenericPublishExecutor) Publish(
	ctx context.Context,
	client api.GenericPublishServiceClient,
	req *api.PublishRequest,
) (resp *testapi.PublishResponse, err error) {
	step, ctx := build.StartStep(ctx, "Publish")
	defer func() { step.End(err) }()

	if req == nil {
		err = fmt.Errorf("PublishRequest is nil")
		return
	}

	if client == nil {
		err = fmt.Errorf("GenericPublishServiceClient is nil")
		return
	}

	common.WriteProtoToStepLog(ctx, step, req, "publish request")
	PublishOp, err := client.Publish(ctx, req, grpc.EmptyCallOption{})
	if err != nil {
		err = errors.Annotate(err, "publish failure: ").Err()
		return
	}

	opResp, err := common.ProcessLro(ctx, PublishOp)
	if err != nil {
		err = errors.Annotate(err, "publish lro failure: ").Err()
		return
	}

	resp = &testapi.PublishResponse{}
	if err = opResp.UnmarshalTo(resp); err != nil {
		err = errors.Annotate(err, "publish lro response unmarshalling failed: ").Err()
		return
	}

	return
}
