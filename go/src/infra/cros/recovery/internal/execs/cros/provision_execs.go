// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"
)

// provisionExec performs provisioning of the device.
//
// To prevent reboot of device please provide action exec argument 'no_reboot'.
// To provide custom image data please use 'os_name', 'os_bucket', 'os_image_path'.
func provisionExec(ctx context.Context, info *execs.ExecInfo) error {
	sv, err := info.Versioner().Cros(ctx, info.GetDut().Name)
	if err != nil {
		return errors.Annotate(err, "cros provision").Err()
	}
	argsMap := info.GetActionArgs(ctx)
	osImageName := argsMap.AsString(ctx, "os_name", sv.OSImage)
	log.Debugf(ctx, "Used OS image name: %s", osImageName)
	osImageBucket := argsMap.AsString(ctx, "os_bucket", gsCrOSImageBucket)
	log.Debugf(ctx, "Used OS bucket name: %s", osImageBucket)
	osImagePath := argsMap.AsString(ctx, "os_image_path", fmt.Sprintf("%s/%s", osImageBucket, osImageName))
	log.Debugf(ctx, "Used OS image path: %s", osImagePath)
	req := &tlw.ProvisionRequest{
		Resource:        info.RunArgs.ResourceName,
		PreventReboot:   false,
		SystemImagePath: osImagePath,
	}
	if _, ok := argsMap["no_reboot"]; ok {
		req.PreventReboot = true
		log.Debugf(ctx, "Cros provision will be perform without reboot.")
	}
	log.Debugf(ctx, "Cros provision OS image path: %s", req.SystemImagePath)
	err = info.RunArgs.Access.Provision(ctx, req)
	return errors.Annotate(err, "cros provision").Err()
}

// Download image to the USB-drive.
//
// To provide custom image data please use 'os_name', 'os_bucket', 'os_image_path'.
func downloadImageToUSBExec(ctx context.Context, info *execs.ExecInfo) error {
	sv, err := info.Versioner().Cros(ctx, info.GetDut().Name)
	if err != nil {
		return errors.Annotate(err, "cros provision").Err()
	}
	argsMap := info.GetActionArgs(ctx)
	osImageName := argsMap.AsString(ctx, "os_name", sv.OSImage)
	log.Debugf(ctx, "Used OS image name: %s", osImageName)
	osImageBucket := argsMap.AsString(ctx, "os_bucket", gsCrOSImageBucket)
	log.Debugf(ctx, "Used OS bucket name: %s", osImageBucket)
	osImagePath := argsMap.AsString(ctx, "os_image_path", fmt.Sprintf("%s/%s", osImageBucket, osImageName))
	log.Debugf(ctx, "Used OS image path: %s", osImagePath)
	// Requesting convert GC path to caches service path.
	// Example: `http://Addr:8082/download/chromeos-image-archive/board-release/R99-XXXXX.XX.0/`
	downloadPath, err := info.RunArgs.Access.GetCacheUrl(ctx, info.GetDut().Name, osImagePath)
	if err != nil {
		return errors.Annotate(err, "download image to usb-drive").Err()
	}
	// Path provided by TLS cannot be used for downloading and/or extracting the image file.
	// But we can utilize the address of caching service and apply some string manipulation to construct the URL that can be used for this.
	// Example: `http://Addr:8082/extract/chromeos-image-archive/board-release/R99-XXXXX.XX.0/chromiumos_test_image.tar.xz?file=chromiumos_test_image.bin`
	extractPath := strings.Replace(downloadPath, "/download/", "/extract/", 1)
	image := fmt.Sprintf("%s/chromiumos_test_image.tar.xz?file=chromiumos_test_image.bin", extractPath)
	log.Debugf(ctx, "Download image for USB-drive: %s", image)
	err = info.NewServod().Set(ctx, "download_image_to_usb_dev", image)
	return errors.Annotate(err, "download image to usb-drive").Err()
}

