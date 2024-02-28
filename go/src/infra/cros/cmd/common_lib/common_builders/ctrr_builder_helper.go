// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common_builders

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/known/anypb"

	_go "go.chromium.org/chromiumos/config/go"
	"go.chromium.org/chromiumos/config/go/test/api"
	testapi_metadata "go.chromium.org/chromiumos/config/go/test/api/metadata"
	"go.chromium.org/chromiumos/config/go/test/artifact"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"

	"infra/cros/cmd/common_lib/common"
)

// addDevicesInfoToKeyvals modifies the keyvals within CrosTestRunnerRequest_Params.
// Adds the metadata key for containers as well as the build_target values from devices.
func (constructor *CftCrosTestRunnerRequestConstructor) addDevicesInfoToKeyvals(keyvals map[string]string) {
	if _, ok := keyvals["build_target"]; !ok && constructor.Cft.GetPrimaryDut().GetContainerMetadataKey() != "" {
		keyvals["build_target"] = constructor.Cft.GetPrimaryDut().GetContainerMetadataKey()
	}
	if constructor.Cft.GetPrimaryDut() != nil && constructor.Cft.GetPrimaryDut().GetDutModel().GetBuildTarget() != "" {
		keyvals["primary-board"] = constructor.Cft.GetPrimaryDut().GetDutModel().GetBuildTarget()
	}
	companionBoards := []string{}
	for _, companion := range constructor.Cft.GetCompanionDuts() {
		companionBoards = append(companionBoards, companion.GetDutModel().GetBuildTarget())
	}
	if len(companionBoards) > 0 {
		keyvals["companion-boards"] = strings.Join(companionBoards, ",")
	}
}

// buildPrimaryDutProvision attempts to use the PrimaryDut from CftTestRequest
// to construct a cros-dut and provision task request.
func (constructor *CftCrosTestRunnerRequestConstructor) buildPrimaryDutProvision(orderedTasks *[]*api.CrosTestRunnerDynamicRequest_Task) {
	if !constructor.Cft.GetStepsConfig().GetHwTestConfig().GetSkipStartingDutService() {
		AppendDutTask(orderedTasks, BuildCrosDutRequest(common.PrimaryDevice))
	}

	constructor.buildProvision(common.PrimaryDevice, constructor.Cft.GetPrimaryDut(), orderedTasks)
}

// buildCompanionDutProvisions attempts to use the CompanionDuts from CftTestRequest
// to construct multiple cros-dut and provision task requests.
func (constructor *CftCrosTestRunnerRequestConstructor) buildCompanionDutProvisions(orderedTasks *[]*api.CrosTestRunnerDynamicRequest_Task) {
	deviceIds := map[string]struct{}{}
	for _, dut := range constructor.Cft.GetCompanionDuts() {
		deviceId := "companionDevice_" + dut.GetDutModel().GetBuildTarget()
		if _, ok := deviceIds[deviceId]; ok {
			// deviceId already exists, try postfixing
			// Standard within swarming when there are duplicate boards
			// is to postfix with `_2`. (e.g. `brya | brya_2`)
			postfix := 2
			for {
				if _, ok := deviceIds[fmt.Sprintf("%s_%d", deviceId, postfix)]; !ok {
					deviceId = fmt.Sprintf("%s_%d", deviceId, postfix)
					break
				}
				postfix += 1
			}
		}
		deviceIds[deviceId] = struct{}{}
		if !constructor.Cft.GetStepsConfig().GetHwTestConfig().GetSkipStartingDutService() {
			AppendDutTask(orderedTasks, BuildCrosDutRequest(deviceId))
		}

		constructor.buildProvision(deviceId, dut, orderedTasks)
	}
}

