// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"
)

const (
	// command to check whether the wifi device has been recogonized
	// and its device driver been loaded by the kernel.
	wifiDetectCmd = `lsmod | grep -E iwl\|rtw\|mt792\|ath\|mwifiex`
)

// auditWiFiExec will validate wifi chip and update state.
//
// Detect if the DUT has wifi device listed in the output of 'lspci' command.
func auditWiFiExec(ctx context.Context, info *execs.ExecInfo) error {
	r := info.DefaultRunner()
	wifi := info.GetChromeos().GetWifi()
	if wifi == nil {
		return errors.Reason("audit wifi: data is not present in dut info").Err()
	}
	_, err := r(ctx, time.Minute, wifiDetectCmd)
	if err == nil {
		// successfully detected
		wifi.State = tlw.HardwareState_HARDWARE_NORMAL
		log.Infof(ctx, "set wifi state to be: %s", tlw.HardwareState_HARDWARE_NORMAL)
		return nil
	}
	if execs.SSHErrorInternal.In(err) || execs.SSHErrorCLINotFound.In(err) {
		wifi.State = tlw.HardwareState_HARDWARE_UNSPECIFIED
		return errors.Annotate(err, "audit wifi").Err()
	}
	if wifi.GetChipName() != "" {
		// If wifi chip is not detected, but was expected by setup info then we
		// set needs_replacement as it is probably a hardware issue.
		wifi.State = tlw.HardwareState_HARDWARE_NEED_REPLACEMENT
	} else {
		// the wifi state cannot be determined due to cmd failed
		// therefore, set it to HardwareStateNotDetected
		wifi.State = tlw.HardwareState_HARDWARE_NOT_DETECTED
	}
	log.Infof(ctx, "set wifi state to be: %s", wifi.State)
	return errors.Annotate(err, "audit wifi").Err()
}

func init() {
	execs.Register("cros_audit_wifi", auditWiFiExec)
}
