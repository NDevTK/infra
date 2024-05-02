// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package jobs

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/protoadapt"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"

	shivasUtil "infra/cmd/shivas/utils"
	"infra/device_manager/internal/controller"
	"infra/device_manager/internal/external"
	"infra/device_manager/internal/frontend"
	"infra/device_manager/internal/model"
	"infra/libs/fleet/device"
	ufspb "infra/unifiedfleet/api/v1/models"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
)

const ufsServiceURI = "ufs.api.cr.dev"

// setupContext sets up context with a UFS namespace.
func setupContext(ctx context.Context, namespace string) context.Context {
	md := metadata.Pairs(ufsUtil.Namespace, namespace)
	return metadata.NewOutgoingContext(ctx, md)
}

// ImportUFSDevices registers the cron to trigger import for all Device
// information from UFS.
func ImportUFSDevices(ctx context.Context, serviceClients frontend.ServiceClients) error {
	// TODO (b/331644796): Import non-OS device data
	// TODO (b/328662436): Collect metrics
	ctx = setupContext(ctx, ufsUtil.OSNamespace)
	ufsClient, err := external.NewUFSClient(ctx, ufsServiceURI)
	if err != nil {
		return err
	}
	serviceClients.UFSClient = ufsClient
	lses, err := getAllMachineLSEs(ctx, serviceClients.UFSClient)
	if err != nil {
		return err
	}
	logging.Debugf(ctx, "ImportUFSDevices: found %d DUTs in UFS OS namespace", len(lses))

	sUnits, err := getAllSchedulingUnits(ctx, serviceClients.UFSClient)
	if err != nil {
		return err
	}
	logging.Debugf(ctx, "ImportUFSDevices: found %d SUs in UFS OS namespace", len(sUnits))

	// Contains active hostnames of SchedulingUnits and individual MachineLSEs.
	// A device marked inactive simply means they should not be exposed or be used
	// by the scheduling layer. For example, the components of a SchedulingUnit
	// should not be individually schedulable.
	var activeDUTs []string

	// map for MachineLSEs associated with SchedulingUnit for easy search
	lseInSUnitMap := make(map[string]bool)
	for _, su := range sUnits {
		if len(su.GetMachineLSEs()) > 0 {
			activeDUTs = append(activeDUTs, su.GetName())
			for _, lseName := range su.GetMachineLSEs() {
				lseInSUnitMap[lseName] = true
			}
		}
	}
	logging.Debugf(ctx, "ImportUFSDevices: %d SchedulingUnit DUTs to be marked inactive", len(lseInSUnitMap))

	// add all individual MachineLSEs as active DUTs
	for _, lse := range lses {
		if !lseInSUnitMap[ufsUtil.RemovePrefix(lse.GetName())] {
			activeDUTs = append(activeDUTs, lse.GetName())
		}
	}
	logging.Debugf(ctx, "ImportUFSDevices: found %d active devices to update", len(activeDUTs))

	// get inactive DUTs
	inactiveDUTs, err := getInactiveDevices(ctx, serviceClients, activeDUTs)
	if err != nil {
		return err
	}

	// loop through all active and inactive MachineLSEs and upsert as Devices
	wg := sync.WaitGroup{}
	for _, d := range inactiveDUTs {
		wg.Add(1)
		go upsertDeviceData(ctx, &wg, serviceClients, d, false)
	}

	for _, dutName := range activeDUTs {
		wg.Add(1)
		go upsertDeviceData(ctx, &wg, serviceClients, dutName, true)
	}
	wg.Wait()

	return nil
}

// getAllMachineLSEs gets all MachineLSEs
func getAllMachineLSEs(ctx context.Context, ic ufsAPI.FleetClient) ([]*ufspb.MachineLSE, error) {
	res, err := shivasUtil.BatchList(ctx, ic, listMachineLSEs, []string{}, 0, true, false)
	if err != nil {
		return nil, errors.Annotate(err, "getAllMachineLSEs").Err()
	}
	lses := make([]*ufspb.MachineLSE, len(res))
	for i, r := range res {
		lses[i] = r.(*ufspb.MachineLSE)
	}
	return lses, nil
}

// listMachineLSEs calls ListMachineLSEs in UFS to get a list of MachineLSEs
func listMachineLSEs(ctx context.Context, ic ufsAPI.FleetClient, pageSize int32, pageToken, filter string, keysOnly, full bool) ([]protoadapt.MessageV1, string, error) {
	req := &ufsAPI.ListMachineLSEsRequest{
		PageSize:  pageSize,
		PageToken: pageToken,
		Filter:    filter,
		KeysOnly:  keysOnly,
		Full:      full,
	}
	res, err := ic.ListMachineLSEs(ctx, req)
	if err != nil {
		return nil, "", errors.Annotate(err, "listMachineLSEs").Err()
	}
	protos := make([]protoadapt.MessageV1, len(res.GetMachineLSEs()))
	for i, kvm := range res.GetMachineLSEs() {
		protos[i] = kvm
	}
	return protos, res.GetNextPageToken(), nil
}

