// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package inventory

import (
	"context"
	"math/rand"
	"time"

	"go.chromium.org/luci/common/data/stringset"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"

	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
	"infra/appengine/crosskylabadmin/internal/app/gitstore"
)

type duoClient struct {
	gc *gitStoreClient
	ic *invServiceClient

	// A number in [0, 100] indicate the write traffic (deploy/update)
	// duplicated to inventory v2 service.
	writeTrafficRatio int
	// A number in [0, 100] indicate the read traffic fanning out to inventory
	// v2 service.
	readTrafficRatio int

	// The uuids of migration test devices.
	testingDeviceUUIDs stringset.Set

	// The uuids of migration test devices.
	testingDeviceNames stringset.Set

	// If we still write to v1.
	inventoryV2Only bool
}

func newDuoClient(ctx context.Context, gs *gitstore.InventoryStore, host string, readTrafficRatio, writeTrafficRatio int, testingUUIDs, testingNames []string, inventoryV2Only bool) (inventoryClient, error) {
	gc, err := newGitStoreClient(ctx, gs)
	if err != nil {
		return nil, errors.Annotate(err, "create git client").Err()
	}
	ic, err := newInvServiceClient(ctx, host)
	if err != nil {
		logging.Infof(ctx, "Failed to create inventory client of the duo client. Just return the git store client")
		return gc, nil
	}
	return &duoClient{
		gc:                 gc.(*gitStoreClient),
		ic:                 ic.(*invServiceClient),
		readTrafficRatio:   readTrafficRatio,
		writeTrafficRatio:  writeTrafficRatio,
		testingDeviceUUIDs: stringset.NewFromSlice(testingUUIDs...),
		testingDeviceNames: stringset.NewFromSlice(testingNames...),
		inventoryV2Only:    inventoryV2Only,
	}, nil
}

func (client *duoClient) willReadFromV2(req *fleet.GetDutInfoRequest) bool {
	if req.MustFromV1 {
		return false
	}
	if client.testingDeviceUUIDs.Has(req.GetId()) {
		return true
	}
	if client.testingDeviceNames.Has(req.GetHostname()) {
		return true
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(100) < client.readTrafficRatio
}

func (client *duoClient) getDutInfo(ctx context.Context, req *fleet.GetDutInfoRequest) ([]byte, time.Time, error) {
	if client.willReadFromV2(req) {
		dut, now, err := client.ic.getDutInfo(ctx, req)
		logging.Infof(ctx, "[v2] GetDutInfo result: %#v: %s", req, err)
		return dut, now, err
	}
	return client.gc.getDutInfo(ctx, req)
}
