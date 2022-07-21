// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package inventory

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	"go.chromium.org/chromiumos/infra/proto/go/device"
	"go.chromium.org/luci/common/data/stringset"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/retry"
	"go.chromium.org/luci/gae/service/datastore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	api "infra/appengine/cros/lab_inventory/api/v1"
	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
	dsinventory "infra/appengine/crosskylabadmin/internal/app/frontend/datastore/inventory"
	"infra/appengine/crosskylabadmin/internal/app/gitstore"
	"infra/libs/skylab/common/heuristics"
	"infra/libs/skylab/inventory"
)

const (
	servoHostAttributeKey = "servo_host"
	servoPortAttributeKey = "servo_port"
)

type inventoryClient interface {
	addManyDUTsToFleet(context.Context, []*inventory.CommonDeviceSpecs, bool) (string, []*inventory.CommonDeviceSpecs, error)
	getAssetsFromRegistration(ctx context.Context, assetList *api.AssetIDList) (*api.AssetResponse, error)
	updateAssetsInRegistration(ctx context.Context, assetList *api.AssetList) (*api.AssetResponse, error)
	updateDUTSpecs(context.Context, *inventory.CommonDeviceSpecs, *inventory.CommonDeviceSpecs, bool) (string, error)
	deleteDUTsFromFleet(context.Context, []string) (string, []string, error)
	getDutInfo(context.Context, *fleet.GetDutInfoRequest) ([]byte, time.Time, error)
	deviceConfigsExists(context.Context, []*device.ConfigId) (map[int32]bool, error)
}

type gitStoreClient struct {
	store *gitstore.InventoryStore
}

func newGitStoreClient(ctx context.Context, gs *gitstore.InventoryStore) (inventoryClient, error) {
	return &gitStoreClient{
		store: gs,
	}, nil
}

func (client *gitStoreClient) addManyDUTsToFleet(ctx context.Context, nds []*inventory.CommonDeviceSpecs, pickServoPort bool) (string, []*inventory.CommonDeviceSpecs, error) {
	return addManyDUTsToFleet(ctx, client.store, nds, pickServoPort)
}

func (client *gitStoreClient) getAssetsFromRegistration(ctx context.Context, assetList *api.AssetIDList) (*api.AssetResponse, error) {
	// Nothing to get from registration for gitStoreClient
	return nil, errors.New("gitStoreClient - nothing to get from registration")
}

func (client *gitStoreClient) updateAssetsInRegistration(ctx context.Context, assetList *api.AssetList) (*api.AssetResponse, error) {
	// Nothing to update in registration for gitStoreClient
	return nil, errors.New("gitStoreClient - nothing to update in registration")
}

func (client *gitStoreClient) updateDUTSpecs(ctx context.Context, od, nd *inventory.CommonDeviceSpecs, pickServoPort bool) (string, error) {
	return updateDUTSpecs(ctx, client.store, od, nd, pickServoPort)
}

func (client *gitStoreClient) deleteDUTsFromFleet(ctx context.Context, ids []string) (string, []string, error) {
	return deleteDUTsFromFleet(ctx, client.store, ids)
}

func (client *gitStoreClient) deviceConfigsExists(ctx context.Context, configIds []*device.ConfigId) (map[int32]bool, error) {
	// Nothing to check for device configs in gitStore
	return nil, errors.New("gitStoreClient - nothing to check for device configs in gitStore")
}

func (client *gitStoreClient) getDutInfo(ctx context.Context, req *fleet.GetDutInfoRequest) ([]byte, time.Time, error) {
	var dut *dsinventory.DeviceUnderTest
	var now time.Time
	var err error
	if req.Id != "" {
		dut, err = dsinventory.GetSerializedDUTByID(ctx, req.Id)
	} else {
		dut, err = dsinventory.GetSerializedDUTByHostname(ctx, req.Hostname)
	}
	if err != nil {
		if datastore.IsErrNoSuchEntity(err) {
			return nil, now, status.Errorf(codes.NotFound, err.Error())
		}
		return nil, now, err
	}
	return dut.Data, dut.Updated, nil
}

