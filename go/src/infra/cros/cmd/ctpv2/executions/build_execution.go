// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package executions

import (
	"container/list"
	"context"
	"fmt"
	"log"

	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/steps"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"

	"infra/cros/cmd/common_lib/common"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
	"infra/cros/cmd/cros_test_runner/protos"
	"infra/cros/cmd/ctpv2/data"
	"infra/cros/cmd/ctpv2/internal/configs"

	"go.chromium.org/luci/common/errors"
)

// TODO : Re-structure different execution flow properly later.
// LuciBuildExecution represents build executions.
func LuciBuildExecution() {
	// Set input property reader functions
	var ctrCipdInfoReader func(context.Context) *protos.CipdVersionInfo
	build.MakePropertyReader(common.HwTestCtrInputPropertyName, &ctrCipdInfoReader)
	input := &steps.CTPv2BinaryBuildInput{}

	// Set output props writer functions
	// TODO: add the fields to the response that is responsible for
	// feeding the test results to upstream.
	var writeOutputProps func(*steps.CTPv2BinaryBuildOutput)
	var mergeOutputProps func(*steps.CTPv2BinaryBuildOutput)

	build.Main(input, &writeOutputProps, &mergeOutputProps,
		func(ctx context.Context, args []string, st *build.State) error {
			log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmsgprefix)
			logging.Infof(ctx, "have input %v", input)
			ctrCipdInfo := ctrCipdInfoReader(ctx)
			logging.Infof(ctx, "have ctr info: %v", ctrCipdInfo)
			logging.Infof(ctx, "ctr label: %s", ctrCipdInfo.GetVersion().GetCipdLabel())
			resp := &steps.CTPv2BinaryBuildOutput{}
			_, err := executeFiltersInLuciBuild(ctx, input.Ctpv2Request, ctrCipdInfo.GetVersion().GetCipdLabel(), st)
			// TODO (azrahman): add compressed result for upstream
			// if resp != nil {
			// 	m, _ := proto.Marshal(resp)
			// 	var b bytes.Buffer
			// 	w := zlib.NewWriter(&b)
			// 	_, _ = w.Write(m)
			// 	_ = w.Close()
			//
			// 	resp.CompressedResult = base64.StdEncoding.EncodeToString(b.Bytes())
			// }
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

func executeFiltersInLuciBuild(
	ctx context.Context,
	req *api.CTPv2Request,
	ctrCipdVersion string,
	buildState *build.State) (*api.CTPv2Response, error) {
	// Validation
	if ctrCipdVersion == "" {
		return nil, fmt.Errorf("Cros-tool-runner cipd version cannot be empty for hw test execution.")
	}

	// Create ctr
	ctrCipdInfo := crostoolrunner.CtrCipdInfo{
		Version:        ctrCipdVersion,
		CtrCipdPackage: common.CtrCipdPackage,
	}

	ctr := &crostoolrunner.CrosToolRunner{
		CtrCipdInfo:       ctrCipdInfo,
		EnvVarsToPreserve: []string{},
	}

	dockerKeyFile, err := common.LocateFile([]string{common.LabDockerKeyFileLocation, common.VmLabDockerKeyFileLocation})
	if err != nil {
		return nil, fmt.Errorf("unable to locate dockerKeyFile during initialization: %w", err)
	}

	executorCfg := configs.NewExecutorConfig(ctr, nil)
	cmdCfg := configs.NewCommandConfig(executorCfg)

	sk := &data.FilterStateKeeper{
		CtpV2Req:              req,
		DockerKeyFileLocation: dockerKeyFile,
		Ctr:                   ctr,
		ContainerInfoQueue:    list.New(),
	}

	// Generate config
	ctpv2Config := configs.NewCtpv2ExecutionConfig(getTotalFilters(req, common.DefaultKarbonFilterNames, common.DefaultKoffeeFilterNames), configs.LuciBuildFilterExecutionConfigType, cmdCfg, sk)
	err = ctpv2Config.GenerateConfig(ctx)
	if err != nil {
		return nil, errors.Annotate(err, "error during generating filter configs: ").Err()
	}

	// Execute config
	err = ctpv2Config.Execute(ctx)
	if err != nil {
		return sk.CtpV2Response, errors.Annotate(err, "error during executing hw test configs: ").Err()
	}

	return sk.CtpV2Response, nil
}

func getTotalFilters(req *api.CTPv2Request, defaultKarbonFilterNames []string, defaultKoffeeFilterNames []string) int {
	filterSet := map[string]bool{}

	for _, filterName := range defaultKarbonFilterNames {
		filterSet[filterName] = true
	}

	for _, filterName := range defaultKoffeeFilterNames {
		filterSet[filterName] = true
	}

	for _, filter := range req.GetKarbonFilters() {
		filterSet[filter.GetContainer().GetName()] = true
	}

	for _, filter := range req.GetKoffeeFilters() {
		filterSet[filter.GetContainer().GetName()] = true
	}

	return len(filterSet)
}
