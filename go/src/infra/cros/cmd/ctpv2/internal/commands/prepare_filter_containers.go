// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"infra/cros/cmd/common_lib/common"
	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/ctpv2/data"

	buildapi "go.chromium.org/chromiumos/config/go/build/api"
	"go.chromium.org/chromiumos/config/go/test/api"
	testapi "go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"
)

// PrepareFilterContainersInfoCmd represents prepare filter containers info cmd.
type PrepareFilterContainersInfoCmd struct {
	*interfaces.AbstractSingleCmdByNoExecutor

	// Deps
	CtpV2Req *testapi.CTPv2Request

	// Updates
	ContainerInfoQueue   *list.List
	ContainerMetadataMap map[string]*buildapi.ContainerImageInfo
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *PrepareFilterContainersInfoCmd) ExtractDependencies(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

	var err error
	switch sk := ski.(type) {
	case *data.FilterStateKeeper:
		err = cmd.extractDepsFromFilterStateKeepr(ctx, sk)

	default:
		return fmt.Errorf("StateKeeper '%T' is not supported by cmd type %s.", sk, cmd.GetCommandType())
	}

	if err != nil {
		return errors.Annotate(err, "error during extracting dependencies for command %s: ", cmd.GetCommandType()).Err()
	}

	return nil
}

// UpdateStateKeeper updates the state keeper with info from the cmd.
func (cmd *PrepareFilterContainersInfoCmd) UpdateStateKeeper(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

	var err error
	switch sk := ski.(type) {
	case *data.FilterStateKeeper:
		err = cmd.updateLocalTestStateKeeper(ctx, sk)
	}

	if err != nil {
		return errors.Annotate(err, "error during updating for command %s: ", cmd.GetCommandType()).Err()
	}

	return nil
}

func (cmd *PrepareFilterContainersInfoCmd) extractDepsFromFilterStateKeepr(
	ctx context.Context,
	sk *data.FilterStateKeeper) error {

	if sk.CtpV2Req == nil {
		return fmt.Errorf("Cmd %q missing dependency: CtpV2Req", cmd.GetCommandType())
	}

	cmd.CtpV2Req = sk.CtpV2Req
	return nil
}

func (cmd *PrepareFilterContainersInfoCmd) updateLocalTestStateKeeper(
	ctx context.Context,
	sk *data.FilterStateKeeper) error {

	if cmd.ContainerInfoQueue != nil {
		sk.ContainerInfoQueue = cmd.ContainerInfoQueue
	}

	if cmd.ContainerMetadataMap != nil {
		sk.ContainerMetadataMap = cmd.ContainerMetadataMap
	}

	return nil
}

// Execute executes the command.
func (cmd *PrepareFilterContainersInfoCmd) Execute(ctx context.Context) error {
	var err error
	step, ctx := build.StartStep(ctx, "Prepare containers for filters")
	defer func() { step.End(err) }()

	// -- Create final container metadata map --

	// TODO (dbeckett): currently there is a bit of a race between getting the metadata to run test-finder, and the boards provided
	// to the request. We don't want to run test-finder per board as that's extremely expensive to setup/act on, however
	// CFT design has test-finder being board specific. For initial MVP we will just use the first board in the request to
	// get the container MD from, but this will need to be solved long term.
	board, gcsPath, err := gcsInfo(cmd.CtpV2Req)
	if err != nil {
		return err
	}
	buildContainerMetadata, err := common.FetchImageData(ctx, board, gcsPath)
	if err != nil {
		return errors.Annotate(err, "failed to fetch container image data: ").Err()
	}

	finalMetadataMap := createContainerImagesInfoMap(ctx, cmd.CtpV2Req, buildContainerMetadata)
	cmd.ContainerMetadataMap = finalMetadataMap

	// Write to log
	mapData, err := json.MarshalIndent(finalMetadataMap, "", "\t")
	if err != nil {
		logging.Infof(
			ctx,
			"error during writing container metadata map to log: %s",
			err.Error())
	}
	step.Log("Final container metadata map").Write(mapData)

	// -- Create ctp filters from default and input filters --

	ctpFilters := make([]*api.CTPFilter, 0)

	karbonFilters, err := common.ConstructCtpFilters(ctx, common.DefaultKarbonFilterNames, finalMetadataMap, cmd.CtpV2Req.GetKarbonFilters())
	if err != nil {
		return errors.Annotate(err, "failed to create karbon filters: ").Err()
	}
	ctpFilters = append(ctpFilters, karbonFilters...)

	koffeeFilters, err := common.ConstructCtpFilters(ctx, common.DefaultKoffeeFilterNames, finalMetadataMap, cmd.CtpV2Req.GetKoffeeFilters())
	if err != nil {
		return errors.Annotate(err, "failed to create koffee filters: ").Err()
	}
	ctpFilters = append(ctpFilters, koffeeFilters...)

	filterData, err := json.MarshalIndent(ctpFilters, "", "\t")
	if err != nil {
		logging.Infof(
			ctx,
			"error during writing ctp filters to log: %s",
			err.Error())
	}
	step.Log("Final Ctp filters list").Write(filterData)

	// -- Create container info queue --

	containerInfoList := list.New()

	for _, filter := range ctpFilters {
		containerInfoList.PushBack(CtpFilterToContainerInfo(filter))
	}
	step.Log("Container Info queue").Write(common.ListToJson(containerInfoList))

	cmd.ContainerInfoQueue = containerInfoList

	return nil
}

