// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common_executors

import (
	"context"
	"fmt"
	"sync"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"

	"infra/cros/cmd/common_lib/common"
	"infra/cros/cmd/common_lib/common_commands"
	"infra/cros/cmd/common_lib/containers"
	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"

	"go.chromium.org/chromiumos/config/go/test/api"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
)

// ContainerExecutor represents executor
// for all container related commands.
type ContainerExecutor struct {
	*interfaces.AbstractExecutor

	Ctr              *crostoolrunner.CrosToolRunner
	WaitGroups       []*sync.WaitGroup
	LogChannels      []chan<- bool
	ContainerChannel chan struct {
		string
		interfaces.ContainerInterface
	}
}

func NewContainerExecutor(ctr *crostoolrunner.CrosToolRunner) *ContainerExecutor {
	absExec := interfaces.NewAbstractExecutor(ContainerExecutorType)
	return &ContainerExecutor{AbstractExecutor: absExec, Ctr: ctr, WaitGroups: []*sync.WaitGroup{}, LogChannels: []chan<- bool{}, ContainerChannel: make(chan struct {
		string
		interfaces.ContainerInterface
	})}
}

func (ex *ContainerExecutor) ExecuteCommand(ctx context.Context, cmdInterface interfaces.CommandInterface) error {
	switch cmd := cmdInterface.(type) {
	case *common_commands.ContainerStartCmd:
		return ex.startContainerCommandExecution(ctx, cmd)
	case *common_commands.ContainerCloseLogsCmd:
		return ex.CloseLogs()
	case *common_commands.ContainerReadLogsCmd:
		return ex.ReadLogs(ctx)

	default:
		return fmt.Errorf(
			"Command type %s, %T, %v is not supported by %s executor type!",
			cmd.GetCommandType(),
			cmdInterface,
			cmdInterface,
			ex.GetExecutorType())
	}
}

// startContainerCommandExecution executes the container start command.
func (ex *ContainerExecutor) startContainerCommandExecution(
	ctx context.Context,
	cmd *common_commands.ContainerStartCmd) error {
	var err error
	step, ctx := build.StartStep(ctx, fmt.Sprintf("Container Start: %s", cmd.ContainerRequest.DynamicIdentifier))
	defer func() {
		step.End(err)
	}()

	if cmd.ContainerRequest == nil {
		return fmt.Errorf("Cannot start container with nil container request.")
	}

	containerInstance, endpoint, err := ex.Start(
		ctx,
		cmd.ContainerRequest.Container,
		interfaces.ContainerType(cmd.ContainerRequest.DynamicIdentifier),
		cmd.ContainerRequest.DynamicIdentifier,
		cmd.ContainerImage)

	if err != nil {
		return errors.Annotate(err, "Start container cmd err: ").Err()
	}
	cmd.ContainerInstance = containerInstance
	cmd.Endpoint = endpoint

	// Send container for logs to be read.
	ex.ContainerChannel <- struct {
		string
		interfaces.ContainerInterface
	}{cmd.ContainerRequest.DynamicIdentifier, containerInstance}

	return err
}

func (ex *ContainerExecutor) ReadLogs(ctx context.Context) error {
	var err error
	step, ctx := build.StartStep(ctx, "Read Container Logs")

	go func() {
		defer func() {
			step.End(err)
		}()

		for recv := range ex.ContainerChannel {
			ex.streamLogAsync(ctx, step, recv.string, recv.ContainerInterface)
		}
	}()

	return nil
}

// Start starts the container.
func (ex *ContainerExecutor) Start(
	ctx context.Context,
	template *api.Template,
	containerType interfaces.ContainerType,
	containerPrefix string,
	containerImage string) (interfaces.ContainerInterface, *labapi.IpEndpoint, error) {

	containerInstance := containers.NewContainer(
		containerType,
		containerPrefix,
		containerImage,
		ex.Ctr,
		true)

	// Process container.
	serverAddress, err := containerInstance.ProcessContainer(ctx, template)
	if err != nil {
		return nil, nil, errors.Annotate(err, "error processing container: ").Err()
	}
	endpoint, err := common.GetIpEndpoint(serverAddress)
	if err != nil {
		return nil, nil, err
	}

	return containerInstance, endpoint, nil
}

// streamLog kicks off streaming the containers log and stores its channel and waitgroup.
func (ex *ContainerExecutor) streamLogAsync(ctx context.Context, step *build.Step, identifier string, containerInstance interfaces.ContainerInterface) (wg *sync.WaitGroup) {
	logsLoc, err := containerInstance.GetLogsLocation()
	if err != nil {
		logging.Infof(ctx, "error during getting container log location: %s", err)
	}
	containerLog := step.Log(fmt.Sprintf("%s Log", identifier))

	taskDone, wg, err := common.StreamLogAsync(ctx, logsLoc, containerLog)
	if err != nil {
		logging.Infof(ctx, "Warning: error during reading container log: %s", err)
		return
	}

	ex.LogChannels = append(ex.LogChannels, taskDone)
	ex.WaitGroups = append(ex.WaitGroups, wg)

	return
}

// CloseLogs signals to the streaming logs through their channels that they can close.
func (ex *ContainerExecutor) CloseLogs() error {
	for _, logChannel := range ex.LogChannels {
		logChannel <- true
	}
	for _, waitGroup := range ex.WaitGroups {
		waitGroup.Wait()
	}

	close(ex.ContainerChannel)

	return nil
}
