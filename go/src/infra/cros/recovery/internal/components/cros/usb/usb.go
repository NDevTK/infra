// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package usb

import (
	"context"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components"
	"infra/cros/recovery/logger"
)

// UsbReset resets USB devices. usbId is one of the following
// - PPPP:VVVV - product and vendor id
// - BBB/DDD   - bus and device number
// - "Product" - product name
func UsbReset(ctx context.Context, run components.Runner, log logger.Logger, usbId string) error {
	usbResetCmd := "usbreset " + usbId
	_, err := run(ctx, time.Minute, usbResetCmd)
	if err != nil {
		return errors.Annotate(err, "usb reset").Err()
	}
	log.Debugf("USB is successfully reset: %s", usbId)
	return nil
}
