// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package btpeer

import (
	"context"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components/btpeer"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/execs/wifirouter/ssh"
)

// enableInitrdExec enables initrd on the btpeer.
func enableInitrdExec(ctx context.Context, info *execs.ExecInfo) error {
	runner := btpeer.NewSshRunner(info.GetAccess(), info.GetActiveResource())

	// Enable building initrd images in the kernel.
	if err := btpeer.AddLineToFile(ctx, runner.Run, "/etc/default/raspberrypi-kernel", "INITRD=Yes"); err != nil {
		return errors.Annotate(err, "enable initrd: failed to enable initrd building on Raspberry Pi").Err()
	}

	if err := btpeer.BuildInitrd(ctx, runner.Run); err != nil {
		return errors.Annotate(err, "enable initrd: failed to build initial initrd image").Err()
	}

	// Tell the kernel to use initrd image that we just built and renamed.
	if err := btpeer.AddLineToFile(ctx, runner.Run, "/boot/config.txt", "initramfs initrd.img followkernel"); err != nil {
		return errors.Annotate(err, "enable initrd: failed to build initial initrd image").Err()
	}

	argsMap := info.GetActionArgs(ctx)
	rebootTime := argsMap.AsDuration(ctx, "wait_reboot", 300, time.Second)
	if err := ssh.Reboot(ctx, runner, 10*time.Second, 10*time.Second, rebootTime); err != nil {
		return errors.Annotate(err, "enable initrd: failed to reboot btpeer").Err()
	}

	// Check that we find initrd messages in the dmesg after rebooting.
	if _, err := runner.Run(ctx, 30*time.Second, "dmesg -T | grep \"initrd\""); err != nil {
		return errors.Annotate(err, "enable initrd: failed to verify initrd is enabled on device after reboot").Err()
	}

	return nil
}

// disableInitrdExec disables initrd/initramfs on the btpeer.
func disableInitrdExec(ctx context.Context, info *execs.ExecInfo) error {
	runner := btpeer.NewSshRunner(info.GetAccess(), info.GetActiveResource())

	// Remove initramfs line from boot config.
	if err := btpeer.RemoveLineFromFile(ctx, runner.Run, "/boot/config.txt", "initramfs initrd.img followkernel"); err != nil {
		return errors.Annotate(err, "disable initrd: failed to build initial initrd image").Err()
	}

	argsMap := info.GetActionArgs(ctx)
	rebootTime := argsMap.AsDuration(ctx, "wait_reboot", 300, time.Second)
	if err := ssh.Reboot(ctx, runner, 10*time.Second, 10*time.Second, rebootTime); err != nil {
		return errors.Annotate(err, "disable initr: failed to reboot btpeer").Err()
	}

	// Check that the initrd message is not in dmesg.
	if _, err := runner.Run(ctx, 30*time.Second, "dmesg -T | grep \"initrd\""); err == nil {
		return errors.Annotate(err, "disable initrd: failed to verify initrd is disabled on device after reboot").Err()
	}

	return nil
}

func init() {
	execs.Register("btpeer_enable_initrd", enableInitrdExec)
	execs.Register("btpeer_disable_initrd", disableInitrdExec)
}
