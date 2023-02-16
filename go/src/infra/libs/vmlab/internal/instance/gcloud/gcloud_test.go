// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package gcloud

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"infra/libs/vmlab/api"
)

type fakeRandomGenerator struct{}

func (c fakeRandomGenerator) GetRandHex(l int) (string, error) {
	return "aaaaaa", nil
}

type instanceCreateSuccessCommander struct {
	Commands *[][]string
}

func (c instanceCreateSuccessCommander) GetCommandOutput(command string, args ...string) ([]byte, error) {
	// Unused field are deleted.
	if c.Commands != nil {
		thisCommand := []string{}
		thisCommand = append(thisCommand, command)
		thisCommand = append(thisCommand, args...)
		*c.Commands = append(*c.Commands, thisCommand)
	}
	return []byte(`
[
  {
    "networkInterfaces": [
      {
        "accessConfigs": [
          {
            "natIP": "8.8.8.8"
          }
        ],
        "networkIP": "192.168.0.1"
      }
    ]
  }
]
`), nil
}

func TestCreateWithPublicIpAddress(t *testing.T) {
	gcloud, _ := New()
	mockExecCommand := instanceCreateSuccessCommander{
		Commands: &[][]string{},
	}
	execCommand = mockExecCommand
	random = fakeRandomGenerator{}
	instance, err := gcloud.Create(
		&api.CreateVmInstanceRequest{
			Config: &api.Config{
				Backend: &api.Config_GcloudBackend{
					GcloudBackend: &api.Config_GCloudBackend{
						Project:        "vmlab-project",
						Zone:           "us-west-2",
						MachineType:    "n2-standard-4",
						InstancePrefix: "vmlab-",
						PublicIp:       true,
						Image: &api.GceImage{
							Project: "imagestorage-project",
							Name:    "betty-arc-r-release-r110-111111111111",
						},
					},
				},
			},
		})
	expectedCommand := []string{
		"gcloud", "compute", "instance", "create", "vmlab-aaaaaa",
		"--project=vmlab-project", "--image=betty-arc-r-release-r110-111111111111",
		"--image-project=imagestorage-project", "--machine-type=n2-standard-4", "--no-scopes",
		"--zone=us-west-2", "--format=json", "--network=default", "--subnet=default"}
	if diff := cmp.Diff(*mockExecCommand.Commands, [][]string{expectedCommand}); diff != "" {
		t.Errorf("Executed wrong command: %v", diff)
	}
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if !strings.HasPrefix(instance.GetName(), "vmlab-") {
		t.Errorf("Instance name incorrect: %v", instance)
	}
	expectedSshTarget := &api.AddressPort{Address: "8.8.8.8", Port: 22}
	if diff := cmp.Diff(instance.GetSsh(), expectedSshTarget, protocmp.Transform()); diff != "" {
		t.Errorf("Got wrong ssh target: %v Diff is:\n%v", instance.GetSsh(), diff)
	}
}

func TestCreateWithInternalIpAddress(t *testing.T) {
	gcloud, _ := New()
	mockExecCommand := instanceCreateSuccessCommander{
		Commands: &[][]string{},
	}
	execCommand = mockExecCommand
	random = fakeRandomGenerator{}
	instance, err := gcloud.Create(
		&api.CreateVmInstanceRequest{
			Config: &api.Config{
				Backend: &api.Config_GcloudBackend{
					GcloudBackend: &api.Config_GCloudBackend{
						Project:        "vmlab-project",
						Zone:           "us-west-2",
						MachineType:    "n2-standard-4",
						InstancePrefix: "vmlab-",
						PublicIp:       false,
						Image: &api.GceImage{
							Project: "imagestorage-project",
							Name:    "betty-arc-r-release-r110-111111111111",
						},
					},
				},
			},
		})
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	expectedCommand := []string{
		"gcloud", "compute", "instance", "create", "vmlab-aaaaaa",
		"--project=vmlab-project", "--image=betty-arc-r-release-r110-111111111111",
		"--image-project=imagestorage-project", "--machine-type=n2-standard-4", "--no-scopes",
		"--zone=us-west-2", "--format=json", "--network=default", "--subnet=default",
		"--no-address"}
	if diff := cmp.Diff(*mockExecCommand.Commands, [][]string{expectedCommand}); diff != "" {
		t.Errorf("Executed wrong command: %v", diff)
	}
	if !strings.HasPrefix(instance.GetName(), "vmlab-") {
		t.Errorf("Instance name incorrect: %v", instance)
	}
	expectedSshTarget := &api.AddressPort{Address: "192.168.0.1", Port: 22}
	if diff := cmp.Diff(instance.GetSsh(), expectedSshTarget, protocmp.Transform()); diff != "" {
		t.Errorf("Got wrong ssh target: %v Diff is:\n%v", instance.GetSsh(), diff)
	}
}

func TestCreateMissingProject(t *testing.T) {
	gcloud, _ := New()
	_, err := gcloud.Create(
		&api.CreateVmInstanceRequest{
			Config: &api.Config{
				Backend: &api.Config_GcloudBackend{
					GcloudBackend: &api.Config_GCloudBackend{},
				},
			},
		})
	if err.Error() != "invalid config argument: project must be set" {
		t.Errorf("expect project not set error, but got: %s", err.Error())
	}
}
