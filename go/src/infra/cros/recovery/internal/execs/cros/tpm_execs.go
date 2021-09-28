// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"fmt"
	"strings"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
)

const (
	// Expected value of tpm dev-signed firmware and kernel version.
	devTpmFirmwareAndKernalVersion = "0x00010001"
)

// isOnDevTPMVersion is a helper function for both isOnDevTPMKernelVersionExec and isOnDevTPMFirmwareVersionExec.
func isOnDevTPMVersion(ctx context.Context, args *execs.RunArgs, actionArgs []string, tpmField string) error {
	r := args.Access.Run(ctx, args.ResourceName, fmt.Sprintf("crossystem tpm_%sver", tpmField))
	if r.ExitCode != 0 {
		return errors.Reason(`on dev tpm %s version: unable to get "tpm_%sver" from crossystem: failed with code: %d, %q`, tpmField, tpmField, r.ExitCode, r.Stderr).Err()
	}
	actualTpmVersion := strings.TrimSpace(r.Stdout)
	if actualTpmVersion != devTpmFirmwareAndKernalVersion {
		return errors.Reason(`on dev tpm %s version: unexpected "tpm_%sver" value: %s, expected: %s. This error may cause firmware provision fail due to the rollback protection`, tpmField, tpmField, actualTpmVersion, devTpmFirmwareAndKernalVersion).Err()
	}
	return nil
}

// isOnDevTPMKernelVersionExec confirms dut is operating dev's tpm kernel version.
//
// For dev-signed firmware, tpm_kernver reported from
// crossystem should always be 0x10001. Firmware update on DUTs with
// incorrect tpm_kernver may fail due to firmware rollback protection.
func isOnDevTPMKernelVersionExec(ctx context.Context, args *execs.RunArgs, actionArgs []string) error {
	return isOnDevTPMVersion(ctx, args, actionArgs, "kern")
}

// isOnDevTPMFirmwareVersionExec confirms dut is operating dev's tpm firmware version.
//
// For dev-signed firmware, tpm_fwver reported from
// crossystem should always be 0x10001. Firmware update on DUTs with
// incorrect tmp_fwver may fail due to firmware rollback protection.
func isOnDevTPMFirmwareVersionExec(ctx context.Context, args *execs.RunArgs, actionArgs []string) error {
	return isOnDevTPMVersion(ctx, args, actionArgs, "fw")
}

func init() {
	execs.Register("cros_is_on_dev_tpm_firmware_version", isOnDevTPMFirmwareVersionExec)
	execs.Register("cros_is_on_dev_tpm_kernel_version", isOnDevTPMKernelVersionExec)
}
