// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package inventory

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/retry"
	"go.chromium.org/luci/gae/service/info"
	"go.chromium.org/luci/grpc/prpc"
	"go.chromium.org/luci/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	api "infra/appengine/cros/lab_inventory/api/v1"
	invV1Api "infra/appengine/crosskylabadmin/api/fleet/v1"
)

type invServiceClient struct {
	client api.InventoryClient
}

func newInvServiceClient(ctx context.Context, host string) (inventoryClient, error) {
	t, err := auth.GetRPCTransport(ctx, auth.AsCredentialsForwarder)
	if err != nil {
		return nil, errors.Annotate(err, "failed to get RPC transport to inventory service").Err()
	}
	ic := api.NewInventoryPRPCClient(&prpc.Client{
		C:    &http.Client{Transport: t},
		Host: host,
		Options: &prpc.Options{
			UserAgent: fmt.Sprintf("%s/%s", info.AppID(ctx), info.VersionID(ctx)),
		},
	})

	return &invServiceClient{client: ic}, nil
}

func (c *invServiceClient) logInfo(ctx context.Context, t string, s ...interface{}) {
	logging.Infof(ctx, fmt.Sprintf("[InventoryV2Client]: %s", t), s...)
}

func (c *invServiceClient) getDutInfo(ctx context.Context, req *invV1Api.GetDutInfoRequest) ([]byte, time.Time, error) {
	var devID *api.DeviceID
	now := time.Now().UTC()
	if req.Id != "" {
		devID = &api.DeviceID{Id: &api.DeviceID_ChromeosDeviceId{ChromeosDeviceId: req.Id}}
	} else {
		devID = &api.DeviceID{Id: &api.DeviceID_Hostname{Hostname: req.Hostname}}
	}
	var rsp *api.GetCrosDevicesResponse
	f := func() (err error) {
		rsp, err = c.client.GetCrosDevices(ctx, &api.GetCrosDevicesRequest{
			Ids: []*api.DeviceID{devID},
		})
		return err
	}
	err := retry.Retry(ctx, transientErrorRetries(), f, retry.LogCallback(ctx, "GetDutInfo v2"))
	if err != nil {
		return nil, now, err
	}
	if devices := rsp.GetData(); len(devices) == 1 {
		dut, err := api.AdaptToV1DutSpec(devices[0])
		if err != nil {
			return nil, now, errors.Annotate(err, "adapt to v1 format").Err()
		}
		data, err := proto.Marshal(dut)
		if err != nil {
			return nil, now, errors.Annotate(err, "marshal dut proto to bytes").Err()
		}
		return data, now, nil
	}
	if devices := rsp.GetFailedDevices(); len(devices) == 1 {
		if msg := devices[0].ErrorMsg; strings.Contains(msg, "No such host:") || strings.Contains(msg, "datastore: no such entity") {
			return nil, now, status.Errorf(codes.NotFound, msg)
		}
		return nil, now, errors.Reason(devices[0].ErrorMsg).Err()
	}
	return nil, now, errors.Reason("GetDutInfo %#v failed. No data responsed in either passed or failed list!", req).Err()
}
