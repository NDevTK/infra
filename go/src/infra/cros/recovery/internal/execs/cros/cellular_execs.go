// Copyright 2023 The ChromiumOS Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components/cros/cellular"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/logger/metrics"
	"infra/cros/recovery/tlw"
)

func init() {
	execs.Register("carrier_not_in", carrierNotInExec)
	execs.Register("cros_audit_cellular_modem", auditCellularModemExec)
	execs.Register("cros_audit_cellular_connection", auditCellularConnectionExec)
	execs.Register("cros_has_mmcli", hasModemManagerCLIExec)
	execs.Register("cros_has_modemmanager_job", hasModemManagerJobExec)
	execs.Register("cros_modemmanager_running", modemManagerRunningExec)
	execs.Register("cros_modem_state_not_in", modemStateNotInExec)
	execs.Register("cros_restart_modemmanager", restartModemManagerExec)
	execs.Register("set_cellular_modem_state", setCellularModemStateExec)
	execs.Register("has_cellular_info", hasCellularInfoExec)
}

// hasModemManagerCLIExec validates that mmcli is present on the DUT
func hasModemManagerCLIExec(ctx context.Context, info *execs.ExecInfo) error {
	if !cellular.HasModemManagerCLI(ctx, info.DefaultRunner(), info.GetExecTimeout()) {
		return errors.Reason("has modem manager cli: mmcli is not found on device").Err()
	}
	return nil
}

// hasModemManagerJobExec validates that modemmanager job is known by upstart and present on the DUT.
func hasModemManagerJobExec(ctx context.Context, info *execs.ExecInfo) error {
	if !cellular.HasModemManagerJob(ctx, info.DefaultRunner(), info.GetExecTimeout()) {
		return errors.Reason("has modem manager job: modemmanager is not found on device").Err()
	}
	return nil
}

// modemManagerRunningExec ensures modemmanager is running on the DUT and starts it if it's not already.
func modemManagerRunningExec(ctx context.Context, info *execs.ExecInfo) error {
	runner := info.DefaultRunner()
	argsMap := info.GetActionArgs(ctx)
	waitTimeout := argsMap.AsDuration(ctx, "wait_timeout", 10, time.Second)
	startTimeout := argsMap.AsDuration(ctx, "start_timeout", 30, time.Second)
	if cellular.WaitForModemManager(ctx, runner, waitTimeout) == nil {
		return nil
	}

	if err := cellular.StartModemManager(ctx, runner, startTimeout); err != nil {
		return errors.Annotate(err, "start modemmanager").Err()
	}

	if err := cellular.WaitForModemManager(ctx, runner, waitTimeout); err != nil {
		return errors.Annotate(err, "wait for modemmanager to start").Err()
	}
	return nil
}

// restartModemManagerExec restarts modemmanagr on the DUT.
func restartModemManagerExec(ctx context.Context, info *execs.ExecInfo) error {
	runner := info.DefaultRunner()
	argsMap := info.GetActionArgs(ctx)
	waitTimeout := argsMap.AsDuration(ctx, "wait_timeout", 10, time.Second)
	restartTimeout := argsMap.AsDuration(ctx, "restart_timeout", 30, time.Second)
	if err := cellular.RestartModemManager(ctx, runner, restartTimeout); err != nil {
		return errors.Annotate(err, "restart modemmanager").Err()
	}

	if err := cellular.WaitForModemManager(ctx, runner, waitTimeout); err != nil {
		return errors.Annotate(err, "wait for modemmanager to start").Err()
	}
	return nil
}

// hasCellularInfoExec validates that cellular data is populated in the dut info.
func hasCellularInfoExec(ctx context.Context, info *execs.ExecInfo) error {
	if c := info.GetChromeos().GetCellular(); c == nil {
		return errors.Reason("has cellular info: cellular data is not present in dut info").Err()
	}
	return nil
}

// setCellularModemStateExec sets the DUT's modem state to the requested value.
func setCellularModemStateExec(ctx context.Context, info *execs.ExecInfo) error {
	c := info.GetChromeos().GetCellular()
	if c == nil {
		return errors.Reason("set cellular modem state: cellular data is not present in dut info").Err()
	}

	actionMap := info.GetActionArgs(ctx)
	state := strings.ToUpper(actionMap.AsString(ctx, "state", ""))
	if state == "" {
		return errors.Reason("set cellular modem state: state is not provided").Err()
	}
	s, ok := tlw.HardwareState_value[state]
	if !ok {
		return errors.Reason("set cellular modem state state: state %q is invalid", state).Err()
	}

	c.ModemState = tlw.HardwareState(s)
	return nil
}

