// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"fmt"

	"google.golang.org/protobuf/types/known/durationpb"
)

// ProvisionBtpeerConfig creates configuration to provision a btpeer with a new OS.
func ProvisionBtpeerConfig(imagePath string) *Configuration {
	return &Configuration{
		PlanNames: []string{
			PlanBluetoothPeer,
		},
		Plans: map[string]*Plan{
			PlanBluetoothPeer: btpeerProvisionPlan(imagePath),
		},
	}
}

// btpeerProvisionPlan provisions a btpeer with a new OS. The call flow is as followed:
//
//  1. If the device is not already partitioned for AB updates, then re-partition is by:
//     a. Verify device has space for new B partitions
//     b. Verify Verify device has standard partition scheme
//     c. Enable initrd
//     d. Shrink rootfs
//     e. Disable initrd
//     f. Create new AB Partitions
//
//  2. If the device is A/B partitioned
//     a. Temp Boot into A partition
//     b. Set permanent boot partition as A
//     c. Download OS Image to  device
//     d. Flash B partitions with new OS Image
//     e. Temp Boot into B partition
//     f. Set permanent boot partition as B
func btpeerProvisionPlan(imagePath string) *Plan {
	return &Plan{
		CriticalActions: []string{
			"Device is SSHable",
			"Device is AB partitioned",
			"Provision OS",
		},
		Actions: map[string]*Action{
			"Device is SSHable": {
				Docs: []string{
					"Try to wait device to be sshable.",
					"Waiting time 150 seconds.",
				},
				ExecName:   "cros_ssh",
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Enable initrd": {
				Docs: []string{
					"Enable initrd on the btpeers so we can add a pre-mount hook for rootfs.",
				},
				Conditions: []string{},
				Dependencies: []string{
					"Device is SSHable",
				},
				ExecName:    "btpeer_enable_initrd",
				RunControl:  RunControl_ALWAYS_RUN,
				ExecTimeout: &durationpb.Duration{Seconds: 600},
			},
			"Disable initrd": {
				Docs: []string{
					"Disables use of initrd on the btpeer.",
				},
				Conditions: []string{},
				Dependencies: []string{
					"Device is SSHable",
				},
				ExecName:    "btpeer_disable_initrd",
				RunControl:  RunControl_ALWAYS_RUN,
				ExecTimeout: &durationpb.Duration{Seconds: 600},
			},
			"Verify device has standard partition scheme": {
				Docs: []string{
					"Verifies that the device has the default raspberry PI partitioning scheme",
					"consisting of 1 FAT32 boot partition and 1 EXT4 rootfs partition.",
				},
				Conditions: []string{},
				Dependencies: []string{
					"Device is SSHable",
				},
				ExecName:   "btpeer_device_has_standard_partitions",
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Verify device has space for new B partitions": {
				Docs: []string{
					"Verifies that the device has space to add the new",
					"BOOT_B/ROOT_B partitions",
				},
				Conditions: []string{},
				Dependencies: []string{
					"Device is SSHable",
				},
				ExecName:   "btpeer_has_partition_room",
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Shrink rootfs": {
				Docs: []string{
					"Shrinks the device rootfs so it can be partitioned.",
				},
				Conditions: []string{},
				Dependencies: []string{
					"Device is SSHable",
					"Verify device has space for new B partitions",
					"Verify device has standard partition scheme",
					"Enable initrd",
				},
				ExecName:    "btpeer_shrink_rootfs",
				RunControl:  RunControl_ALWAYS_RUN,
				ExecTimeout: &durationpb.Duration{Seconds: 1000},
			},
			"Temp Boot into A partition": {
				Docs: []string{
					"Temporarily boots into the A partition on the device so we can ",
					"verify it's ok before setting it as the permanent boot device.",
				},
				Conditions: []string{},
				Dependencies: []string{
					"Device is SSHable",
				},
				ExecName: "btpeer_temp_boot_into_partition",
				ExecExtraArgs: []string{
					"boot_partition_label:BOOT_A",
				},
				RunControl:  RunControl_ALWAYS_RUN,
				ExecTimeout: &durationpb.Duration{Seconds: 600},
			},
			"Set permanent boot partition as A": {
				Docs: []string{
					"Sets boot partition A as the default boot partition to use on reboot.",
				},
				Conditions: []string{},
				Dependencies: []string{
					"Device is SSHable",
				},
				ExecName: "btpeer_set_permanent_boot_partition",
				ExecExtraArgs: []string{
					"boot_partition_label:BOOT_A",
				},
				RunControl:  RunControl_ALWAYS_RUN,
				ExecTimeout: &durationpb.Duration{Seconds: 600},
			},
			"Temp Boot into B partition": {
				Docs: []string{
					"Temporarily boots into the B partition on the device so we can ",
					"verify it's ok before setting it as the permanent boot device.",
				},
				Conditions: []string{},
				Dependencies: []string{
					"Device is SSHable",
				},
				ExecName: "btpeer_temp_boot_into_partition",
				ExecExtraArgs: []string{
					"boot_partition_label:BOOT_B",
				},
				RunControl:  RunControl_ALWAYS_RUN,
				ExecTimeout: &durationpb.Duration{Seconds: 600},
			},
			"Set permanent boot partition as B": {
				Docs: []string{
					"Sets boot partition B as the default boot partition to use on reboot.",
				},
				Conditions: []string{},
				Dependencies: []string{
					"Device is SSHable",
				},
				ExecName: "btpeer_set_permanent_boot_partition",
				ExecExtraArgs: []string{
					"boot_partition_label:BOOT_B",
				},
				RunControl:  RunControl_ALWAYS_RUN,
				ExecTimeout: &durationpb.Duration{Seconds: 600},
			},
			"Download OS Image to device": {
				Docs: []string{
					"Downloads a the OS image to the device",
				},
				Conditions: []string{},
				Dependencies: []string{
					"Device is SSHable",
				},
				ExecName: "btpeer_download_image",
				ExecExtraArgs: []string{
					fmt.Sprintf("image_path:%s", imagePath),
				},
				RunControl:  RunControl_ALWAYS_RUN,
				ExecTimeout: &durationpb.Duration{Seconds: 960},
			},
			"Flash B partitions with downloaded OS image": {
				Docs: []string{
					"Flashes a new OS onto the BOOT/ROOT_B partitions",
				},
				Conditions: []string{},
				Dependencies: []string{
					"Device is SSHable",
				},
				ExecName:    "btpeer_provision_device",
				RunControl:  RunControl_ALWAYS_RUN,
				ExecTimeout: &durationpb.Duration{Seconds: 2400},
			},
			"Create new AB Partitions": {
				Docs: []string{
					"Takes a normal-partitioned raspberry PI partitions it for AB booting.",
				},
				Conditions: []string{},
				Dependencies: []string{
					"Device is SSHable",
				},
				ExecName:    "btpeer_partition_device",
				ExecTimeout: &durationpb.Duration{Seconds: 1000},
				RunControl:  RunControl_ALWAYS_RUN,
			},
			"AB Partition device": {
				Docs: []string{
					"Takes a device with the default partitioning scheme and re-partitions it.",
				},
				Conditions: []string{},
				Dependencies: []string{
					"Device is SSHable",
					"Verify device has space for new B partitions",
					"Verify device has standard partition scheme",
					"Enable initrd",
					"Shrink rootfs",
					"Disable initrd",
					"Create new AB Partitions",
				},
				ExecName:   "sample_pass",
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Device is AB partitioned": {
				Docs: []string{
					"Checks that the raspberry pi is AB partitioned",
				},
				Conditions: []string{},
				Dependencies: []string{
					"Device is SSHable",
				},
				ExecName: "btpeer_has_partitions_with_labels",
				ExecExtraArgs: []string{
					"labels:BOOT_A,ROOT_A,BOOT_B,ROOT_B",
				},
				RecoveryActions: []string{
					"AB Partition device",
				},
				RunControl: RunControl_ALWAYS_RUN,
			},
			"Provision OS": {
				Docs: []string{
					"Provisions a new OS",
				},
				Conditions: []string{},
				Dependencies: []string{
					"Device is SSHable",
					"Temp Boot into A partition",
					"Set permanent boot partition as A",
					"Download OS Image to device",
					"Flash B partitions with downloaded OS image",
					"Temp Boot into B partition",
					"Set permanent boot partition as B",
				},
				ExecName:   "sample_pass",
				RunControl: RunControl_ALWAYS_RUN,
			},
		},
	}
}
