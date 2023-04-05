// Copyright 2022 The Chromium Authors
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
	"infra/cros/recovery/internal/retry"
	"infra/cros/recovery/logger/metrics"
	"infra/cros/recovery/tlw"

	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func servoUSBHasCROSStableImageExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	// This is the max number of times we'll try reading the ChromeOS image
	servodReadCROSImageRetryLimit := argsMap.AsInt(ctx, "retry_count", 1)
	// retryInterval is the timeout between retries for reading the ChromeOS image
	retryInterval := argsMap.AsDuration(ctx, "retry_interval", 1, time.Second)
	usbFileCheck := argsMap.AsBool(ctx, "usb_file_check", false)
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
	usbPath, newUSBState, err := servo.USBDrivePath(ctx, usbFileCheck, run, servod, logger)
	if err != nil {
		log.Debugf(ctx, "Servo USB Has CROS Stable Image: could not read usb key path")
		// Skip state if it is not specified.
		if newUSBState != tlw.HardwareState_HARDWARE_UNSPECIFIED {
			sh.UsbkeyState = newUSBState
			log.Infof(ctx, "New Servo USB-key state: %s", sh.GetUsbkeyState().String())
		}
		return errors.Annotate(err, "servo usb-key has cros stable image").Err()
	}
	var imageName string
	var getImageName = func() (err error) {
		imageName, err = servo.ChromeOSImageNameFromUSBDrive(ctx, usbPath, run, servod, logger)
		return err
	}
	if err := retry.LimitCount(ctx, servodReadCROSImageRetryLimit, retryInterval, getImageName, "get ChromeOS image name from usb"); err != nil {
		return errors.Annotate(err, "servo usb-key has cros stable image").Err()
	}
	if strings.Contains(expectedImage, imageName) {
		log.Infof(ctx, "The image %q found on USB-key and match to stable version", imageName)
		return nil
	}
	return errors.Reason("servo usb-key has cros stable image: expected %q but found %q", expectedImage, imageName).Err()
}

func servoUSBKeyIsDetectedExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	fileCheck := argsMap.AsBool(ctx, "file_check", false)
	servoHost := info.GetChromeos().GetServo()
	if servoHost.GetName() == "" {
		return errors.Reason("servo USB key is detected: servo host is not specified").Err()
	}
	servodRun := info.NewRunner(servoHost.GetName())
	servod := info.NewServod()
	usbPath, newUSBState, err := servo.USBDrivePath(ctx, fileCheck, servodRun, servod, info.NewLogger())
	var usbDetectionObservationValue string
	defer func() {
		if usbDetectionObservationValue != "" {
			info.AddObservation(metrics.NewStringObservation("usb_detection", usbDetectionObservationValue))
		}
	}()
	if err != nil {
		// Skip state if it is not specified.
		if newUSBState != tlw.HardwareState_HARDWARE_UNSPECIFIED {
			servoHost.UsbkeyState = newUSBState
			log.Infof(ctx, "New Servo USB-key state: %s", servoHost.GetUsbkeyState().String())
		}
		usbDetectionObservationValue = "usbkey_detection_failed"
		log.Debugf(ctx, "Fail to detect path to USB key connected to the servo from servo-host.")
		return errors.Annotate(err, "servo USB key is detected").Err()
	}
	// The replacement state can be overwritten by the audit logic.
	if servoHost.GetUsbkeyState() != tlw.HardwareState_HARDWARE_NEED_REPLACEMENT {
		servoHost.UsbkeyState = tlw.HardwareState_HARDWARE_NORMAL
	}
	usbDetectionObservationValue = "usbkey_detected"
	log.Debugf(ctx, "USB key is detected from servo-host as: %q.", usbPath)
	return nil
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
	execs.Register("servo_usbkey_is_detected", servoUSBKeyIsDetectedExec)
}
