// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package btpeer

import (
	"context"
	"fmt"
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
	// localOSImageStorePath is the path to the local image file after downloading it to the device.
	localOSImageStorePath = "/tmp/rpi.img"
	// Filesystem labels.
	bootALabel = "BOOT_A"
	rootALabel = "ROOT_A"
	bootBLabel = "BOOT_B"
	rootBLabel = "ROOT_B"
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

// createPartitionsExec creates two new "B"  partitions in the free space on the
// btpeer so the device can be AB updated.
//
// The final partioning looks like:
//
//	BOOT_A(FAT32) - existing partition
//	ROOT_A(EXT4)  - existing partition
//	BOOT_B(FAT32) - newly created
//	ROOT_B(EXT4)  - newly created
//
// The A partitions correspond to the "recovery" partitions while the B partitions
// are the ones that the B new OS versions will actually be flashed onto.
func createPartitionsExec(ctx context.Context, info *execs.ExecInfo) error {
	runner := btpeer.NewSshRunner(info.GetAccess(), info.GetActiveResource())
	helper, err := btpeer.NewPartitionHelperForPath(ctx, runner.Run, "/")
	if err != nil {
		return errors.Annotate(err, "create partitions: failed to create partition helper").Err()
	}

	// Add label to existing BOOT_A partition.
	if err := btpeer.SetLabelForFAT32Path(ctx, runner.Run, "/boot", bootALabel); err != nil {
		return errors.Annotate(err, "create partitions: failed to set BOOT_A label").Err()
	}

	// Add label to existing ROOT_A partition.
	if err := btpeer.SetLabelForEXTPath(ctx, runner.Run, "/", rootALabel); err != nil {
		return errors.Annotate(err, "create partitions: failed to set ROOT_A label").Err()
	}

	// Create the B boot partition with a fixed size of 1GB.
	if bootB, err := helper.CreateNewPrimaryPartitionBytes(ctx, "fat32", int64(1e9)); err != nil {
		return errors.Annotate(err, "create partitions: failed to create boot partition").Err()
		// Initialize an empty filesystem on the newly created boot partition and add the BOOT_B label.
	} else if err := btpeer.InitFAT32FS(ctx, runner.Run, bootB, bootBLabel); err != nil {
		return errors.Annotate(err, "create partitions: failed initialize ROOT_B fs").Err()
	}

	// Create the B root partition using all of the remaining free space.
	if rootB, err := helper.CreateNewPrimaryPartitionPercent(ctx, "ext4", 100); err != nil {
		return errors.Annotate(err, "create partitions: failed to create root partition").Err()
		// Initialize an empty filesystem on the newly created root partition and add the ROOT_B label.
	} else if err := btpeer.InitEXT4FS(ctx, runner.Run, rootB, rootBLabel); err != nil {
		return errors.Annotate(err, "create partitions: failed initialize ROOT_B fs").Err()
	}

	return nil
}

// downloadImageExec downloads an OS Image and extracts (.xz) it.
//
// Currently  this downloads the images directly from a url as we don't have a
// GCS bucket set up for these images. It also assumes images are .xz'd but can
// be easily changed once we land on a final image upload scheme.
func downloadImageExec(ctx context.Context, info *execs.ExecInfo) error {
	runner := btpeer.NewSshRunner(info.GetAccess(), info.GetActiveResource())
	argsMap := info.GetActionArgs(ctx)
	imagePath := argsMap.AsString(ctx, "local_image_path", localOSImageStorePath)
	downloadTimeout := argsMap.AsDuration(ctx, "download_timeout", 180, time.Second)
	img := argsMap.AsString(ctx, "image_path", "")
	unxzTimeout := argsMap.AsDuration(ctx, "unxz_timeout", 300, time.Second)

	if img == "" {
		return errors.Reason("download image: required image_path argument missing").Err()
	}

	if _, _, _, err := ssh.WgetURL(ctx, runner, downloadTimeout, img, "-O", imagePath+".xz"); err != nil {
		return errors.Annotate(err, "download image: failed to download image with wget").Err()
	}

	if _, err := runner.Run(ctx, unxzTimeout, "unxz", imagePath+".xz"); err != nil {
		return errors.Annotate(err, "download image: failed to unxz image").Err()
	}

	return nil
}

// hasPartitionsWithLabelsExec verifies that partitions can be found with the requested labels.
func hasPartitionsWithLabelsExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	runner := btpeer.NewSshRunner(info.GetAccess(), info.GetActiveResource())

	labels := argsMap.AsStringSlice(ctx, "labels", []string{})
	if len(labels) == 0 {
		return errors.Reason("partitions with labels exist: required argument: 'labels' not provided").Err()
	}

	for _, label := range labels {
		device, err := btpeer.GetFSWithLabel(ctx, runner.Run, label)
		if err != nil {
			return errors.Annotate(err, "partitions with labels exist: failed to find partition with label: %q", label).Err()
		}
		log.Infof(ctx, "Found device: %q with label: %q", label, device)
	}
	return nil
}

