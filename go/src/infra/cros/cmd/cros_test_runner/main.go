// Copyright 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner/steps"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"

	"infra/cros/cmd/cros_test_runner/common"
	"infra/cros/cmd/cros_test_runner/internal/configs"
	"infra/cros/cmd/cros_test_runner/internal/data"
	"infra/cros/cmd/cros_test_runner/internal/tools/crostoolrunner"
	"infra/cros/cmd/cros_test_runner/protos"
)

func main() {
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
			// TODO (azrahman): After stablizing in prod, move log data gs root to cft/new proto.
			err := executeHwTests(ctx, input.CftTestRequest, ctrCipdInfo.GetVersion().GetCipdLabel(), input.GetConfig().GetOutput().GetLogDataGsRoot(), st)
			if err != nil {
				logging.Infof(ctx, "error found: %s", err)
			}

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
	buildState *build.State) error {

	// Validation
	if ctrCipdVersion == "" {
		return fmt.Errorf("Cros-tool-runner cipd version cannot be empty for hw test execution.")
	}
	if gsRoot == "" {
		return fmt.Errorf("GS root cannot be empty for hw test execution.")
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
		return fmt.Errorf("Provided key %q does not exist in provided container metadata.", metadataKey)
	}
	containerImagesMap := metadataMap.GetImages()
	containerCfg := configs.NewCftContainerConfig(ctr, containerImagesMap)
	executorCfg := configs.NewExecutorConfig(ctr, containerCfg)
	cmdCfg := configs.NewCommandConfig(executorCfg)

	// Create state keeper
	gcsurl := common.GetGcsUrl(gsRoot)
	sk := &data.HwTestStateKeeper{
		BuildState:            buildState,
		CftTestRequest:        req,
		Ctr:                   ctr,
		DockerKeyFileLocation: common.LabDockerKeyFileLocation,
		GcsPublishSrcDir:      os.Getenv("TEMPDIR"),
		GcsUrl:                gcsurl,
		StainlessUrl:          common.GetStainlessUrl(gcsurl),
		TesthausUrl:           common.GetTesthausUrl(gcsurl),
	}

	// Generate config
	hwTestConfig := configs.NewTestExecutionConfig(configs.HwTestExecutionConfigType, cmdCfg, sk)
	err := hwTestConfig.GenerateConfig(ctx)
	if err != nil {
		return errors.Annotate(err, "error during generating hw test configs: ").Err()
	}

	// Execute config
	err = hwTestConfig.Execute(ctx)
	if err != nil {
		return errors.Annotate(err, "error during executing hw test configs: ").Err()
	}
	return nil
}
