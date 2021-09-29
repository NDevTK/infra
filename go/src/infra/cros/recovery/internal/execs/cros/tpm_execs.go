// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
)

const (
	// Expected value of tpm dev-signed firmware and kernel version.
	devTpmFirmwareAndKernelVersion = "0x00010001"
)

// isOnDevTPMKernelVersionExec verifies dev's tpm kernel version is match to expected value.
//
// For dev-signed firmware, tpm_kernver reported from
// crossystem should always be 0x10001. Firmware update on DUTs with
// incorrect tpm_kernver may fail due to firmware rollback protection.
func matchDevTPMKernelVersionExec(ctx context.Context, args *execs.RunArgs, actionArgs []string) error {
	if err := matchCrosSystemValueToExpectation(ctx, args, "tpm_kernver", devTpmFirmwareAndKernelVersion); err != nil {
		return errors.Annotate(err, "match dev tpm kernel version: dev tpm kernel version mismatch").Err()
	}
	return nil
}

// matchDevTPMFirmwareVersionExec verifies dev's tpm firmware version is match to expected value.
//
// For dev-signed firmware, tpm_fwver reported from
// crossystem should always be 0x10001. Firmware update on DUTs with
// incorrect tmp_fwver may fail due to firmware rollback protection.
func matchDevTPMFirmwareVersionExec(ctx context.Context, args *execs.RunArgs, actionArgs []string) error {
	if err := matchCrosSystemValueToExpectation(ctx, args, "tpm_fwver", devTpmFirmwareAndKernelVersion); err != nil {
		return errors.Annotate(err, "match dev tpm firmware version: dev tpm firmware version mismatch").Err()
	}
	return nil
}

func init() {
	execs.Register("cros_match_dev_tpm_firmware_version", matchDevTPMFirmwareVersionExec)
	execs.Register("cros_match_dev_tpm_kernel_version", matchDevTPMKernelVersionExec)
}
