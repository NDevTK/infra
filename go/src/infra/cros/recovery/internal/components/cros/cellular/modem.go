// Copyright 2023 The ChromiumOS Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cellular

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components"
	"infra/cros/recovery/internal/retry"
)

const (
	modemManagerJob           = "modemmanager"
	detectCmd                 = "mmcli -m a -J"
	expectedCmd               = "cros_config /modem firmware-variant"
	mmcliCliPresentCmd        = "which mmcli"
	modemManagerJobPresentCmd = "initctl status modemmanager"
	restartModemManagerCmd    = "restart modemmanager"
	startModemManagerCmd      = "start modemmanager"
)

// IsExpected returns true if cellular modem is expected to exist on the DUT.
func IsExpected(ctx context.Context, runner components.Runner) bool {
	if _, err := runner(ctx, 5*time.Second, expectedCmd); err != nil {
		return false
	}
	return true
}

// HasModemManagerCLI returns true if mmcli is present on the DUT.
func HasModemManagerCLI(ctx context.Context, runner components.Runner, timeout time.Duration) bool {
	if _, err := runner(ctx, timeout, mmcliCliPresentCmd); err != nil {
		return false
	}
	return true
}

// HasModemManagerJob returns true if modemmanager job is present on the DUT.
func HasModemManagerJob(ctx context.Context, runner components.Runner, timeout time.Duration) bool {
	if _, err := runner(ctx, timeout, modemManagerJobPresentCmd); err != nil {
		return false
	}
	return true
}

// StartModemManager starts modemmanager via upstart.
func StartModemManager(ctx context.Context, runner components.Runner, timeout time.Duration) error {
	if _, err := runner(ctx, timeout, startModemManagerCmd); err != nil {
		return errors.Annotate(err, "start modemmanager").Err()
	}
	return nil
}

// RestartModemManager restarts modemmanager via upstart.
func RestartModemManager(ctx context.Context, runner components.Runner, timeout time.Duration) error {
	if _, err := runner(ctx, timeout, restartModemManagerCmd); err != nil {
		return errors.Annotate(err, "restart modemmanager").Err()
	}
	return nil
}

// WaitForModemManager waits for the modemmanager job to be running via upstart.
func WaitForModemManager(ctx context.Context, runner components.Runner, timeout time.Duration) error {
	cmd := fmt.Sprintf("status %s", modemManagerJob)
	return retry.WithTimeout(ctx, time.Second, timeout, func() error {
		if output, err := runner(ctx, 5*time.Second, cmd); err != nil {
			return errors.Annotate(err, "get modemmanager status").Err()
		} else if !strings.Contains(output, "start/running") {
			return errors.Reason("modemmanager not running").Err()
		}
		return nil
	}, "wait for modemmanager")
}

// ModemInfo is a simplified version of the JSON output from ModemManager to get the modem connection state information.
type ModemInfo struct {
	Modem *struct {
		Generic *struct {
			State string `state:"callbox,omitempty"`
		} `json:"generic,omitempty"`
	} `modem:"modem,omitempty"`
}

// WaitForModemInfo polls for a modem to appear on the DUT, which can take up to two minutes on reboot.
func WaitForModemInfo(ctx context.Context, runner components.Runner, timeout time.Duration) (*ModemInfo, error) {
	var info *ModemInfo
	if err := retry.WithTimeout(ctx, time.Second, timeout, func() error {
		output, err := runner(ctx, 5*time.Second, detectCmd)
		if err != nil {
			return errors.Annotate(err, "call mmcli").Err()
		}

		// Note: info is defined in outer scope as retry.WithTimeout only allows returning errors.
		info, err = parseModemInfo(ctx, output)
		if err != nil {
			return errors.Annotate(err, "parse mmcli response").Err()
		}

		if info == nil || info.Modem == nil {
			return errors.Reason("no modem found on DUT").Err()
		}

		return nil
	}, "wait for modem"); err != nil {
		return nil, errors.Annotate(err, "wait for modem info: wait for ModemManager to export modem").Err()
	}

	return info, nil
}

// parseModemInfo unmarshals the modem properties json output from mmcli.
func parseModemInfo(ctx context.Context, output string) (*ModemInfo, error) {
	info := &ModemInfo{}
	if err := json.Unmarshal([]byte(output), info); err != nil {
		return nil, err
	}
	return info, nil
}
