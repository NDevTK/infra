// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package rpc_services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/anypb"

	longrunning "go.chromium.org/chromiumos/config/go/longrunning"
	api "go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/logging"

	"infra/cmd/shivas/utils"
	"infra/cros/recovery/docker"
	"infra/cros/satlab/common/services/ufs"
	"infra/cros/satlab/common/site"
	"infra/cros/satlab/common/utils/misc"
	ufsApi "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
)

// StartServod start Docker servod container.
func (s *SatlabRpcServiceServer) StartServod(ctx context.Context, in *api.StartServodRequest) (*longrunning.Operation, error) {
	err := s.validateStartServodRequest(ctx, in)
	if err != nil {
		logging.Infof(ctx, "validate request fail:  %s\n", err)
		return nil, err
	}
	r, err := s.startServodContainer(ctx, in)
	if err != nil {
		logging.Infof(ctx, "start servod fail:  %s\n", err)
		return nil, err
	}
	logging.Infof(ctx, "start servod container response %#v\n", r)

	startRes := &api.StartServodResponse{}
	startResAnypb, err := anypb.New(startRes)

	if err != nil {
		return nil, err
	}

	return &longrunning.Operation{
		Done: true,
		Result: &longrunning.Operation_Response{
			Response: startResAnypb,
		},
	}, nil
}

// validateStartServodRequest check the request and fill the missing value from UFS.
func (s *SatlabRpcServiceServer) validateStartServodRequest(ctx context.Context, in *api.StartServodRequest) error {
	logging.Infof(ctx, "validating start servod request: %#v\n", in)
	if in.GetServodDockerContainerName() == "" {
		return fmt.Errorf("validateStartServodRequest: servod docker container name is required")
	}
	// if board, mode, serial, servo port is missing; fill the information from UFS.
	if in.GetBoard() == "" || in.GetModel() == "" || in.GetSerialName() == "" || in.GetServodPort() == 0 {
		if err := s.fillDutServoInfo(ctx, in); err != nil {
			return err
		}
	}
	// if servod docker image is missing, fill it from env var.
	if in.GetServodDockerImagePath() == "" {
		in.ServodDockerImagePath = fmt.Sprintf(
			"%s/servod:%s",
			misc.GetEnv("SERVOD_REGISTRY_URI", "us-docker.pkg.dev/chromeos-partner-moblab/common-core"),
			misc.GetEnv("SERVOD_CONTAINER_LABEL", "release"),
		)
	}
	logging.Infof(ctx, "validated start servod request: %#v\n", in)
	return nil
}

// getDutNameFromServodDockerContainerName extract DUT name from servod container name.
func (s *SatlabRpcServiceServer) getDutNameFromServodDockerContainerName(c string) (string, error) {
	if strings.HasSuffix(c, "-docker_servod") {
		return strings.TrimSuffix(c, "-docker_servod"), nil
	}
	return "", fmt.Errorf("getDutNameFromServodDockerContainerName: servod docker container name should end with `-docker_servod`")
}

