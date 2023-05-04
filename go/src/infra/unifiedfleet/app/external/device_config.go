// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package external

import (
	"context"

	"go.chromium.org/chromiumos/infra/proto/go/device"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	invV2Api "infra/appengine/cros/lab_inventory/api/v1"
	ufsdevice "infra/unifiedfleet/api/v1/models/chromeos/device"
	"infra/unifiedfleet/app/model/configuration"
)

// DeviceConfigClient handles read operations for DeviceConfigs
// These functions match the functions in the configuration package, and if
// datastore is the only source for DeviceConfigs, consider deleting this code
type DeviceConfigClient interface {
	GetDeviceConfig(ctx context.Context, cfgID *ufsdevice.ConfigId) (*ufsdevice.Config, error)
	DeviceConfigsExists(ctx context.Context, cfgIDs []*ufsdevice.ConfigId) ([]bool, error)
}

// InventoryDeviceConfigClient exposes methods needed to read from inventory.
// This is used when we dual read from inventory and UFS sources.
type InventoryDeviceConfigClient interface {
	DeviceConfigsExists(ctx context.Context, in *invV2Api.DeviceConfigsExistsRequest, opts ...grpc.CallOption) (*invV2Api.DeviceConfigsExistsResponse, error)
	GetDeviceConfig(ctx context.Context, in *invV2Api.GetDeviceConfigRequest, opts ...grpc.CallOption) (*device.Config, error)
}

// DualDeviceConfigClient uses both inventory and UFS data sources to fetch
// device configs. If it is able to detect a device config in either data
// source, it treat it as existing.
type DualDeviceConfigClient struct {
	inventoryClient InventoryDeviceConfigClient
}

// GetDeviceConfig fetches a specific device config.
func (c *DualDeviceConfigClient) GetDeviceConfig(ctx context.Context, cfgID *ufsdevice.ConfigId) (*ufsdevice.Config, error) {
	crosCfgID, err := ufsToCrosCfgIDProto(cfgID)
	if err != nil {
		return nil, errors.Annotate(err, "failed to convert between ufs and inventory proto, likely proto versions are out of sync").Err()
	}

	req := &invV2Api.GetDeviceConfigRequest{
		ConfigId: crosCfgID,
	}

	resp, err := c.inventoryClient.GetDeviceConfig(ctx, req)
	if err != nil || resp == nil {
		logging.Debugf(ctx, "request for cfg: %v was not found with error: %s. falling back to ufs datastore", cfgID, err)

		return configuration.GetDeviceConfig(ctx, cfgID)
	}

	return crosToUFSDeviceConfigProto(resp)
}

// DeviceConfigsExists detects whether any number of configs exist. The return
// is an array of booleans, where the ith boolean represents the existence of
// the ith config.
func (c *DualDeviceConfigClient) DeviceConfigsExists(ctx context.Context, cfgIDs []*ufsdevice.ConfigId) ([]bool, error) {
	return nil, nil
}

// ufsToCrosCfgIDProto naively marshalls then unmarshalls proto to convert
// between the format cros inventory expects and the format UFS expects. This
// is dangerous and can break when the protos are out of sync.
func ufsToCrosCfgIDProto(ufsCfgID *ufsdevice.ConfigId) (*device.ConfigId, error) {
	s, err := proto.Marshal(ufsCfgID)
	if err != nil {
		return nil, err
	}
	var crosCfgID device.ConfigId
	err = proto.Unmarshal(s, &crosCfgID)

	return &crosCfgID, err
}

// crosToUFSDeviceConfigProto naively marshalls then unmarshalls proto to convert
// between the format cros inventory expects and the format UFS expects. This
// is dangerous and can break when the protos are out of sync.
func crosToUFSDeviceConfigProto(crosCfgID *device.Config) (*ufsdevice.Config, error) {
	s, err := proto.Marshal(crosCfgID)
	if err != nil {
		return nil, err
	}
	var ufsCfgID ufsdevice.Config
	err = proto.Unmarshal(s, &ufsCfgID)

	return &ufsCfgID, err
}
