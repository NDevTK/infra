// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package rpm

import (
	"context"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components/cros"
	"infra/cros/recovery/internal/components/cros/rpm"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/tlw"
)

// rpmAuditWithoutBatteryExec verifies whether RPM is in working order
// when battery is absent.
func rpmAuditWithoutBatteryExec(ctx context.Context, info *execs.ExecInfo) error {
	if info.GetChromeos().GetRpmOutlet() == nil {
		return errors.Reason("rpm audit: outlet not present").Err()
	}
	argsMap := info.GetActionArgs(ctx)
	downTimeout := argsMap.AsDuration(ctx, "down_timeout", 120, time.Second)
	bootTimeout := argsMap.AsDuration(ctx, "boot_timeout", 150, time.Second)
	waitInterval := argsMap.AsDuration(ctx, "wait_interval", 5, time.Second)
	log := info.NewLogger()
	ping := info.DefaultPinger()
	run := info.DefaultRunner()
	log.Debugf("Set RPM off ...")
	if err := rpmPowerOffExec(ctx, info); err != nil {
		return errors.Annotate(err, "rpm audit").Err()
	}
	// RPM service is single thread, perform the action OFF can take up-to 60 seconds.
	log.Debugf("Start waiting until the device goes down...")
	if waitDownErr := cros.WaitUntilNotPingable(ctx, downTimeout, waitInterval, 2, ping, log); waitDownErr != nil {
		info.GetChromeos().GetRpmOutlet().State = tlw.RPMOutlet_WRONG_CONFIG
		log.Debugf("Failed to power down the host: restoring RPM to ON state.")
		rpmPowerOnExec(ctx, info)
		return errors.Annotate(waitDownErr, "rpm audit: resource still pingable").Err()
	}
	log.Debugf("Set RPM on ...")
	if err := rpmPowerOnExec(ctx, info); err != nil {
		return errors.Annotate(err, "rpm audit").Err()
	}
	log.Debugf("Start waiting until the device goes up...")
	if err := cros.WaitUntilSSHable(ctx, bootTimeout, waitInterval, run, log); err != nil {
		info.GetChromeos().GetRpmOutlet().State = tlw.RPMOutlet_WRONG_CONFIG
		return errors.Annotate(err, "rpm audit: resource did not booted").Err()
	}
	log.Debugf("Validation finished.")
	info.GetChromeos().GetRpmOutlet().State = tlw.RPMOutlet_WORKING
	return nil
}

// rpmAuditWithBatteryExec verifies whether RPM is in working order
// when battery is present.
func rpmAuditWithBatteryExec(ctx context.Context, info *execs.ExecInfo) error {
	if info.GetChromeos().GetRpmOutlet() == nil {
		return errors.Reason("rpm audit: outlet not present").Err()
	}
	if err := rpmPowerOffExec(ctx, info); err != nil {
		return errors.Annotate(err, "rpm audit").Err()
	}
	run := info.DefaultRunner()
	ping := info.NewPinger(info.GetDut().Name)
	am := info.GetActionArgs(ctx)
	timeOut := am.AsDuration(ctx, "timeout", 120, time.Second)
	waitInterval := am.AsDuration(ctx, "wait_interval", 5, time.Second)
	if err := rpm.ValidatePowerState(ctx, run, ping, false, timeOut, waitInterval); err != nil {
		return errors.Annotate(err, "rpm audit with battery").Err()
	}
	if err := rpmPowerOnExec(ctx, info); err != nil {
		return errors.Annotate(err, "rpm audit").Err()
	}
	if err := rpm.ValidatePowerState(ctx, run, ping, true, timeOut, waitInterval); err != nil {
		return errors.Annotate(err, "rpm audit with battery").Err()
	}
	return nil
}

func init() {
	execs.Register("rpm_audit_without_battery", rpmAuditWithoutBatteryExec)
	execs.Register("rpm_audit_with_battery", rpmAuditWithBatteryExec)
}
