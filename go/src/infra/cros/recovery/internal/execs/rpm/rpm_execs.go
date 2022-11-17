// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package rpm

import (
	"context"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"
)

// hasRpmInfoExec verifies if rpm info is present for DUT.
func hasRpmInfoExec(ctx context.Context, info *execs.ExecInfo) error {
	if r := info.GetChromeos().GetRpmOutlet(); r != nil {
		// TODO(otabek@): set fixed number to check and add accept argument value.
		if r.GetHostname() != "" && r.GetOutlet() != "" {
			return nil
		}
		r.State = tlw.RPMOutlet_MISSING_CONFIG
	}
	return errors.Reason("has rpm info: not present or incorrect").Err()
}

// rpmPowerCycleExec performs power cycle the device by RPM.
// This function use RPM service built-in cycle interface which has an 5 seconds interval between power state change.
func rpmPowerCycleExec(ctx context.Context, info *execs.ExecInfo) error {
	if err := info.RPMAction(ctx, info.GetDut().Name, info.GetChromeos().GetRpmOutlet(), tlw.RunRPMActionRequest_CYCLE); err != nil {
		return errors.Annotate(err, "rpm power cycle").Err()
	}
	log.Debugf(ctx, "RPM power cycle finished with success.")
	return nil
}

// rpmPowerOffExec performs power off the device by RPM.
func rpmPowerOffExec(ctx context.Context, info *execs.ExecInfo) error {
	if err := info.RPMAction(ctx, info.GetDut().Name, info.GetChromeos().GetRpmOutlet(), tlw.RunRPMActionRequest_OFF); err != nil {
		return errors.Annotate(err, "rpm power off").Err()
	}
	log.Debugf(ctx, "RPM power OFF finished with success.")
	return nil
}

// rpmPowerOffExec performs power on the device by RPM.
func rpmPowerOnExec(ctx context.Context, info *execs.ExecInfo) error {
	if err := info.RPMAction(ctx, info.GetDut().Name, info.GetChromeos().GetRpmOutlet(), tlw.RunRPMActionRequest_ON); err != nil {
		return errors.Annotate(err, "rpm power on").Err()
	}
	log.Debugf(ctx, "RPM power ON finished with success.")
	return nil
}

// hasRpmInfoDeviceExec verifies if rpm info is present for DUT.
func hasRpmInfoDeviceExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	deviceType := argsMap.AsString(ctx, "device_type", "")
	switch deviceType {
	case "dut":
		if r := info.GetChromeos().GetRpmOutlet(); r != nil {
			// TODO(otabek@): set fixed number to check and add accept argument value.
			if r.GetHostname() != "" && r.GetOutlet() != "" {
				return nil
			}
			r.State = tlw.RPMOutlet_MISSING_CONFIG
		}
		return errors.Reason("has rpm info: rpm for dut not present or incorrect").Err()

	case "chameleon":
		c, err := activeChameleon(info)
		if err != nil {
			return errors.Annotate(err, "has rpm info chameleon:").Err()
		}
		if r := c.GetRPMOutlet(); r != nil {
			if r.GetHostname() != "" && r.GetOutlet() != "" {
				return nil
			}
			r.State = tlw.RPMOutlet_MISSING_CONFIG
		}
		return errors.Reason("has rpm info chameleon: chameleon rpm not present or incorrect").Err()
	}
	return errors.Reason("has rpm info: device_type not specified or incorrect").Err()
}

