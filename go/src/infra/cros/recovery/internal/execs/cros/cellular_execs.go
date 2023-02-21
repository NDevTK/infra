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
	execs.Register("cros_audit_cellular", auditCellularExec)
	execs.Register("cros_has_mmcli", hasModemManagerCLIExec)
	execs.Register("cros_has_modemmanager_job", hasModemManagerJobExec)
	execs.Register("cros_modemmanager_running", modemManagerRunningExec)
	execs.Register("cros_restart_modemmanager", restartModemManagerExec)
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

// auditCellularExec will validate cellular modem and connectivity state.
func auditCellularExec(ctx context.Context, info *execs.ExecInfo) error {
	c := info.GetChromeos().GetCellular()
	if c == nil {
		return errors.Reason("audit cellular: cellular data is not present in dut info").Err()
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

	err = errors.Annotate(err, "audit cellular").Err()
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
