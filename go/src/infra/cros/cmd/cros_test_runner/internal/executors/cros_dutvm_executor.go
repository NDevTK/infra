// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package executors

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.chromium.org/chromiumos/config/go/test/api"
	testapi "go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"
	"google.golang.org/protobuf/types/known/durationpb"

	"infra/cros/cmd/cros_test_runner/common"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
	"infra/libs/vmlab"
	vmlabapi "infra/libs/vmlab/api"
)

// CrosDutVmExecutor implements the execution of the steps defined in supported
// DutVm-related commands.
type CrosDutVmExecutor struct {
	*interfaces.AbstractExecutor

	// Dependencies for Injection
	Container   interfaces.ContainerInterface
	ImageApi    vmlabapi.ImageApi
	InstanceApi vmlabapi.InstanceApi
}

func NewCrosDutVmExecutor(container interfaces.ContainerInterface) *CrosDutVmExecutor {
	absExec := interfaces.NewAbstractExecutor(CrosDutVmExecutorType)
	return &CrosDutVmExecutor{AbstractExecutor: absExec, Container: container}
}

// getImageApi takes a provider id and returns the image API interface
func (ex *CrosDutVmExecutor) getImageApi() (vmlabapi.ImageApi, error) {
	var err error
	if ex.ImageApi == nil {
		ex.ImageApi, err = vmlab.NewImageApi(vmlabapi.ProviderId_CLOUDSDK)
	}
	return ex.ImageApi, err
}

// getInstanceApi takes a provider id and returns the instance API interface
func (ex *CrosDutVmExecutor) getInstanceApi(providerId vmlabapi.ProviderId) (vmlabapi.InstanceApi, error) {
	var err error
	if ex.InstanceApi == nil {
		ex.InstanceApi, err = vmlab.NewInstanceApi(providerId)
	}
	return ex.InstanceApi, err
}

// getProviderId returns the VM instance provider ID based on build flags
func getProviderId(ctx context.Context, buildState *build.State) vmlabapi.ProviderId {
	if buildState == nil {
		return vmlabapi.ProviderId_GCLOUD
	}
	experiments := buildState.Build().GetInput().GetExperiments()
	for _, v := range experiments {
		if v == common.VmLeaserExperimentStr {
			return vmlabapi.ProviderId_VM_LEASER
		}
	}
	return vmlabapi.ProviderId_GCLOUD
}

func (ex *CrosDutVmExecutor) ExecuteCommand(
	ctx context.Context,
	cmdInterface interfaces.CommandInterface) error {

	switch cmd := cmdInterface.(type) {
	case *commands.DutVmGetImageCmd:
		return ex.vmGetImageCommandExecution(ctx, cmd)
	case *commands.DutVmLeaseCmd:
		return ex.vmLeaseCommandExecution(ctx, cmd)
	case *commands.DutVmReleaseCmd:
		return ex.vmReleaseCommandExecution(ctx, cmd)
	// For MVP, cros-dut container is the same for both HW and VM tests. In future
	// VM tests may have a dedicated cros-dut-vm container.
	case *commands.DutServiceStartCmd:
		return ex.dutStartCommandExecution(ctx, cmd)
	default:
		return fmt.Errorf("Command type %s is not supported by %s executor type!", cmd.GetCommandType(), ex.GetExecutorType())
	}
}

// vmGetImageCommandExecution executes the "Get Dut VM GCE image" step.
func (ex *CrosDutVmExecutor) vmGetImageCommandExecution(
	ctx context.Context,
	cmd *commands.DutVmGetImageCmd) error {
	var err error
	step, ctx := build.StartStep(ctx, "Get Dut VM GCE image")
	defer func() { step.End(err) }()

	imageApi, err := ex.getImageApi()
	if err != nil {
		return err
	}

	buildPath := common.GetValueFromRequestKeyvals(ctx, cmd.CftTestRequest, "build")
	logging.Infof(ctx, "call VMLab to get GCE image based on build path in CftTestRequest: %s, ", buildPath)
	// For MVP, this call may take minutes: If image doesn't exist, GetImage
	// will create an image on the fly.
	gceImage, err := imageApi.GetImage(buildPath, true)

	if gceImage != nil {
		logging.Infof(ctx, "got gceImage from VMLab: %v", gceImage)
	}
	cmd.DutVmGceImage = gceImage
	return err
}

// vmReleaseCommandExecution executes the "Release Dut VM" step. This step is
// non-critical: all errors will be logged as warnings.
func (ex *CrosDutVmExecutor) vmReleaseCommandExecution(
	ctx context.Context,
	cmd *commands.DutVmReleaseCmd) error {
	var err error
	step, ctx := build.StartStep(ctx, "Release Dut VM")
	defer func() { step.End(err) }()

	providerId := getProviderId(ctx, cmd.BuildState)
	instanceApi, err := ex.getInstanceApi(providerId)
	if err != nil {
		logging.Warningf(ctx, "failed to get instance API from vmlab: %v", err)
		return nil
	}
	logging.Infof(ctx, "got instance api: %v", instanceApi)

	err = instanceApi.Delete(ctx, cmd.DutVm)
	if err == nil {
		logging.Infof(ctx, "successfully released Dut VM: %s", cmd.DutVm.Name)
	} else {
		logging.Warningf(ctx, "failed to delete instance by vmlab: %v", err)
	}
	return nil
}

