// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package servo

import (
	"context"
	"strings"
	"time"

	labApi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components/servo"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/execs/servo/topology"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/logger/metrics"

	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func servoUSBHasCROSStableImageExec(ctx context.Context, info *execs.ExecInfo) error {
	sv, err := info.Versioner().Cros(ctx, info.GetDut().Name)
	if err != nil {
		return errors.Annotate(err, "servo usb-key has cros stable image").Err()
	}
	expectedImage := sv.OSImage
	if expectedImage == "" {
		return errors.Reason("servo usb-key has cros stable image: stable image is not specified").Err()
	}
	sh := info.GetChromeos().GetServo()
	if sh.GetName() == "" {
		return errors.Reason("servo usb-key has cros stable image: servo is not present as part of dut info").Err()
	}
	run := info.NewRunner(sh.GetName())
	servod := info.NewServod()
	logger := info.NewLogger()
	usbPath, err := servo.USBDrivePath(ctx, false, run, servod, logger)
	if err != nil {
		return errors.Annotate(err, "servo usb-key has cros stable image").Err()
	}
	imageName, err := servo.ChromeOSImageNameFromUSBDrive(ctx, usbPath, run, servod, logger)
	if err != nil {
		return errors.Annotate(err, "servo usb-key has cros stable image").Err()
	}
	if strings.Contains(expectedImage, imageName) {
		log.Infof(ctx, "The image %q found on USB-key and match to stable version", imageName)
		return nil
	}
	return errors.Reason("servo usb-key has cros stable image: expected %q but found %q", expectedImage, imageName).Err()
}

// createMetricsRecordWhenNewUSBDriveFound creates new metrics record when new USB drive detected in setup.
// TODO(gregorynisbet): refactor to reduce copy of better interaction with metrics, avoid passing info.
func createMetricsRecordWhenNewUSBDriveFound(ctx context.Context, info *execs.ExecInfo, newDevice *labApi.UsbDrive) {
	action := &metrics.Action{
		// TODO(b/248635230): When karte' Search API is capable of taking in asset tag,
		// change the query to use asset tag instead of using hostname.
		Hostname:   info.GetDut().Name,
		ActionKind: metrics.USBDriveDetectionKind,
		StartTime:  newDevice.GetFirstSeenTime().AsTime(),
		StopTime:   newDevice.GetFirstSeenTime().AsTime(),
		Status:     metrics.ActionStatusSuccess,
		Observations: []*metrics.Observation{
			metrics.NewStringObservation("serial", newDevice.GetSerial()),
			metrics.NewStringObservation("manufacturer", newDevice.GetManufacturer()),
		},
	}
	if info.GetMetrics() == nil {
		log.Debugf(ctx, "Skip creating metrics with %#v", action)
		return
	}
	if err := info.GetMetrics().Create(ctx, action); err != nil {
		log.Debugf(ctx, "Fail to create metrics for %q: %s", action.ActionKind, err)
	}
}

// createMetricsRecordWhenUSBDriveReplaced creates new metrics record when detected that device was replace or removed from setup.
// TODO(b/248635230): refactor to reduce copy of better interaction with metrics, avoid passing info.
func createMetricsRecordWhenUSBDriveReplaced(ctx context.Context, info *execs.ExecInfo, oldDevice, newDevice *labApi.UsbDrive) {
	newTime := time.Now()
	if newDevice != nil {
		newTime = newDevice.GetFirstSeenTime().AsTime()
	}
	duration := newTime.Sub(oldDevice.GetFirstSeenTime().AsTime())
	if info.GetMetrics() == nil {
		return
	}
	action := &metrics.Action{
		Hostname:   info.GetDut().Name,
		ActionKind: metrics.USBDriveReplacedKind,
		StartTime:  newTime,
		StopTime:   newTime,
		Status:     metrics.ActionStatusSuccess,
		Observations: []*metrics.Observation{
			metrics.NewStringObservation("serial", oldDevice.GetSerial()),
			metrics.NewStringObservation("manufacturer", oldDevice.GetManufacturer()),
			metrics.NewStringObservation("duration", duration.String()),
		},
	}
	if info.GetMetrics() == nil {
		log.Debugf(ctx, "Skip creating metrics with %#v", action)
		return
	}
	if err := info.GetMetrics().Create(ctx, action); err != nil {
		log.Debugf(ctx, "Fail to create metrics for %q: %s", action.ActionKind, err)
	}
}

// servoUpdateUSBKeyHistoryExec will update the inventory record for the servo's usbkey stick with the latest information.
func servoUpdateUSBKeyHistoryExec(ctx context.Context, info *execs.ExecInfo) error {
	sh := info.GetChromeos().GetServo()
	if sh.GetName() == "" {
		return errors.Reason("servo update usbkey history: servo is not present as part of dut info").Err()
	}
	servoSerial := sh.GetSerialNumber()
	usbDrives, err := topology.USBDrives(ctx, info.NewRunner(sh.GetName()), servoSerial)
	if err != nil {
		return errors.Annotate(err, "servo update usbkey history").Err()
	} else if len(usbDrives) > 1 {
		return errors.Reason("servo update usbkey history: too many usb-drive detected").Err()
	}
	var newDevice *labApi.UsbDrive
	if len(usbDrives) == 1 {
		newDevice = usbDrives[0]
		// If we have device then we set time of detection of it.
		newDevice.FirstSeenTime = timestamppb.New(time.Now())
	}
	oldDevice := sh.GetUsbDrive()
	// Two cases where we need to update the record for usbkey.
	if oldDevice.GetSerial() == "" && newDevice.GetSerial() == "" {
		log.Debugf(ctx, "USB drive not  found and was not in the setup.")
	} else if oldDevice.GetSerial() == "" {
		log.Debugf(ctx, "New USB drive detected.")
		createMetricsRecordWhenNewUSBDriveFound(ctx, info, newDevice)
		// Updating inventory record.
		sh.UsbDrive = newDevice
	} else if newDevice.GetSerial() == "" {
		log.Debugf(ctx, "USB drive removed from the servo.")
		createMetricsRecordWhenUSBDriveReplaced(ctx, info, oldDevice, newDevice)
	} else if oldDevice.GetSerial() != newDevice.GetSerial() {
		log.Debugf(ctx, "USB drive replaced to new one.")
		createMetricsRecordWhenNewUSBDriveFound(ctx, info, newDevice)
		createMetricsRecordWhenUSBDriveReplaced(ctx, info, oldDevice, newDevice)
		// Updating inventory record.
		sh.UsbDrive = newDevice
	}
	return nil
}

func init() {
	execs.Register("servo_usbkey_has_stable_image", servoUSBHasCROSStableImageExec)
	execs.Register("servo_update_usbkey_history", servoUpdateUSBKeyHistoryExec)
}