const (
	// provisionFailed - A flag file to indicate provision failures.
	// The file's location in stateful means that on successful update
	// it will be removed.  Thus, if this file exists, it indicates that
	// we've tried and failed in a previous attempt to update.
	// The file will be created every time a OS provision is kicked off.
	// TODO(b/229309510): Remove old marker file when new marker file is in use.
	provisionFailed       = "/var/tmp/provision_failed"
	provisionFailedMarker = "/mnt/stateful_partition/unencrypted/provision_failed"
)

// isLastProvisionSuccessfulExec confirms that the DUT successfully finished
// its last provision job.
//
// At the start of any update (e.g. for a Provision job), the code creates
// a marker file. The file will be removed if an update finishes successfully.
// Thus, the presence of the file indicates that a prior update failed.
// The verifier tests for the existence of the marker file and fails if
// it still exists.
func isLastProvisionSuccessfulExec(ctx context.Context, info *execs.ExecInfo) error {
	run := info.DefaultRunner()
	if _, err := run(ctx, 20*time.Second, fmt.Sprintf("test -f %s", provisionFailed)); err == nil {
		return errors.Reason("last provision successful: found fail provision marker on %q", provisionFailed).Err()
	}
	if _, err := run(ctx, 20*time.Second, fmt.Sprintf("test -f %s", provisionFailedMarker)); err == nil {
		return errors.Reason("last provision successful: found fail provision marker on %q", provisionFailedMarker).Err()
	}
	return nil
}

// chartOSMap is the required version for chart devices.
// It is used until new stable version tool is ready.
var chartOSMap = map[string]string{
	"scarlet,dru":   "scarlet-release/R102-14695.107.0",
	"kukui,krane":   "kukui-release/R106-15054.14.0",
	"nami,pantheon": "nami-release/R106-15054.42.0",
}

// getChartOS returns the required os of chart if predefined in chartOSMap
func getChartOS(ctx context.Context, info *execs.ExecInfo) (string, error) {
	b := fmt.Sprintf("%s,%s", info.GetChromeos().GetBoard(), info.GetChromeos().GetModel())
	expectedOS, ok := chartOSMap[b]
	if !ok {
		return "", errors.Reason("os not defined for %s device", b).Err()
	}
	return expectedOS, nil
}

// isCameraboxTabletOnOSVersionExec check if the tablet is on the required os
// version.
func isCameraboxTabletOnOSVersionExec(ctx context.Context, info *execs.ExecInfo) error {
	expectedOS, err := getChartOS(ctx, info)
	if err != nil {
		return errors.Annotate(err, "camerabox tablet match os version").Err()
	}
	log.Debugf(ctx, "Expected version: %s", expectedOS)
	fromDevice, err := releaseBuildPath(ctx, info.DefaultRunner())
	if err != nil {
		return errors.Annotate(err, "camerabox tablet match os version").Err()
	}
	log.Debugf(ctx, "Version on device: %s", fromDevice)
	if fromDevice != expectedOS {
		return errors.Reason("camerabox tablet os version: mismatch, expected %q, found %q", expectedOS, fromDevice).Err()
	}
	return nil
}

// provisionCameraboxTabletExec
func provisionCameraboxTabletExec(ctx context.Context, info *execs.ExecInfo) error {
	chartOSName, err := getChartOS(ctx, info)
	if err != nil {
		return errors.Annotate(err, "camerabox tablet match os version").Err()
	}
	info.ActionArgs = []string{fmt.Sprintf("os_name:%s", chartOSName)}
	return provisionExec(ctx, info)
}

func init() {
	execs.Register("cros_provision", provisionExec)
	execs.Register("servo_download_image_to_usb", downloadImageToUSBExec)
	execs.Register("cros_is_last_provision_successful", isLastProvisionSuccessfulExec)
	execs.Register("is_camerabox_tablet_on_os_version", isCameraboxTabletOnOSVersionExec)
	execs.Register("provision_camerabox_tablet", provisionCameraboxTabletExec)
}