// fillDutServoInfo fills missing dut servo related information such as board, model, serial etc.
func (s *SatlabRpcServiceServer) fillDutServoInfo(ctx context.Context, in *api.StartServodRequest) error {
	ctx = utils.SetupContext(ctx, site.GetNamespace(""))
	ufsClient, err := ufs.NewUFSClientWithDefaultOptions(ctx, site.GetUFSService(s.dev))
	if err != nil {
		return fmt.Errorf("fillDutServoInfo: error connecting to UFS: %w", err)
	}
	dutName, err := s.getDutNameFromServodDockerContainerName(in.GetServodDockerContainerName())
	if err != nil {
		return fmt.Errorf("fillDutServoInfo: %w", err)
	}
	dut, err := ufsClient.GetMachineLSE(ctx, &ufsApi.GetMachineLSERequest{
		Name: ufsUtil.AddPrefix(ufsUtil.MachineLSECollection, dutName),
	})
	if err != nil {
		return fmt.Errorf("error fetching DUT %s from UFS: %w", dutName, err)
	}
	servo := dut.GetChromeosMachineLse().GetDeviceLse().GetDut().GetPeripherals().GetServo()
	// Fill in servo serial if missing from the request.
	if in.GetSerialName() == "" {
		in.SerialName = servo.GetServoSerial()
	}
	// Fill in servo port if missing from the request.
	if in.GetServodPort() == 0 {
		in.ServodPort = servo.GetServoPort()
	}
	if in.GetAllowDualV4() == "" {
		in.AllowDualV4 = servo.GetServoSetup().String()
	}
	if len(dut.GetMachines()) == 0 {
		return fmt.Errorf("fillDutServoInfo: fetched DUT %s has no machineId", dutName)
	}
	machine, err := ufsClient.GetMachine(
		ctx,
		&ufsApi.GetMachineRequest{
			Name: ufsUtil.AddPrefix(ufsUtil.MachineCollection, dut.GetMachines()[0]),
		},
	)
	if err != nil {
		return fmt.Errorf("fillDutServoInfo: error fetching machine %s from UFS: %w", machine, err)
	}
	// Fill in board if missing from the request.
	if in.GetBoard() == "" {
		in.Board = machine.GetChromeosMachine().GetBuildTarget()
	}
	// Fill in model if missing from the request.
	if in.GetModel() == "" {
		in.Model = machine.GetChromeosMachine().GetModel()
	}
	return nil
}

// startServodContainers runs the servod container with a validated request.
func (s *SatlabRpcServiceServer) startServodContainer(ctx context.Context, in *api.StartServodRequest) (*docker.StartResponse, error) {
	c, err := docker.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("startServodContainer: Fail to create docker client: %w", err)
	}
	// Force remove servod container if existed.
	// Ignore error if container does not exist.
	err = c.Remove(ctx, in.GetServodDockerContainerName(), true)
	if err != nil {
		logging.Infof(ctx, "fail to remove container `%s`. Non-fatal\n", in.GetServodDockerContainerName())
	}

	containerArgs := &docker.ContainerArgs{
		Detached:   true,
		Network:    "default_satlab",
		Privileged: true,
		ImageName:  in.GetServodDockerImagePath(),
		EnvVar:     generateEnvVars(in),
		Exec:       getExecCmd(in),
		Volumes:    generateVols(in),
	}
	return c.Start(ctx, in.GetServodDockerContainerName(), containerArgs, time.Minute)
}

// getExecCmd return the Docker container exec command for StartServodRequest.
func getExecCmd(in *api.StartServodRequest) []string {
	if in.GetDebug() == "true" {
		return []string{"tail", "-f", "/dev/null"}
	}
	return []string{"bash", "/start_servod.sh"}
}

// generateVols return the array of mounting volumes for servod container.
func generateVols(in *api.StartServodRequest) []string {
	return []string{
		"/dev:/dev",
		fmt.Sprintf("%s_log:/var/log/servod_9999/", in.GetServodDockerContainerName()),
	}
}

// generateVols returns the array of env vars for servod container from StartServodRequest.
func generateEnvVars(in *api.StartServodRequest) []string {
	containerEnvVars := []string{
		fmt.Sprintf("BOARD=%s", in.GetBoard()),
		fmt.Sprintf("MODEL=%s", in.GetModel()),
		fmt.Sprintf("SERIAL=%s", in.GetSerialName()),
		fmt.Sprintf("PORT=%d", in.GetServodPort()),
	}
	if in.GetAllowDualV4() != "" {
		containerEnvVars = append(containerEnvVars, fmt.Sprintf("DUAL_V4=%s", in.GetAllowDualV4()))
	}
	if in.GetRecoveryMode() != "" {
		containerEnvVars = append(containerEnvVars, "REC_MODE=1")
	}
	return containerEnvVars
}
