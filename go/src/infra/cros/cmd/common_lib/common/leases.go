// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"context"
	"time"

	ufspb "infra/unifiedfleet/api/v1/models"
	ufsapi "infra/unifiedfleet/api/v1/rpc"

	schedukepb "go.chromium.org/chromiumos/config/go/test/scheduling"
	"go.chromium.org/luci/auth/client/authcli"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
)

const (
	defaultPool             = "DUT_POOL_QUOTA"
	leasePriority           = 1
	leasesSchedulingAccount = "leases"
	leaseSchedulingWindow   = 2 * time.Hour
	pollingInterval         = 30 * time.Second
	schedukeTaskKey         = 1
)

// DeviceInfo contains details about the physical lab setup and machine of a
// particular Swarming device.
type DeviceInfo struct {
	Name     string
	LabSetup *ufspb.MachineLSE
	Machine  *ufspb.Machine
}

// LeaseInfo contains details about a particular lease of a Swarming device.
type LeaseInfo struct {
	Device *DeviceInfo
	Build  *buildbucketpb.Build
}

// Abandon sends a cancellation request to Scheduke for the given device names,
// releasing all leased devices for the current user if no devices are
// specified.
func Abandon(ctx context.Context, flags *authcli.Flags, deviceNames []string) error {
	user, err := getUserEmail(ctx, flags)
	if err != nil {
		return err
	}

	sc, err := NewLocalSchedukeClient(ctx)
	if err != nil {
		return err
	}

	return sc.CancelTasks(nil, []string{user}, deviceNames)
}

// Info returns device information for the device with the given name.
func Info(ctx context.Context, deviceName string, flags *authcli.Flags, uc ufsapi.FleetClient) (*DeviceInfo, error) {
	info := &DeviceInfo{Name: deviceName}
	err := addDeviceInfo(ctx, info, uc)
	if err != nil {
		return nil, err
	}
	return info, nil
}

// Lease leases a device from Scheduke and returns information about the device,
// and a bool indicating whether full device information was retrieved.
func Lease(ctx context.Context, flags *authcli.Flags, dims map[string][]string, mins int64, uc ufsapi.FleetClient) (*LeaseInfo, bool, error) {
	deviceName, err := leaseDeviceFromScheduke(ctx, flags, dims, mins)
	if err != nil {
		return nil, false, err
	}
	info := &LeaseInfo{Device: &DeviceInfo{Name: deviceName}}
	// Swallow any UFS errors since the lease has been secured at this point.
	err = addDeviceInfo(ctx, info.Device, uc)
	fullInfoRetrieved := err == nil
	return info, fullInfoRetrieved, nil
}

// Leases retrieves device information for each in-flight lease for the current
// user, and a bool indicating whether full device information was retrieved.
func Leases(ctx context.Context, flags *authcli.Flags, uc ufsapi.FleetClient) ([]*LeaseInfo, bool, error) {
	leaseStates, err := listLeasesFromScheduke(ctx, flags)
	if err != nil {
		return nil, false, err
	}

	var info []*LeaseInfo
	fullInfoRetrieved := true
	for _, ls := range leaseStates {
		// Leases returned from listLeasesFromScheduke include still-pending leases;
		// we only want to return active ones here.
		if ls.State != schedukepb.TaskState_LAUNCHED {
			continue
		}
		li := &LeaseInfo{Device: &DeviceInfo{Name: ls.DeviceName}}
		// Swallow any UFS errors since at least the device name has been retrieved
		// at this point.
		err = addDeviceInfo(ctx, li.Device, uc)
		fullInfoRetrieved = fullInfoRetrieved && (err == nil)
		info = append(info, li)
	}

	return info, fullInfoRetrieved, nil
}