// getAllSchedulingUnits gets all SchedulingUnits
func getAllSchedulingUnits(ctx context.Context, ic ufsAPI.FleetClient) ([]*ufspb.SchedulingUnit, error) {
	res, err := shivasUtil.BatchList(ctx, ic, listSchedulingUnits, []string{}, 0, false, true)
	if err != nil {
		return nil, errors.Annotate(err, "getAllSchedulingUnits").Err()
	}
	sUnits := make([]*ufspb.SchedulingUnit, len(res))
	for i, r := range res {
		sUnits[i] = r.(*ufspb.SchedulingUnit)
	}
	return sUnits, nil
}

// listSchedulingUnits calls ListSchedulingUnits in UFS to get a list of SchedulingUnits
func listSchedulingUnits(ctx context.Context, ic ufsAPI.FleetClient, pageSize int32, pageToken, filter string, keysOnly, full bool) ([]protoadapt.MessageV1, string, error) {
	req := &ufsAPI.ListSchedulingUnitsRequest{
		PageSize:  pageSize,
		PageToken: pageToken,
		Filter:    filter,
		KeysOnly:  keysOnly,
	}
	res, err := ic.ListSchedulingUnits(ctx, req)
	if err != nil {
		return nil, "", err
	}
	protos := make([]protoadapt.MessageV1, len(res.GetSchedulingUnits()))
	for i, m := range res.GetSchedulingUnits() {
		protos[i] = m
	}
	return protos, res.GetNextPageToken(), nil
}

// getAllDMDevices gets all Devices in the Device Manager DB
func getAllDMDevices(ctx context.Context, db *sql.DB) ([]model.Device, error) {
	var (
		pageNumber int = 0
		pageSize   int = 1000
		res        []model.Device
		devices    []model.Device
		err        error
	)
	for (pageNumber == 0) || (pageNumber != 0 && len(devices) != 0) {
		devices, err = model.ListDevices(ctx, db, pageNumber, pageSize)
		if err != nil {
			return nil, err
		}
		res = append(res, devices...)
		pageNumber += 1
	}
	return res, nil
}

// upsertDeviceData upserts to db and publishes a device event with UFS device data
func upsertDeviceData(ctx context.Context, wg *sync.WaitGroup, serviceClients frontend.ServiceClients, name string, active bool) {
	// catch panic and continue
	defer func() {
		if err := recover(); err != nil {
			logging.Debugf(ctx, "panic occurred: %s", err)
		}
		wg.Done()
	}()

	// process Device, upsert to db, and publish DeviceEvent
	deviceModel := model.Device{
		ID:              ufsUtil.RemovePrefix(name),
		DeviceType:      "DEVICE_TYPE_PHYSICAL",
		LastUpdatedTime: time.Now(),
		IsActive:        active,
	}

	// only get dims for active Devices
	if active {
		r := func(e error) { logging.Debugf(ctx, "sanitize dimensions: %s\n", e) }
		dims, err := device.GetOSResourceDims(ctx, serviceClients.UFSClient, r, name)
		if err != nil {
			return
		}
		schedLabels := make(model.SchedulableLabels)
		for k, v := range dims {
			schedLabels[k] = model.LabelValues{
				Values: v,
			}
		}
		deviceModel.SchedulableLabels = schedLabels
	}

	// upsert deviceModel to DM db
	err := model.UpsertDevice(ctx, serviceClients.DBClient.Conn, deviceModel)
	if err != nil {
		return
	}

	err = controller.PublishDeviceEvent(ctx, serviceClients.PubSubClient, deviceModel)
	if err != nil {
		return
	}
}

// getInactiveDevices marks inactive Devices as inactive and returns the list.
//
// This function takes a list of Devices and compares them with the list of
// Devices managed by Device Manager. It marks Devices that are not in the
// active list as inactive and returns the list.
func getInactiveDevices(ctx context.Context, serviceClients frontend.ServiceClients, activeDevices []string) ([]string, error) {
	dmDevices, err := getAllDMDevices(ctx, serviceClients.DBClient.Conn)
	if err != nil {
		return nil, err
	}
	logging.Debugf(ctx, "getAllDMDevices: found %d Devices in DM database", len(dmDevices))

	// create a map of active Device names
	activeMap := make(map[string]struct{}, len(dmDevices))
	for _, activeDevice := range activeDevices {
		activeMap[ufsUtil.RemovePrefix(activeDevice)] = struct{}{}
	}

	// the set difference of active Devices - all DM Devices = inactive Devices
	var inactiveDevices []string
	for _, dmDevice := range dmDevices {
		// only process DM Devices that are currently marked active
		if _, found := activeMap[dmDevice.ID]; !found {
			inactiveDevices = append(inactiveDevices, dmDevice.ID)
		}
	}
	logging.Debugf(ctx, "decommDevices: found %d inactive DUTs", len(inactiveDevices))

	return inactiveDevices, nil
}
