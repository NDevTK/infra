// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package fake

import (
	"context"

	ufsdevice "infra/unifiedfleet/api/v1/models/chromeos/device"
)

// DeviceConfigClient is a fake impl for testing
type DeviceConfigClient struct {
}

// GetDeviceConfig fetches a specific device config.
func (c *DeviceConfigClient) GetDeviceConfig(ctx context.Context, cfgID *ufsdevice.ConfigId) (*ufsdevice.Config, error) {
	return nil, nil
}

// DeviceConfigsExists detects whether any number of configs exist. The return
// is an array of booleans, where the ith boolean represents the existence of
// the ith config.
func (c *DeviceConfigClient) DeviceConfigsExists(ctx context.Context, cfgIDs []*ufsdevice.ConfigId) ([]bool, error) {
	return nil, nil
}
