// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package external

import (
	"context"

	ufsdevice "infra/unifiedfleet/api/v1/models/chromeos/device"
)

// DeviceConfigClient handles read operations for DeviceConfigs
// These functions match the functions in the configuration package, and if
// datastore is the only source for DeviceConfigs, consider deleting this code
type DeviceConfigClient interface {
	GetDeviceConfig(ctx context.Context, cfgID *ufsdevice.ConfigId) (*ufsdevice.Config, error)
	DeviceConfigsExists(ctx context.Context, cfgIDs []*ufsdevice.ConfigId) ([]bool, error)
}

// DualDeviceConfigClient uses both inventory and UFS data sources to fetch
// device configs. If it is able to detect a device config in either data
// source, it treat it as existing.
type DualDeviceConfigClient struct{}

// GetDeviceConfig fetches a specific device config.
func (c *DualDeviceConfigClient) GetDeviceConfig(ctx context.Context, cfgID *ufsdevice.ConfigId) (*ufsdevice.Config, error) {
	return nil, nil
}

// DeviceConfigsExists detects whether any number of configs exist. The return
// is an array of booleans, where the ith boolean represents the existence of
// the ith config.
func (c *DualDeviceConfigClient) DeviceConfigsExists(ctx context.Context, cfgIDs []*ufsdevice.ConfigId) ([]bool, error) {
	return nil, nil
}
