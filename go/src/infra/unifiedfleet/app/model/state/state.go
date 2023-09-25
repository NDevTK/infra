// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package state

import (
	"context"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/gae/service/datastore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	ufspb "infra/unifiedfleet/api/v1/models"
	ufsds "infra/unifiedfleet/app/model/datastore"
	"infra/unifiedfleet/app/util"
)

// RecordKind is the datastore entity kind of state.
const RecordKind string = "State"

// RecordEntity is a datastore entity that tracks dhcp.
type RecordEntity struct {
	_kind  string                `gae:"$kind,State"`
	_extra datastore.PropertyMap `gae:"-,extra"`
	// refer to the hostname
	ResourceName string `gae:"$id"`
	ResourceType string `gae:"resource_type"`
	State        string `gae:"state"`
	// ufspb.StateRecord cannot be directly used as it contains pointer (timestamp).
	StateRecord []byte `gae:",noindex"`
}

// GetProto returns the unmarshaled DHCP.
func (e *RecordEntity) GetProto() (proto.Message, error) {
	var p ufspb.StateRecord
	if err := proto.Unmarshal(e.StateRecord, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func newRecordEntity(ctx context.Context, pm proto.Message) (ufsds.FleetEntity, error) {
	p := pm.(*ufspb.StateRecord)
	if p.GetResourceName() == "" {
		return nil, errors.Reason("Empty resource name in state record").Err()
	}
	s, err := proto.Marshal(p)
	if err != nil {
		return nil, errors.Annotate(err, "fail to marshal state record %s", p).Err()
	}
	return &RecordEntity{
		ResourceName: p.GetResourceName(),
		ResourceType: util.GetPrefix(p.GetResourceName()),
		State:        p.GetState().String(),
		StateRecord:  s,
	}, nil
}

// GetStateRecord returns the state for a given resource name.
func GetStateRecord(ctx context.Context, id string) (*ufspb.StateRecord, error) {
	pm, err := ufsds.Get(ctx, &ufspb.StateRecord{ResourceName: id}, newRecordEntity)
	if err == nil {
		return pm.(*ufspb.StateRecord), err
	}
	return nil, err
}

// UpdateStateRecord updates a state record in datastore.
func UpdateStateRecord(ctx context.Context, stateRecord *ufspb.StateRecord) (*ufspb.StateRecord, error) {
	stateRecord.UpdateTime = ptypes.TimestampNow()
	pm, err := ufsds.PutSingle(ctx, stateRecord, newRecordEntity)
	if err == nil {
		return pm.(*ufspb.StateRecord), err
	}
	return nil, err
}

// ListStateRecords lists all the states
func ListStateRecords(ctx context.Context, pageSize int32, pageToken string, filterMap map[string][]interface{}) (res []*ufspb.StateRecord, nextPageToken string, err error) {
	q, err := ufsds.ListQuery(ctx, RecordKind, pageSize, pageToken, filterMap, false)
	if err != nil {
		return nil, "", err
	}
	var nextCur datastore.Cursor
	err = datastore.Run(ctx, q, func(ent *RecordEntity, cb datastore.CursorCB) error {
		pm, err := ent.GetProto()
		if err != nil {
			logging.Errorf(ctx, "Failed to UnMarshal: %s", err)
			return nil
		}
		res = append(res, pm.(*ufspb.StateRecord))
		if len(res) >= int(pageSize) {
			if nextCur, err = cb(); err != nil {
				return err
			}
			return datastore.Stop
		}
		return nil
	})
	if err != nil {
		logging.Errorf(ctx, "Failed to List state records %s", err)
		return nil, "", status.Errorf(codes.Internal, ufsds.InternalError)
	}
	if nextCur != nil {
		nextPageToken = nextCur.String()
	}
	return
}

// ImportStateRecords creates or updates a batch of state records in datastore
func ImportStateRecords(ctx context.Context, states []*ufspb.StateRecord) (*ufsds.OpResults, error) {
	protos := make([]proto.Message, len(states))
	utime := ptypes.TimestampNow()
	for i, m := range states {
		m.UpdateTime = utime
		protos[i] = m
	}
	return ufsds.Insert(ctx, protos, newRecordEntity, true, true)
}

func queryAllState(ctx context.Context) ([]ufsds.FleetEntity, error) {
	var entities []*RecordEntity
	q := datastore.NewQuery(RecordKind)
	if err := datastore.GetAll(ctx, q, &entities); err != nil {
		return nil, err
	}
	fe := make([]ufsds.FleetEntity, len(entities))
	for i, e := range entities {
		fe[i] = e
	}
	return fe, nil
}

// GetAllStates returns all states in datastore.
func GetAllStates(ctx context.Context) (*ufsds.OpResults, error) {
	return ufsds.GetAll(ctx, queryAllState)
}

// DeleteStates deletes a batch of states
func DeleteStates(ctx context.Context, resourceNames []string) *ufsds.OpResults {
	protos := make([]proto.Message, len(resourceNames))
	for i, m := range resourceNames {
		protos[i] = &ufspb.StateRecord{
			ResourceName: m,
		}
	}
	return ufsds.DeleteAll(ctx, protos, newRecordEntity)
}

// GetStateIndexedFieldName returns the index name
func GetStateIndexedFieldName(input string) (string, error) {
	var field string
	input = strings.TrimSpace(input)
	switch strings.ToLower(input) {
	case util.StateFilterName:
		field = "state"
	case util.ResourceTypeFilterName:
		field = "resource_type"
	default:
		return "", status.Errorf(codes.InvalidArgument, "Invalid field name %s - field name for state record are state/resourcetype", input)
	}
	return field, nil
}

// BatchUpdateStates updates the states to UFS.
//
// This can be used inside a transaction
func BatchUpdateStates(ctx context.Context, states []*ufspb.StateRecord) ([]*ufspb.StateRecord, error) {
	protos := make([]proto.Message, len(states))
	utime := ptypes.TimestampNow()
	for i, s := range states {
		s.UpdateTime = utime
		protos[i] = s
	}
	_, err := ufsds.PutAll(ctx, protos, newRecordEntity, true)
	if err == nil {
		return states, err
	}
	return nil, err
}
