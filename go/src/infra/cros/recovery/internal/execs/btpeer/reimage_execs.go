// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package btpeer

import (
	"context"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components/btpeer"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/execs/wifirouter/ssh"
	"infra/cros/recovery/internal/log"
)

const (
	// rootASizeGB is the size of the new A partition rootFS after shrinking (in GB).
	rootASizeGB = 16
	// bPartitionsSizeGB is the combined space for the B root+boot partitions (in GB), this
	// is smaller than A to make sure that we can fit it all on a 32GB SD card.
	bPartitionsSizeGB = 12
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

// shrinkRootFSExec shrinks the root partition using a pre-mount hook that is executed
// prior to mounting the root partition.
func shrinkRootFSExec(ctx context.Context, info *execs.ExecInfo) error {
	runner := btpeer.NewSshRunner(info.GetAccess(), info.GetActiveResource())

	argsMap := info.GetActionArgs(ctx)
	rootSize := int64(argsMap.AsInt(ctx, "root_a_size", rootASizeGB) * 1e9)

	clean, err := btpeer.CreateShrinkInitrdHook(ctx, runner.Run, rootSize)
	if err != nil {
		return errors.Annotate(err, "shrink rootfs: failed to create shrink hook").Err()
	}

	// Reboot device to run hook.
	if err := ssh.Reboot(ctx, runner, 10*time.Second, 10*time.Second, 10*time.Minute); err != nil {
		return errors.Annotate(err, "shrink rootfs: failed to reboot device").Err()
	}

	if err := clean(ctx); err != nil {
		return errors.Annotate(err, "shrink rootfs: failed to clean up shrink hook").Err()
	}
	return nil
}

// hasRoomToPartition verifies that the device has enough unused space to create the A/B partitioning.
func hasRoomToPartitionExec(ctx context.Context, info *execs.ExecInfo) error {
	runner := btpeer.NewSshRunner(info.GetAccess(), info.GetActiveResource())

	argsMap := info.GetActionArgs(ctx)
	rootSize := int64(argsMap.AsInt(ctx, "root_a_size", rootASizeGB) * 1e9)

	usedFSSpace, err := btpeer.GetUsedFSSpace(ctx, runner.Run, "/")
	if err != nil {
		return errors.Reason("has room to partition: failed to get current used bytes").Err()
	} else if usedFSSpace > rootSize {
		return errors.Reason("has room to partition: current file system %dB is too large to resize, need %dB", usedFSSpace, rootSize).Err()
	}
	log.Infof(ctx, "ROOT_A used space: %dB", usedFSSpace)

	helper, err := btpeer.NewPartitionHelperForPath(ctx, runner.Run, "/")
	if err != nil {
		return errors.Annotate(err, "has room to partition: failed to create partition helper").Err()
	}

	freeSpace, err := helper.GetFreeSpace(ctx)
	if err != nil {
		return errors.Annotate(err, "has room to partition: failed to create partition helper").Err()
	}
	log.Infof(ctx, "ROOT_A partition free space: %dB", freeSpace)

	partitionSize, err := helper.GetPartitionSize(ctx, "/")
	if err != nil {
		return errors.Annotate(err, "has room to partition: failed to get current partition size").Err()
	}
	log.Infof(ctx, "ROOT_A partition size: %dB", partitionSize)

	// Calculate final space available for the new partition.
	newPartitionSize := int64(argsMap.AsInt(ctx, "b_partition_size", bPartitionsSizeGB) * 1e9)
	partitionSpace := partitionSize + freeSpace - rootSize
	if partitionSpace < newPartitionSize {
		return errors.Reason("has room to partition: not enough room for new partitions, have %dB space available, want %dB", partitionSpace, newPartitionSize).Err()
	}

	return nil
}

// hasStandardPartitioningExec checks that the device has the normal OS partition scheme consisting of
// 1 boot (fat32) and 1 rootfs (ext4) partition.
func hasStandardPartitioningExec(ctx context.Context, info *execs.ExecInfo) error {
	runner := btpeer.NewSshRunner(info.GetAccess(), info.GetActiveResource())
	helper, err := btpeer.NewPartitionHelperForPath(ctx, runner.Run, "/")
	if err != nil {
		return errors.Annotate(err, "has standard partitioning: failed to create partition helper").Err()
	}

	parts, err := helper.GetPartitionInfo(ctx, "%", true)
	if err != nil {
		return errors.Annotate(err, "has standard partitioning: failed to get partition info").Err()
	}

	if len(parts) != 2 {
		return errors.Reason("has standard partitioning: expected 2 partitions, got %d", len(parts)).Err()
	}
	if !strings.EqualFold(parts[0].Type, "fat32") {
		return errors.Reason("has standard partitioning: expected first partition type fat32, got %q", parts[0].Type).Err()
	}
	if !strings.EqualFold(parts[1].Type, "ext4") {
		return errors.Reason("has standard partitioning: expected second partition type ext4, got %q", parts[1].Type).Err()
	}

	return nil
}

func init() {
	execs.Register("btpeer_enable_initrd", enableInitrdExec)
	execs.Register("btpeer_disable_initrd", disableInitrdExec)
	execs.Register("btpeer_has_partition_room", hasRoomToPartitionExec)
	execs.Register("btpeer_shrink_rootfs", shrinkRootFSExec)
	execs.Register("btpeer_device_has_standard_partitions", hasStandardPartitioningExec)
}