// rpmPowerCycleDeviceExec performs power cycle the device by RPM.
// This function use RPM service built-in cycle interface which has an 5 seconds interval between power state change.
func rpmPowerCycleDeviceExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	deviceType := argsMap.AsString(ctx, "device_type", "")
	switch deviceType {
	case "dut":
		if err := info.RPMAction(ctx, info.GetDut().Name, info.GetChromeos().GetRpmOutlet(), tlw.RunRPMActionRequest_CYCLE); err != nil {
			return errors.Annotate(err, "rpm power cycle dut").Err()
		}
		log.Debugf(ctx, "RPM power cycle dut finished with success.")
		return nil
	case "chameleon":
		c, err := activeChameleon(info)
		if err != nil {
			return errors.Annotate(err, "rpm power cycle chameleon").Err()
		}
		if err := info.RPMAction(ctx, c.GetName(), c.GetRPMOutlet(), tlw.RunRPMActionRequest_CYCLE); err != nil {
			return errors.Annotate(err, "rpm power cycle chameleon").Err()
		}
		log.Debugf(ctx, "RPM power cycle %s finished with success.", c.GetName())
		return nil
	}
	return errors.Reason("RPM power cycle: device_type not specified or incorrect").Err()
}

// rpmPowerOffDeviceExec performs power off the device by RPM.
func rpmPowerOffDeviceExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	deviceType := argsMap.AsString(ctx, "device_type", "")
	switch deviceType {
	case "dut":
		if err := info.RPMAction(ctx, info.GetDut().Name, info.GetChromeos().GetRpmOutlet(), tlw.RunRPMActionRequest_OFF); err != nil {
			return errors.Annotate(err, "rpm power off dut").Err()
		}
		log.Debugf(ctx, "RPM power OFF dut finished with success.")
		return nil
	case "chameleon":
		c, err := activeChameleon(info)
		if err != nil {
			return errors.Annotate(err, "rpm power cycle chameleon").Err()
		}
		if err := info.RPMAction(ctx, c.GetName(), c.GetRPMOutlet(), tlw.RunRPMActionRequest_OFF); err != nil {
			return errors.Annotate(err, "rpm power off chameleon").Err()
		}
		log.Debugf(ctx, "RPM power OFF chameleon finished with success.")
		return nil
	}
	return errors.Reason("RPM power cycle: device_type not specified or incorrect").Err()

}

// rpmPowerOffExec performs power on the device by RPM.
func rpmPowerOnDeviceExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	deviceType := argsMap.AsString(ctx, "device_type", "")
	switch deviceType {
	case "dut":
		if err := info.RPMAction(ctx, info.GetDut().Name, info.GetChromeos().GetRpmOutlet(), tlw.RunRPMActionRequest_ON); err != nil {
			return errors.Annotate(err, "rpm power dut on").Err()
		}
		log.Debugf(ctx, "RPM power ON dut finished with success.")
		return nil
	case "chameleon":
		c, err := activeChameleon(info)
		if err != nil {
			return errors.Annotate(err, "rpm power on chameleon").Err()
		}
		if err := info.RPMAction(ctx, c.GetName(), c.GetRPMOutlet(), tlw.RunRPMActionRequest_ON); err != nil {
			return errors.Annotate(err, "rpm power on chameleon").Err()
		}
		log.Debugf(ctx, "RPM power ON chameleon finished with success.")
		return nil
	}
	return errors.Reason("RPM power on: device_type not specified or incorrect").Err()

}

// activeChameleon finds active chameleon related to the executed plan.
func activeChameleon(info *execs.ExecInfo) (*tlw.Chameleon, error) {
	if c := info.GetChromeos().GetChameleon(); c != nil {
		if c.GetName() == info.GetActiveResource() {
			return c, nil
		}
	}
	return nil, errors.Reason("chameleon: chameleon `%s` not found", info.GetActiveResource()).Err()
}

func init() {
	execs.Register("has_rpm_info", hasRpmInfoExec)
	execs.Register("rpm_power_cycle", rpmPowerCycleExec)
	execs.Register("rpm_power_off", rpmPowerOffExec)
	execs.Register("rpm_power_on", rpmPowerOnExec)
	execs.Register("device_has_rpm_info", hasRpmInfoDeviceExec)
	execs.Register("device_rpm_power_cycle", rpmPowerCycleDeviceExec)
	execs.Register("device_rpm_power_off", rpmPowerOffDeviceExec)
	execs.Register("device_rpm_power_on", rpmPowerOnDeviceExec)
}
