// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package executions

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"

	api_common "go.chromium.org/chromiumos/infra/proto/go/test_platform/common"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner/steps"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"
	"google.golang.org/protobuf/proto"

	"infra/cros/cmd/common_lib/common"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
	"infra/cros/cmd/cros_test_runner/data"
	"infra/cros/cmd/cros_test_runner/internal/configs"
	"infra/cros/cmd/cros_test_runner/protos"
)

// TODO : Re-structure different execution flow properly later.
// HwExecution represents hw executions.
func HwExecution() {
	// Set input property reader functions
	var ctrCipdInfoReader func(context.Context) *protos.CipdVersionInfo
	build.MakePropertyReader(common.HwTestCtrInputPropertyName, &ctrCipdInfoReader)
	input := &steps.RunTestsRequest{}

	// Set output props writer functions
	var writeOutputProps func(*steps.RunTestsResponse)
	var mergeOutputProps func(*steps.RunTestsResponse)

	build.Main(input, &writeOutputProps, &mergeOutputProps,
		func(ctx context.Context, args []string, st *build.State) error {
			log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmsgprefix)
			logging.Infof(ctx, "have input %v", input)
			ctrCipdInfo := ctrCipdInfoReader(ctx)
			logging.Infof(ctx, "have ctr info: %v", ctrCipdInfo)
			logging.Infof(ctx, "ctr label: %s", ctrCipdInfo.GetVersion().GetCipdLabel())
			resp := &steps.RunTestsResponse{}
			// TODO (azrahman): After stablizing in prod, move log data gs root to cft/new proto.
			var skylabResult *skylab_test_runner.Result
			var err error
			if input.CrosTestRunnerRequest != nil {
				// If the request is a CrosTestRunner dynamic request...
				skylabResult, err = executeHwTestsV2(ctx, nil, input.CrosTestRunnerRequest, ctrCipdInfo.GetVersion().GetCipdLabel(), input.GetConfig().GetOutput().GetLogDataGsRoot(), st)
			} else if input.CftTestRequest.TranslateTrv2Request {
				// If the request is a CrosTestRunner non-dynamic request with translation flag...
				builder := &common.CrosTestRunnerRequestBuilder{}
				constructor := &common.CftCrosTestRunnerRequestConstructor{
					Cft: input.CftTestRequest,
				}
				crosTestRunnerRequest := builder.Build(constructor)
				skylabResult, err = executeHwTestsV2(ctx, input.CftTestRequest, crosTestRunnerRequest, ctrCipdInfo.GetVersion().GetCipdLabel(), input.GetConfig().GetOutput().GetLogDataGsRoot(), st)
			} else {
				// If the request is a CrosTestRunner non-dynamic request...
				skylabResult, err = executeHwTests(ctx, input.CftTestRequest, ctrCipdInfo.GetVersion().GetCipdLabel(), input.GetConfig().GetOutput().GetLogDataGsRoot(), st)
			}
			if skylabResult != nil {
				m, _ := proto.Marshal(skylabResult)
				var b bytes.Buffer
				w := zlib.NewWriter(&b)
				_, _ = w.Write(m)
				_ = w.Close()
				resp.CompressedResult = base64.StdEncoding.EncodeToString(b.Bytes())
			}
			if err != nil {
				if common.GlobalNonInfraError != nil {
					err = common.GlobalNonInfraError
				} else {
					err = build.AttachStatus(err, buildbucketpb.Status_INFRA_FAILURE, nil)
				}
				logging.Infof(ctx, "error found: %s", err)
				st.SetSummaryMarkdown(err.Error())
				resp.ErrorSummaryMarkdown = err.Error()
			}

			writeOutputProps(resp)
			return err
		},
	)
}

