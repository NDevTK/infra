// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package adb

import (
	"context"
	"fmt"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components"
	"infra/cros/recovery/logger"
)

// EnableWiFi enables WiFi on DUT.
func EnableWiFi(ctx context.Context, run components.Runner, log logger.Logger, serialNumber string) error {
	const adbEnableWiFiCmd = "adb -s %s shell svc wifi enable"
	cmd := fmt.Sprintf(adbEnableWiFiCmd, serialNumber)
	_, err := run(ctx, time.Minute, cmd)
	if err != nil {
		return errors.Annotate(err, "enable wifi").Err()
	}
	log.Debugf("wifi is enabled on the device: %q", serialNumber)
	return nil
}

// ConnectToWiFiNetwork connects DUT to WiFi network.
func ConnectToWiFiNetwork(ctx context.Context, run components.Runner, log logger.Logger, serialNumber string, wifiSSID string, wifiSecurityType string, wifiPassword string) error {
	const adbConnectNetworkCmd = "adb -s %s shell cmd wifi connect-network"
	cmd := fmt.Sprintf(adbConnectNetworkCmd, serialNumber)
	_, err := run(ctx, time.Minute, cmd, wifiSSID, wifiSecurityType, wifiPassword)
	if err != nil {
		return errors.Annotate(err, "connect wifi network").Err()
	}
	log.Debugf("device %q is connected to wifi network", serialNumber)
	return nil
}
