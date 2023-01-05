// Copyright 2021 The Chromium OS Authors. All rights reserved.  Use
// of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"fmt"
	"regexp"
	"time"

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
	return true, nil
}

func ProcessUptime(uptimeVal string) (*time.Duration, error) {
	// uptimePattern is a decimal number, possibly containing a decimal point.
	uptimePattern := regexp.MustCompile(`^\s*(\d+\.?\d*)\s+(\d+\.?\d*)\s*$`)
	parts := uptimePattern.FindStringSubmatch(uptimeVal)
	if len(parts) != 3 {
		// 'parts' consists of the complete match, as well as the two
		// captured groups for the two numerical quantities returned
		// upon reading /proc/uptime. Hence, in a correct read, the
		// array should have exactly three elements.
		return nil, errors.Reason("process uptime: fail to read value from %s", uptimeVal).Err()
	}
	dur, err := time.ParseDuration(fmt.Sprintf("%ss", parts[1]))
	if err != nil {
		return nil, errors.Annotate(err, "process uptime").Err()
	}
	return &dur, nil
}
