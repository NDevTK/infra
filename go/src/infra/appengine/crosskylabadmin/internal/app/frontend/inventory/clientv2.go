// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package inventory

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"go.chromium.org/chromiumos/infra/proto/go/device"
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
	"infra/libs/skylab/inventory"

	"go.chromium.org/chromiumos/infra/proto/go/lab"
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

func (c *invServiceClient) addManyDUTsToFleet(ctx context.Context, nds []*inventory.CommonDeviceSpecs, pickServoPort bool) (string, []*inventory.CommonDeviceSpecs, error) {
	// In case there's any panic happens in the new code.
	defer func() {
		if r := recover(); r != nil {
			logging.Errorf(ctx, "Recovered in addManyDUTsToFleet(%s)", r)
			debug.PrintStack()
		}
	}()

	c.logInfo(ctx, "Access inventory service as user: %s", auth.CurrentUser(ctx))

	devicesToAdd, labstations, _, err := api.ImportFromV1DutSpecs(nds)
	if err != nil {
		logging.Errorf(ctx, "failed to import DUT specs: %s", err)
		return "", nil, err
	}
	if len(devicesToAdd) == 0 {
		devicesToAdd = labstations
	}

	var rsp *api.AddCrosDevicesResponse
	f := func() error {
		rsp, err = c.client.AddCrosDevices(ctx, &api.AddCrosDevicesRequest{
			Devices:       devicesToAdd,
			PickServoPort: pickServoPort,
		})
		return err
	}
	err = retry.Retry(ctx, transientErrorRetries(), f, retry.LogCallback(ctx, "addManyDUTsToFleet v2"))
	if err != nil {
		return "", nil, err
	}

	passedDevSpecs := getPassedDevSpecs(nds, rsp.GetPassedDevices())
	failedDevices := rsp.GetFailedDevices()
	msgs := make([]string, 0, len(failedDevices))
	for _, d := range failedDevices {
		msgs = append(msgs, d.GetErrorMsg())
	}
	if len(msgs) > 0 {
		err = errors.Reason(strings.Join(msgs, ",")).Err()
	}
	return "URL N/A", passedDevSpecs, err
}

func (c *invServiceClient) getAssetsFromRegistration(ctx context.Context, assetIDList *api.AssetIDList) (*api.AssetResponse, error) {
	// In case there's any panic happens in the new code.
	defer func() {
		if r := recover(); r != nil {
			logging.Errorf(ctx, "Recovered in getAssetsFromRegistration(%s)", r)
			debug.PrintStack()
		}
	}()
	c.logInfo(ctx, "Getting Assets from Registration System as user : %s", auth.CurrentUser(ctx))
	// getting existing assets from registration system
	return c.client.GetAssets(ctx, assetIDList)
}

func (c *invServiceClient) updateAssetsInRegistration(ctx context.Context, assetList *api.AssetList) (*api.AssetResponse, error) {
	// In case there's any panic happens in the new code.
	defer func() {
		if r := recover(); r != nil {
			logging.Errorf(ctx, "Recovered in updateAssetsInRegistration(%s)", r)
			debug.PrintStack()
		}
	}()
	c.logInfo(ctx, "Updating Registration System as user : %s", auth.CurrentUser(ctx))
	// Update existing assets to registration system
	return c.client.UpdateAssets(ctx, assetList)
}

func getPassedDevSpecs(allDevSpecs []*inventory.CommonDeviceSpecs, passedDevices []*api.DeviceOpResult) []*inventory.CommonDeviceSpecs {
	// The response only has hostname/id, so we need to get other data from the
	// input.
	nameToSpec := map[string]*inventory.CommonDeviceSpecs{}
	passedDevSpecs := make([]*inventory.CommonDeviceSpecs, 0, len(allDevSpecs))
	for i := range allDevSpecs {
		nameToSpec[allDevSpecs[i].GetHostname()] = allDevSpecs[i]
	}

	for _, d := range passedDevices {
		pd := nameToSpec[d.GetHostname()]
		id := d.GetId()
		pd.Id = &id // The ID may be newly assigned by inventory service.
		passedDevSpecs = append(passedDevSpecs, pd)
	}
	return passedDevSpecs
}

func (c *invServiceClient) updateDUTSpecs(ctx context.Context, od, nd *inventory.CommonDeviceSpecs, pickServoPort bool) (string, error) {
	defer func() {
		if r := recover(); r != nil {
			logging.Errorf(ctx, "Recovered in updateDUTSpecs(%s)", r)
		}
	}()

	devicesToUpdate, labstations, _, err := api.ImportFromV1DutSpecs([]*inventory.CommonDeviceSpecs{nd})
	if err != nil {
		logging.Errorf(ctx, "failed to import DUT specs: %s", err)
		return "", err
	}
	if len(devicesToUpdate) == 0 {
		devicesToUpdate = labstations
	}

	cleanPreDeployFields(devicesToUpdate)

	f := func() error {
		if rsp, err := c.client.UpdateCrosDevicesSetup(ctx, &api.UpdateCrosDevicesSetupRequest{
			Devices:       devicesToUpdate,
			PickServoPort: pickServoPort,
			// TODO (guocb) Add reason why update.
		}); err != nil {
			return err
		} else if len(rsp.FailedDevices) > 0 {
			// There's only one device under updating.
			return errors.Reason(rsp.FailedDevices[0].ErrorMsg).Err()
		}
		return nil
	}
	err = retry.Retry(ctx, transientErrorRetries(), f, retry.LogCallback(ctx, "updateDUTSpecs v2"))
	if err != nil {
		return "", errors.Annotate(err, "update DUT spects").Err()
	}

	return "URL N/A", nil
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

func (c *invServiceClient) deviceConfigsExists(ctx context.Context, configIds []*device.ConfigId) (map[int32]bool, error) {
	// In case there's any panic happens in the new code.
	defer func() {
		if r := recover(); r != nil {
			logging.Errorf(ctx, "Recovered in deviceConfigsExists(%s)", r)
			debug.PrintStack()
		}
	}()
	// checking if device configs exists in Inventory V2
	resp, err := c.client.DeviceConfigsExists(ctx, &api.DeviceConfigsExistsRequest{ConfigIds: configIds})
	if err != nil {
		c.logInfo(ctx, "Unable to check if device configs exists", err)
		return nil, err
	}
	return resp.GetExists(), err
}

// cleanPreDeployFields resets values for the fields re-generated during deployment.
func cleanPreDeployFields(devices []*lab.ChromeOSDevice) {
	for _, d := range devices {
		if dut := d.GetDut(); dut != nil {
			servo := dut.GetPeripherals().GetServo()
			if servo != nil {
				servo.ServoType = ""
				servo.ServoTopology = nil
			}
		}
	}
}
