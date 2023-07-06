// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package internal

import (
	"context"
	"errors"
	"fmt"
	managers "infra/cros/cmd/cros_test_platformV2/docker_managers"
	"infra/cros/cmd/cros_test_platformV2/executor"
	"infra/cros/cmd/cros_test_platformV2/tools/gcs"

	"go.chromium.org/chromiumos/config/go/test/api"
)

func translateRequest(req *api.CTPv2Request) *api.InternalTestplan {
	internalStruct := &api.InternalTestplan{}
	hwReqs := hwRequirements(req)
	// TODO, there is a bitch of a gap between the input request (where build is part of legacySW)
	// And the current definition of channels (which is basically dev/beta/stable) in the suiteMetadata.
	// It might not even be needed as part of the metadata. Need to investigate more.
	internalStruct.SuiteMetadata = &api.SuiteMetadata{
		HwRequirements: hwReqs,
		Pool:           req.Pool,
	}

	return internalStruct
}

func hwRequirements(req *api.CTPv2Request) (hwReqs []*api.HWRequirements) {
	switch hw := req.HwTargets.Targets.(type) {
	case *api.HWTargets_LegacyHw:
		fmt.Println(hw)
		hwReq := &api.HWRequirements{
			HwDefinition: []*api.SwarmingDefinition{},
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

func getFirstBoardFromLegacy(req *api.HWTargets) string {
	switch hw := req.Targets.(type) {
	case *api.HWTargets_LegacyHw:
		if len(hw.LegacyHw.Board) == 0 {
			return ""
		}
		return hw.LegacyHw.Board[0]
	default:
		return ""
	}
}
func buildExecutors(ctx context.Context, req *api.CTPv2Request, cloud bool) ([]executor.Executor, error) {
	execs := []executor.Executor{}

	// TODO currently there is a bit of a race between getting the metadata to run test-finder, and the boards provided
	// to the request. We don't want to run test-finder per board as that's extremely expensive to setup/act on, however
	// CFT design has test-finder being board specific. For initial MVP we will just use the first board in the request to
	// get the container MD from, but this will need to be solved long term.
	board := getFirstBoardFromLegacy(req.HwTargets)
	if board == "" {
		return nil, errors.New("no board provided in legacy request")
	}
	// TODO, hardcodes....
	containerMetadata, err := gcs.FetchImageData(ctx, board, "release", "R111-1111-1")
	if err != nil {
		return nil, fmt.Errorf("unable to fetch image data: %s", err)
	}

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
