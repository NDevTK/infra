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

	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner/steps"
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
			skylabResult, err := executeHwTests(ctx, input.CftTestRequest, ctrCipdInfo.GetVersion().GetCipdLabel(), input.GetConfig().GetOutput().GetLogDataGsRoot(), st)
			if skylabResult != nil {
				m, _ := proto.Marshal(skylabResult)
				var b bytes.Buffer
				w := zlib.NewWriter(&b)
				_, _ = w.Write(m)
				_ = w.Close()
				resp.CompressedResult = base64.StdEncoding.EncodeToString(b.Bytes())
			}
			if err != nil {
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
	if ctrCipdVersion == "" {
		return nil, fmt.Errorf("Cros-tool-runner cipd version cannot be empty for hw test execution.")
	}
	if gsRoot == "" {
		return nil, fmt.Errorf("GS root cannot be empty for hw test execution.")
	}

	// Create ctr
	ctrCipdInfo := crostoolrunner.CtrCipdInfo{
		Version:        ctrCipdVersion,
		CtrCipdPackage: common.CtrCipdPackage,
	}

	ctr := &crostoolrunner.CrosToolRunner{
		CtrCipdInfo:       ctrCipdInfo,
		EnvVarsToPreserve: configs.GetHwConfigsEnvVars(),
	}

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

	companionBoards := []string{}
	for _, companion := range sk.CftTestRequest.CompanionDuts {
		companionBoards = append(companionBoards, companion.GetDutModel().GetBuildTarget())
	}
	sk.CftTestRequest.AutotestKeyvals["companion-boards"] = strings.Join(companionBoards, ",")

	// For demonstration/logging purposes.
	_ = sk.Injectables.Set("req", req)
	_ = sk.Injectables.Set("botDims", buildState.Build().GetInfra().GetSwarming().GetBotDimensions())
	_ = sk.Injectables.Set("gcs-url", gcsurl)
	_ = sk.Injectables.Set("testhaus-url", common.GetTesthausUrl(gcsurl))

	// Generate config
	hwTestConfig := configs.NewTrv2ExecutionConfig(configs.HwTestExecutionConfigType, cmdCfg, sk, req.GetStepsConfig())
	err = hwTestConfig.GenerateConfig(ctx)
	if err != nil {
		return sk.SkylabResult, errors.Annotate(err, "error during generating hw test configs: ").Err()
	}

	// Execute config
	err = hwTestConfig.Execute(ctx)
	sk.Injectables.LogStorageToBuild(ctx, buildState)
	if err != nil {
		return sk.SkylabResult, errors.Annotate(err, "error during executing hw test configs: ").Err()
	}
	return sk.SkylabResult, nil
}
