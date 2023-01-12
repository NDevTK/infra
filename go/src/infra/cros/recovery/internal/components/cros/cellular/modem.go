// Copyright 2023 The ChromiumOS Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cellular

import (
	"context"
	"encoding/json"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components"
	"infra/cros/recovery/internal/retry"
)

const (
	detectCmd   = "mmcli -m a -J"
	expectedCmd = "cros_config /modem firmware-variant"
)

// IsExpected returns true if cellular modem is expected to exist on the DUT.
func IsExpected(ctx context.Context, runner components.Runner) bool {
	if _, err := runner(ctx, 5*time.Second, expectedCmd); err != nil {
		return false
	}
	return true
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
