// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/execs/cros/amt"
)

// flexAMTPresent returns true if Intel AMT (vPro) is present.
func flexAMTPresent(ctx context.Context, info *execs.ExecInfo) error {
	client := getFlexAMTClient()
	present, err := client.AMTPresent()
	if err != nil {
		return errors.Annotate(err, "flex AMT present").Err()
	}
	if !present {
		return errors.Reason("flex AMT present: not found").Err()
	}
	return nil
}

// flexAMTPowerOff powers the DUT off using Intel AMT (vPro).
func flexAMTPowerOff(ctx context.Context, info *execs.ExecInfo) error {
	client := getFlexAMTClient()
	return errors.Annotate(client.PowerOff(), "flex AMT power-off").Err()
}

// flexAMTPowerOn powers the DUT on using Intel AMT (vPro).
func flexAMTPowerOn(ctx context.Context, info *execs.ExecInfo) error {
	client := getFlexAMTClient()
	return errors.Annotate(client.PowerOn(), "flex AMT power-off").Err()
}

// Configure and return an AMTClient.
func getFlexAMTClient() amt.AMTClient {
	//TODO(josephsussman): Get these from somewhere else.
	return amt.NewAMTClient("192.168.231.218", "admin", "P@ssword1")
}

func init() {
	execs.Register("cros_flex_amt_present", flexAMTPresent)
	execs.Register("cros_flex_amt_power_off", flexAMTPowerOff)
	execs.Register("cros_flex_amt_power_on", flexAMTPowerOn)
}
