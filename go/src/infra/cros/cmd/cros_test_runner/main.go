// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"log"

	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner/steps"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"

	"infra/cros/cmd/cros_test_runner/internal/configs"
	"infra/cros/cmd/cros_test_runner/internal/data"
)

func main() {
	input := &steps.RunTestsRequest{}
	var writeOutputProps func(*steps.RunTestsResponse)
	var mergeOutputProps func(*steps.RunTestsResponse)

	build.Main(input, &writeOutputProps, &mergeOutputProps,
		func(ctx context.Context, args []string, st *build.State) error {
			log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmsgprefix)
			logging.Infof(ctx, "have input %v", input)
			err := executeHwTests(ctx, input.CftTestRequest)
			if err != nil {
				logging.Infof(ctx, "error found: %s", err)
			}

			// TODO (azrahman): write output properties
			return err
		},
	)
}

// executeHwTests executes hw tests
func executeHwTests(ctx context.Context, req *skylab_test_runner.CFTTestRequest) error {
	// Create configs
	executorCfg := configs.NewExecutorConfig()
	cmdCfg := configs.NewCommandConfig(executorCfg)

	// Create state keeper
	sk := &data.HwTestStateKeeper{CftTestRequest: req}

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
