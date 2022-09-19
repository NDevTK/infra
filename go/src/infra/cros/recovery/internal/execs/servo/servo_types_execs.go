// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package servo

import (
	"context"
	"regexp"
	"strings"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components/servo"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"
)

// servoVerifyV3Exec verifies whether the servo attached to the servo host is servo_v3.
func servoVerifyV3Exec(ctx context.Context, info *execs.ExecInfo) error {
	// The "-servo" suffix will exist only when the setup is for type V3,
	// (i.e. there is no labstation present).
	if info.GetChromeos().GetServo().GetName() == "" {
		return errors.Reason("servo verify v3: name is empty").Err()
	}
	const servoSuffix = "-servo"
	if strings.HasSuffix(info.GetChromeos().GetServo().GetName(), servoSuffix) {
		return nil
	}
	return errors.Reason("servo verify v3: servo hostname does not carry suffix %q, this is not V3.", servoSuffix).Err()
}

// servoVerifyV4Exec verifies whether the servo attached to the servo
// host is of type V4 or V4p1.
func servoVerifyV4Exec(ctx context.Context, info *execs.ExecInfo) error {
	sType, err := WrappedServoType(ctx, info)
	if err != nil {
		log.Debugf(ctx, "Servo Verify V4: could not determine the servo type")
		return errors.Annotate(err, "servo verify v4").Err()
	}
	if !sType.IsV4() {
		log.Debugf(ctx, "Servo Verify V4: servo type is neither V4, or V4P1.")
		return errors.Reason("servo verify v4: servo type %q is not V4.", sType).Err()
	}
	return nil
}

// servoVerifyV4Exec verifies whether the servo attached to the servo
// host is of type V4p1.
func servoVerifyV4p1Exec(ctx context.Context, info *execs.ExecInfo) error {
	sType, err := WrappedServoType(ctx, info)
	if err != nil {
		log.Debugf(ctx, "Servo Verify V4p1: could not determine the servo type")
		return errors.Annotate(err, "servo verify v4p1").Err()
	}
	if !sType.IsV4() {
		log.Debugf(ctx, "Servo Verify V4: servo type is not V4P1.")
		return errors.Reason("servo verify v4p1: servo type %q is not V4p1.", sType).Err()
	}
	return nil
}

// servoVerifyV4p1BySerialNumberExec verifies whether the servo attached to the servo
// host is of type v4p1 based on its serial number.
func servoVerifyV4p1BySerialNumberExec(ctx context.Context, info *execs.ExecInfo) error {
	const servoV4p1SerialPrefix = "SERVOV4P1"
	sn := info.GetChromeos().GetServo().GetSerialNumber()
	if !strings.HasPrefix(sn, servoV4p1SerialPrefix) {
		return errors.Reason("servo verify v4p1 by serial number: the serial number %s does not have expected prefix %s.", sn, servoV4p1SerialPrefix).Err()
	}
	return nil
}

// servoVerifyServoMicroExec verifies whether the servo attached to
// the servo host is of type servo micro.
func servoVerifyServoMicroExec(ctx context.Context, info *execs.ExecInfo) error {
	sType, err := WrappedServoType(ctx, info)
	if err != nil {
		log.Debugf(ctx, "Servo Verify Servo Micro: could not determine the servo type")
		return errors.Annotate(err, "servo verify servo micro").Err()
	}
	if !sType.IsMicro() {
		log.Debugf(ctx, "Servo Verify servo micro: servo type is not servo micro.")
		return errors.Reason("servo verify servo micro: servo type %q is not servo micro.", sType).Err()
	}
	return nil
}

