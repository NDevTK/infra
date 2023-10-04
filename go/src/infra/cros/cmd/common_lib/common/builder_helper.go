// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"fmt"
	"strings"

	_go "go.chromium.org/chromiumos/config/go"
	"go.chromium.org/chromiumos/config/go/test/api"
	testapi_metadata "go.chromium.org/chromiumos/config/go/test/api/metadata"
	"go.chromium.org/chromiumos/config/go/test/artifact"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"google.golang.org/protobuf/types/known/anypb"
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
func (constructor *CftCrosTestRunnerRequestConstructor) buildPrimaryDutProvision(orderedTasks *[]*skylab_test_runner.CrosTestRunnerRequest_Task) {
	if !constructor.Cft.GetStepsConfig().GetHwTestConfig().GetSkipStartingDutService() {
		AppendDutTask(orderedTasks, BuildCrosDutRequest(PrimaryDevice))
	}

	if !constructor.Cft.GetStepsConfig().GetHwTestConfig().GetSkipProvision() {
		AppendProvisionTask(orderedTasks,
			BuildProvisionContainerRequest(PrimaryDevice, IsAndroidProvisionState(constructor.Cft.GetPrimaryDut().GetProvisionState())),
			BuildProvisionRequest(PrimaryDevice, constructor.Cft.GetPrimaryDut()))
	}
}

// buildCompanionDutProvisions attempts to use the CompanionDuts from CftTestRequest
// to construct multiple cros-dut and provision task requests.
func (constructor *CftCrosTestRunnerRequestConstructor) buildCompanionDutProvisions(orderedTasks *[]*skylab_test_runner.CrosTestRunnerRequest_Task) {
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

		if !constructor.Cft.GetStepsConfig().GetHwTestConfig().GetSkipProvision() {
			AppendProvisionTask(orderedTasks,
				BuildProvisionContainerRequest(deviceId, IsAndroidProvisionState(dut.GetProvisionState())),
				BuildProvisionRequest(deviceId, dut))
		}
	}
}

// buildTestExecution attempts to construct a Test task request.
func (constructor *CftCrosTestRunnerRequestConstructor) buildTestExecution(orderedTasks *[]*skylab_test_runner.CrosTestRunnerRequest_Task) {
	if !constructor.Cft.GetStepsConfig().GetHwTestConfig().GetSkipTestExecution() {
		isCqRun := IsCqRun(constructor.Cft.GetTestSuites())
		platform := GetBotProvider()
		AppendTestTask(orderedTasks,
			BuildTestContainerRequest(isCqRun, platform),
			BuildTestRequest())
	}
}

// buildPublishes attempts to construct out each publish task request.
func (constructor *CftCrosTestRunnerRequestConstructor) buildPublishes(orderedTasks *[]*skylab_test_runner.CrosTestRunnerRequest_Task) {
	if !constructor.Cft.GetStepsConfig().GetHwTestConfig().GetSkipAllResultPublish() {
		constructor.buildRdbPublish(orderedTasks)
		constructor.buildGcsPublish(orderedTasks)
		constructor.buildCpconPublish(orderedTasks)
	}
}

