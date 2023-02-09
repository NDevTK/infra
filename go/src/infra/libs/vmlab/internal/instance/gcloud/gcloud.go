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

func randHex(l int) string {
	b := make([]byte, l/2+1)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[:l]
}

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

var execCommand commander = realCommander{}

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

	// TODO(fqj): support internal IP address
	if !gcloudConfig.GetPublicIp() {
		return nil, errors.New("only public IP address is supported.")
	}

	// Instance name is a str that prefixes with instance_prefix, followed by a few alphabet or numeric characters.
	name := fmt.Sprintf("%s%s", gcloudConfig.GetInstancePrefix(), randHex(25))

	// TODO(fqj): implement tags
	out, err := execCommand.GetCommandOutput(
		"gcloud", "compute", "instance", "create", name,
		"--project="+gcloudConfig.GetProject(),
		"--image="+gcloudConfig.GetImage().GetName(), "--image-project="+gcloudConfig.GetImage().GetProject(),
		"--machine-type="+gcloudConfig.GetMachineType(), "--no-scopes",
		"--zone="+gcloudConfig.GetZone(), "--format=json", "--network=default", "--subnet=default")
	if err != nil {
		return nil, fmt.Errorf("failed to launch instance: %w", err)
	}
	var gcloudResult []gcloudResponseInstance
	if err := json.Unmarshal(out, &gcloudResult); err != nil {
		return nil, fmt.Errorf("unable to parse gcloud result: %w", err)
	}

	publicIpAddress := gcloudResult[0].NetworkInterfaces[0].AccessConfigs[0].NatIP

	return &api.VmInstance{
		Name: name,
		Ssh: &api.AddressPort{
			Address: publicIpAddress,
			Port:    22,
		},
	}, nil
}

func (g *gcloudInstanceApi) Delete(ins *api.VmInstance) error {
	return errors.New("not implemented")
}

func (g *gcloudInstanceApi) Cleanup(req *api.CleanupVmInstancesRequest) error {
	return errors.New("not implemented")
}
