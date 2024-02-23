// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package executions

import (
	"container/list"
	"context"
	"fmt"
	"log"
	"sync"

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
			_, err := executeRequests(ctx, input, ctrCipdInfo.GetVersion().GetCipdLabel(), st)
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

func executeRequests(
	ctx context.Context,
	input *steps.CTPv2BinaryBuildInput,
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

	sk := &data.PrePostFilterStateKeeper{
		DockerKeyFileLocation: dockerKeyFile,
		Ctr:                   ctr,
		CtpV1Requests:         input.GetRequests(),
		CtpV2Request:          input.GetCtpv2Request(),
	}

	ctpv2PreConfig := configs.NewCtpv2ExecutionConfig(0, configs.Ctpv2PreExecutionConfigType, cmdCfg, sk)
	err = ctpv2PreConfig.GenerateConfig(ctx)
	if err != nil {
		return nil, errors.Annotate(err, "error during generating pre execution configs: ").Err()
	}

	ctpv2PostConfig := configs.NewCtpv2ExecutionConfig(0, configs.Ctpv2PostExecutionConfigType, cmdCfg, sk)
	err = ctpv2PostConfig.GenerateConfig(ctx)
	if err != nil {
		return nil, errors.Annotate(err, "error during generating post execution configs: ").Err()
	}

	// Execute pre configs
	err = ctpv2PreConfig.Execute(ctx)
	if err != nil {
		return &api.CTPv2Response{}, errors.Annotate(err, "error during executing pre execution configs: ").Err()
	}

	// Execute Ctpv2 Reqs
	executeCtpv2Reqs(ctx, sk.CtpV2Request, buildState, ctr)

	// Execute post configs
	err = ctpv2PostConfig.Execute(ctx)
	if err != nil {
		return &api.CTPv2Response{}, errors.Annotate(err, "error during executing post execution configs: ").Err()
	}

	return &api.CTPv2Response{}, nil
}

func executeCtpv2Reqs(ctx context.Context,
	ctpv2Req *api.CTPv2Request, buildState *build.State, ctr *crostoolrunner.CrosToolRunner) {
	var err error
	step, ctx := build.StartStep(ctx, "Suite Executions (async)")
	defer func() { step.End(err) }()

	wg := &sync.WaitGroup{}
	contInfoMap := data.NewContainerInfoMap()
	for _, ctpReq := range ctpv2Req.GetRequests() {
		wg.Add(1)
		// TODO (azrahman): handle errors through channels
		go executeFiltersInLuciBuild(ctx, ctpReq, buildState, wg, ctr, contInfoMap)
	}
	wg.Wait()
}

func executeFiltersInLuciBuild(
	ctx context.Context,
	req *api.CTPRequest,
	buildState *build.State,
	wg *sync.WaitGroup, ctr *crostoolrunner.CrosToolRunner, contInfoMap *data.ContainerInfoMap) error {
	defer wg.Done()
	var err error
	step, ctx := build.StartStep(ctx, req.GetSuiteRequest().GetTestSuite().GetName())
	defer func() { step.End(err) }()

	executorCfg := configs.NewExecutorConfig(ctr, nil)
	cmdCfg := configs.NewCommandConfig(executorCfg)

	sk := &data.FilterStateKeeper{
		CtpReq:             req,
		Ctr:                ctr,
		ContainerInfoQueue: list.New(),
		BuildState:         buildState,
		Scheduler:          req.GetSchedulerInfo().GetScheduler(),
		ContainerInfoMap:   contInfoMap,
	}

	nFilters := getTotalFilters(ctx, req, common.MakeDefaultFilters(ctx, req.GetSuiteRequest()), common.DefaultKoffeeFilterNames)
	logging.Infof(ctx, "nfilters: %s", nFilters)
	// Generate config
	ctpv2Config := configs.NewCtpv2ExecutionConfig(nFilters, configs.LuciBuildFilterExecutionConfigType, cmdCfg, sk)
	err = ctpv2Config.GenerateConfig(ctx)
	if err != nil {
		return errors.Annotate(err, "error during generating filter configs: ").Err()
	}

	// Execute config
	err = ctpv2Config.Execute(ctx)
	if err != nil {
		return errors.Annotate(err, "error during executing hw test configs: ").Err()
	}

	return nil
}

func getTotalFilters(ctx context.Context, req *api.CTPRequest, defaultKarbonFilterNames []string, defaultKoffeeFilterNames []string) int {
	filterSet := map[string]bool{}
	logging.Infof(ctx, "n defaultKarbonFilterNames: %s", len(defaultKarbonFilterNames))
	logging.Infof(ctx, "Given Karbon: %s And Koffee: %s", defaultKarbonFilterNames, defaultKoffeeFilterNames)

	for _, filterName := range defaultKarbonFilterNames {
		filterSet[filterName] = true
	}

	for _, filterName := range defaultKoffeeFilterNames {
		filterSet[filterName] = true
	}

	for _, filter := range req.GetKarbonFilters() {
		filterSet[filter.GetContainerInfo().GetContainer().GetName()] = true
	}

	for _, filter := range req.GetKoffeeFilters() {
		filterSet[filter.GetContainerInfo().GetContainer().GetName()] = true
	}

	return len(filterSet)
}
