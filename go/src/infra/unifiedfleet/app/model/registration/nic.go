// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package registration

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"go.chromium.org/gae/service/datastore"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	fleet "infra/unifiedfleet/api/v1/proto"
	fleetds "infra/unifiedfleet/app/model/datastore"
)

// NicKind is the datastore entity kind Nic.
const NicKind string = "Nic"

// NicEntity is a datastore entity that tnics Nic.
type NicEntity struct {
	_kind    string `gae:"$kind,Nic"`
	ID       string `gae:"$id"`
	SwitchID string `gae:"switch_id"`
	// fleet.Nic cannot be directly used as it contains pointer.
	Nic []byte `gae:",noindex"`
}

// GetProto returns the unmarshaled Nic.
func (e *NicEntity) GetProto() (proto.Message, error) {
	var p fleet.Nic
	if err := proto.Unmarshal(e.Nic, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func newNicEntity(ctx context.Context, pm proto.Message) (fleetds.FleetEntity, error) {
	p := pm.(*fleet.Nic)
	if p.GetName() == "" {
		return nil, errors.Reason("Empty Nic ID").Err()
	}
	nic, err := proto.Marshal(p)
	if err != nil {
		return nil, errors.Annotate(err, "fail to marshal Nic %s", p).Err()
	}
	return &NicEntity{
		ID:       p.GetName(),
		SwitchID: p.GetSwitchInterface().GetSwitch(),
		Nic:      nic,
	}, nil
}

// CreateNic creates a new nic in datastore.
func CreateNic(ctx context.Context, nic *fleet.Nic) (*fleet.Nic, error) {
	return putNic(ctx, nic, false)
}

// UpdateNic updates nic in datastore.
func UpdateNic(ctx context.Context, nic *fleet.Nic) (*fleet.Nic, error) {
	return putNic(ctx, nic, true)
}

// GetNic returns nic for the given id from datastore.
func GetNic(ctx context.Context, id string) (*fleet.Nic, error) {
	pm, err := fleetds.Get(ctx, &fleet.Nic{Name: id}, newNicEntity)
	if err == nil {
		return pm.(*fleet.Nic), err
	}
	return nil, err
}

// QueryNicByPropertyName query's Nic Entity in the datastore
//
// If keysOnly is true, then only key field is populated in returned nics
func QueryNicByPropertyName(ctx context.Context, propertyName, id string, keysOnly bool) ([]*fleet.Nic, error) {
	q := datastore.NewQuery(NicKind).KeysOnly(keysOnly)
	var entities []*NicEntity
	if err := datastore.GetAll(ctx, q.Eq(propertyName, id), &entities); err != nil {
		logging.Errorf(ctx, "Failed to query from datastore: %s", err)
		return nil, status.Errorf(codes.Internal, fleetds.InternalError)
	}
	if len(entities) == 0 {
		logging.Infof(ctx, "No nics found for the query: %s", id)
		return nil, nil
	}
	nics := make([]*fleet.Nic, 0, len(entities))
	for _, entity := range entities {
		if keysOnly {
			nics = append(nics, &fleet.Nic{
				Name: entity.ID,
			})
		} else {
			pm, perr := entity.GetProto()
			if perr != nil {
				logging.Errorf(ctx, "Failed to unmarshal proto: %s", perr)
				continue
			}
			nics = append(nics, pm.(*fleet.Nic))
		}
	}
	return nics, nil
}

// ListNics lists the nics
//
// Does a query over Nic entities. Returns up to pageSize entities, plus non-nil cursor (if
// there are more results). pageSize must be positive.
func ListNics(ctx context.Context, pageSize int32, pageToken string) (res []*fleet.Nic, nextPageToken string, err error) {
	q, err := fleetds.ListQuery(ctx, NicKind, pageSize, pageToken)
	if err != nil {
		return nil, "", err
	}
	var nextCur datastore.Cursor
	err = datastore.Run(ctx, q, func(ent *NicEntity, cb datastore.CursorCB) error {
		pm, err := ent.GetProto()
		if err != nil {
			logging.Errorf(ctx, "Failed to UnMarshal: %s", err)
			return nil
		}
		res = append(res, pm.(*fleet.Nic))
		if len(res) >= int(pageSize) {
			if nextCur, err = cb(); err != nil {
				return err
			}
			return datastore.Stop
		}
		return nil
	})
	if err != nil {
		logging.Errorf(ctx, "Failed to List Nics %s", err)
		return nil, "", status.Errorf(codes.Internal, fleetds.InternalError)
	}
	if nextCur != nil {
		nextPageToken = nextCur.String()
	}
	return
}

// DeleteNic deletes the nic in datastore
func DeleteNic(ctx context.Context, id string) error {
	return fleetds.Delete(ctx, &fleet.Nic{Name: id}, newNicEntity)
}

func putNic(ctx context.Context, nic *fleet.Nic, update bool) (*fleet.Nic, error) {
	nic.UpdateTime = ptypes.TimestampNow()
	pm, err := fleetds.Put(ctx, nic, newNicEntity, update)
	if err == nil {
		return pm.(*fleet.Nic), err
	}
	return nil, err
}

// BatchUpdateNics updates nics in datastore.
//
// This is a non-atomic operation and doesnt check if the object already exists before
// update. Must be used within a Transaction where objects are checked before update.
// Will lead to partial updates if not used in a transaction.
func BatchUpdateNics(ctx context.Context, nics []*fleet.Nic) ([]*fleet.Nic, error) {
	return putAllNic(ctx, nics, true)
}

func putAllNic(ctx context.Context, nics []*fleet.Nic, update bool) ([]*fleet.Nic, error) {
	protos := make([]proto.Message, len(nics))
	updateTime := ptypes.TimestampNow()
	for i, nic := range nics {
		nic.UpdateTime = updateTime
		protos[i] = nic
	}
	_, err := fleetds.PutAll(ctx, protos, newNicEntity, update)
	if err == nil {
		return nics, err
	}
	return nil, err
}

// ImportNics creates or updates a batch of nics in datastore.
func ImportNics(ctx context.Context, nics []*fleet.Nic) (*fleetds.OpResults, error) {
	protos := make([]proto.Message, len(nics))
	utime := ptypes.TimestampNow()
	for i, m := range nics {
		m.UpdateTime = utime
		protos[i] = m
	}
	return fleetds.Insert(ctx, protos, newNicEntity, true, true)
}