// buildRdbPublish attempts to construct a RdbPublish task.
func (constructor *CftCrosTestRunnerRequestConstructor) buildRdbPublish(orderedTasks *[]*skylab_test_runner.CrosTestRunnerRequest_Task) {
	if !constructor.Cft.GetStepsConfig().GetHwTestConfig().GetSkipRdbPublish() {
		rdbPublishMetadata, _ := anypb.New(&testapi_metadata.PublishRdbMetadata{
			Sources: &testapi_metadata.PublishRdbMetadata_Sources{
				GsPath:            constructor.Cft.GetPrimaryDut().GetProvisionState().GetSystemImage().GetSystemImagePath().GetPath() + SourceMetadataPath,
				IsDeploymentDirty: constructor.Cft.GetPrimaryDut().GetProvisionState().GetFirmware() != nil || len(constructor.Cft.GetPrimaryDut().GetProvisionState().GetPackages()) > 0,
			},
			TestResult: &artifact.TestResult{},
		})
		AppendPublishTask(orderedTasks,
			BuildPublishContainerRequest("rdb-publish", api.CrosPublishTemplate_PUBLISH_RDB, nil),
			BuildPublishRequest(RdbPublishTestArtifactDir, rdbPublishMetadata, []*skylab_test_runner.DynamicDep{
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
func (constructor *CftCrosTestRunnerRequestConstructor) buildGcsPublish(orderedTasks *[]*skylab_test_runner.CrosTestRunnerRequest_Task) {
	if !constructor.Cft.GetStepsConfig().GetHwTestConfig().GetSkipGcsPublish() {
		gcsPublishMetadata, _ := anypb.New(&api.PublishGcsMetadata{
			GcsPath: &_go.StoragePath{
				HostType: _go.StoragePath_GS,
			},
		})
		AppendPublishTask(orderedTasks,
			BuildPublishContainerRequest("gcs-publish", api.CrosPublishTemplate_PUBLISH_GCS, []*skylab_test_runner.DynamicDep{
				{
					Key:   "crosPublish.publishSrcDir",
					Value: "env-TEMPDIR",
				},
			}),
			BuildPublishRequest(GcsPublishTestArtifactsDir, gcsPublishMetadata, []*skylab_test_runner.DynamicDep{
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
func (constructor *CftCrosTestRunnerRequestConstructor) buildCpconPublish(orderedTasks *[]*skylab_test_runner.CrosTestRunnerRequest_Task) {
	if constructor.Cft.GetStepsConfig().GetHwTestConfig().GetRunCpconPublish() {
		cpconMetadata, _ := anypb.New(&api.PublishTkoMetadata{})
		AppendPublishTask(orderedTasks,
			BuildPublishContainerRequest("cpcon-publish", api.CrosPublishTemplate_PUBLISH_CPCON, nil),
			BuildPublishRequest(CpconPublishTestArtifactsDir, cpconMetadata, []*skylab_test_runner.DynamicDep{
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
func BuildCrosDutRequest(deviceId string) *skylab_test_runner.ContainerRequest {
	return &skylab_test_runner.ContainerRequest{
		DynamicIdentifier: CrosDut + "-" + deviceId,
		Container: &api.Template{
			Container: &api.Template_CrosDut{
				CrosDut: &api.CrosDutTemplate{},
			},
		},
		ContainerImageKey: CrosDut,
		DynamicDeps: []*skylab_test_runner.DynamicDep{
			{
				Key:   "crosDut.cacheServer",
				Value: PrimaryDevice + ".dut.cacheServer.address",
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
func BuildProvisionRequest(deviceId string, device *skylab_test_runner.CFTTestRequest_Device) *skylab_test_runner.ProvisionRequest {
	var installRequest *api.InstallRequest
	var serviceAddress string
	if IsAndroidProvisionState(device.GetProvisionState()) {
		serviceAddress = AndroidProvision + "-" + deviceId
		installRequest = &api.InstallRequest{
			PreventReboot: false,
			Metadata:      device.GetProvisionState().GetProvisionMetadata(),
		}
	} else {
		serviceAddress = CrosProvision + "-" + deviceId
		crosProvisionMetadata, _ := anypb.New(&api.CrOSProvisionMetadata{})
		installRequest = &api.InstallRequest{
			ImagePath:     device.GetProvisionState().GetSystemImage().GetSystemImagePath(),
			PreventReboot: false,
			Metadata:      crosProvisionMetadata,
		}
	}
	return &skylab_test_runner.ProvisionRequest{
		ServiceAddress: &labapi.IpEndpoint{},
		InstallRequest: installRequest,
		DynamicDeps: []*skylab_test_runner.DynamicDep{
			{
				Key:   "serviceAddress",
				Value: serviceAddress,
			},
		},
		Target: deviceId,
	}
}

// BuildProvisionContainerRequest constructs a ContainerRequest for a certain deviceId
// with variations for android devices.
func BuildProvisionContainerRequest(deviceId string, isAndroid bool) *skylab_test_runner.ContainerRequest {
	var container *api.Template
	var imageKey string
	var key string
	if isAndroid {
		key = "androidProvision"
		imageKey = AndroidProvision
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
		imageKey = CrosProvision
		container = &api.Template{
			Container: &api.Template_CrosProvision{
				CrosProvision: &api.CrosProvisionTemplate{
					InputRequest: &api.CrosProvisionRequest{},
				},
			},
		}
	}
	return &skylab_test_runner.ContainerRequest{
		DynamicIdentifier: imageKey + "-" + deviceId,
		Container:         container,
		ContainerImageKey: imageKey,
		DynamicDeps: []*skylab_test_runner.DynamicDep{
			{
				Key:   key + ".inputRequest.dut",
				Value: deviceId + ".dut",
			},
			{
				Key:   key + ".inputRequest.dutServer",
				Value: CrosDut + "-" + deviceId,
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

// BuildTestContainerRequest constructs a ContainerRequest
// with the parameters for cros-test.
func BuildTestContainerRequest(isCqRun bool, platform SwarmingBotProvider) *skylab_test_runner.ContainerRequest {
	key := CrosTest
	if isCqRun && platform == BotProviderGce {
		key = "cros-test-cq-light"
	}
	return &skylab_test_runner.ContainerRequest{
		DynamicIdentifier: CrosTest,
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
func BuildTestRequest() *skylab_test_runner.TestRequest {
	return &skylab_test_runner.TestRequest{
		ServiceAddress: &labapi.IpEndpoint{},
		TestRequest:    &api.CrosTestRequest{},
		DynamicDeps: []*skylab_test_runner.DynamicDep{
			{
				Key:   "serviceAddress",
				Value: CrosTest,
			},
			{
				Key:   "testRequest.testSuites",
				Value: "req.params.testSuites",
			},
			{
				Key:   "testRequest.primary",
				Value: PrimaryDevice,
			},
			{
				Key:   "testRequest.companions",
				Value: CompanionDevices,
			},
		},
	}
}

// BuildPublishContainerRequest constructs a ContainerRequest for cros-publish
// using parameters marking the publish type.
func BuildPublishContainerRequest(identifier string, publishType api.CrosPublishTemplate_PublishType, deps []*skylab_test_runner.DynamicDep) *skylab_test_runner.ContainerRequest {
	return &skylab_test_runner.ContainerRequest{
		DynamicIdentifier: identifier,
		Container: &api.Template{
			Container: &api.Template_CrosPublish{
				CrosPublish: &api.CrosPublishTemplate{
					PublishType: publishType,
				},
			},
		},
		ContainerImageKey: CrosPublish,
		DynamicDeps:       deps,
	}
}

// BuildPublishRequest constructs a PublishRequest with provided dependencies.
func BuildPublishRequest(artifactPath string, metadata *anypb.Any, deps []*skylab_test_runner.DynamicDep) *skylab_test_runner.PublishRequest {
	return &skylab_test_runner.PublishRequest{
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
	orderedTasks *[]*skylab_test_runner.CrosTestRunnerRequest_Task,
	containerRequest *skylab_test_runner.ContainerRequest) {

	*orderedTasks = append(*orderedTasks, &skylab_test_runner.CrosTestRunnerRequest_Task{
		OrderedContainerRequests: []*skylab_test_runner.ContainerRequest{
			containerRequest,
		},
	})
}

// AppendProvisionTask takes a provision container request and
// a provision request and appends it to orderedTasks.
func AppendProvisionTask(
	orderedTasks *[]*skylab_test_runner.CrosTestRunnerRequest_Task,
	containerRequest *skylab_test_runner.ContainerRequest,
	provisionRequest *skylab_test_runner.ProvisionRequest) {

	*orderedTasks = append(*orderedTasks, &skylab_test_runner.CrosTestRunnerRequest_Task{
		OrderedContainerRequests: []*skylab_test_runner.ContainerRequest{
			containerRequest,
		},
		Task: &skylab_test_runner.CrosTestRunnerRequest_Task_Provision{
			Provision: provisionRequest,
		},
	})
}

// AppendTestTask takes a test container request and
// a test request and appends it to orderedTasks.
func AppendTestTask(
	orderedTasks *[]*skylab_test_runner.CrosTestRunnerRequest_Task,
	containerRequest *skylab_test_runner.ContainerRequest,
	testRequest *skylab_test_runner.TestRequest) {

	*orderedTasks = append(*orderedTasks, &skylab_test_runner.CrosTestRunnerRequest_Task{
		OrderedContainerRequests: []*skylab_test_runner.ContainerRequest{
			containerRequest,
		},
		Task: &skylab_test_runner.CrosTestRunnerRequest_Task_Test{
			Test: testRequest,
		},
	})
}

// AppendPublishTask takes a publish container request and
// a publish request and appends it to orderedTasks.
func AppendPublishTask(
	orderedTasks *[]*skylab_test_runner.CrosTestRunnerRequest_Task,
	containerRequest *skylab_test_runner.ContainerRequest,
	publishRequest *skylab_test_runner.PublishRequest,
	required bool) {

	*orderedTasks = append(*orderedTasks, &skylab_test_runner.CrosTestRunnerRequest_Task{
		Required: required,
		OrderedContainerRequests: []*skylab_test_runner.ContainerRequest{
			containerRequest,
		},
		Task: &skylab_test_runner.CrosTestRunnerRequest_Task_Publish{
			Publish: publishRequest,
		},
	})
}