// vmLeaseCommandExecution executes the "Lease Dut VM" step.
func (ex *CrosDutVmExecutor) vmLeaseCommandExecution(
	ctx context.Context,
	cmd *commands.DutVmLeaseCmd) error {
	var err error
	step, ctx := build.StartStep(ctx, "Lease Dut VM")
	defer func() { step.End(err) }()

	providerId := getProviderId(ctx, cmd.BuildState)
	instanceApi, err := ex.getInstanceApi(providerId)
	if err != nil {
		logging.Warningf(ctx, "failed to get instance API from vmlab: %v", err)
		return err
	}
	logging.Infof(ctx, "got instance api: %v", instanceApi)

	var request *vmlabapi.CreateVmInstanceRequest
	switch providerId {
	case vmlabapi.ProviderId_VM_LEASER:
		request = constructVmLeaserPayload(ctx, cmd)
	case vmlabapi.ProviderId_GCLOUD:
		request = constructGcloudPayload(ctx, cmd)
	default:
		request = constructGcloudPayload(ctx, cmd)
	}

	logging.Infof(ctx, "call VmLab to lease an instance with request %v", request)
	instance, err := instanceApi.Create(ctx, request)
	if err != nil {
		return errors.Annotate(err, "Lease dut vm err: ").Err()
	}
	logging.Infof(ctx, "leased instance from VmLab %v", instance)
	cmd.DutVm = instance

	logging.Infof(ctx, "wait for SSH to become available")
	common.WaitDutVmBoot(ctx, instance.GetSsh().GetAddress())

	logging.Infof(ctx, "completed wait for SSH")
	return err
}

// dutStartCommandExecution executes the "Start Dut service" step. This is
// mostly the same as how to start dut service for HW tests, with the exception
// that the cache server address needs to be switched from localhost to an IP
// address in order to be accessible from the Dut VM.
func (ex *CrosDutVmExecutor) dutStartCommandExecution(
	ctx context.Context,
	cmd *commands.DutServiceStartCmd) error {

	var err error
	step, ctx := build.StartStep(ctx, "Dut service start")
	defer func() { step.End(err) }()

	template := &api.Template{
		Container: &api.Template_CrosDut{
			CrosDut: &testapi.CrosDutTemplate{
				CacheServer: cmd.CacheServerAddress,
				DutAddress:  cmd.DutSshAddress,
			},
		},
	}

	// Process container.
	serverAddress, err := ex.Container.ProcessContainer(ctx, template)
	if err != nil {
		return errors.Annotate(err, "error processing container: ").Err()
	}

	// Process dut server address.
	dutServerAddress, err := common.GetIpEndpoint(serverAddress)

	logging.Infof(ctx, "Cros-dut started at address: %v", dutServerAddress)
	cmd.DutServerAddress = dutServerAddress
	return err
}

// constructGcloudPayload returns a VM lease request intended to use the Gcloud
// provider backend.
//
// TODO(b/274006123): remove hardcoded configs
func constructGcloudPayload(ctx context.Context, cmd *commands.DutVmLeaseCmd) *vmlabapi.CreateVmInstanceRequest {
	tags := make(map[string]string, 0)
	tags["swarming-bot-name"] = os.Getenv("SWARMING_BOT_ID")
	return &vmlabapi.CreateVmInstanceRequest{
		Config: &vmlabapi.Config{
			Backend: &vmlabapi.Config_GcloudBackend{
				GcloudBackend: &vmlabapi.Config_GCloudBackend{
					Project:        "chromeos-gce-tests",
					Zone:           "us-central1-a",
					MachineType:    "n2-standard-4",
					InstancePrefix: "ctsprototype-",
					PublicIp:       false,
					Image:          cmd.DutVmGceImage,
					Network:        "chromeos-gce-tests",
					Subnet:         "us-central1",
				},
			},
		},
		Tags: tags,
	}
}

// constructVmLeaserPayload returns a VM lease request intended to use the VM
// Leaser provider backend.
//
// TODO(b/274006123): remove hardcoded configs
func constructVmLeaserPayload(ctx context.Context, cmd *commands.DutVmLeaseCmd) *vmlabapi.CreateVmInstanceRequest {
	img := fmt.Sprintf("projects/%v/global/images/%v", cmd.DutVmGceImage.GetProject(), cmd.DutVmGceImage.GetName())
	d, _ := time.ParseDuration("1d")
	return &vmlabapi.CreateVmInstanceRequest{
		Config: &vmlabapi.Config{
			Backend: &vmlabapi.Config_VmLeaserBackend_{
				VmLeaserBackend: &vmlabapi.Config_VmLeaserBackend{
					Env: vmlabapi.Config_VmLeaserBackend_ENV_STAGING,
					VmRequirements: &api.VMRequirements{
						GceImage:       img,
						GceRegion:      "us-central1-a",
						GceProject:     "chromeos-gce-tests",
						GceMachineType: "n2-standard-4",
						GceNetwork:     "global/networks/chromeos-gce-tests",
						GceSubnet:      "regions/us-central1/subnetworks/us-central1",
						GceDiskSize:    20,
					},
					LeaseDuration: durationpb.New(d),
				},
			},
		},
	}
}
