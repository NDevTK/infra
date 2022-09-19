// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/internal/retry"
)

const (
	// The path to get the value of the Flag 2
	VPD_CACHE = `/mnt/stateful_partition/unencrypted/cache/vpd/full-v2.txt`
)

// isEnrollmentInCleanState confirms that the device's enrollment state is clean
//
// Verify that the device's enrollment state is clean.
//
// There are two "flags" that generate 3 possible enrollment states here.
// Flag 1 - The presence of install attributes file in
//
//	/home/.shadow/install_attributes.pb
//
// Flag 2 - The value of "check_enrollment" from VPD. Can be obtained by
//
//	reading the cache file in
//	/mnt/stateful_partition/unencrypted/cache/vpd/full-v2.txt
//
// The states:
// State 1 - Device is enrolled, means flag 1 is true and in flag 2 check_enrollment=1
// State 2 - Device is consumer owned, means flag 1 is true and in flag 2 check_enrollment=0
// State 3 - Device is enrolled and has been powerwashed, means flag 1 is
//
//	false. If the value in flag 2 is check_enrollment=1 then the
//	device will perform forced re-enrollment check and depending
//	on the response from the server might force the device to enroll
//	again. If the value is check_enrollment=0, then device can be
//	used like a new device.
//
// We consider state 1, and first scenario(check_enrollment=1) of state 3
// as unacceptable state here as they may interfere with normal tests.
func isEnrollmentInCleanStateExec(ctx context.Context, info *execs.ExecInfo) error {
	run := info.DefaultRunner()
	command := fmt.Sprintf(`grep "check_enrollment" %s`, VPD_CACHE)
	result, err := run(ctx, time.Minute, command)
	if err == nil {
		log.Debugf(ctx, "Enrollment state in VPD cache: %s", result)
		if result != `"check_enrollment"="0"` {
			return errors.Reason("enrollment in clean state: failed, The device is enrolled, it may interfere with some tests").Err()
		}
		return nil
	}
	// In any case it returns a non zero value, it means we can't verify enrollment state, but we cannot say the device is enrolled
	// Only trigger the enrollment in clean state when we can confirm the device is enrolled.
	log.Errorf(ctx, "Unexpected error occurred during verify enrollment state in VPD cache, skipping verify process.")
	return nil
}

// enrollmentCleanupExec cleans up the enrollment state on the
// ChromeOS device.
func enrollmentCleanupExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	run := info.NewRunner(info.GetDut().Name)
	// 1. Reset VPD enrollment state
	repairTimeout := argsMap.AsDuration(ctx, "repair_timeout", 120, time.Second)
	log.Debugf(ctx, "enrollment cleanup: using repair timeout :%s", repairTimeout)
	run(ctx, repairTimeout, "/usr/sbin/update_rw_vpd check_enrollment", "0")
	// 2. clear tpm owner state
	clearTpmOwnerTimeout := argsMap.AsDuration(ctx, "clear_tpm_owner_timeout", 60, time.Second)
	log.Debugf(ctx, "enrollment cleanup: using clear tpm owner timeout :%s", clearTpmOwnerTimeout)
	if _, err := run(ctx, clearTpmOwnerTimeout, "crossystem", "clear_tpm_owner_request=1"); err != nil {
		log.Debugf(ctx, "enrollment cleanup: unable to clear TPM.")
		return errors.Annotate(err, "enrollment cleanup").Err()
	}
	filesToRemove := []string{
		"/home/chronos/.oobe_completed",
		"/home/chronos/Local\\ State",
		"/var/cache/shill/default.profile",
	}
	dirsToRemove := []string{
		"/home/.shadow/*",
		filepath.Join("/var/cache/shill/default.profile", "*"),
		"/var/lib/whitelist/*", // nocheck
		"/var/cache/app_pack",
		"/var/lib/tpm",
	}
	// We do not care about any errors that might be returned by the
	// following two command executions.
	fileDeletionTimeout := argsMap.AsDuration(ctx, "file_deletion_timeout", 120, time.Second)
	run(ctx, fileDeletionTimeout, "sudo", "rm", "-rf", strings.Join(filesToRemove, " "), strings.Join(dirsToRemove, " "))
	run(ctx, fileDeletionTimeout, "sync")
	rebootTimeout := argsMap.AsDuration(ctx, "reboot_timeout", 10, time.Second)
	log.Debugf(ctx, "enrollment cleanup: using reboot timeout :%s", rebootTimeout)
	if err := SimpleReboot(ctx, run, rebootTimeout, info); err != nil {
		return errors.Annotate(err, "enrollment cleanup").Err()
	}
	// Finally, we will read the TPM status, and will check whether it
	// has been cleared or not.
	tpmTimeout := argsMap.AsDuration(ctx, "tpm_timeout", 150, time.Second)
	log.Debugf(ctx, "enrollment cleanup: using tpm timeout :%s", tpmTimeout)
	retry.WithTimeout(ctx, time.Second, tpmTimeout, func() error {
		tpmStatus := NewTpmStatus(ctx, run, repairTimeout)
		if tpmStatus.hasSuccess() {
			return nil
		}
		return errors.Reason("enrollment cleanup: failed to read TPM status.").Err()
	}, "wait to read tpm status")
	tpmStatus := NewTpmStatus(ctx, run, repairTimeout)
	isOwned, err := tpmStatus.isOwned()
	if err != nil {
		return errors.Reason("enrollment cleanup: failed to read TPM status.").Err()
	}
	if isOwned {
		return errors.Reason("enrollment cleanup: failed to clear TPM.").Err()
	}
	return nil
}

func init() {
	execs.Register("cros_is_enrollment_in_clean_state", isEnrollmentInCleanStateExec)
	execs.Register("cros_enrollment_cleanup", enrollmentCleanupExec)
}