// executeHwTests executes hw tests
func executeHwTests(
	ctx context.Context,
	req *skylab_test_runner.CFTTestRequest,
	ctrCipdVersion string,
	gsRoot string,
	buildState *build.State) (*skylab_test_runner.Result, error) {

	// Validation
	err := validateHwExecution(ctrCipdVersion, gsRoot)
	if err != nil {
		return nil, err
	}

	// Create ctr
	ctr := setupCtr(ctrCipdVersion)

	// Create configs
	metadataContainers := req.GetContainerMetadata().GetContainers()
	metadataKey := req.GetPrimaryDut().GetContainerMetadataKey()
	metadataMap, ok := metadataContainers[metadataKey]
	if !ok {
		return nil, fmt.Errorf("Provided key %q does not exist in provided container metadata.", metadataKey)
	}
	dockerKeyFile, err := common.LocateFile([]string{common.LabDockerKeyFileLocation, common.VmLabDockerKeyFileLocation})
	if err != nil {
		return nil, fmt.Errorf("unable to locate dockerKeyFile during initialization: %w", err)
	}
	cqRun := common.IsCqRun(req.TestSuites)
	containerImagesMap := metadataMap.GetImages()
	containerCfg := configs.NewContainerConfig(ctr, containerImagesMap, cqRun)
	executorCfg := configs.NewExecutorConfig(ctr, containerCfg)
	cmdCfg := configs.NewCommandConfig(executorCfg)

	// Create state keeper
	gcsurl := common.GetGcsUrl(gsRoot)
	sk := data.NewHwTestStateKeeper()
	sk.BuildState = buildState
	sk.CftTestRequest = req
	sk.Ctr = ctr
	sk.DockerKeyFileLocation = dockerKeyFile
	sk.GcsPublishSrcDir = os.Getenv("TEMPDIR")
	sk.GcsUrl = gcsurl
	sk.TesthausUrl = common.GetTesthausUrl(gcsurl)
	sk.ContainerImages = containerImagesMap

	if sk.CftTestRequest.GetPrimaryDut() != nil {
		sk.CftTestRequest.AutotestKeyvals["primary-board"] = sk.CftTestRequest.GetPrimaryDut().GetDutModel().GetBuildTarget()
	}
	companionBoards := []string{}
	for _, companion := range sk.CftTestRequest.GetCompanionDuts() {
		companionBoards = append(companionBoards, companion.GetDutModel().GetBuildTarget())
	}
	sk.CftTestRequest.AutotestKeyvals["companion-boards"] = strings.Join(companionBoards, ",")

	// For demonstration/logging purposes.
	common.LogWarningIfErr(ctx, sk.Injectables.Set("req", req))
	common.LogWarningIfErr(ctx, sk.Injectables.Set("botDims", buildState.Build().GetInfra().GetSwarming().GetBotDimensions()))
	common.LogWarningIfErr(ctx, sk.Injectables.Set("gcs-url", gcsurl))
	common.LogWarningIfErr(ctx, sk.Injectables.Set("testhaus-url", common.GetTesthausUrl(gcsurl)))

	// Generate config
	hwTestConfig := configs.NewTrv2ExecutionConfig(configs.HwTestExecutionConfigType, cmdCfg, sk, req.GetStepsConfig())
	err = hwTestConfig.GenerateConfig(ctx)
	if err != nil {
		return sk.SkylabResult, errors.Annotate(err, "error during generating hw test configs: ").Err()
	}

	// Execute config
	err = hwTestConfig.Execute(ctx)
	// For demonstration/logging purposes.
	sk.Injectables.LogStorageToBuild(ctx, buildState)
	if err != nil {
		return sk.SkylabResult, errors.Annotate(err, "error during executing hw test configs: ").Err()
	}
	return sk.SkylabResult, nil
}

