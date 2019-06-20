// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package queries contains convenient datastore queries.
package queries

import (
	"github.com/google/uuid"
	"go.chromium.org/gae/service/datastore"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"golang.org/x/net/context"

	"infra/appengine/drone-queen/api"
	"infra/appengine/drone-queen/internal/config"
	"infra/appengine/drone-queen/internal/entities"
)

// CreateNewDrone creates a new Drone datastore entity with a unique ID.
func CreateNewDrone(ctx context.Context) (entities.DroneID, error) {
	const maxAttempts = 10
	var id entities.DroneID
	retry := errors.New("retry")
	for i := 1; ; i++ {
		f := func(ctx context.Context) error {
			proposed := uuid.New().String()
			key := datastore.MakeKey(ctx, entities.DroneKind, proposed)
			res, err := datastore.Exists(ctx, key)
			if err != nil {
				return errors.Annotate(err, "check if drone %s exists", id).Err()
			}
			if res.Any() {
				if i == maxAttempts {
					return errors.Reason("max attempts finding unique ID").Err()
				}
				return retry
			}
			id = entities.DroneID(proposed)
			drone := entities.Drone{
				ID: id,
			}
			if err := datastore.Put(ctx, &drone); err != nil {
				return err
			}
			return nil
		}
		if err := datastore.RunInTransaction(ctx, f, nil); err != nil {
			if err == retry {
				retryUniqueUUID.Add(ctx, 1, config.Instance(ctx))
				continue
			}
			return "", errors.Annotate(err, "create new drone").Err()
		}
		return id, nil
	}
}

// GetDroneDUTs gets the DUTs assigned to a drone.  This does not have
// to be run in a transaction, but caveat emptor.
func GetDroneDUTs(ctx context.Context, d entities.DroneID) ([]*entities.DUT, error) {
	q := datastore.NewQuery(entities.DUTKind)
	q = q.Eq(entities.AssignedDroneField, d)
	q = q.Ancestor(entities.DUTGroupKey(ctx))
	var duts []*entities.DUT
	if err := datastore.GetAll(ctx, q, &duts); err != nil {
		return nil, errors.Annotate(err, "get drone %v DUTs:", d).Err()
	}
	return duts, nil
}

// GetUnassignedDUTs gets at most the specified number of unassigned
// DUTs.  This does not have to be run in a transaction, but caveat
// emptor.  If n is less than zero, return no DUTs.
func GetUnassignedDUTs(ctx context.Context, n int32) ([]*entities.DUT, error) {
	if n < 0 {
		return nil, nil
	}
	q := datastore.NewQuery(entities.DUTKind)
	q = q.Eq(entities.AssignedDroneField, "")
	q = q.Eq(entities.DrainingField, false)
	q = q.Ancestor(entities.DUTGroupKey(ctx))
	q = q.Limit(n)
	var duts []*entities.DUT
	if err := datastore.GetAll(ctx, q, &duts); err != nil {
		return nil, errors.Annotate(err, "get %v unassigned DUTs", n).Err()
	}
	return duts, nil
}

// AssignNewDUTs assigns new DUTs to the drone according to its load
// indicators and current DUTs.  Returns the list of all DUTs assigned
// to the drone.
//
// This function needs to be run within a datastore transaction.
func AssignNewDUTs(ctx context.Context, d entities.DroneID, li *api.ReportDroneRequest_LoadIndicators) ([]*entities.DUT, error) {
	currentDUTs, err := GetDroneDUTs(ctx, d)
	if err != nil {
		return nil, errors.Annotate(err, "assign new DUTs to %v", d).Err()
	}
	dutsNeeded := uint32ToInt(li.GetDutCapacity()) - len(currentDUTs)
	newDUTs, err := GetUnassignedDUTs(ctx, int32(dutsNeeded))
	if err != nil {
		return nil, errors.Annotate(err, "assign new DUTs to %v", d).Err()
	}
	logging.Infof(ctx, "Got unassigned DUTs to assign: %v", entities.FormatDUTs(newDUTs))
	for _, dut := range newDUTs {
		dut.AssignedDrone = d
	}
	currentDUTs = append(currentDUTs, newDUTs...)
	if err := datastore.Put(ctx, newDUTs); err != nil {
		return nil, errors.Annotate(err, "assign new DUTs to %v", d).Err()
	}
	return currentDUTs, nil
}

// uint32ToInt converts a uint32 to an int.  In case of overflow, panic.
func uint32ToInt(a uint32) int {
	b := int(a)
	if b < 0 {
		panic(a)
	}
	return b
}