// provisionExec flashes a raspberry pi image to the ROOT_B/BOOT_B partitions.
func provisionExec(ctx context.Context, info *execs.ExecInfo) error {
	runner := btpeer.NewSshRunner(info.GetAccess(), info.GetActiveResource())
	argsMap := info.GetActionArgs(ctx)
	imagePath := argsMap.AsString(ctx, "local_image_path", localOSImageStorePath)

	// First find the partitions we are going to flash.
	rootB, err := btpeer.GetFSWithLabel(ctx, runner.Run, rootBLabel)
	if err != nil {
		return errors.Annotate(err, "provision: failed to find ROOT_B partition").Err()
	}
	bootB, err := btpeer.GetFSWithLabel(ctx, runner.Run, bootBLabel)
	if err != nil {
		return errors.Annotate(err, "provision: failed to find BOOT_B partition").Err()
	}

	// Load the image loopback device so we can interact with it as if it's just a normal device.
	device, err := btpeer.LoadImageAsLoopbackDevice(ctx, runner.Run, imagePath)
	if err != nil {
		return errors.Annotate(err, "provision: failed to get and associate image").Err()
	}

	imgBootDev := fmt.Sprintf("%sp1", device)
	imgRootDev := fmt.Sprintf("%sp2", device)

	// Change partition labels on image before flashing.
	// Do this here so if something goes wrong during the flashing the partition
	// The labels will still exist so we can find the correct partition next time.
	if err := btpeer.SetEXTLabel(ctx, runner.Run, imgRootDev, rootBLabel); err != nil {
		return errors.Annotate(err, "provision: failed to label ROOT_B").Err()
	}

	if err := btpeer.SetFAT32Label(ctx, runner.Run, imgBootDev, bootBLabel); err != nil {
		return errors.Annotate(err, "provision: failed to label BOOT_B").Err()
	}

	// Flash boot partition.
	if err := btpeer.FlashImage(ctx, runner.Run, imgBootDev, bootB); err != nil {
		return errors.Annotate(err, "provision: failed to flash BOOT_B parition").Err()
	}

	// Flash root partition.
	if err := btpeer.FlashImage(ctx, runner.Run, imgRootDev, rootB); err != nil {
		return errors.Annotate(err, "provision: failed to flash ROOT_B parition").Err()
	}

	// Finally, update the partition IDs to match the new device.
	// Without this, the raspberry PI will search for partitions that match the UUID in the image since it's expecting
	// that we dd the entire .img to the device rather than individual partitions.
	// This is similar to what is already done in the default raspberry pi OS during first boot:
	// https://github.com/RPi-Distro/raspi-config/blob/bookworm/usr/lib/raspi-config/init_resize.sh#L90-L91

	// Mount the newly flashed partitions so we can update the files.
	imgBootMount, err := btpeer.MountDevice(ctx, runner.Run, bootB)
	if err != nil {
		return errors.Annotate(err, "update partition id: failed to mount image boot partition").Err()
	}

	imgRootMount, err := btpeer.MountDevice(ctx, runner.Run, rootB)
	if err != nil {
		return errors.Annotate(err, "update partition id: failed to mount image boot partition").Err()
	}

	// Replace old boot and root partition IDs in fstab.
	if err := btpeer.ReplacePartID(ctx, runner.Run, imgRootDev, rootB, imgRootMount+"/etc/fstab"); err != nil {
		return errors.Annotate(err, "update partition id: failed to replace old image root PART ID").Err()
	}

	if err := btpeer.ReplacePartID(ctx, runner.Run, imgBootDev, bootB, imgRootMount+"/etc/fstab"); err != nil {
		return errors.Annotate(err, "update partition id: failed to replace old image root PART ID").Err()
	}

	// Replace old root partition ID in cmdline.txt.
	if err := btpeer.ReplacePartID(ctx, runner.Run, imgRootDev, rootB, imgBootMount+"/cmdline.txt"); err != nil {
		return errors.Annotate(err, "update partition id: failed to replace old image root PART ID").Err()
	}

	return nil
}

