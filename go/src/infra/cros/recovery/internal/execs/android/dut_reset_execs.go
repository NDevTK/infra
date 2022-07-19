// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package android

import (
	"context"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components/adb"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/internal/retry"
)

// resetDutExec resets DUT.
func resetDutExec(ctx context.Context, info *execs.ExecInfo) error {
	serialNumber := info.GetAndroid().GetSerialNumber()
	err := adb.EnableDeviceTestHarnessMode(ctx, newRunner(info), info.NewLogger(), serialNumber)
	if err != nil {
		return errors.Annotate(err, "dut reset").Err()
	}
	return nil
}

// waitTillDutOfflineExec waits till DUT disconnects from associated host.
func waitTillDutOfflineExec(ctx context.Context, info *execs.ExecInfo) error {
	serialNumber := info.GetAndroid().GetSerialNumber()
	argsMap := info.GetActionArgs(ctx)
	retryCount := argsMap.AsInt(ctx, "retry_count", 10)
	waitInRetry := argsMap.AsDuration(ctx, "wait_in_retry", 1, time.Second)
	run := newRunner(info)
	logger := info.NewLogger()
	log.Debugf(ctx, "Waiting till attached device %q offline: retry_count=%d, wait_in_retry=%s", serialNumber, retryCount, waitInRetry)
	return retry.LimitCount(ctx, retryCount, waitInRetry, func() error {
		if adb.IsDeviceAccessible(ctx, run, logger, serialNumber) == nil {
			return errors.Reason("wait for offline dut failed: %q.", serialNumber).Err()
		}
		return nil
	}, "attached DUT offline")
}

// waitTillDutOnlineExec waits till DUT online.
func waitTillDutOnlineExec(ctx context.Context, info *execs.ExecInfo) error {
	serialNumber := info.GetAndroid().GetSerialNumber()
	err := adb.WaitForDevice(ctx, newRunner(info), info.NewLogger(), serialNumber)
	if err != nil {
		return errors.Annotate(err, "wait for online dut").Err()
	}
	return nil
}

func init() {
	execs.Register("android_dut_reset", resetDutExec)
	execs.Register("android_wait_for_offline_dut", waitTillDutOfflineExec)
	execs.Register("android_wait_for_online_dut", waitTillDutOnlineExec)
}
