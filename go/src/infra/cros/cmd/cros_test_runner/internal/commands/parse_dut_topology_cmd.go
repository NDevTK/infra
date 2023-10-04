// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"
	"infra/cros/cmd/common_lib/common"
	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/cros_test_runner/data"
	"strings"

	"go.chromium.org/chromiumos/config/go/test/api"
	testapi "go.chromium.org/chromiumos/config/go/test/api"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"
	"golang.org/x/exp/slices"
)

// ParseDutTopologyCmd represents build input validation command.
type ParseDutTopologyCmd struct {
	*interfaces.AbstractSingleCmdByNoExecutor

	// Deps
	DutTopology     *labapi.DutTopology
	PrimaryBoard    string
	CompanionBoards []string

	// Updates
	Devices           map[string]*testapi.CrosTestRequest_Device
	DevicesMetadata   map[string]*skylab_test_runner.CFTTestRequest_Device
	DeviceIdentifiers []string
}

type DeviceInfo struct {
	Device   *testapi.CrosTestRequest_Device
	Metadata *skylab_test_runner.CFTTestRequest_Device
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *ParseDutTopologyCmd) ExtractDependencies(ctx context.Context, ski interfaces.StateKeeperInterface) error {
	var err error
	switch sk := ski.(type) {
	case *data.HwTestStateKeeper:
		err = cmd.extractDepsFromHwTestStateKeeper(ctx, sk)
	case *data.LocalTestStateKeeper:
		err = cmd.extractDepsFromHwTestStateKeeper(ctx, &sk.HwTestStateKeeper)

	default:
		return fmt.Errorf("StateKeeper '%T' is not supported by cmd type %s.", sk, cmd.GetCommandType())
	}

	if err != nil {
		return errors.Annotate(err, "error during extracting dependencies for command %s: ", cmd.GetCommandType()).Err()
	}

	return nil
}

// UpdateStateKeeper updates the state keeper with info from the cmd.
func (cmd *ParseDutTopologyCmd) UpdateStateKeeper(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

	var err error
	switch sk := ski.(type) {
	case *data.HwTestStateKeeper:
		err = cmd.updateHwTestStateKeeper(ctx, sk)
	case *data.LocalTestStateKeeper:
		err = cmd.updateHwTestStateKeeper(ctx, &sk.HwTestStateKeeper)

	default:
		return fmt.Errorf("StateKeeper '%T' is not supported by cmd type %s.", sk, cmd.GetCommandType())
	}

	if err != nil {
		return errors.Annotate(err, "error during updating for command %s: ", cmd.GetCommandType()).Err()
	}

	return nil
}

// Execute executes the command.
func (cmd *ParseDutTopologyCmd) Execute(ctx context.Context) error {
	var err error
	step, ctx := build.StartStep(ctx, "Parse DutTopology")
	defer func() { step.End(err) }()

	common.WriteProtoToStepLog(ctx, step, cmd.DutTopology, "DutTopology")
	cmd.Devices = make(map[string]*testapi.CrosTestRequest_Device)
	cmd.DevicesMetadata = make(map[string]*skylab_test_runner.CFTTestRequest_Device)
	cmd.DeviceIdentifiers = []string{}

	devicePool := []*DeviceInfo{}
	for _, dut := range cmd.DutTopology.GetDuts() {
		device, deviceMetadata := parseDut(dut)
		devicePool = append(devicePool, &DeviceInfo{
			Device:   device,
			Metadata: deviceMetadata,
		})
	}

	// Match primary board to dut.
	if cmd.PrimaryBoard != "" {
		info, err := cmd.matchDut(devicePool, cmd.PrimaryBoard)
		if err != nil {
			return fmt.Errorf("Failed to match primaryDevice, %s", err)
		}
		cmd.appendDevice("primaryDevice", info)
	}

	for _, companionBoard := range cmd.CompanionBoards {
		info, err := cmd.matchDut(devicePool, companionBoard)
		if err != nil {
			return fmt.Errorf("Failed to match companionDevice, %s", err)
		}
		deviceId := "companionDevice_" + info.Metadata.GetDutModel().GetBuildTarget()
		if _, ok := cmd.Devices[deviceId]; ok {
			// deviceId already exists, try postfixing
			// Standard within swarming when there are duplicate boards
			// is to postfix with `_2`. (e.g. `brya | brya_2`)
			postfix := 2
			for {
				if _, ok := cmd.Devices[fmt.Sprintf("%s_%d", deviceId, postfix)]; !ok {
					deviceId = fmt.Sprintf("%s_%d", deviceId, postfix)
					break
				}
				postfix += 1
			}
		}
		cmd.appendDevice(deviceId, info)
	}

	return nil
}

// appendDevice handles storing deviceInfo within top-level stores.
func (cmd *ParseDutTopologyCmd) appendDevice(deviceId string, deviceInfo *DeviceInfo) {
	cmd.DeviceIdentifiers = append(cmd.DeviceIdentifiers, deviceId)
	cmd.Devices[deviceId] = deviceInfo.Device
	cmd.DevicesMetadata[deviceId] = deviceInfo.Metadata
}

// matchDut finds a dut within the dutPool that contains a board that matches the requested board.
func (cmd *ParseDutTopologyCmd) matchDut(dutPool []*DeviceInfo, board string) (*DeviceInfo, error) {
	foundIndex := -1
	for i, deviceMetadataPair := range dutPool {
		if deviceMetadataPair.Metadata.GetDutModel().GetBuildTarget() == board {
			foundIndex = i
			break
		}
	}
	if foundIndex == -1 {
		return nil, fmt.Errorf("Failed to find board_target %s within dut_topology", board)
	}
	match := dutPool[foundIndex]
	dutPool = slices.Delete(dutPool, foundIndex, foundIndex+1)
	return match, nil
}

