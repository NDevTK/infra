// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package service

import (
	"context"
	"infra/cros/cmd/cros_test_runner/executions"
	"infra/cros/cmd/cros_test_runner/internal/data"

	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
)

type CrosTestRunnerService struct {
	ServerStartRequest *skylab_test_runner.CrosTestRunnerServerStartRequest
	req                *skylab_test_runner.ExecuteRequest
	sk                 *data.LocalTestStateKeeper
}

func NewCrosTestRunnerService(execReq *skylab_test_runner.ExecuteRequest, serverSK *data.LocalTestStateKeeper) (*CrosTestRunnerService, error) {
	executeSK := &data.LocalTestStateKeeper{Args: &data.LocalArgs{
		HostName:             serverSK.HostName,
		SkipBuildDutTopology: serverSK.DutTopology != nil,
	}}
	executeSK.DutTopology = serverSK.DutTopology
	executeSK.DockerKeyFileLocation = serverSK.DockerKeyFileLocation
	executeSK.GcsUrl = serverSK.GcsUrl
	executeSK.StainlessUrl = serverSK.StainlessUrl
	executeSK.TesthausUrl = serverSK.TesthausUrl
	executeSK.GcsPublishSrcDir = serverSK.GcsPublishSrcDir

	executeSK.Args.HostName = serverSK.HostName
	duts := serverSK.DutTopology.GetDuts()
	if serverSK.DutTopology != nil && len(duts) > 0 {
		executeSK.Args.SkipCacheServer = duts[0].GetCacheServer() != nil
		executeSK.Args.SkipSshReverseTunnel = duts[0].GetCacheServer() != nil
		executeSK.Args.SkipSshTunnel = duts[0].GetChromeos().Ssh != nil
	}
	cftTestRequest := execReq.GetCftTestRequest()
	if cftTestRequest != nil {
		stepConfig := cftTestRequest.GetStepsConfig().GetHwTestConfig()
		if stepConfig != nil {
			executeSK.Args.SkipBuildDutTopology = serverSK.DutTopology != nil || stepConfig.GetSkipLoadingDutTopology()
			executeSK.Args.SkipDutServer = stepConfig.GetSkipStartingDutService()
			executeSK.Args.SkipProvision = stepConfig.GetSkipProvision()
			// TODO: Support test finder for server execution.
			// Skipped for initial implementation.
			executeSK.Args.SkipTestFinder = true
			executeSK.Args.SkipTest = stepConfig.GetSkipTestExecution()
		}
	}

	return &CrosTestRunnerService{
		req: execReq,
		sk:  executeSK,
	}, nil
}

func (crs *CrosTestRunnerService) Execute(ctx context.Context, logPath string) (*skylab_test_runner.ExecuteResponse, error) {
	crs.sk.CftTestRequest = crs.req.GetCftTestRequest()

	testPlan := crs.req.GetTestPlan()
	if testPlan != nil {
		crs.sk.TestArgs = testPlan.TestArgs
	}

	return executions.LocalExecution(crs.sk, crs.req.CtrCipdVersion, logPath)
}