// buildProvision checks for each possible type of provision that might occur
// and calls into the corresponding provision builder function.
func (constructor *CftCrosTestRunnerRequestConstructor) buildProvision(
	deviceId string,
	dut *skylab_test_runner.CFTTestRequest_Device,
	orderedTasks *[]*api.CrosTestRunnerDynamicRequest_Task) {

	if !constructor.Cft.GetStepsConfig().GetHwTestConfig().GetSkipProvision() {
		AppendProvisionTask(orderedTasks,
			BuildProvisionContainerRequest(deviceId, IsAndroidProvisionState(dut.GetProvisionState())),
			BuildProvisionRequest(deviceId, dut))

		if ContainsFwProvisionState(dut.GetProvisionState()) {
			AppendProvisionTask(orderedTasks,
				BuildFwProvisionContainerRequest(deviceId),
				BuildFwProvisionRequest(deviceId, dut))
		}
	}
}

// buildTestExecution attempts to construct a Test task request.
func (constructor *CftCrosTestRunnerRequestConstructor) buildTestExecution(orderedTasks *[]*api.CrosTestRunnerDynamicRequest_Task) {
	if !constructor.Cft.GetStepsConfig().GetHwTestConfig().GetSkipTestExecution() {
		isCqRun := common.IsCqRun(constructor.Cft.GetTestSuites())
		platform := common.GetBotProvider()
		AppendTestTask(orderedTasks,
			BuildTestContainerRequest(isCqRun, platform),
			BuildTestRequest())
	}
}

// buildPublishes attempts to construct out each publish task request.
func (constructor *CftCrosTestRunnerRequestConstructor) buildPublishes(orderedTasks *[]*api.CrosTestRunnerDynamicRequest_Task) {
	if !constructor.Cft.GetStepsConfig().GetHwTestConfig().GetSkipAllResultPublish() {
		constructor.buildRdbPublish(orderedTasks)
		constructor.buildGcsPublish(orderedTasks)
		constructor.buildCpconPublish(orderedTasks)
	}
}

// buildRdbPublish attempts to construct a RdbPublish task.
func (constructor *CftCrosTestRunnerRequestConstructor) buildRdbPublish(orderedTasks *[]*api.CrosTestRunnerDynamicRequest_Task) {
	if !constructor.Cft.GetStepsConfig().GetHwTestConfig().GetSkipRdbPublish() {
		rdbPublishMetadata, _ := anypb.New(&testapi_metadata.PublishRdbMetadata{
			Sources: &testapi_metadata.PublishRdbMetadata_Sources{
				GsPath:            constructor.Cft.GetPrimaryDut().GetProvisionState().GetSystemImage().GetSystemImagePath().GetPath() + common.SourceMetadataPath,
				IsDeploymentDirty: constructor.Cft.GetPrimaryDut().GetProvisionState().GetFirmware() != nil || len(constructor.Cft.GetPrimaryDut().GetProvisionState().GetPackages()) > 0,
			},
			TestResult: &artifact.TestResult{},
		})
		AppendPublishTask(orderedTasks,
			BuildPublishContainerRequest("rdb-publish", api.CrosPublishTemplate_PUBLISH_RDB, nil),
			BuildPublishRequest(common.RdbPublishTestArtifactDir, rdbPublishMetadata, []*api.DynamicDep{
				{
					Key:   "serviceAddress",
					Value: "rdb-publish",
				},
				{
					Key:   "publishRequest.metadata.currentInvocationId",
					Value: "invocation-id",
				},
				{
					Key:   "publishRequest.metadata.testResult",
					Value: "rdb-test-result",
				},
			}),
			false)
	}
}

// buildRdbPublish attempts to construct a GcsPublish task.
func (constructor *CftCrosTestRunnerRequestConstructor) buildGcsPublish(orderedTasks *[]*api.CrosTestRunnerDynamicRequest_Task) {
	if !constructor.Cft.GetStepsConfig().GetHwTestConfig().GetSkipGcsPublish() {
		gcsPublishMetadata, _ := anypb.New(&api.PublishGcsMetadata{
			GcsPath: &_go.StoragePath{
				HostType: _go.StoragePath_GS,
			},
		})
		AppendPublishTask(orderedTasks,
			BuildPublishContainerRequest("gcs-publish", api.CrosPublishTemplate_PUBLISH_GCS, []*api.DynamicDep{
				{
					Key:   "crosPublish.publishSrcDir",
					Value: "env-TEMPDIR",
				},
			}),
			BuildPublishRequest(common.GcsPublishTestArtifactsDir, gcsPublishMetadata, []*api.DynamicDep{
				{
					Key:   "serviceAddress",
					Value: "gcs-publish",
				},
				{
					Key:   "publishRequest.metadata.gcsPath.path",
					Value: "gcs-url",
				},
			}),
			true)
	}
}

