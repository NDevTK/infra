// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strconv"
	"time"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	"go.chromium.org/chromiumos/config/go/test/api"
	schedulingAPI "go.chromium.org/chromiumos/config/go/test/scheduling"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server/auth"

	"infra/device_manager/internal/model"
)

const (
	deviceEventsPubsubTopic string = "device-events-v1"
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

// UpdateDevice updates a Device in a transaction.
func UpdateDevice(ctx context.Context, tx *sql.Tx, device model.Device, cloudProject string) error {
	err := model.UpdateDevice(ctx, tx, device)
	if err != nil {
		logging.Errorf(ctx, "UpdateDevice: failed to update Device %s: %s", device.ID, err)
		return err
	}
	logging.Debugf(ctx, "UpdateDevice: updated Device %s successfully", device.ID)

	// Send message to PubSub Device events stream
	logging.Debugf(ctx, "getting rpc credentials")
	creds, err := auth.GetPerRPCCredentials(ctx, auth.AsSelf, auth.WithScopes(auth.CloudOAuthScopes...))
	if err != nil {
		return errors.Annotate(err, "UpdateDevice: failed to get AsSelf credentails").Err()
	}
	client, err := pubsub.NewClient(
		ctx, cloudProject,
		option.WithGRPCDialOption(grpc.WithPerRPCCredentials(creds)),
	)
	if err != nil {
		logging.Errorf(ctx, "UpdateDevice: cannot set up PubSub client: %s", err)
		return err
	}

	logging.Debugf(ctx, "checking topic existence")
	topic := client.Topic(deviceEventsPubsubTopic)
	ok, err := topic.Exists(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("UpdateDevice: topic %s not found", deviceEventsPubsubTopic)
	}

	logging.Debugf(ctx, "marshaling message")
	var msg []byte
	msg, err = proto.Marshal(&schedulingAPI.DeviceEvent{
		EventTime:   time.Now().Unix(),
		DeviceId:    device.ID,
		DeviceReady: false,
	})
	if err != nil {
		return fmt.Errorf("proto.Marshal err: %w", err)
	}

	rsp := topic.Publish(ctx, &pubsub.Message{
		Data: msg,
	})

	_, err = rsp.Get(ctx)
	if err != nil {
		logging.Debugf(ctx, "UpdateDevice: failed to publish to PubSub %s", err)
	}

	topic.Stop()

	return nil
}

// IsDeviceAvailable checks if a device state is available.
func IsDeviceAvailable(ctx context.Context, state api.DeviceState) bool {
	return state == api.DeviceState_DEVICE_STATE_AVAILABLE
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

// convertAPIDeviceAddressToDBFormat takes a DeviceAddress and converts it to string.
//
// The format is defined by the DeviceAddress proto. It does a basic join of
// Host and Port using the net package.
func convertAPIDeviceAddressToDBFormat(ctx context.Context, addr *api.DeviceAddress) string {
	return net.JoinHostPort(addr.GetHost(), fmt.Sprint(addr.GetPort()))
}

// convertDeviceTypeToAPIFormat takes a string and converts it to DeviceType.
func convertDeviceTypeToAPIFormat(ctx context.Context, deviceType string) api.DeviceType {
	return api.DeviceType(api.DeviceType_value[deviceType])
}

// convertDeviceStateToAPIFormat takes a string and converts it to DeviceState.
func convertDeviceStateToAPIFormat(ctx context.Context, state string) api.DeviceState {
	return api.DeviceState(api.DeviceState_value[state])
}
