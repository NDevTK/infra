// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"strconv"
	"strings"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"
)

const (
	// moSysSkuCmd will retrieve the SKU label of the DUT.
	moSysSkuCmd = "mosys platform sku"
)

// updateDeviceSKUExec updates device's SKU label if not present in inventory
// or keep it the same if the args.DUT already has the value for SKU label.
func updateDeviceSKUExec(ctx context.Context, args *execs.RunArgs, actionArgs []string) error {
	r := args.NewRunner(args.ResourceName)
	skuLabelOutput, err := r(ctx, moSysSkuCmd)
	if err != nil {
		log.Debug(ctx, "Device sku label not found in the DUT.")
		return errors.Annotate(err, "update device sku label").Err()
	}
	args.DUT.DeviceSku = skuLabelOutput
	return nil
}

const (
	// gsctool version command that used to check the RW and RO version
	cr50VersionCmd = "gsctool -a -f"
)

// isCr50FirmwareExistExec checks if the cr 50 firmware exists on the DUT by running the gsctool version command.
func isCr50FirmwareExistExec(ctx context.Context, args *execs.RunArgs, actionArgs []string) error {
	r := args.NewRunner(args.ResourceName)
	if _, err := r(ctx, cr50VersionCmd); err != nil {
		return errors.Annotate(err, "cr50 fimrware exist").Err()
	}
	return nil
}

// updateCr50Exec will update the DUT's Cr50Phase state into the corresponding Cr50 state.
func updateCr50Exec(ctx context.Context, args *execs.RunArgs, actionArgs []string) error {
	r := args.NewRunner(args.ResourceName)
	rwVersion, err := cr50FWVersion(ctx, r, "RW")
	if err != nil {
		args.DUT.Cr50Phase = tlw.Cr50PhaseUnspecified
		return errors.Annotate(err, "update cr 50").Err()
	}
	rwVersionComponents := strings.Split(rwVersion, ".")
	if len(rwVersionComponents) < 2 {
		args.DUT.Cr50Phase = tlw.Cr50PhaseUnspecified
		return errors.Reason("update cr 50: the number of version component in the rw version is incorrect.").Err()
	}
	majorRwVersion, err := strconv.ParseInt(rwVersionComponents[1], 10, 64)
	if err != nil {
		args.DUT.Cr50Phase = tlw.Cr50PhaseUnspecified
		return errors.Annotate(err, "update cr 50").Err()
	}
	if majorRwVersion%2 != 0 {
		args.DUT.Cr50Phase = tlw.Cr50PhasePVT
		log.Info(ctx, "update DUT's Cr50 to be %s", tlw.Cr50PhasePVT)
	} else {
		args.DUT.Cr50Phase = tlw.Cr50PhasePREPVT
		log.Info(ctx, "update DUT's Cr50 to be %s", tlw.Cr50PhasePREPVT)
	}
	return nil
}

func init() {
	execs.Register("cros_update_device_sku", updateDeviceSKUExec)
	execs.Register("cros_is_cr_50_firmware_exist", isCr50FirmwareExistExec)
	execs.Register("cros_update_cr_50", updateCr50Exec)
}
