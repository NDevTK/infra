// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package servod

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"infra/cros/cmd/satlab/internal/site"
	"infra/cros/recovery/docker"
	ufsModels "infra/unifiedfleet/api/v1/models"
	ufsCros "infra/unifiedfleet/api/v1/models/chromeos/lab"
	ufsApi "infra/unifiedfleet/api/v1/rpc"
)

// FakeUFSClient is a mockUFS client that returns hardcoded data. It can also return bad/non-existent data
type FakeUFSClient struct {
	badData bool
}

// GetDUT returns a hardcoded DUT with a servo connected
func (f *FakeUFSClient) GetDut(context.Context, *ufsApi.GetMachineLSERequest) (*ufsModels.MachineLSE, error) {
	if f.badData {
		return &ufsModels.MachineLSE{}, nil
	}

	servo := ufsCros.Servo{
		ServoSerial:         "ufsSerial",
		DockerContainerName: "ufsDockerContainerName",
	}

	return &ufsModels.MachineLSE{
		Machines: []string{"ufsAssetId"},
		Lse: &ufsModels.MachineLSE_ChromeosMachineLse{
			ChromeosMachineLse: &ufsModels.ChromeOSMachineLSE{
				ChromeosLse: &ufsModels.ChromeOSMachineLSE_DeviceLse{
					DeviceLse: &ufsModels.ChromeOSDeviceLSE{
						Device: &ufsModels.ChromeOSDeviceLSE_Dut{
							Dut: &ufsCros.DeviceUnderTest{
								Peripherals: &ufsCros.Peripherals{
									Servo: &servo,
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

// GetAsset returns a hardcoded asset
func (f *FakeUFSClient) GetAsset(context.Context, *ufsApi.GetAssetRequest) (*ufsModels.Asset, error) {
	if f.badData {
		return &ufsModels.Asset{}, nil
	}

	return &ufsModels.Asset{
		Model: "ufsModel",
		Info:  &ufsModels.AssetInfo{ReferenceBoard: "ufsBoard"},
	}, nil
}

// TestStartServodCmd tests the innerRun function of our command with fake UFS and docker clients, comparing the args we use to actually launch docker w/what is expected
// It tests input entirely, partially, and not at all user given
func TestStartServodCmd(t *testing.T) {
	t.Parallel()

	type test struct {
		inputCommand          *startServodRun
		expectedContainerName string
		expectedContainerArgs docker.ContainerArgs
	}

	tests := []test{
		{
			&startServodRun{host: "eli", board: "board", model: "model", servoSerial: "serial", servodContainerName: "docker_servod", noServodProcess: false},
			"docker_servod",
			*buildServodContainerArgs(ServodContainerOptions{"docker_servod", "board", "model", "serial", true}),
		},
		{
			&startServodRun{host: "eli", board: "board"},
			"ufsDockerContainerName",
			*buildServodContainerArgs(ServodContainerOptions{"docker_servod", "board", "ufsModel", "ufsSerial", true}),
		},
		{
			&startServodRun{host: "eli"},
			"ufsDockerContainerName",
			*buildServodContainerArgs(ServodContainerOptions{"docker_servod", "ufsBoard", "ufsModel", "ufsSerial", true}),
		},
		{
			&startServodRun{host: "eli", board: "board", model: "model", servoSerial: "serial", servodContainerName: "docker_servod", noServodProcess: true},
			"docker_servod",
			*buildServodContainerArgs(ServodContainerOptions{"docker_servod", "board", "model", "serial", false}),
		},
	}

	for _, tc := range tests {
		fakeDockerClient := NewFakeDockerClient(false, false)

		err := tc.inputCommand.runOrchestratedCommand(context.Background(), &fakeDockerClient, &FakeUFSClient{})
		if err != nil {
			t.Errorf("Unexpected err: %v", err)
		}

		if diff := cmp.Diff(fakeDockerClient.runningContainers[tc.expectedContainerName], tc.expectedContainerArgs); diff != "" {
			t.Errorf("Docker container launched with unexpected diff: %s", diff)
		}
	}
}

// TestStartServodCmdFailedFetch tests behavior when we get data from UFS but it is not complete
func TestStartServodCmdFailedFetch(t *testing.T) {
	t.Parallel()

	fakeDockerClient := NewFakeDockerClient(false, false)
	fakeUFSClient := FakeUFSClient{badData: true}

	err := (&startServodRun{host: "eli"}).runOrchestratedCommand(context.Background(), &fakeDockerClient, &fakeUFSClient)

	if err == nil {
		t.Errorf("Expected command to error")
	}
}

// fakeHostBoxGetter is a function of type `HostBoxIdentifierGetter` injected to make sure we use it correctly
func fakeHostBoxGetter() (string, error) {
	return "1234", nil
}

// TestValidate tests behavior of validate under different user inputs
func TestValidate(t *testing.T) {
	t.Parallel()

	type test struct {
		inputCommand   *startServodRun
		inputArgs      []string
		expectedOutput *startServodRun
		expectedErr    bool
	}

	tests := []test{
		{&startServodRun{host: ""}, []string{}, nil, true},
		{&startServodRun{host: "eli"}, []string{}, &startServodRun{host: "satlab-1234-eli"}, false},
		{&startServodRun{host: "eli", commonFlags: site.CommonFlags{SatlabID: "5678"}}, []string{}, &startServodRun{host: "satlab-5678-eli"}, false},
		{&startServodRun{host: "satlab-1234-eli"}, []string{}, &startServodRun{host: "satlab-1234-eli"}, false},
		{&startServodRun{host: "satlab-1234-eli"}, []string{"bad"}, nil, true},
	}

	for _, tc := range tests {
		err := tc.inputCommand.validate("1234", tc.inputArgs)
		if tc.expectedErr != (err != nil) {
			t.Errorf("Expected tc to error: %t, but was the opposite for input %+v", tc.expectedErr, tc)
		}

		if !tc.expectedErr {
			if diff := cmp.Diff(tc.inputCommand.host, tc.expectedOutput.host); diff != "" {
				t.Errorf("unexpected diff: %s for input %+v", diff, tc.inputCommand)
			}
		}
	}
}
