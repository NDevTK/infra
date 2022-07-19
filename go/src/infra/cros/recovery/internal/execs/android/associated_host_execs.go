// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package android

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components"
	"infra/cros/recovery/internal/components/adb"
	"infra/cros/recovery/internal/components/cros"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
)

const (
	flagsDir   = "/var/lib/servod/"
	inUseFlag  = flagsDir + "%s_in_use"
	rebootFlag = flagsDir + "%s_reboot"
)

func newPinger(info *execs.ExecInfo) components.Pinger {
	hostName := info.GetAndroid().GetAssociatedHostname()
	return info.NewPinger(hostName)
}

func newRunner(info *execs.ExecInfo) components.Runner {
	hostName := info.GetAndroid().GetAssociatedHostname()
	return info.NewRunner(hostName)
}

// pingAssociatedHostExec verifies that associated host of the DUT is pingable.
func pingAssociatedHostExec(ctx context.Context, info *execs.ExecInfo) error {
	return cros.WaitUntilPingable(ctx, info.ActionTimeout, cros.PingRetryInteval, 2, newPinger(info), info.NewLogger())
}

// sshAssociatedHostExec verifies ssh access to the associated host of the DUT.
func sshAssociatedHostExec(ctx context.Context, info *execs.ExecInfo) error {
	return cros.WaitUntilSSHable(ctx, info.ActionTimeout, cros.SSHRetryInteval, newRunner(info), info.NewLogger())
}

// isAssociatedHostLabstationExec verifies that adb is installed at the DUT associated host.
func isAssociatedHostLabstationExec(ctx context.Context, info *execs.ExecInfo) error {
	board, err := cros.ReleaseBoard(ctx, newRunner(info), info.NewLogger())
	if err != nil {
		return errors.Annotate(err, "associated host is labstation").Err()
	}
	if !strings.Contains(board, "labstation") {
		return errors.Reason("associated host is not labstation").Err()
	}
	return nil
}

// hasADBVendorKeyExec verifies that adb vendor key is provisioned at the DUT associated host.
func hasADBVendorKeyExec(ctx context.Context, info *execs.ExecInfo) error {
	if len(info.ActionArgs) != 1 {
		return errors.Reason("invalid number of arguments: adb vendor key is required").Err()
	}
	return adb.CheckADBVendorKey(ctx, newRunner(info), info.NewLogger(), info.ActionArgs[0])
}

// hasADBInstalledExec verifies that adb is installed at the DUT associated host.
func hasADBInstalledExec(ctx context.Context, info *execs.ExecInfo) error {
	path, err := adb.ADBInstallPath(ctx, newRunner(info), info.NewLogger())
	if err != nil {
		return err
	}
	log.Debugf(ctx, "Adb path at the associated host: %q", path)
	if path == "" {
		return errors.Reason("adb is not installed at the dut associated host").Err()
	}
	return nil
}

// startADBServerExec ensures that adb server is running on the DUT associated host.
func startADBServerExec(ctx context.Context, info *execs.ExecInfo) error {
	if len(info.ActionArgs) != 1 {
		return errors.Reason("invalid number of arguments: adb vendor key is required").Err()
	}
	vendorKey := info.ActionArgs[0]
	return adb.StartADBServer(ctx, newRunner(info), info.NewLogger(), vendorKey)
}

// killADBServerExec kills adb server if it is running on the DUT associated host.
func killADBServerExec(ctx context.Context, info *execs.ExecInfo) error {
	return adb.KillADBServer(ctx, newRunner(info), info.NewLogger())
}

// isFileSystemWritable checks whether the stateful file systems are writable.
func isFileSystemWritableExec(ctx context.Context, info *execs.ExecInfo) error {
	// N.B. Order matters here:  Encrypted stateful is loop-mounted from a file in unencrypted stateful,
	// so we don't test for errors in encrypted stateful if unencrypted fails.
	testDirs := []string{"/mnt/stateful_partition", "/var/tmp"}
	return cros.IsFileSystemWritable(ctx, newRunner(info), info.NewLogger(), testDirs)
}

// createInUseFlagExec creates in-use flag file.
func createInUseFlagExec(ctx context.Context, info *execs.ExecInfo) error {
	const createInUseFlagCmd = "touch " + inUseFlag
	run := newRunner(info)
	serialNumber := info.GetAndroid().GetSerialNumber()
	if _, err := run(ctx, time.Minute, fmt.Sprintf(createInUseFlagCmd, serialNumber)); err != nil {
		log.Debugf(ctx, "Create in-use flag file: %s", err)
	}
	// Ignore errors.
	return nil
}

// removeInUseFlagExec removes in-use flag file.
func removeInUseFlagExec(ctx context.Context, info *execs.ExecInfo) error {
	const removeInUseFlagCmd = "rm -f " + inUseFlag
	run := newRunner(info)
	serialNumber := info.GetAndroid().GetSerialNumber()
	if _, err := run(ctx, time.Minute, fmt.Sprintf(removeInUseFlagCmd, serialNumber)); err != nil {
		log.Debugf(ctx, "Remove in-use file flag: %s", err)
	}
	// Ignore errors.
	return nil
}

// scheduleRebootExec creates a file to request reboot of the DUT associated host.
func scheduleRebootExec(ctx context.Context, info *execs.ExecInfo) error {
	const createRebootFlagCmd = "touch " + rebootFlag
	run := newRunner(info)
	serialNumber := info.GetAndroid().GetSerialNumber()
	if _, err := run(ctx, time.Minute, fmt.Sprintf(createRebootFlagCmd, serialNumber)); err != nil {
		log.Debugf(ctx, "Schedule a reboot request: %s", err)
	}
	// Ignore errors.
	return nil
}

func init() {
	execs.Register("android_associated_host_ping", pingAssociatedHostExec)
	execs.Register("android_associated_host_ssh", sshAssociatedHostExec)
	execs.Register("android_associated_host_is_labstation", isAssociatedHostLabstationExec)
	execs.Register("android_associated_host_has_vendor_key", hasADBVendorKeyExec)
	execs.Register("android_associated_host_has_adb", hasADBInstalledExec)
	execs.Register("android_associated_host_start_adb", startADBServerExec)
	execs.Register("android_associated_host_stop_adb", killADBServerExec)
	execs.Register("android_associated_host_fs_is_writable", isFileSystemWritableExec)
	execs.Register("android_associated_host_lock", createInUseFlagExec)
	execs.Register("android_associated_host_unlock", removeInUseFlagExec)
	execs.Register("android_associated_host_schedule_reboot", scheduleRebootExec)
}
