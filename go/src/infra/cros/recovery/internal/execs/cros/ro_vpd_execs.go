// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"fmt"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
)

// ROVPDKeys is the list of keys from RO_VPD that should be persisted and flashed in the recovery process.
var ROVPDKeys = []string{"wifi_sar"}

const (
	// readROVPDValuesCmdGlob reads the value of VPD key from RO_VPD partition by name.
	readROVPDValuesCmd = "vpd -i RO_VPD -g %s"
	writeVPDValuesCmd  = "vpd -s %s=%s"
)

// isROVPDSkuNumberRequired confirms that this device is required to have sku_number in RO_VPD.
func isROVPDSkuNumberRequired(ctx context.Context, info *execs.ExecInfo) error {
	r := info.DefaultRunner()
	hasSkuNumber, err := r(ctx, time.Minute, "cros_config /cros-healthd/cached-vpd has-sku-number")
	if err != nil {
		return errors.Annotate(err, "is RO_VPD sku_number required").Err()
	}
	// If has-sku-number is not true (e.g. unspecified or false), sku_number is not required.
	if hasSkuNumber != "true" {
		return errors.Reason("sku_number is not required in RO_VPD.").Err()
	}
	return nil
}

// verifyROVPDSkuNumber confirms that the key 'sku_number' is present in RO_VPD.
func verifyROVPDSkuNumber(ctx context.Context, info *execs.ExecInfo) error {
	r := info.DefaultRunner()
	if _, err := r(ctx, time.Minute, "vpd -i RO_VPD -g sku_number"); err != nil {
		return errors.Annotate(err, "verify sku_number in RO_VPD").Err()
	}
	log.Infof(ctx, "sku_number is present in RO_VPD")
	return nil
}

// setFakeROVPDSkuNumber sets a fake sku_number in RO_VPD.
//
// @params: actionArgs should be in the format:
// ["sku_number:FAKE-SKU"]
func setFakeROVPDSkuNumber(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	if !argsMap.Has("sku_number") {
		return errors.Reason("set fake sku_number: fake value is not specified.").Err()
	}
	skuNumber := argsMap.AsString(ctx, "sku_number", "")

	r := info.DefaultRunner()
	cmd := fmt.Sprintf(writeVPDValuesCmd, "sku_number", skuNumber)
	if _, err := r(ctx, time.Minute, cmd); err != nil {
		return errors.Annotate(err, "cannot set fake sku_number").Err()
	}
	log.Infof(ctx, "set fake sku_number successfully")
	return nil
}

// updateROVPDToInv reads RO_VPD values from ROVPDKeys from the resource and updates DUT info.
func updateROVPDToInv(ctx context.Context, info *execs.ExecInfo) error {
	r := info.DefaultRunner()
	if info.GetChromeos().RoVpdMap == nil {
		info.GetChromeos().RoVpdMap = make(map[string]string)
	}
	for _, key := range ROVPDKeys {
		cmd := fmt.Sprintf(readROVPDValuesCmd, key)
		if value, err := r(ctx, time.Minute, cmd); err == nil {
			info.GetChromeos().RoVpdMap[key] = value
		}
	}
	log.Infof(ctx, "recorded RO_VPD values successfully")
	return nil
}

// matchROVPDToInv matches RO_VPD values from ROVPDKeys from the resource to values in the Inventory.
func matchROVPDToInv(ctx context.Context, info *execs.ExecInfo) error {
	r := info.DefaultRunner()
	for k, v := range info.GetChromeos().RoVpdMap {
		cmd := fmt.Sprintf(readROVPDValuesCmd, k)
		value, err := r(ctx, time.Minute, cmd)
		if err != nil {
			return errors.Annotate(err, "cannot read RO_VPD key").Err()
		}
		if value != v {
			return errors.Annotate(err, "RO_VPD had a bad value").Err()
		}
	}
	log.Infof(ctx, "RO_VPD values are correct")
	return nil
}

// setROVPD sets RO_VPD values from ROVPDKeys from the Inventory to the resource.
func setROVPD(ctx context.Context, info *execs.ExecInfo) error {
	r := info.DefaultRunner()
	for k, v := range info.GetChromeos().RoVpdMap {
		cmd := fmt.Sprintf(writeVPDValuesCmd, k, v)
		if _, err := r(ctx, time.Minute, cmd); err != nil {
			return errors.Annotate(err, "cannot set RO_VPD key").Err()
		}

	}
	log.Infof(ctx, "set RO_VPD values successfully")
	return nil
}

func init() {
	execs.Register("cros_is_ro_vpd_sku_number_required", isROVPDSkuNumberRequired)
	execs.Register("cros_verify_ro_vpd_sku_number", verifyROVPDSkuNumber)
	execs.Register("cros_set_fake_ro_vpd_sku_number", setFakeROVPDSkuNumber)
	execs.Register("cros_update_ro_vpd_inventory", updateROVPDToInv)
	execs.Register("cros_match_ro_vpd_inventory", matchROVPDToInv)
	execs.Register("cros_set_ro_vpd", setROVPD)
}
