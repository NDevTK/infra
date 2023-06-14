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

func Execute(inputPath string, cloud bool) (*api.TestSuite, error) {
	request, err := parsers.ReadInput(inputPath)
	if err != nil {
		fmt.Printf("Unable to parse: %s", err)
		// return nil, fmt.Errorf("unable to parse request: %s", err)
	}
	resp := &api.TestSuite{}

	ctx := context.Background()

	// Build Exectors. Right now do CTRv2 + Both filters.
	// Note: the filter impls are currently nil.
	executors, _ := buildExecutors(ctx, request, resp, cloud)

	// Run the same commands for each
	for _, executor := range executors {
		// For CTR, init = start Server async. For services it will be pull/prep container/launch
		err := executor.Execute(ctx, "init")
		fmt.Println("Executing Executor Init.")
		if err != nil {
			fmt.Println("Error")
		}
	}
	// Run the same commands for each
	for _, executor := range executors {
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

func createContainerManagerExecutor(ctx context.Context, cloud bool) (managers.ContainerManager, executor.Executor) {
	var containerMgr managers.ContainerManager
	var e executor.Executor
	if cloud {
		containerMgr = managers.NewCloudDummy()
		e = executor.NewCloudContainerExecutor(containerMgr)
	} else {
		containerMgr = managers.NewCtrManager()
		e = executor.NewCtrExecutor(containerMgr)
	}

	return containerMgr, e
}

func buildExecutors(ctx context.Context, req *api.CTPRequest2, resp *api.TestSuite, cloud bool) ([]executor.Executor, error) {
	execs := []executor.Executor{}

	// First must always be the container Manager. All further executors will require the manager.
	contMang, contExec := createContainerManagerExecutor(ctx, cloud)
	execs = append(execs, contExec)

	for _, filter := range req.KarbonFilters {
		execs = append(execs, executor.NewKarbonExecutor(contMang, resp, filter))

	}
	for _, filter := range req.KoffeeFilters {
		execs = append(execs, executor.NewKoffeeExecutor(contMang, resp, filter))
	}

	//Can expand to more as needed.
	return execs, nil
}
