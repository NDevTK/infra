// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package gcloud

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"

	"infra/libs/vmlab/api"
)

// TODO(b/250961857): implement gcloud provider
// gcloudInstanceApi implements api.InstanceApi. The struct itself doesn't need
// to be public.
type gcloudInstanceApi struct{}

type gcloudResponseInstance struct {
	NetworkInterfaces []gcloudNetworkInterface
}

type gcloudNetworkInterface struct {
	AccessConfigs []gcloudAccessConfig
	NetworkIP     string
}

type gcloudAccessConfig struct {
	NatIP string
}

// An interface that wraps some function of os/exec to allow unittest mock.
type commander interface {
	GetCommandOutput(string, ...string) ([]byte, error)
}

type realCommander struct{}

func (c realCommander) GetCommandOutput(command string, args ...string) ([]byte, error) {
	return exec.Command(command, args...).Output()
}

// A utility method that extracts explicit message from an error, especially for
// the error returned by exec.Cmd.Output().
func extractErrorMessage(err error) string {
	if ee, ok := err.(*exec.ExitError); ok {
		return string(ee.Stderr)
	}
	return err.Error()
}

// An interface that wraps some function of crypto/rand to allow unittest mock.
type randomGenerator interface {
	GetRandHex(l int) (string, error)
}

type realRandomGenerator struct{}

func (c realRandomGenerator) GetRandHex(l int) (string, error) {
	b := make([]byte, l/2+1)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b)[:l], nil
}

var execCommand commander = realCommander{}
var random randomGenerator = realRandomGenerator{}

// New constructs a new api.InstanceApi with gcloud backend.
func New() (api.InstanceApi, error) {
	return &gcloudInstanceApi{}, nil
}

func checkGCloudConfig(gcloudConfig *api.Config_GCloudBackend) error {
	if gcloudConfig.GetProject() == "" {
		return errors.New("project must be set")
	}
	if gcloudConfig.GetZone() == "" {
		return errors.New("zone must be set")
	}
	if gcloudConfig.GetMachineType() == "" {
		return errors.New("machine type must be set")
	}
	if gcloudConfig.GetInstancePrefix() == "" {
		return errors.New("instance prefix must be set")
	}
	if gcloudConfig.GetImage().GetName() == "" {
		return errors.New("image name must be set")
	}
	if gcloudConfig.GetImage().GetProject() == "" {
		return errors.New("image project must be set")
	}
	return nil
}

func (g *gcloudInstanceApi) Create(req *api.CreateVmInstanceRequest) (*api.VmInstance, error) {
	gcloudConfig := req.GetConfig().GetGcloudBackend()
	if gcloudConfig == nil {
		return nil, fmt.Errorf("invalid argument: bad backend: want gcloud, got %v", req.GetConfig())
	}
	if err := checkGCloudConfig(gcloudConfig); err != nil {
		return nil, fmt.Errorf("invalid config argument: %w", err)
	}

	gcloudArgs := []string{}

	// Instance name is a str that prefixes with instance_prefix, followed by a few alphabet or numeric characters.
	randName, err := random.GetRandHex(25)
	if err != nil {
		return nil, fmt.Errorf("failed to generate instance name: %w", err)
	}
	name := fmt.Sprintf("%s%s", gcloudConfig.GetInstancePrefix(), randName)
	gcloudArgs = append(gcloudArgs, "compute", "instances", "create", name)

	gcloudArgs = append(gcloudArgs,
		"--project="+gcloudConfig.GetProject(),
		"--image="+gcloudConfig.GetImage().GetName(), "--image-project="+gcloudConfig.GetImage().GetProject(),
		"--machine-type="+gcloudConfig.GetMachineType(), "--no-scopes",
		"--zone="+gcloudConfig.GetZone(), "--format=json")

	if gcloudConfig.GetNetwork() != "" {
		gcloudArgs = append(gcloudArgs, "--network="+gcloudConfig.GetNetwork())
	}
	if gcloudConfig.GetSubnet() != "" {
		gcloudArgs = append(gcloudArgs, "--subnet="+gcloudConfig.GetSubnet())
	}

	if !gcloudConfig.GetPublicIp() {
		gcloudArgs = append(gcloudArgs, "--no-address")
	}

	// TODO(fqj): implement tags
	out, err := execCommand.GetCommandOutput("gcloud", gcloudArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to launch instance: %v", extractErrorMessage(err))
	}
	var gcloudResult []gcloudResponseInstance
	if err := json.Unmarshal(out, &gcloudResult); err != nil {
		return nil, fmt.Errorf("unable to parse gcloud result: %w\n output is: %v\n", err, string(out))
	}

	ipAddress := gcloudResult[0].NetworkInterfaces[0].NetworkIP
	if gcloudConfig.GetPublicIp() {
		ipAddress = gcloudResult[0].NetworkInterfaces[0].AccessConfigs[0].NatIP
	}

	return &api.VmInstance{
		Name: name,
		Ssh: &api.AddressPort{
			Address: ipAddress,
			Port:    22,
		},
		Config: req.GetConfig(),
	}, nil
}

func (g *gcloudInstanceApi) Delete(ins *api.VmInstance) error {
	gcloudConfig := ins.GetConfig().GetGcloudBackend()
	if gcloudConfig == nil {
		return fmt.Errorf("invalid argument: bad backend: want gcloud, got %v", ins.GetConfig())
	}
	if ins.GetName() == "" {
		return errors.New("instance name must be set")
	}
	if gcloudConfig.GetProject() == "" {
		return errors.New("project must be set")
	}
	if gcloudConfig.GetZone() == "" {
		return errors.New("zone must be set")
	}

	gcloudArgs := []string{"compute", "instances", "delete", ins.GetName()}
	gcloudArgs = append(gcloudArgs,
		"--project="+gcloudConfig.GetProject(),
		"--zone="+gcloudConfig.GetZone(), "--quiet")

	_, err := execCommand.GetCommandOutput("gcloud", gcloudArgs...)
	if err != nil {
		return fmt.Errorf("failed to launch instance: %v", extractErrorMessage(err))
	}

	return nil
}

func (g *gcloudInstanceApi) Cleanup(req *api.CleanupVmInstancesRequest) error {
	return errors.New("not implemented")
}
