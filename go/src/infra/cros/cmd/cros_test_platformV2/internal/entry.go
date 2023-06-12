// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package internal

import (
	"context"

	managers "infra/cros/cmd/cros_test_platformV2/docker_managers"
	"infra/cros/cmd/cros_test_platformV2/executor"

	parsers "infra/cros/cmd/cros_test_platformV2/tools"

	"fmt"

	"go.chromium.org/chromiumos/config/go/test/api"
)

func Execute(inputPath string) (*api.TestSuite, error) {
	request, err := parsers.ReadInput(inputPath)
	if err != nil {
		return nil, fmt.Errorf("unable to parse request: %s", err)
	}
	resp := &api.TestSuite{}

	ctx := context.Background()

	// Build Exectors. Right now do CTRv2 + Both filters.
	// Note: the filter impls are currently nil.
	executors, _ := buildExecutors(ctx, request, resp)

	// Run the same commands for each
	for _, executor := range executors {
		// For CTR, init = start Server async. For services it will be pull/prep container/launch
		err := executor.Execute(ctx, "init")
		// Gcloud auth for CTR (kinda odd...). For services, it will be `call the service.`
		err = executor.Execute(ctx, "run")
		if err != nil {
			fmt.Println("Error")
		}
	}

	// After all execs are run, stop them all. TODO, this probably needs to be a bit smarter defered.
	for _, executor := range executors {
		err := executor.Execute(ctx, "stop")
		if err != nil {
			fmt.Println("Error")
		}
	}
	return nil, nil

}

func createContainerManagerExecutor(ctx context.Context, kind string) (managers.ContainerManager, executor.Executor) {
	var containerMgr managers.ContainerManager
	var e executor.Executor
	if kind == "ctr" {
		containerMgr = managers.NewCTRDummy()
		e = executor.NewCtrExecutor(containerMgr)
	} else {
		containerMgr = managers.NewCloudDummy()
		e = executor.NewCloudContainerExecutor(containerMgr)
	}

	return containerMgr, e
}

func buildExecutors(ctx context.Context, req *api.CTPRequest2, resp *api.TestSuite) ([]executor.Executor, error) {
	execs := []executor.Executor{}

	// First must always be the container Manager. All further executors will require the manager.
	contMang, contExec := createContainerManagerExecutor(ctx, "ctr")
	execs = append(execs, contExec)
	// Actual Filters
	// TODO, change this to a loop. The request contains a list of filters, so check that, and build the Executor from that.
	// the executors should not need the full resp, but rather info from the resp, like name, build.
	execs = append(execs, executor.NewKarbonExecutor(contMang, resp))
	execs = append(execs, executor.NewKoffeeExecutor(contMang, resp))

	//Can expand to more as needed.
	return execs, nil
}