// setTempBootPartitionExec temporarily boots into the specified partition.
//
// This is done prior to changing the permanent boot partition to verify that the
// partition is valid and bootable before making any permanent changes.
func setTempBootPartitionExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	partition := argsMap.AsString(ctx, "boot_partition_label", "")
	if partition == "" {
		return errors.Reason("set temp boot partition: required arg boot_partition_label not provided.").Err()
	}

	rebootTime := argsMap.AsDuration(ctx, "wait_reboot", 300, time.Second)
	runner := btpeer.NewSshRunner(info.GetAccess(), info.GetActiveResource())
	if err := btpeer.TempBootIntoPartition(ctx, runner.Run, partition, rebootTime); err != nil {
		return errors.Annotate(err, "set temp boot partition: failed to boot into requested partition").Err()
	}
	return nil
}

// setPermanentBootPartitionExec permanently boots into the requested partition.
//
// See: https://github.com/raspberrypi/documentation/blob/develop/documentation/asciidoc/computers/config_txt/autoboot.adoc
func setPermanentBootPartitionExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	partition := argsMap.AsString(ctx, "boot_partition_label", "")
	if partition == "" {
		return errors.Reason("set temp boot partition: required arg boot_partition_label not provided.").Err()
	}

	rebootTime := argsMap.AsDuration(ctx, "wait_reboot", 300, time.Second)
	runner := btpeer.NewSshRunner(info.GetAccess(), info.GetActiveResource())
	// Raspberry PI bootloader will search the first FAT32 partition for
	// an autoboot.txt file which points to which partition contains the boot partition.
	// This is sometimes a separate RECOVERY partition, but for us it is just BOOT_A partition
	if err := btpeer.PermBootIntoPartition(ctx, runner.Run, bootALabel, partition, rebootTime); err != nil {
		return errors.Reason("set perm boot partition: failed to set perm boot partition.").Err()
	}
	return nil
}

func init() {
	execs.Register("btpeer_enable_initrd", enableInitrdExec)
	execs.Register("btpeer_disable_initrd", disableInitrdExec)
	execs.Register("btpeer_has_partition_room", hasRoomToPartitionExec)
	execs.Register("btpeer_shrink_rootfs", shrinkRootFSExec)
	execs.Register("btpeer_partition_device", createPartitionsExec)
	execs.Register("btpeer_has_partitions_with_labels", hasPartitionsWithLabelsExec)
	execs.Register("btpeer_device_has_standard_partitions", hasStandardPartitioningExec)
	execs.Register("btpeer_download_image", downloadImageExec)
	execs.Register("btpeer_set_permanent_boot_partition", setPermanentBootPartitionExec)
	execs.Register("btpeer_temp_boot_into_partition", setTempBootPartitionExec)
	execs.Register("btpeer_provision_device", provisionExec)
}