func (cmd *ParseDutTopologyCmd) extractDepsFromHwTestStateKeeper(ctx context.Context, sk *data.HwTestStateKeeper) error {
	if sk.DutTopology == nil {
		return fmt.Errorf("Cmd %q missing dependency: DutTopology", cmd.GetCommandType())
	}
	cmd.DutTopology = sk.DutTopology

	primaryBoard := common.GetValueFromRequestKeyvals(ctx, sk.CftTestRequest, sk.CrosTestRunnerRequest, "primary-board")
	if primaryBoard == "" {
		logging.Infof(ctx, "Cmd %s missing non-required dependency: primary-board", cmd.GetCommandType())
	}
	cmd.PrimaryBoard = primaryBoard

	companionBoards := common.GetValueFromRequestKeyvals(ctx, sk.CftTestRequest, sk.CrosTestRunnerRequest, "companion-boards")
	if companionBoards == "" {
		logging.Infof(ctx, "Cmd %s missing non-required dependency: companion-boards", cmd.GetCommandType())
		cmd.CompanionBoards = []string{}
	} else {
		cmd.CompanionBoards = strings.Split(companionBoards, ",")
	}

	return nil
}

func (cmd *ParseDutTopologyCmd) updateHwTestStateKeeper(
	ctx context.Context,
	sk *data.HwTestStateKeeper) error {

	sk.CompanionDevices = []*testapi.CrosTestRequest_Device{}
	sk.CompanionDevicesMetadata = []*skylab_test_runner.CFTTestRequest_Device{}
	sk.Devices = map[string]*testapi.CrosTestRequest_Device{}

	for deviceId, device := range cmd.Devices {
		deviceMetadata := cmd.DevicesMetadata[deviceId]
		if deviceId == "primaryDevice" {
			sk.PrimaryDevice = device
			sk.PrimaryDeviceMetadata = deviceMetadata
		} else {
			sk.CompanionDevices = append(sk.CompanionDevices, device)
			sk.CompanionDevicesMetadata = append(sk.CompanionDevicesMetadata, deviceMetadata)
		}
		sk.Devices[deviceId] = device
		if err := sk.Injectables.Set(deviceId, device); err != nil {
			logging.Infof(ctx, "Warning: cmd %s failed to set %s in the injectable storage, %s", cmd.GetCommandType(), deviceId, err)
		}
		if err := sk.Injectables.Set(deviceId+"Metadata", deviceMetadata); err != nil {
			logging.Infof(ctx, "Warning: cmd %s failed to set %s in the injectable storage, %s", cmd.GetCommandType(), deviceId+"Metadata", err)
		}
		if err := sk.Injectables.Set(deviceId+"ProvisionResponse", &testapi.InstallResponse{}); err != nil {
			logging.Infof(ctx, "Warning: cmd %s failed to set %s in the injectable storage, %s", cmd.GetCommandType(), deviceId+"ProvisionResponse", err)
		}
	}

	sk.DeviceIdentifiers = cmd.DeviceIdentifiers
	if err := sk.Injectables.Set("companionDevices", sk.CompanionDevices); err != nil {
		logging.Infof(ctx, "Warning: cmd %s failed to set companionDevices in the injectable storage, %s", cmd.GetCommandType(), err)
	}
	if err := sk.Injectables.Set("companionDevicesMetadata", sk.CompanionDevicesMetadata); err != nil {
		logging.Infof(ctx, "Warning: cmd %s failed to set companionDevicesMetadata in the injectable storage, %s", cmd.GetCommandType(), err)
	}

	return nil
}

// parseDut extracts ssh and board/model info from the lab dut and constructs
// a pair of CFT compatible objects.
func parseDut(dut *labapi.Dut) (*testapi.CrosTestRequest_Device, *skylab_test_runner.CFTTestRequest_Device) {
	var ssh *labapi.IpEndpoint
	var model *labapi.DutModel
	var device *testapi.CrosTestRequest_Device
	var deviceMetadata *skylab_test_runner.CFTTestRequest_Device
	switch dutType := dut.DutType.(type) {
	case *labapi.Dut_Chromeos:
		ssh = dutType.Chromeos.GetSsh()
		model = dutType.Chromeos.GetDutModel()
	case *labapi.Dut_Android_:
		// AssociatedHostname points to the ssh address of the
		// labstation associated with the android device, not
		// the android device itself. This value is used in
		// provisioning but testing expects the actual android
		// device address, which will need to be discovered
		// with some pre-test service.
		ssh = dutType.Android.GetAssociatedHostname()
		ssh.Port = 22
		model = dutType.Android.GetDutModel()
	case *labapi.Dut_Devboard_:
		ssh = dutType.Devboard.GetServo().GetServodAddress()
		model = &labapi.DutModel{
			BuildTarget: dutType.Devboard.BoardType,
			ModelName:   dutType.Devboard.BoardType,
		}
	default:
	}
	device = &api.CrosTestRequest_Device{
		Dut:       dut,
		DutServer: ssh,
	}
	deviceMetadata = &skylab_test_runner.CFTTestRequest_Device{
		DutModel: model,
	}

	return device, deviceMetadata
}

func NewParseDutTopologyCmd() *ParseDutTopologyCmd {
	abstractCmd := interfaces.NewAbstractCmd(ParseDutTopologyCmdType)
	abstractSingleCmdByNoExecutor := &interfaces.AbstractSingleCmdByNoExecutor{AbstractCmd: abstractCmd}
	return &ParseDutTopologyCmd{AbstractSingleCmdByNoExecutor: abstractSingleCmdByNoExecutor}
}