// carrierNotInExec validates that the DUT cellular network carrier is not in a provided list.
func carrierNotInExec(ctx context.Context, info *execs.ExecInfo) error {
	c := info.GetChromeos().GetCellular()
	if c == nil {
		return errors.Reason("carrier not in: cellular data is not present in dut info").Err()
	}

	if c.Carrier == "" {
		return errors.Reason("carrier not in: DUT carrier label is empty").Err()
	}

	argsMap := info.GetActionArgs(ctx)
	carriers := argsMap.AsStringSlice(ctx, "carriers", []string{})
	for _, carrier := range carriers {
		if strings.EqualFold(carrier, c.Carrier) {
			return errors.Reason("carrier not in: carrier %q is not allowed", c.Carrier).Err()
		}
	}

	return nil
}

// modemStateNotInExec verifies that the modem state exported by ModemManager is not in the provided list.
func modemStateNotInExec(ctx context.Context, info *execs.ExecInfo) error {
	c := info.GetChromeos().GetCellular()
	if c == nil {
		return errors.Reason("modem state not in: cellular data is not present in dut info").Err()
	}

	argsMap := info.GetActionArgs(ctx)
	timeout := argsMap.AsDuration(ctx, "wait_modem_timeout", 15, time.Second)
	modemInfo, err := cellular.WaitForModemInfo(ctx, info.DefaultRunner(), timeout)
	if err != nil {
		return errors.Reason("modem state not in: no modem exported by ModemManager").Err()
	}

	if modemInfo.GetState() == "" {
		return errors.Reason("modem state not in: modem state is empty").Err()
	}

	modemState := strings.ToUpper(modemInfo.GetState())
	states := argsMap.AsStringSlice(ctx, "states", []string{})
	for _, state := range states {
		if strings.ToUpper(state) == modemState {
			return errors.Reason("modem state not in: modem state %q not allowed", modemState).Err()
		}
	}

	return nil
}

// auditCellularConnectionExec verifies that the device is able to connect to the provided cellular network.
func auditCellularConnectionExec(ctx context.Context, info *execs.ExecInfo) error {
	runner := info.DefaultRunner()
	argsMap := info.GetActionArgs(ctx)
	waitTimeout := argsMap.AsDuration(ctx, "wait_timeout", 120, time.Second)
	if err := cellular.ConnectToDefaultService(ctx, runner, waitTimeout); err != nil {
		return errors.Annotate(err, "audit cellular connection").Err()
	}
	return nil
}

// auditCellularModem will validate cellular modem hardware state.
func auditCellularModemExec(ctx context.Context, info *execs.ExecInfo) error {
	c := info.GetChromeos().GetCellular()
	if c == nil {
		return errors.Reason("audit cellular modem: cellular data is not present in dut info").Err()
	}
	expected := cellular.IsExpected(ctx, info.DefaultRunner())

	// if no cellular is expected then set total timeout to be much lower otherwise we will add
	// ~2 minutes to every repair even ones that don't require a modem.
	argsMap := info.GetActionArgs(ctx)
	timeout := argsMap.AsDuration(ctx, "wait_manager_when_not_expected", 120, time.Second)
	if !expected {
		timeout = argsMap.AsDuration(ctx, "wait_manager_when_expected", 15, time.Second)
	}

	modemInfo, err := cellular.WaitForModemInfo(ctx, info.DefaultRunner(), timeout)
	if err == nil {
		// found modem, try to get connection status.
		connectionState := "UNKNOWN"
		if modemInfo.Modem.Generic != nil && modemInfo.Modem.Generic.State != "" {
			connectionState = strings.ToUpper(modemInfo.Modem.Generic.State)
		}

		// only report connection state for devices where modem was found.
		info.AddObservation(metrics.NewStringObservation("cellularConnectionState", connectionState))
		c.ModemState = tlw.HardwareState_HARDWARE_NORMAL
		return nil
	}

	err = errors.Annotate(err, "audit cellular modem").Err()
	if execs.SSHErrorInternal.In(err) {
		c.ModemState = tlw.HardwareState_HARDWARE_UNSPECIFIED
		return err
	}

	if expected {
		// no modem detected but was expected by setup info
		// then we set needs_replacement as it is probably a hardware issue.
		c.ModemState = tlw.HardwareState_HARDWARE_NEED_REPLACEMENT
		return err
	}

	// not found and not expected, don't report an error, instead just log it
	log.Errorf(ctx, err.Error())
	c.ModemState = tlw.HardwareState_HARDWARE_NOT_DETECTED
	return nil
}
