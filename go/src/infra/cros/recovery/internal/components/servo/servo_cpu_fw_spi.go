// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package servo

import (
	"context"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components"
)

// CpuFwSpiState specifies value to set for cpu_fw_spi or ccd_cpu_fw_spi.
type CpuFwSpiState string

const (
	// Value for set cpu_fw_spi on.
	CpuFwSpiValueON CpuFwSpiState = "on"
	// Value to set cpu_fw_spi off.
	CpuFwSpiValueOFF CpuFwSpiState = "off"
)

// SetCpuFwSpiState enable/disable SPI mode via servo.
func SetCpuFwSpiState(ctx context.Context, servod components.Servod, deviceType string, val CpuFwSpiState) error {
	servodControlMap := map[string]string{
		"servo_micro": "cpu_fw_spi",
		"ccd_gsc":     "ccd_cpu_fw_spi",
		"ccd_cr50":    "ccd_cpu_fw_spi",
	}
	if _, ok := servodControlMap[deviceType]; !ok {
		return errors.Reason("set cpu fw spi state: unsupported device type: %s", deviceType).Err()
	}
	if err := servod.Set(ctx, servodControlMap[deviceType], string(val)); err != nil {
		return errors.Annotate(err, "set cup fw spi %q", val).Err()
	}
	return nil
}
