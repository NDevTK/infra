// Copyright 2018 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package harness manages the setup and teardown of various Swarming
// bot resources for running lab tasks, like results directories and
// host info.
package harness

import (
	"context"
	"log"

	"github.com/golang/protobuf/proto"
	"go.chromium.org/chromiumos/infra/proto/go/lab"
	"go.chromium.org/luci/common/errors"

	invV2 "infra/appengine/cros/lab_inventory/api/v1"
	"infra/cmd/skylab_swarming_worker/internal/swmbot"
	"infra/cmd/skylab_swarming_worker/internal/swmbot/harness/resultsdir"
	"infra/cmd/skylab_swarming_worker/internal/swmbot/harness/schedulingunit"
	"infra/libs/skylab/inventory"
	ufspb "infra/unifiedfleet/api/v1/models"
	chromeosLab "infra/unifiedfleet/api/v1/models/chromeos/lab"
)

// closer interface to wrap Close method with providing context.
type closer interface {
	Close(ctx context.Context) error
}

// Info holds information about the Swarming bot harness.
type Info struct {
	*swmbot.Info

	ResultsDir string
	DUTs       []*DUTHarness
	// err tracks errors during setup to simplify error handling
	// logic.
	err error

	closers []closer
}

// Close closes and flushes out the harness resources.  This is safe
// to call multiple times.
func (i *Info) Close(ctx context.Context) error {
	var errs []error
	for _, dh := range i.DUTs {
		if err := dh.Close(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	for n := len(i.closers) - 1; n >= 0; n-- {
		if err := i.closers[n].Close(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Annotate(errors.MultiError(errs), "close harness").Err()
	}
	return nil
}

// Open opens and sets up the bot and task harness needed for Autotest
// jobs.  An Info struct is returned with necessary fields, which must
// be closed.
func Open(ctx context.Context, b *swmbot.Info, o ...Option) (i *Info, err error) {
	i = &Info{
		Info: b,
	}
	defer func(i *Info) {
		if err != nil {
			_ = i.Close(ctx)
		}
	}(i)
	// Make result dir for swarming bot, which will be uploaded to GS once the
	// task completes.
	i.makeResultsDir()
	if i.err != nil {
		return nil, errors.Annotate(i.err, "open harness").Err()
	}
	i.loadDUTHarness(ctx)
	if i.err != nil {
		return nil, errors.Annotate(i.err, "load DUTHarness").Err()
	}
	for _, dh := range i.DUTs {
		for _, o := range o {
			o(dh)
		}
		// Load DUT's info(e.g. labels, attributes, stable_versions) from UFS/inventory.
		d, sv := dh.loadDUTInfo(ctx)
		// Load DUT's info from bot state file on drone.
		dh.loadLocalState(ctx)
		// Convert DUT's info into RTD(e.g. autotest) friendly format, a.k.a host_info_store.
		hi := dh.makeHostInfo(d, sv)
		dh.addBotInfoToHostInfo(hi)
		// Make a sub-dir for each DUT, which will be consumed by lucifer later.
		dh.makeDUTResultsDir()
		// Copying host_info_store file into DUT's result dir.
		dh.exposeHostInfo(hi)
		if dh.err != nil {
			return nil, errors.Annotate(dh.err, "open DUTharness").Err()
		}
	}
	return i, nil
}

func (i *Info) loadDUTHarness(ctx context.Context) {
	if i.err != nil {
		return
	}
	if i.Info.IsSchedulingUnit {
		su, err := schedulingunit.GetSchedulingUnitFromUFS(ctx, i.Info, i.Info.BotDUTID)
		if err != nil {
			i.err = errors.Annotate(err, "Failed to get Scheduling unit from UFS").Err()
			return
		}
		for _, hostname := range su.GetMachineLSEs() {
			d := makeDUTHarness(i.Info)
			d.DUTName = hostname
			i.DUTs = append(i.DUTs, d)
		}
	} else {
		d := makeDUTHarness(i.Info)
		d.DUTID = i.Info.BotDUTID
		i.DUTs = append(i.DUTs, d)
	}
}

func (i *Info) makeResultsDir() {
	if i.err != nil {
		return
	}
	path := i.Info.ResultsDir()
	rdc, err := resultsdir.Open(path)
	if err != nil {
		i.err = err
		return
	}
	log.Printf("Created results directory %s", path)
	i.closers = append(i.closers, rdc)
	i.ResultsDir = path
}

// TODO(xixuan): move it to lib.
func getStatesFromLabel(dutID string, l *inventory.SchedulableLabels) *lab.DutState {
	state := lab.DutState{
		Id: &lab.ChromeOSDeviceID{Value: dutID},
	}
	p := l.GetPeripherals()
	if p != nil {
		state.Servo = lab.PeripheralState(p.GetServoState())
		state.RpmState = lab.PeripheralState(p.GetRpmState())
		if p.GetChameleon() {
			state.Chameleon = lab.PeripheralState_WORKING
		}
		if p.GetAudioLoopbackDongle() {
			state.AudioLoopbackDongle = lab.PeripheralState_WORKING
		}
		state.WorkingBluetoothBtpeer = p.GetWorkingBluetoothBtpeer()
		switch l.GetCr50Phase() {
		case inventory.SchedulableLabels_CR50_PHASE_PVT:
			state.Cr50Phase = lab.DutState_CR50_PHASE_PVT
		case inventory.SchedulableLabels_CR50_PHASE_PREPVT:
			state.Cr50Phase = lab.DutState_CR50_PHASE_PREPVT
		}
		switch l.GetCr50RoKeyid() {
		case "prod":
			state.Cr50KeyEnv = lab.DutState_CR50_KEYENV_PROD
		case "dev":
			state.Cr50KeyEnv = lab.DutState_CR50_KEYENV_DEV
		}

		state.StorageState = lab.HardwareState(int32(p.GetStorageState()))
		state.ServoUsbState = lab.HardwareState(int32(p.GetServoUsbState()))
		state.BatteryState = lab.HardwareState(int32(p.GetBatteryState()))
	}
	return &state
}

func getMetaFromSpecs(dutID string, specs *inventory.CommonDeviceSpecs) *invV2.DutMeta {
	attr := specs.GetAttributes()
	dutMeta := invV2.DutMeta{
		ChromeosDeviceId: dutID,
	}
	for _, kv := range attr {
		if kv.GetKey() == "serial_number" {
			dutMeta.SerialNumber = kv.GetValue()
		}
		if kv.GetKey() == "HWID" {
			dutMeta.HwID = kv.GetValue()
		}
	}
	dutMeta.DeviceSku = specs.GetLabels().GetSku()
	return &dutMeta
}

func getLabMetaFromLabel(dutID string, l *inventory.SchedulableLabels) (labconfig *invV2.LabMeta) {
	labMeta := invV2.LabMeta{
		ChromeosDeviceId: dutID,
	}
	p := l.GetPeripherals()
	if p != nil {
		labMeta.ServoType = p.GetServoType()
		labMeta.SmartUsbhub = p.GetSmartUsbhub()
		labMeta.ServoTopology = convertServoTopology(p.GetServoTopology())
	}

	return &labMeta
}

// TODO (xixuan): will remove the above duplicated functions when UFS feature for OS lab is launched.
func getUFSDutMetaFromSpecs(dutID string, specs *inventory.CommonDeviceSpecs) *ufspb.DutMeta {
	attr := specs.GetAttributes()
	dutMeta := &ufspb.DutMeta{
		ChromeosDeviceId: dutID,
		Hostname:         specs.GetHostname(),
	}
	for _, kv := range attr {
		if kv.GetKey() == "serial_number" {
			dutMeta.SerialNumber = kv.GetValue()
		}
		if kv.GetKey() == "HWID" {
			dutMeta.HwID = kv.GetValue()
		}
	}
	dutMeta.DeviceSku = specs.GetLabels().GetSku()
	return dutMeta
}

func getUFSLabMetaFromSpecs(dutID string, specs *inventory.CommonDeviceSpecs) (labconfig *ufspb.LabMeta) {
	labMeta := &ufspb.LabMeta{
		ChromeosDeviceId: dutID,
		Hostname:         specs.GetHostname(),
	}
	p := specs.GetLabels().GetPeripherals()
	if p != nil {
		labMeta.ServoType = p.GetServoType()
		labMeta.SmartUsbhub = p.GetSmartUsbhub()
		labMeta.ServoTopology = copyServoTopology(convertServoTopology(p.GetServoTopology()))
	}

	return labMeta
}

func getUFSDutComponentStateFromSpecs(dutID string, specs *inventory.CommonDeviceSpecs) *chromeosLab.DutState {
	state := &chromeosLab.DutState{
		Id:       &chromeosLab.ChromeOSDeviceID{Value: dutID},
		Hostname: specs.GetHostname(),
	}
	l := specs.GetLabels()
	p := l.GetPeripherals()
	if p != nil {
		state.Servo = chromeosLab.PeripheralState(p.GetServoState())
		state.RpmState = chromeosLab.PeripheralState(p.GetRpmState())
		if p.GetChameleon() {
			state.Chameleon = chromeosLab.PeripheralState_WORKING
		}
		if p.GetAudioLoopbackDongle() {
			state.AudioLoopbackDongle = chromeosLab.PeripheralState_WORKING
		}
		state.WorkingBluetoothBtpeer = p.GetWorkingBluetoothBtpeer()
		switch l.GetCr50Phase() {
		case inventory.SchedulableLabels_CR50_PHASE_PVT:
			state.Cr50Phase = chromeosLab.DutState_CR50_PHASE_PVT
		case inventory.SchedulableLabels_CR50_PHASE_PREPVT:
			state.Cr50Phase = chromeosLab.DutState_CR50_PHASE_PREPVT
		}
		switch l.GetCr50RoKeyid() {
		case "prod":
			state.Cr50KeyEnv = chromeosLab.DutState_CR50_KEYENV_PROD
		case "dev":
			state.Cr50KeyEnv = chromeosLab.DutState_CR50_KEYENV_DEV
		}

		state.StorageState = chromeosLab.HardwareState(int32(p.GetStorageState()))
		state.ServoUsbState = chromeosLab.HardwareState(int32(p.GetServoUsbState()))
		state.BatteryState = chromeosLab.HardwareState(int32(p.GetBatteryState()))
	}
	return state
}

func copyServoTopology(topology *lab.ServoTopology) *chromeosLab.ServoTopology {
	if topology == nil {
		return nil
	}
	s := proto.MarshalTextString(topology)
	var newTopology chromeosLab.ServoTopology
	err := proto.UnmarshalText(s, &newTopology)
	if err != nil {
		log.Printf("cannot unmarshal servo topology: %s", err.Error())
		return nil
	}
	return &newTopology
}

func newServoTopologyItem(i *inventory.ServoTopologyItem) *lab.ServoTopologyItem {
	if i == nil {
		return nil
	}
	return &lab.ServoTopologyItem{
		Type:         i.GetType(),
		SysfsProduct: i.GetSysfsProduct(),
		Serial:       i.GetSerial(),
		UsbHubPort:   i.GetUsbHubPort(),
	}
}

func convertServoTopology(st *inventory.ServoTopology) *lab.ServoTopology {
	var t *lab.ServoTopology
	if st != nil {
		var children []*lab.ServoTopologyItem
		for _, child := range st.GetChildren() {
			children = append(children, newServoTopologyItem(child))
		}
		t = &lab.ServoTopology{
			Main:     newServoTopologyItem(st.Main),
			Children: children,
		}
	}
	return t
}