// executeHwTestsV2 uses the dynamic CrosTestRunner request to construct
// a hardware test execution environment.
func executeHwTestsV2(
	ctx context.Context,
	cft *skylab_test_runner.CFTTestRequest,
	req *skylab_test_runner.CrosTestRunnerRequest,
	ctrCipdVersion string,
	gsRoot string,
	buildState *build.State) (*skylab_test_runner.Result, error) {

	// Validation
	err := validateHwExecution(ctrCipdVersion, gsRoot)
	if err != nil {
		return nil, err
	}

	// Create ctr
	ctr := setupCtr(ctrCipdVersion)

	// Create configs
	metadataContainers := req.GetParams().GetContainerMetadata().GetContainers()
	buildTarget, ok := req.GetParams().GetKeyvals()["build_target"]
	if !ok {
		return nil, fmt.Errorf("Provided keyvals in CrosTestRunnerRequest missing key 'build_target'")
	}
	metadataMap, ok := metadataContainers[buildTarget]
	if !ok {
		return nil, fmt.Errorf("Provided key %q does not exist in provided container metadata.", buildTarget)
	}
	dockerKeyFile, err := common.LocateFile([]string{common.LabDockerKeyFileLocation, common.VmLabDockerKeyFileLocation})
	if err != nil {
		return nil, fmt.Errorf("unable to locate dockerKeyFile during initialization: %w", err)
	}
	containerImagesMap := metadataMap.GetImages()
	executorCfg := configs.NewExecutorConfig(ctr, nil)
	cmdCfg := configs.NewCommandConfig(executorCfg)

	// Create state keeper
	gcsurl := common.GetGcsUrl(gsRoot)
	sk := data.NewHwTestStateKeeper()
	sk.BuildState = buildState
	sk.CrosTestRunnerRequest = req
	sk.CftTestRequest = cft
	sk.Ctr = ctr
	sk.DockerKeyFileLocation = dockerKeyFile
	sk.GcsPublishSrcDir = os.Getenv("TEMPDIR")
	sk.GcsUrl = gcsurl
	sk.TesthausUrl = common.GetTesthausUrl(gcsurl)
	sk.ContainerImages = containerImagesMap

	common.LogWarningIfErr(ctx, sk.Injectables.Set("req", req))
	common.LogWarningIfErr(ctx, sk.Injectables.Set("botDims", buildState.Build().GetInfra().GetSwarming().GetBotDimensions()))
	common.LogWarningIfErr(ctx, sk.Injectables.Set("gcs-url", gcsurl))
	common.LogWarningIfErr(ctx, sk.Injectables.Set("testhaus-url", common.GetTesthausUrl(gcsurl)))

	populateRequestQueues(sk, req)

	// Generate config
	hwTestConfig := configs.NewTrv2ExecutionConfig(configs.HwTestExecutionConfigType, cmdCfg, sk, &api_common.CftStepsConfig{})
	err = hwTestConfig.GenerateConfig(ctx)
	if err != nil {
		return sk.SkylabResult, errors.Annotate(err, "error during generating hw test configs: ").Err()
	}

	// Execute config
	err = hwTestConfig.Execute(ctx)
	// For debugging purposes, logs the final state of the Injectables
	// Storage to the top level of the buildState.
	sk.Injectables.LogStorageToBuild(ctx, buildState)
	if err != nil {
		return sk.SkylabResult, errors.Annotate(err, "error during executing hw test configs: ").Err()
	}
	return sk.SkylabResult, nil
}

// populateRequestQueues parses through the OrderedTasks of a CrosTestRunnerRequest
// to populate corresponding queues of requests.
func populateRequestQueues(sk *data.HwTestStateKeeper, req *skylab_test_runner.CrosTestRunnerRequest) {
	if req != nil {
		for _, taskRequest := range req.OrderedTasks {
			for _, containerRequest := range taskRequest.OrderedContainerRequests {
				sk.ContainerQueue.PushBack(containerRequest)
			}

			switch typedRequest := taskRequest.Task.(type) {
			case *skylab_test_runner.CrosTestRunnerRequest_Task_Provision:
				sk.ProvisionQueue.PushBack(typedRequest.Provision)
			case *skylab_test_runner.CrosTestRunnerRequest_Task_PreTest:
				sk.PreTestQueue.PushBack(typedRequest.PreTest)
			case *skylab_test_runner.CrosTestRunnerRequest_Task_Test:
				sk.TestQueue.PushBack(typedRequest.Test)
			case *skylab_test_runner.CrosTestRunnerRequest_Task_PostTest:
				sk.PostTestQueue.PushBack(typedRequest.PostTest)
			case *skylab_test_runner.CrosTestRunnerRequest_Task_Publish:
				sk.PublishQueue.PushBack(typedRequest.Publish)
			default:
			}
		}
	}
}

// validateHwExecution ensures values are set for HW.
func validateHwExecution(ctrCipdVersion, gsRoot string) error {
	if ctrCipdVersion == "" {
		return fmt.Errorf("Cros-tool-runner cipd version cannot be empty for hw test execution.")
	}
	if gsRoot == "" {
		return fmt.Errorf("GS root cannot be empty for hw test execution.")
	}

	return nil
}

// setupCtr creates an instance for CrosToolRunner.
func setupCtr(ctrCipdVersion string) *crostoolrunner.CrosToolRunner {
	ctrCipdInfo := crostoolrunner.CtrCipdInfo{
		Version:        ctrCipdVersion,
		CtrCipdPackage: common.CtrCipdPackage,
	}

	return &crostoolrunner.CrosToolRunner{
		CtrCipdInfo:       ctrCipdInfo,
		EnvVarsToPreserve: configs.GetHwConfigsEnvVars(),
	}
}
