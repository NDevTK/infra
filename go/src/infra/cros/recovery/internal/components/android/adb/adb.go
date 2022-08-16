// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package adb

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components"
	"infra/cros/recovery/logger"
)

var (
	adbVersionRegex = regexp.MustCompile(`Android Debug Bridge version (\d+\.\d+\.\d+)`)
	adbPathRegex    = regexp.MustCompile(`Installed as ([/\w]+adb)`)
)

// AdbVersionInfo holds info about installed Android Debug Bridge.
type adbVersionInfo struct {
	version string
	path    string
}

// getADBVersionInfo reads adb version info.
func getADBVersionInfo(ctx context.Context, run components.Runner, log logger.Logger) (*adbVersionInfo, error) {
	const adbVersionCmd = "adb version"
	output, err := run(ctx, time.Minute, adbVersionCmd)
	if err != nil {
		return nil, errors.Annotate(err, "get adb version info").Err()
	}
	adb := &adbVersionInfo{
		version: adbVersionRegex.FindString(output),
		path:    adbPathRegex.FindString(output),
	}
	log.Debugf("adb info: %#v", adb)
	return adb, nil
}

// CheckADBVendorKey checks adb vendor private key.
func CheckADBVendorKey(ctx context.Context, run components.Runner, log logger.Logger, vendorKey string) error {
	const adbCheckVendorKeyCmd = "openssl rsa -in %s -check -noout"
	cmd := fmt.Sprintf(adbCheckVendorKeyCmd, vendorKey)
	_, err := run(ctx, time.Minute, cmd)
	if err != nil {
		return errors.Annotate(err, "check adb vendor key").Err()
	}
	return nil
}

// StartADBServer ensures that there is adb server running.
func StartADBServer(ctx context.Context, run components.Runner, log logger.Logger, vendorKey string) error {
	const adbStartServerCmd = "ADB_VENDOR_KEYS=%s adb start-server"
	// Ensure adb run path exists since it will get wiped on every associated host reboot.
	if _, err := run(ctx, time.Minute, "mkdir -p /run/arc/adb"); err != nil {
		return errors.Annotate(err, "start adb server").Err()
	}
	cmd := fmt.Sprintf(adbStartServerCmd, vendorKey)
	if _, err := run(ctx, time.Minute, cmd); err != nil {
		return errors.Annotate(err, "start adb server").Err()
	}
	log.Debugf("adb server is running")
	return nil
}

// KillADBServer kills adb server if it is running.
func KillADBServer(ctx context.Context, run components.Runner, log logger.Logger) error {
	const adbKillServerCmd = "adb kill-server"
	if _, err := run(ctx, time.Minute, adbKillServerCmd); err != nil {
		return errors.Annotate(err, "kill adb server").Err()
	}
	log.Debugf("adb server is not running")
	return nil
}

// ADBVersion returns adb version number.
func ADBVersion(ctx context.Context, run components.Runner, log logger.Logger) (string, error) {
	adb, err := getADBVersionInfo(ctx, run, log)
	if err != nil {
		return "", errors.Annotate(err, "get adb version").Err()
	}
	log.Debugf("adb version: %q", adb.version)
	return adb.version, nil
}

// ADBInstallPath returns adb installation path.
func ADBInstallPath(ctx context.Context, run components.Runner, log logger.Logger) (string, error) {
	adb, err := getADBVersionInfo(ctx, run, log)
	if err != nil {
		return "", errors.Annotate(err, "get adb install path").Err()
	}
	log.Debugf("adb path: %q", adb.path)
	return adb.path, nil
}

// RestartADBDAsRoot restarts adbd on device with root permissions.
func RestartADBDAsRoot(ctx context.Context, run components.Runner, log logger.Logger, serialNumber string) error {
	const adbRootCmd = "adb -s %s root"
	cmd := fmt.Sprintf(adbRootCmd, serialNumber)
	if _, err := run(ctx, time.Minute, cmd); err != nil {
		return errors.Annotate(err, "restart adb as root").Err()
	}
	log.Debugf("adb runs with root permissions on the device: %q", serialNumber)
	return nil
}