func NewPrepareFilterContainersInfoCmd() *PrepareFilterContainersInfoCmd {
	abstractCmd := interfaces.NewAbstractCmd(PrepareFilterContainersCmdType)
	abstractSingleCmdByNoExecutor := &interfaces.AbstractSingleCmdByNoExecutor{AbstractCmd: abstractCmd}
	return &PrepareFilterContainersInfoCmd{AbstractSingleCmdByNoExecutor: abstractSingleCmdByNoExecutor}
}

func gcsInfo(req *testapi.CTPv2Request) (string, string, error) {
	board := getFirstBoardFromLegacy(req.Targets)
	if board == "" {
		return "", "", errors.New("no board provided in legacy request")
	}

	gcsPath := getFirstGcsPathFromLegacy(req.Targets)
	if gcsPath == "" {
		return "", "", errors.New("no gcsPath provided in legacy request")
	}

	return board, gcsPath, nil
}

func getFirstBoardFromLegacy(targs []*testapi.Targets) string {
	if len(targs) == 0 {
		return ""
	}
	switch hw := targs[0].HwTarget.Target.(type) {
	case *testapi.HWTarget_LegacyHw:
		return hw.LegacyHw.Board
	default:
		return ""
	}
}

func getFirstGcsPathFromLegacy(targs []*testapi.Targets) string {
	if len(targs) == 0 {
		return ""
	}
	if len(targs[0].SwTargets) == 0 {
		return ""
	}
	switch sw := targs[0].SwTargets[0].SwTarget.(type) {
	case *testapi.SWTarget_LegacySw:
		return sw.LegacySw.GcsPath
	default:
		return ""
	}
}

func createContainerImagesInfoMap(ctx context.Context, req *testapi.CTPv2Request, buildContMetadata map[string]*buildapi.ContainerImageInfo) map[string]*buildapi.ContainerImageInfo {
	// In case of any overlap of container metadata between input and build metadata,
	// the input metadata will be prioritized.
	for _, filter := range req.GetKarbonFilters() {
		buildContMetadata[filter.GetContainer().GetName()] = filter.GetContainer()
	}

	for _, filter := range req.GetKoffeeFilters() {
		buildContMetadata[filter.GetContainer().GetName()] = filter.GetContainer()
	}

	return buildContMetadata
}

// CtpFilterToContainerInfo creates container info from provided ctp filter.
func CtpFilterToContainerInfo(ctpFilter *api.CTPFilter) *data.ContainerInfo {
	return &data.ContainerInfo{
		ImageKey:  ctpFilter.GetContainer().GetName(),
		Request:   common.CreateContainerRequest(ctpFilter),
		ImageInfo: ctpFilter.GetContainer(),
	}
}
