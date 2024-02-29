// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"database/sql"
	"net"
	"strconv"

	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"

	"infra/device_manager/internal/model"
)

// GetDevice gets a Device from the database based on deviceName.
func GetDevice(ctx context.Context, db *sql.DB, deviceName string) (*api.Device, error) {
	device, err := model.GetDeviceByName(ctx, db, deviceName)
	if err != nil {
		return &api.Device{}, err
	}

	addr, err := convertDeviceAddressToAPIFormat(ctx, device.DeviceAddress)
	if err != nil {
		logging.Errorf(ctx, err.Error())
		addr = &api.DeviceAddress{}
	}

	deviceProto := &api.Device{
		Id:      device.ID,
		Address: addr,
		Type:    convertDeviceTypeToAPIFormat(ctx, device.DeviceType),
		State:   convertDeviceStateToAPIFormat(ctx, device.DeviceState),
	}
	return deviceProto, nil
}

// convertDeviceAddressToAPIFormat takes a net address string and converts it.
//
// The format is defined by the DeviceAddress proto. It does a basic split of
// Host and Port and uses the net package. This package supports IPv4 and IPv6.
func convertDeviceAddressToAPIFormat(ctx context.Context, addr string) (*api.DeviceAddress, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return &api.DeviceAddress{}, errors.Annotate(err, "failed to split host and port %s", addr).Err()
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return &api.DeviceAddress{}, errors.Annotate(err, "port %s is not convertible to integer", portStr).Err()
	}

	return &api.DeviceAddress{
		Host: host,
		Port: int32(port),
	}, nil
}

// convertDeviceTypeToAPIFormat takes a string and converts it to DeviceType.
func convertDeviceTypeToAPIFormat(ctx context.Context, deviceType string) api.DeviceType {
	return api.DeviceType(api.DeviceType_value[deviceType])
}

// convertDeviceStateToAPIFormat takes a string and converts it to DeviceState.
func convertDeviceStateToAPIFormat(ctx context.Context, state string) api.DeviceState {
	return api.DeviceState(api.DeviceState_value[state])
}