func addManyDUTsToFleet(ctx context.Context, s *gitstore.InventoryStore, nds []*inventory.CommonDeviceSpecs, pickServoPort bool) (string, []*inventory.CommonDeviceSpecs, error) {
	var respURL string
	newDeviceToID := make(map[*inventory.CommonDeviceSpecs]string)

	f := func() error {
		var ds []*inventory.CommonDeviceSpecs

		for _, nd := range nds {
			ds = append(ds, proto.Clone(nd).(*inventory.CommonDeviceSpecs))
		}

		if err := s.Refresh(ctx); err != nil {
			return errors.Annotate(err, "add dut to fleet").Err()
		}

		// New cache after refreshing store.
		c := newGlobalInvCache(ctx, s)

		for i, d := range ds {
			hostname := d.GetHostname()
			logging.Infof(ctx, "add device to fleet: %s", hostname)
			if _, ok := c.hostnameToID[hostname]; ok {
				logging.Infof(ctx, "dut with hostname %s already exists, skip adding", hostname)
				continue
			}
			if pickServoPort && !hasServoPortAttribute(d) {
				if err := assignNewServoPort(s.Lab.Duts, d); err != nil {
					logging.Infof(ctx, "fail to assign new servo port, skip adding")
					continue
				}
			}
			nid := addDUTToStore(s, d)
			newDeviceToID[nds[i]] = nid
		}

		// TODO(ayatane): Implement this better than just regenerating the cache.
		c = newGlobalInvCache(ctx, s)

		for _, id := range newDeviceToID {
			if _, err := assignDUT(ctx, c, id); err != nil {
				return errors.Annotate(err, "add dut to fleet").Err()
			}
		}

		firstHostname := "<empty>"
		if len(ds) > 0 {
			firstHostname = ds[0].GetHostname()
		}

		url, err := s.Commit(ctx, fmt.Sprintf("Add %d new DUT(s) : %s ...", len(ds), firstHostname))
		if err != nil {
			return errors.Annotate(err, "add dut to fleet").Err()
		}

		respURL = url
		for _, nd := range nds {
			id := newDeviceToID[nd]
			nd.Id = &id
		}
		return nil
	}

	err := retry.Retry(ctx, transientErrorRetries(), f, retry.LogCallback(ctx, "addManyDUTsToFleet"))

	newDevices := make([]*inventory.CommonDeviceSpecs, 0)
	for nd := range newDeviceToID {
		newDevices = append(newDevices, nd)
	}
	return respURL, newDevices, err
}

// updateDUTSpecs updates the DUT specs for an existing DUT in the inventory.
func updateDUTSpecs(ctx context.Context, s *gitstore.InventoryStore, od, nd *inventory.CommonDeviceSpecs, pickServoPort bool) (string, error) {
	var respURL string
	f := func() error {
		// Clone device specs before modifications so that changes don't leak
		// across retries.
		d := proto.Clone(nd).(*inventory.CommonDeviceSpecs)

		if err := s.Refresh(ctx); err != nil {
			return errors.Annotate(err, "add new dut to inventory").Err()
		}

		if pickServoPort && !hasServoPortAttribute(d) {
			if err := assignNewServoPort(s.Lab.Duts, d); err != nil {
				return errors.Annotate(err, "add dut to fleet").Err()
			}
		}

		dut, exists := getDUTByID(s.Lab, od.GetId())
		if !exists {
			return status.Errorf(codes.NotFound, "no DUT with ID %s", od.GetId())
		}
		// TODO(crbug/929776) DUTs under deployment are not marked specially in the
		// inventory yet. This causes two problems:
		// - Another admin task (say repair) may get scheduled on the new bot
		//   before the deploy task we create.
		// - If the deploy task fails, the DUT will still enter the fleet, but may
		//   not be ready for use.
		if !proto.Equal(dut.GetCommon(), od) {
			return errors.Reason("DUT specs update conflict").Err()
		}
		dut.Common = d

		url, err := s.Commit(ctx, fmt.Sprintf("Update DUT %s", od.GetId()))
		if err != nil {
			return errors.Annotate(err, "update DUT specs").Err()
		}

		respURL = url
		return nil
	}
	err := retry.Retry(ctx, transientErrorRetries(), f, retry.LogCallback(ctx, "updateDUTSpecs"))
	return respURL, err
}

// assignNewServoPort adds a valid servo port attribute to d
//
// This function returns an error if d does not have the servo host attribute.
// The assigned servo port value is unique among all the duts with the same
// servo host attribute.
func assignNewServoPort(duts []*inventory.DeviceUnderTest, d *inventory.CommonDeviceSpecs) error {
	if heuristics.LooksLikeLabstation(d.GetHostname()) {
		return nil
	}
	servoHost, found := getAttributeByKey(d, servoHostAttributeKey)
	if !found {
		return errors.Reason("no servo_host attribute in specs").Err()
	}

	used := usedServoPorts(duts, servoHost)
	p, err := findFreePort(used)
	if err != nil {
		return err
	}
	k := servoPortAttributeKey
	v := strconv.Itoa(p)
	d.Attributes = append(d.Attributes, &inventory.KeyValue{Key: &k, Value: &v})
	return nil
}

