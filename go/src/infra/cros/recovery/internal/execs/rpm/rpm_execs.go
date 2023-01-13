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
	_, r, err := deviceHostnameAndRPMOutlet(info, deviceType)
	if err != nil {
		return errors.Annotate(err, "has rpm info: ").Err()
	}
	if r.GetHostname() != "" && r.GetOutlet() != "" {
		return nil
	}
	r.State = tlw.RPMOutlet_MISSING_CONFIG
	return errors.Reason("has rpm info: rpm for %s not present or incorrect", deviceType).Err()
}

// rpmPowerCycleDeviceExec performs power cycle the device by RPM.
// This function use RPM service built-in cycle interface which has an 5 seconds interval between power state change.
func rpmPowerCycleDeviceExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	deviceType := argsMap.AsString(ctx, "device_type", "")
	hostname, r, err := deviceHostnameAndRPMOutlet(info, deviceType)
	if err != nil {
		return errors.Annotate(err, "rpm power cycle:").Err()
	}
	if err := info.RPMAction(ctx, hostname, r, tlw.RunRPMActionRequest_CYCLE); err != nil {
		return errors.Annotate(err, "rpm power cycle ").Err()
	}
	log.Debugf(ctx, "RPM power cycle %s finished with success.", deviceType)
	return nil
}

// rpmPowerOffDeviceExec performs power off the device by RPM.
func rpmPowerOffDeviceExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	deviceType := argsMap.AsString(ctx, "device_type", "")
	hostname, r, err := deviceHostnameAndRPMOutlet(info, deviceType)
	if err != nil {
		return errors.Annotate(err, "rpm power off:").Err()
	}
	if err := info.RPMAction(ctx, hostname, r, tlw.RunRPMActionRequest_OFF); err != nil {
		return errors.Annotate(err, "rpm power off:").Err()
	}
	log.Debugf(ctx, "RPM power OFF %s finished with success.", deviceType)
	return nil

}

// rpmPowerOffExec performs power on the device by RPM.
func rpmPowerOnDeviceExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	deviceType := argsMap.AsString(ctx, "device_type", "")
	hostname, r, err := deviceHostnameAndRPMOutlet(info, deviceType)
	if err != nil {
		return errors.Annotate(err, "rpm power on:").Err()
	}
	if err := info.RPMAction(ctx, hostname, r, tlw.RunRPMActionRequest_ON); err != nil {
		return errors.Annotate(err, "rpm power dut on").Err()
	}
	log.Debugf(ctx, "RPM power ON %s finished with success.", deviceType)
	return nil
}

// activeChameleon finds active chameleon related to the executed plan.
func activeChameleon(info *execs.ExecInfo) (*tlw.Chameleon, error) {
	if c := info.GetChromeos().GetChameleon(); c != nil {
		if c.GetName() == info.GetActiveResource() {
			return c, nil
		}
	}
	return nil, errors.Reason("chameleon: chameleon %q not found", info.GetActiveResource()).Err()
}

// deviceHostnameAndRPMOutlet gets device hostname and its RPMOutlet given device type.
func deviceHostnameAndRPMOutlet(info *execs.ExecInfo, deviceType string) (string, *tlw.RPMOutlet, error) {
	switch deviceType {
	case "":
		return "", nil, errors.Reason("device hostname and rpmoutlet: device type not specified").Err()
	case "dut":
		return info.GetDut().Name, info.GetChromeos().GetRpmOutlet(), nil
	case "chameleon":
		c, err := activeChameleon(info)
		if err != nil {
			return "", nil, errors.Annotate(err, "device hostname and rpmoutlet for %q:", deviceType).Err()
		}
		return c.GetName(), c.GetRPMOutlet(), nil
	default:
		return "", nil, errors.Reason("device hostname and rpmoutlet: %q incorrect device type", deviceType).Err()
	}
}

func init() {
	// TODO(bniche@): retire non device execs after device execs are fully integrated.
	execs.Register("has_rpm_info", hasRpmInfoExec)
	execs.Register("rpm_power_cycle", rpmPowerCycleExec)
	execs.Register("rpm_power_off", rpmPowerOffExec)
	execs.Register("rpm_power_on", rpmPowerOnExec)

	execs.Register("device_has_rpm_info", hasRpmInfoDeviceExec)
	execs.Register("device_rpm_power_cycle", rpmPowerCycleDeviceExec)
	execs.Register("device_rpm_power_off", rpmPowerOffDeviceExec)
	execs.Register("device_rpm_power_on", rpmPowerOnDeviceExec)
}
