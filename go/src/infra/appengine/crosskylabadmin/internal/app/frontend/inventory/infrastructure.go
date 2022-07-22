// Copyright 2019 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package inventory implements the fleet.Inventory service end-points of
// corsskylabadmin.
package inventory

import (
	"context"
	"fmt"

	"go.chromium.org/luci/common/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"infra/appengine/crosskylabadmin/internal/app/config"
	"infra/appengine/crosskylabadmin/internal/app/gitstore"
	"infra/libs/skylab/inventory"
)

// GetDutsByEnvironment returns Duts belong to a given environment.
func GetDutsByEnvironment(ctx context.Context, s *gitstore.InventoryStore) ([]*inventory.DeviceUnderTest, error) {
	c := newGlobalInvCache(ctx, s)
	cfg := config.Get(ctx).Inventory
	d := queenDroneName(cfg.Environment)
	logging.Debugf(ctx, "Using pseudo-drone %s", d)
	server, ok := c.hostnameToDrone[d]
	if !ok {
		return nil, fmt.Errorf("drone (%s) does not exist", d)
	}
	dutUids := server.GetDutUids()
	logging.Debugf(ctx, "server (%s) contains %d duts", server.GetHostname(), len(dutUids))
	duts := make([]*inventory.DeviceUnderTest, 0, len(dutUids))
	for _, duid := range dutUids {
		if d, ok := c.idToDUT[duid]; ok {
			duts = append(duts, d)
		}
	}
	return duts, nil
}

// globalInvCache wraps an InventoryStore and keeps various lookup caches.
// Unlike invCache, this ignores the environment and includes the entire inventory.
type globalInvCache struct {
	store           *gitstore.InventoryStore
	hostnameToID    map[string]string
	droneForDUT     map[string]*inventory.Server
	idToDUT         map[string]*inventory.DeviceUnderTest
	hostnameToDrone map[string]*inventory.Server
}

func newGlobalInvCache(ctx context.Context, s *gitstore.InventoryStore) *globalInvCache {
	ic := globalInvCache{
		store:           s,
		hostnameToID:    make(map[string]string),
		droneForDUT:     make(map[string]*inventory.Server),
		idToDUT:         make(map[string]*inventory.DeviceUnderTest),
		hostnameToDrone: make(map[string]*inventory.Server),
	}
	for _, d := range s.Lab.GetDuts() {
		c := d.GetCommon()
		ic.hostnameToID[c.GetHostname()] = c.GetId()
		ic.idToDUT[c.GetId()] = d
	}
	for _, srv := range s.Infrastructure.GetServers() {
		if !isDrone(srv) {
			continue
		}
		ic.hostnameToDrone[srv.GetHostname()] = srv
		for _, d := range srv.DutUids {
			ic.droneForDUT[d] = srv
		}
	}
	return &ic
}

// assignDUT assigns the given DUT to the queen drone in the current environment.
func assignDUT(ctx context.Context, c *globalInvCache, dutID string) (drone string, _ error) {
	cfg := config.Get(ctx).Inventory
	d := queenDroneName(cfg.Environment)
	logging.Debugf(ctx, "Using pseudo-drone %s for DUT %s", d, dutID)
	if _, ok := c.idToDUT[dutID]; !ok {
		return "", status.Error(codes.NotFound, fmt.Sprintf("DUT %s does not exist", dutID))
	}
	if server, ok := c.droneForDUT[dutID]; ok {
		return "", status.Errorf(codes.InvalidArgument,
			"dut %s is already assigned to drone %s", dutID, server.GetHostname())
	}
	server, ok := c.hostnameToDrone[d]
	if !ok {
		panic(fmt.Sprintf("drone %s does not exist", d))
	}
	server.DutUids = append(server.DutUids, dutID)
	c.droneForDUT[dutID] = server
	c.idToDUT[dutID].RemovalReason = nil
	return d, nil
}