// getAttributeByKey by returns the value for the attribute with the given key,
// and whether the key was found.
func getAttributeByKey(d *inventory.CommonDeviceSpecs, key string) (string, bool) {
	for _, a := range d.Attributes {
		if *a.Key == key {
			return *a.Value, true
		}
	}
	return "", false
}

// findFreePort finds a valid port that is not in use.
//
// This function returns an error if no free port is found.
// This function modifies the slice of ports provided.
func findFreePort(used []int) (int, error) {
	sort.Sort(sort.Reverse(sort.IntSlice(used)))
	// This range is consistent with the range of ports generated by servod:
	// https://chromium.googlesource.com/chromiumos/third_party/hdctools/+/cf5f8027b9d3015db75df4853e37ea7a2f1ac538/servo/servod.py#36
	for p := 9999; p > 9900; p-- {
		if len(used) == 0 || p != used[0] {
			return p, nil
		}
		used = used[1:]
	}
	return -1, errors.Reason("no free valid ports").Err()
}

// usedServoPorts finds the servo ports used by duts with the given servo host
// attribute.
func usedServoPorts(duts []*inventory.DeviceUnderTest, servoHost string) []int {
	used := []int{}
	for _, d := range duts {
		c := d.GetCommon()
		if sh, found := getAttributeByKey(c, servoHostAttributeKey); found && sh == servoHost {
			if p, found := getAttributeByKey(c, servoPortAttributeKey); found {
				ip, err := strconv.ParseInt(p, 10, 32)
				if err != nil {
					// This is not the right place to enforce servo_port correctness for other DUTs.
					// All we care about is that the port we pick will not conflict with
					// this corrupted entry.
					continue
				}
				used = append(used, int(ip))
			}
		}
	}
	return used
}

func hasServoPortAttribute(d *inventory.CommonDeviceSpecs) bool {
	_, found := getAttributeByKey(d, servoPortAttributeKey)
	return found
}

// addDUTToStore adds a new DUT with the given specs to the store.
//
// This function returns the ID for the added DUT.
func addDUTToStore(s *gitstore.InventoryStore, nd *inventory.CommonDeviceSpecs) string {
	// TODO(crbug/912977) DUTs under deployment are not marked specially in the
	// inventory yet. This causes two problems:
	// - Another admin task (say repair) may get scheduled on the new bot
	//   before the deploy task we create.
	// - If the deploy task fails, the DUT will still enter the fleet, but may
	//   not be ready for use.
	s.Lab.Duts = append(s.Lab.Duts, &inventory.DeviceUnderTest{
		Common: nd,
	})
	return *nd.Id
}

func deleteDUTsFromFleet(ctx context.Context, s *gitstore.InventoryStore, ids []string) (string, []string, error) {
	var changeURL string
	var removedIDs []string
	f := func() error {
		if err2 := s.Refresh(ctx); err2 != nil {
			return err2
		}
		removedDUTs := removeDUTWithHostnames(s, ids)
		url, err2 := s.Commit(ctx, fmt.Sprintf("delete %d duts", len(removedDUTs)))
		if gitstore.IsEmptyErr(err2) {
			return nil
		}
		if err2 != nil {
			return err2
		}

		// Captured variables only on success, hence at most once.
		changeURL = url
		removedIDs = make([]string, 0, len(removedDUTs))
		for _, d := range removedDUTs {
			removedIDs = append(removedIDs, d.GetCommon().GetId())
		}
		return nil
	}
	err := retry.Retry(ctx, transientErrorRetries(), f, retry.LogCallback(ctx, "DeleteDut"))
	return changeURL, removedIDs, err
}

// removeDUTWithHostnames deletes duts with the given hostnames.
//
// The function returns the deleted duts.
// If multiple DUTs have the same hostname, that is in hostnames, they are all deleted.
func removeDUTWithHostnames(s *gitstore.InventoryStore, hostnames []string) []*inventory.DeviceUnderTest {
	duts := s.Lab.Duts
	toRemove := stringset.NewFromSlice(hostnames...)
	removedDuts := make([]*inventory.DeviceUnderTest, 0, len(hostnames))
	for i := 0; i < len(duts); {
		d := duts[i]
		h := d.GetCommon().GetHostname()
		if !toRemove.Has(h) {
			i++
			continue
		}
		removedDuts = append(removedDuts, d)
		duts = deleteAtIndex(duts, i)
	}
	s.Lab.Duts = duts
	return removedDuts
}

func deleteAtIndex(duts []*inventory.DeviceUnderTest, i int) []*inventory.DeviceUnderTest {
	copy(duts[i:], duts[i+1:])
	duts[len(duts)-1] = nil
	return duts[:len(duts)-1]
}