// buildRdbPublish attempts to construct a CpconPublish task.
func (constructor *CftCrosTestRunnerRequestConstructor) buildCpconPublish(orderedTasks *[]*api.CrosTestRunnerDynamicRequest_Task) {
	if constructor.Cft.GetStepsConfig().GetHwTestConfig().GetRunCpconPublish() {
		cpconMetadata, _ := anypb.New(&api.PublishTkoMetadata{})
		AppendPublishTask(orderedTasks,
			BuildPublishContainerRequest("cpcon-publish", api.CrosPublishTemplate_PUBLISH_CPCON, nil),
			BuildPublishRequest(common.CpconPublishTestArtifactsDir, cpconMetadata, []*api.DynamicDep{
				{
					Key:   "serviceAddress",
					Value: "cpcon-publish",
				},
				{
					Key:   "publishRequest.metadata.jobName",
					Value: "jobname",
				},
			}),
			true)
	}
}

// BuildCrosDutRequest is a helper function to construct a ContainerRequest
// with the CrosDut template, using a deviceId to dynamically execute it.
func BuildCrosDutRequest(deviceId string) *api.ContainerRequest {
	return &api.ContainerRequest{
		DynamicIdentifier: common.CrosDut + "-" + deviceId,
		Container: &api.Template{
			Container: &api.Template_CrosDut{
				CrosDut: &api.CrosDutTemplate{},
			},
		},
		ContainerImageKey: common.CrosDut,
		DynamicDeps: []*api.DynamicDep{
			{
				Key:   "crosDut.cacheServer",
				Value: common.PrimaryDevice + ".dut.cacheServer.address",
			},
			{
				Key:   "crosDut.dutAddress",
				Value: deviceId + ".dutServer",
			},
		},
	}
}

// BuildProvisionRequest takes a Cft device to construct a ProvisionRequest.
// Checks provision state to determine install request.
func BuildProvisionRequest(deviceId string, device *skylab_test_runner.CFTTestRequest_Device) *api.ProvisionTask {
	var installRequest *api.InstallRequest
	var serviceAddress string
	if IsAndroidProvisionState(device.GetProvisionState()) {
		serviceAddress = common.AndroidProvision + "-" + deviceId
		installRequest = &api.InstallRequest{
			PreventReboot: false,
			Metadata:      device.GetProvisionState().GetProvisionMetadata(),
		}
	} else {
		serviceAddress = common.CrosProvision + "-" + deviceId
		crosProvisionMetadata, _ := anypb.New(&api.CrOSProvisionMetadata{})
		installRequest = &api.InstallRequest{
			ImagePath:     device.GetProvisionState().GetSystemImage().GetSystemImagePath(),
			PreventReboot: false,
			Metadata:      crosProvisionMetadata,
		}
	}
	return &api.ProvisionTask{
		ServiceAddress: &labapi.IpEndpoint{},
		InstallRequest: installRequest,
		DynamicDeps: []*api.DynamicDep{
			{
				Key:   "serviceAddress",
				Value: serviceAddress,
			},
		},
		Target: deviceId,
	}
}

// BuildFwProvisionRequest creates a generic provision request using the FirmwareConfig
// within the device's provided provision state as part of the install request.
func BuildFwProvisionRequest(deviceId string, device *skylab_test_runner.CFTTestRequest_Device) *api.ProvisionTask {
	startUpMetadata, _ := anypb.New(&api.FirmwareProvisionStartupMetadata{})
	installMetadata, _ := anypb.New(&api.FirmwareProvisionInstallMetadata{
		FirmwareConfig: device.GetProvisionState().GetFirmware(),
	})
	return &api.ProvisionTask{
		ServiceAddress: &labapi.IpEndpoint{},
		StartupRequest: &api.ProvisionStartupRequest{
			Metadata: startUpMetadata,
		},
		InstallRequest: &api.InstallRequest{
			Metadata: installMetadata,
		},
		DynamicDeps: []*api.DynamicDep{
			{
				Key:   "serviceAddress",
				Value: common.FwProvision + "-" + deviceId,
			},
			{
				Key:   "startupRequest.dut",
				Value: deviceId + ".dut",
			},
			{
				Key:   "startupRequest.dutServer",
				Value: common.CrosDut + "-" + deviceId,
			},
		},
		Target: deviceId,
	}
}

