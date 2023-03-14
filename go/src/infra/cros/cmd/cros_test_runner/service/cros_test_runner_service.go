// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package service

import (
	"context"
	"infra/cros/cmd/cros_test_runner/internal/data"

	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
)

type CrosTestRunnerService struct {
	ServerStartRequest *skylab_test_runner.CrosTestRunnerServerStartRequest
	req                *skylab_test_runner.ExecuteRequest
}

func NewCrosTestRunnerService(execReq *skylab_test_runner.ExecuteRequest, serverSK *data.LocalTestStateKeeper) (*CrosTestRunnerService, error) {

	// TODO: Construct new state keeper (using provided SK) and configs

	return &CrosTestRunnerService{
		req: execReq,
	}, nil
}

func (crs *CrosTestRunnerService) Execute(ctx context.Context) (*skylab_test_runner.ExecuteResponse, error) {
	// TODO: invoke local test execution flow

	return nil, nil
}
