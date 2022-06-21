// Copyright 2021 The Chromium OS Authors. All rights reserved.  Use
// of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components"
	"infra/cros/recovery/internal/components/cros/power"
	"infra/cros/recovery/internal/components/servo"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"
)

const (
	// The percentage of the battery that is considered to be not
	// enough.
	MinimumBatteryLevel = 80
)

// RecoveryModeRequiredPDOff examines whether servo_v4 role needs to
// snk when booting into recovery mode.
func RecoveryModeRequiredPDOff(ctx context.Context, run components.Runner, pinger components.Pinger, servod components.Servod, dut *tlw.Dut) (bool, error) {
	expectBattery := dut.GetChromeos().GetBattery() != nil
	if !expectBattery {
		log.Debugf(ctx, "recovery mode required PD off: DUT is not expected to have the battery")
		return false, nil
	}
	if p, err := power.ReadPowerInfo(ctx, run); err == nil {
		expectBattery, _ = p.HasBattery()
	}
	if !expectBattery {
		return false, nil
	}
	if pdControlSupported, err := servo.ServoSupportsBuiltInPDControl(ctx, servod); err != nil {
		return false, errors.Annotate(err, "require sink mode in recovery").Err()
	} else if !pdControlSupported {
		log.Debugf(ctx, "Require Sink Mode in Recovery: power delivery is no tsupported on this servo, snk mode is not needed for recovery.")
		return false, nil
	}
	if batteryLevel, err := servo.BatteryChargePercent(ctx, servod); err != nil {
		return false, errors.Reason("require sink mode in recovery: could not read the battery level using servod control.").Err()
	} else if batteryLevel < MinimumBatteryLevel {
		log.Debugf(ctx, "require sink mode in recovery: battery level %d is less than the thresold %d. We will attempt to boot host in recovery mode without changing servo to snk mode. Please note that the host may not be able to see usb drive in recovery mode later due to servo not in snk mode.", batteryLevel, MinimumBatteryLevel)
	}
	return true, nil
}
