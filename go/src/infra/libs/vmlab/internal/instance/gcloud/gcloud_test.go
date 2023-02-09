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

type instanceCreateSuccessCommander struct{}

func (c instanceCreateSuccessCommander) GetCommandOutput(command string, args ...string) ([]byte, error) {
	// Unused field are deleted.
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
	execCommand = instanceCreateSuccessCommander{}
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