// BuildFwProvisionContainerRequest creates a container request for a certain deviceId,
// specifically geared towards supported cros-fw-provisions.
func BuildFwProvisionContainerRequest(deviceId string) *api.ContainerRequest {
	return &api.ContainerRequest{
		DynamicIdentifier: common.FwProvision + "-" + deviceId,
		ContainerImageKey: common.FwProvision,
		Container: &api.Template{
			Container: &api.Template_Generic{
				Generic: &api.GenericTemplate{
					BinaryName: "cros-fw-provision",
					BinaryArgs: []string{
						"server",
						"-port", "0",
					},
					DockerArtifactDir: "/tmp/cros-fw-provision",
					AdditionalVolumes: []string{
						"/creds:/creds",
					},
				},
			},
		},
		DynamicDeps: []*api.DynamicDep{},
	}
}

// BuildProvisionContainerRequest constructs a ContainerRequest for a certain deviceId
// with variations for android devices.
func BuildProvisionContainerRequest(deviceId string, isAndroid bool) *api.ContainerRequest {
	var container *api.Template
	var imageKey string
	var key string
	if isAndroid {
		key = "androidProvision"
		imageKey = common.AndroidProvision
		container = &api.Template{
			Container: &api.Template_Generic{
				Generic: &api.GenericTemplate{
					BinaryName: "android-provision",
					BinaryArgs: []string{
						"server",
						"-port", "0",
					},
					DockerArtifactDir: "/tmp/provision",
					AdditionalVolumes: []string{
						"/creds:/creds",
					},
				},
			},
		}
	} else {
		key = "crosProvision"
		imageKey = common.CrosProvision
		container = &api.Template{
			Container: &api.Template_CrosProvision{
				CrosProvision: &api.CrosProvisionTemplate{
					InputRequest: &api.CrosProvisionRequest{},
				},
			},
		}
	}
	return &api.ContainerRequest{
		DynamicIdentifier: imageKey + "-" + deviceId,
		Container:         container,
		ContainerImageKey: imageKey,
		DynamicDeps: []*api.DynamicDep{
			{
				Key:   key + ".inputRequest.dut",
				Value: deviceId + ".dut",
			},
			{
				Key:   key + ".inputRequest.dutServer",
				Value: common.CrosDut + "-" + deviceId,
			},
		},
	}
}

// IsAndroidProvisionState checks if the metadata of the
// provision state can unmarshal to an android metadata.
func IsAndroidProvisionState(state *api.ProvisionState) bool {
	androidMetadata := &api.AndroidProvisionRequestMetadata{}
	err := state.GetProvisionMetadata().UnmarshalTo(androidMetadata)
	if err != nil {
		return false
	}
	return true
}

// ContainsFwProvisionState checks if there is fw provision info in the
// provision state.
func ContainsFwProvisionState(state *api.ProvisionState) bool {
	return state != nil && state.Firmware != nil
}

// BuildTestContainerRequest constructs a ContainerRequest
// with the parameters for cros-test.
func BuildTestContainerRequest(isCqRun bool, platform common.SwarmingBotProvider) *api.ContainerRequest {
	key := common.CrosTest
	if isCqRun && platform == common.BotProviderGce {
		key = "cros-test-cq-light"
	}
	return &api.ContainerRequest{
		DynamicIdentifier: common.CrosTest,
		Container: &api.Template{
			Container: &api.Template_CrosTest{
				CrosTest: &api.CrosTestTemplate{},
			},
		},
		ContainerImageKey: key,
	}
}

