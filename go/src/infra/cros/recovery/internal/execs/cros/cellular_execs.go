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
	"infra/cros/recovery/logger/metrics"
	"infra/cros/recovery/tlw"
)

func init() {
	execs.Register("cros_audit_cellular", auditCellularExec)
}

// auditCellularExec will validate cellular modem and connectivity state.
func auditCellularExec(ctx context.Context, info *execs.ExecInfo) error {
	cellularState := info.GetChromeos().GetCellular()
	if cellularState == nil {
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
		cellularState.ModemState = tlw.HardwareState_HARDWARE_NORMAL
		return nil
	}

	if execs.SSHErrorInternal.In(err) || execs.SSHErrorCLINotFound.In(err) {
		cellularState.ModemState = tlw.HardwareState_HARDWARE_UNSPECIFIED
	} else if expected {
		// no modem detected but was expected by setup info
		// then we set needs_replacement as it is probably a hardware issue.
		cellularState.ModemState = tlw.HardwareState_HARDWARE_NEED_REPLACEMENT
	} else {
		cellularState.ModemState = tlw.HardwareState_HARDWARE_NOT_DETECTED
	}

	return errors.Annotate(err, "audit cellular").Err()
}
