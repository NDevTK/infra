// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"

	"go.chromium.org/luci/common/errors"

	"infra/cros/dutstate"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/execs/cros/battery"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"
)

// auditBatteryExec confirms that it is able to audit battery info
// and mark the DUT if it needs replacement.
func auditBatteryExec(ctx context.Context, info *execs.ExecInfo) error {
	r := info.DefaultRunner()
	batteryInfo := info.GetChromeos().GetBattery()
	if batteryInfo == nil {
		return errors.Reason("audit battery: data is not present in dut info").Err()
	}
	b, err := battery.ReadBatteryInfo(ctx, r)
	if err != nil {
		batteryInfo.State = tlw.HardwareState_HARDWARE_UNSPECIFIED
		return errors.Annotate(err, "audit battery: dut battery state cannot extracted").Err()
	}
	hardwareState := battery.DetermineHardwareStatus(ctx, b.FullChargeCapacity, b.FullChargeCapacityDesigned)
	log.Infof(ctx, "Battery hardware state: %s", hardwareState)
	if hardwareState == tlw.HardwareState_HARDWARE_UNSPECIFIED {
		return errors.Reason("audit battery: dut battery did not detected or state cannot extracted").Err()
	}
	if hardwareState == tlw.HardwareState_HARDWARE_NEED_REPLACEMENT {
		log.Infof(ctx, "Detected issue with battery on the DUT.")
		batteryInfo.State = tlw.HardwareState_HARDWARE_NEED_REPLACEMENT
		log.Debugf(ctx, "Audit Battery Exec: setting dut state to %s", string(dutstate.NeedsReplacement))
		info.GetDut().State = dutstate.NeedsReplacement
	}
	return nil
}

func init() {
	execs.Register("cros_audit_battery", auditBatteryExec)
}