// BuildTestRequest constructs a TestRequest using
// default dependencies.
func BuildTestRequest() *api.TestTask {
	return &api.TestTask{
		ServiceAddress: &labapi.IpEndpoint{},
		TestRequest:    &api.CrosTestRequest{},
		DynamicDeps: []*api.DynamicDep{
			{
				Key:   "serviceAddress",
				Value: common.CrosTest,
			},
			{
				Key:   "testRequest.testSuites",
				Value: "req.params.testSuites",
			},
			{
				Key:   "testRequest.primary",
				Value: common.PrimaryDevice,
			},
			{
				Key:   "testRequest.companions",
				Value: common.CompanionDevices,
			},
		},
	}
}

// BuildPublishContainerRequest constructs a ContainerRequest for cros-publish
// using parameters marking the publish type.
func BuildPublishContainerRequest(identifier string, publishType api.CrosPublishTemplate_PublishType, deps []*api.DynamicDep) *api.ContainerRequest {
	return &api.ContainerRequest{
		DynamicIdentifier: identifier,
		Container: &api.Template{
			Container: &api.Template_CrosPublish{
				CrosPublish: &api.CrosPublishTemplate{
					PublishType: publishType,
				},
			},
		},
		ContainerImageKey: common.CrosPublish,
		DynamicDeps:       deps,
	}
}

// BuildPublishRequest constructs a PublishRequest with provided dependencies.
func BuildPublishRequest(artifactPath string, metadata *anypb.Any, deps []*api.DynamicDep) *api.PublishTask {
	return &api.PublishTask{
		ServiceAddress: &labapi.IpEndpoint{},
		PublishRequest: &api.PublishRequest{
			ArtifactDirPath: &_go.StoragePath{
				HostType: _go.StoragePath_LOCAL,
				Path:     artifactPath},
			Metadata: metadata,
		},
		DynamicDeps: deps,
	}
}

// AppendDutTask takes a DutContainer request and appends it
// to orderedTasks.
func AppendDutTask(
	orderedTasks *[]*api.CrosTestRunnerDynamicRequest_Task,
	containerRequest *api.ContainerRequest) {

	*orderedTasks = append(*orderedTasks, &api.CrosTestRunnerDynamicRequest_Task{
		OrderedContainerRequests: []*api.ContainerRequest{
			containerRequest,
		},
	})
}

// AppendProvisionTask takes a provision container request and
// a provision request and appends it to orderedTasks.
func AppendProvisionTask(
	orderedTasks *[]*api.CrosTestRunnerDynamicRequest_Task,
	containerRequest *api.ContainerRequest,
	provisionRequest *api.ProvisionTask) {

	*orderedTasks = append(*orderedTasks, &api.CrosTestRunnerDynamicRequest_Task{
		OrderedContainerRequests: []*api.ContainerRequest{
			containerRequest,
		},
		Task: &api.CrosTestRunnerDynamicRequest_Task_Provision{
			Provision: provisionRequest,
		},
	})
}

// AppendTestTask takes a test container request and
// a test request and appends it to orderedTasks.
func AppendTestTask(
	orderedTasks *[]*api.CrosTestRunnerDynamicRequest_Task,
	containerRequest *api.ContainerRequest,
	testRequest *api.TestTask) {

	*orderedTasks = append(*orderedTasks, &api.CrosTestRunnerDynamicRequest_Task{
		OrderedContainerRequests: []*api.ContainerRequest{
			containerRequest,
		},
		Task: &api.CrosTestRunnerDynamicRequest_Task_Test{
			Test: testRequest,
		},
	})
}

// AppendPublishTask takes a publish container request and
// a publish request and appends it to orderedTasks.
func AppendPublishTask(
	orderedTasks *[]*api.CrosTestRunnerDynamicRequest_Task,
	containerRequest *api.ContainerRequest,
	publishRequest *api.PublishTask,
	required bool) {

	*orderedTasks = append(*orderedTasks, &api.CrosTestRunnerDynamicRequest_Task{
		Required: required,
		OrderedContainerRequests: []*api.ContainerRequest{
			containerRequest,
		},
		Task: &api.CrosTestRunnerDynamicRequest_Task_Publish{
			Publish: publishRequest,
		},
	})
}
