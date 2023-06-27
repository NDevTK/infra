// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package internal

import (
	"context"
	"fmt"
	managers "infra/cros/cmd/cros_test_platformV2/docker_managers"
	"infra/cros/cmd/cros_test_platformV2/executor"
	"infra/cros/cmd/cros_test_platformV2/tools/gcs"

	"go.chromium.org/chromiumos/config/go/test/api"
)

func translateRequest(req *api.CTPRequest2) *executor.TestPlanResponse {
	internalStruct := &executor.TestPlanResponse{}
	hwReqs := hwRequirements(req)
	internalStruct.SuiteMetadata = &executor.SuiteMetadata{
		HWRequirements: hwReqs,
		Builds:         []string{req.Build},
		Pool:           "none",
	}

	return internalStruct
}

func hwRequirements(req *api.CTPRequest2) (hwReqs []*executor.HWRequirement) {
	switch hw := req.Variant.(type) {
	case *api.CTPRequest2_Board:
		hwReq := &executor.HWRequirement{
			Board: hw.Board,
		}
		hwReqs = append(hwReqs, hwReq)
		return hwReqs
	default:
		return hwReqs
	}
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

func buildExecutors(ctx context.Context, req *api.CTPRequest2, cloud bool) ([]executor.Executor, error) {
	execs := []executor.Executor{}

	// TODO finalize this mess.
	containerMetadata, err := gcs.FetchImageData(ctx, req.GetBoard(), "release", "R111-1111-1")

	// First must always be the container Manager. All further executors will require the manager.
	contMang, contExec := createContainerManagerExecutor(ctx, cloud)
	execs = append(execs, contExec)

	defaultKarbon, defaultKoffe, err := executor.DefaultFilters(req, containerMetadata)
	if err != nil {
		return nil, fmt.Errorf("Unable to create default filters: %s", err)
	}

	for _, filter := range defaultKarbon {
		exec, err := executor.NewKarbonExecutor(contMang, filter, containerMetadata)
		if err != nil {
			return nil, err
		}
		execs = append(execs, exec)
	}
	for _, filter := range req.KarbonFilters {
		exec, err := executor.NewKarbonExecutor(contMang, filter, containerMetadata)
		if err != nil {
			return nil, err
		}
		execs = append(execs, exec)
	}

	for _, filter := range defaultKoffe {
		exec, err := executor.NewKoffeeExecutor(contMang, filter, containerMetadata)
		if err != nil {
			return nil, err
		}
		execs = append(execs, exec)
	}
	for _, filter := range req.KoffeeFilters {
		exec, err := executor.NewKoffeeExecutor(contMang, filter, containerMetadata)
		if err != nil {
			return nil, err
		}
		execs = append(execs, exec)
	}
	// Can expand to more as needed.
	return execs, nil
}