// servoIsDualSetupConfiguredExec checks whether the servo device has
// been setup in dual mode.
func servoIsDualSetupConfiguredExec(ctx context.Context, info *execs.ExecInfo) error {
	if info.GetDut() != nil && info.GetDut().ExtraAttributes != nil {
		if attrs, ok := info.GetDut().ExtraAttributes[tlw.ExtraAttributeServoSetup]; ok {
			for _, a := range attrs {
				if a == tlw.ExtraAttributeServoSetupDual {
					log.Debugf(ctx, "Servo Is Dual Setup Configured: servo device is configured to be in dual-setup mode.")
					return nil
				}
			}
		}
	}
	return errors.Reason("servo is dual setup configured: servo device is not configured to be in dual-setup mode").Err()
}

// servoVerifyDualSetupExec verifies whether the servo attached to the
// servo host actually exhibits dual setup.
func servoVerifyDualSetupExec(ctx context.Context, info *execs.ExecInfo) error {
	sType, err := WrappedServoType(ctx, info)
	if err != nil {
		return errors.Annotate(err, "servo verify dual setup").Err()
	}
	if !sType.IsDualSetup() {
		return errors.Reason("servo verify dual setup: servo type %q is not dual setup.", sType).Err()
	}
	return nil
}

// servoVerifyServoCCDExec verifies whether the servo attached to
// the servo host is of type servo ccd.
func servoVerifyServoCCDExec(ctx context.Context, info *execs.ExecInfo) error {
	sType, err := WrappedServoType(ctx, info)
	if err != nil {
		log.Debugf(ctx, "Servo Verify Servo CCD: could not determine the servo type")
		return errors.Annotate(err, "servo verify servo type ccd").Err()
	}
	if !sType.IsCCD() {
		log.Debugf(ctx, "Servo Verify servo CCD: servo type is not servo ccd.")
		return errors.Reason("servo verify servo ccd: servo type %q is not servo ccd.", sType).Err()
	}
	return nil
}

// mainDeviceIsGSCExec checks whether or not the servo device is CR50 or TI50.
func mainDeviceIsGSCExec(ctx context.Context, info *execs.ExecInfo) error {
	sType, err := WrappedServoType(ctx, info)
	if err != nil {
		return errors.Annotate(err, "main devices is gsc").Err()
	}
	md := sType.MainDevice()
	switch md {
	case servo.C2D2:
		fallthrough
	case servo.CCD_CR50:
		fallthrough
	case servo.CCD_GSC:
		info.NewLogger().Debugf("Found main device: %q", md)
		return nil
	default:
		return errors.Reason("main devices is gsc: found %q does not match expectations", md).Err()
	}
}

// servoTypeRegexMatchExec checks if servo_type match to provided regex.
func servoTypeRegexMatchExec(ctx context.Context, info *execs.ExecInfo) error {
	actionMap := info.GetActionArgs(ctx)
	regex := actionMap.AsString(ctx, "regex", "")
	if regex == "" {
		return errors.Reason("servo-type regex match: regex is empty").Err()
	}
	servoType, err := servo.GetServoType(ctx, info.NewServod())
	if err != nil {
		return errors.Annotate(err, "servo-type regex match").Err()
	}
	m, err := regexp.MatchString(regex, servoType.String())
	if err != nil {
		return errors.Annotate(err, "servo-type regex match").Err()
	}
	if !m {
		return errors.Reason("servo-type regex match: not match").Err()
	}
	return nil
}

func init() {
	execs.Register("is_servo_v3", servoVerifyV3Exec)
	execs.Register("is_servo_v4", servoVerifyV4Exec)
	execs.Register("is_servo_v4p1", servoVerifyV4p1Exec)
	execs.Register("is_servo_v4p1_by_serial_number", servoVerifyV4p1BySerialNumberExec)
	execs.Register("is_servo_micro", servoVerifyServoMicroExec)
	execs.Register("is_dual_setup_configured", servoIsDualSetupConfiguredExec)
	execs.Register("is_dual_setup", servoVerifyDualSetupExec)
	execs.Register("is_servo_type_ccd", servoVerifyServoCCDExec)
	execs.Register("servo_main_device_is_gcs", mainDeviceIsGSCExec)
	execs.Register("servo_type_regex_match", servoTypeRegexMatchExec)
}
