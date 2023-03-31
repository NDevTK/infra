// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package servo

import (
	"context"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components/servo"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/execs/servo/topology"
	"infra/cros/recovery/internal/log"
)

// servoPowerCycleRootServoExec resets(power-cycle) the servo via smart usbhub.
func servoPowerCycleRootServoExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	// Timeout for resetting the servo. Default to be 30s.
	resetTimeout := argsMap.AsDuration(ctx, "reset_timeout", 30, time.Second)
	// Timeout to wait after resetting the servo. Default to be 20s.
	waitTimeout := argsMap.AsDuration(ctx, "wait_timeout", 20, time.Second)
	run := info.DefaultRunner()
	servoInfo := info.GetChromeos().GetServo()
	var smartUsbhubPresent = false
	defer func() {
		servoInfo.SmartUsbhubPresent = smartUsbhubPresent
	}()
	servoSerial := servoInfo.SerialNumber
	// Get the usb devnum before the reset.
	preResetDevnum, err := topology.GetServoUsbDevnum(ctx, run, servoSerial)
	if err != nil {
		return errors.Annotate(err, "servo power cycle root servo: find the servo").Err()
	}
	log.Infof(ctx, "Servo usb devnum before reset: %s", preResetDevnum)
	// Resetting servo.
	if _, err := run(ctx, 10*time.Second, "which cambronix_power_cycle"); err == nil {
		log.Infof(ctx, "Resetting servo through Cambronix usbhub.")
		if _, err := run(ctx, resetTimeout, "cambronix_power_cycle", servoSerial); err != nil {
			log.Warningf(ctx, `Failed to reset servo with serial: %s. Please ignore this error if the DUT is not connected to a Cambronix usbhub`, servoSerial)
			return errors.Annotate(err, "servo power cycle root servo by cambronix").Err()
		}
	} else {
		log.Infof(ctx, "Resetting servo through smart usbhub.")
		if _, err := run(ctx, resetTimeout, "servodtool", "device", "-s", servoSerial, "power-cycle"); err != nil {
			log.Warningf(ctx, "Failed to reset servo with serial: %s. Please ignore this error if the DUT is not connected to a smart usbhub", servoSerial)
			return errors.Annotate(err, "servo power cycle root servo").Err()
		}
	}
	// Since we are able to run the power cycle servodtool command
	// It implies the smartUsb is present.
	smartUsbhubPresent = true
	log.Debugf(ctx, "Wait %v for servo to come back from reset.", waitTimeout)
	time.Sleep(waitTimeout)
	// Reset authorized flag fror servo-hub for servo v4p1 only.
	if ResetUsbkeyAuthorized(ctx, run, servoSerial, info.GetChromeos().GetServo().ServodType) != nil {
		return errors.Annotate(err, "servo power cycle root servo").Err()
	}
	// Get the usb devnum after the reset.
	postResetDevnum, err := topology.GetServoUsbDevnum(ctx, run, servoSerial)
	if err != nil {
		return errors.Annotate(err, "servo power cycle root servo: after rest").Err()
	}
	log.Infof(ctx, "Servo usb devnum after reset: %s", postResetDevnum)
	if preResetDevnum == "" || postResetDevnum == "" {
		log.Infof(ctx, "Servo reset completed but unable to verify devnum change!")
	} else if preResetDevnum != postResetDevnum {
		log.Infof(ctx, "Reset servo with serial %s completed successfully!", servoSerial)
	} else {
		log.Infof(ctx, "Servo reset completed but devnum is still not changed!")
	}
	return nil
}

// allowsPowerCycleServoExec checks if power-cycle servo is allowed.
func allowsPowerCycleServoExec(ctx context.Context, info *execs.ExecInfo) error {
	const servoV4p1SerialPrefix = "SERVOV4P1"
	run := info.DefaultRunner()
	if strings.HasPrefix(info.GetChromeos().GetServo().GetSerialNumber(), servoV4p1SerialPrefix) {
		// Servo_v4p1 is not allowed to power-cycle except setup with cambrionix.
		if _, cErr := run(ctx, 30*time.Second, "which cambronix_power_cycle"); cErr == nil {
			log.Debugf(ctx, "Power-cycle servo is allowed for setup with cambronix")
			return nil
		} else {
			log.Debugf(ctx, "Fail to check cambrix script: %s", cErr)
		}
		return errors.Reason("allow power-cycle servo: not allowed").Err()
	}
	log.Debugf(ctx, "Power-cycle servo is allowed")
	return nil
}

// servoV4P1NetResetExec reset servo_v4p1 network controller.
func servoV4P1NetResetExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	// Timeout between off/on reset steps.
	resetTimeout := argsMap.AsDuration(ctx, "reset_timeout", 1, time.Second)
	servod := info.NewServod()
	logger := info.NewLogger()
	err := servo.ResetServoV4p1EthernetController(ctx, servod, logger, resetTimeout)
	return errors.Annotate(err, "servo_v4p1 net reset").Err()
}

func init() {
	execs.Register("servo_power_cycle_root_servo", servoPowerCycleRootServoExec)
	execs.Register("servo_v4p1_network_reset", servoV4P1NetResetExec)
	execs.Register("servo_allows_power_cycle_servo", allowsPowerCycleServoExec)
}
