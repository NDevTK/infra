// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"
)

const (
	// moSysCmd will retrieve the sku label of the DUT.
	moSysSkuCmd = "mosys platform sku"
)

// updateDeviceSkuLabelExec will update the DUT's SKU label from either inventory
// or keep it the same if the args.DUT already has the value for SKU label.
func updateDeviceSkuLabelExec(ctx context.Context, args *execs.RunArgs, actionArgs []string) error {
	deviceSkuLabel := args.DUT.DeviceSku
	if deviceSkuLabel != "" {
		log.Info(ctx, "DUT's device sku label found in inventory: %s", deviceSkuLabel)
		return nil
	}
	r := args.NewRunner(args.ResourceName)
	skuLabelOutput, err := r(ctx, moSysSkuCmd)
	if err != nil {
		log.Debug(ctx, "Device sku label not found in the DUT.")
		return errors.Annotate(err, "update device sku label").Err()
	}
	args.DUT.DeviceSku = tlw.DeviceSku(skuLabelOutput)
	return nil
}

func init() {
	execs.Register("cros_update_device_sku_label", updateDeviceSkuLabelExec)
}
