// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configuration

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

// RackLSEPrototypeKind is the datastore entity kind for RackLSEPrototypes.
const RackLSEPrototypeKind string = "RackLSEPrototype"

// RackLSEPrototypeEntity is a datastore entity that tracks a platform.
type RackLSEPrototypeEntity struct {
	_kind string   `gae:"$kind,RackLSEPrototype"`
	ID    string   `gae:"$id"`
	Tags  []string `gae:"tags"`
	// ufspb.RackLSEPrototype cannot be directly used as it contains pointer.
	RackLSEPrototype []byte `gae:",noindex"`
}

// GetProto returns the unmarshaled RackLSEPrototype.
func (e *RackLSEPrototypeEntity) GetProto() (proto.Message, error) {
	var p ufspb.RackLSEPrototype
	if err := proto.Unmarshal(e.RackLSEPrototype, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func newRackLSEPrototypeEntity(ctx context.Context, pm proto.Message) (ufsds.FleetEntity, error) {
	p := pm.(*ufspb.RackLSEPrototype)
	if p.GetName() == "" {
		return nil, errors.Reason("Empty RackLSEPrototype ID").Err()
	}
	rackLSEPrototype, err := proto.Marshal(p)
	if err != nil {
		return nil, errors.Annotate(err, "fail to marshal RackLSEPrototype %s", p).Err()
	}
	return &RackLSEPrototypeEntity{
		ID:               p.GetName(),
		RackLSEPrototype: rackLSEPrototype,
		Tags:             p.GetTags(),
	}, nil
}

// CreateRackLSEPrototype creates a new rackLSEPrototype in datastore.
func CreateRackLSEPrototype(ctx context.Context, rackLSEPrototype *ufspb.RackLSEPrototype) (*ufspb.RackLSEPrototype, error) {
	return putRackLSEPrototype(ctx, rackLSEPrototype, false)
}

// UpdateRackLSEPrototype updates rackLSEPrototype in datastore.
func UpdateRackLSEPrototype(ctx context.Context, rackLSEPrototype *ufspb.RackLSEPrototype) (*ufspb.RackLSEPrototype, error) {
	return putRackLSEPrototype(ctx, rackLSEPrototype, true)
}

// BatchUpdateRackLSEPrototypes updates rackLSEPrototype in datastore.
//
// This is a non-atomic operation and doesnt check if the object already exists before
// update. Must be used within a Transaction where objects are checked before update.
// Will lead to partial updates if not used in a transaction.
func BatchUpdateRackLSEPrototypes(ctx context.Context, prototypes []*ufspb.RackLSEPrototype) ([]*ufspb.RackLSEPrototype, error) {
	protos := make([]proto.Message, len(prototypes))
	updateTime := ptypes.TimestampNow()
	for i, p := range prototypes {
		p.UpdateTime = updateTime
		protos[i] = p
	}
	_, err := ufsds.PutAll(ctx, protos, newRackLSEPrototypeEntity, true)
	if err == nil {
		return prototypes, err
	}
	return nil, err
}

// GetRackLSEPrototype returns rackLSEPrototype for the given id from datastore.
func GetRackLSEPrototype(ctx context.Context, id string) (*ufspb.RackLSEPrototype, error) {
	pm, err := ufsds.Get(ctx, &ufspb.RackLSEPrototype{Name: id}, newRackLSEPrototypeEntity)
	if err == nil {
		return pm.(*ufspb.RackLSEPrototype), err
	}
	return nil, err
}

func getRackLSEPrototypeID(pm proto.Message) string {
	p := pm.(*ufspb.RackLSEPrototype)
	return p.GetName()
}

// BatchGetRackLSEPrototypes returns a batch of rack lse prototypes from datastore.
func BatchGetRackLSEPrototypes(ctx context.Context, ids []string) ([]*ufspb.RackLSEPrototype, error) {
	protos := make([]proto.Message, len(ids))
	for i, n := range ids {
		protos[i] = &ufspb.RackLSEPrototype{Name: n}
	}
	pms, err := ufsds.BatchGet(ctx, protos, newRackLSEPrototypeEntity, getRackLSEPrototypeID)
	if err != nil {
		return nil, err
	}
	res := make([]*ufspb.RackLSEPrototype, len(pms))
	for i, pm := range pms {
		res[i] = pm.(*ufspb.RackLSEPrototype)
	}
	return res, nil
}

// ListRackLSEPrototypes lists the rackLSEPrototypes
//
// Does a query over RackLSEPrototype entities. Returns up to pageSize entities, plus non-nil cursor (if
// there are more results). pageSize must be positive.
func ListRackLSEPrototypes(ctx context.Context, pageSize int32, pageToken string, filterMap map[string][]interface{}, keysOnly bool) (res []*ufspb.RackLSEPrototype, nextPageToken string, err error) {
	// Passing -1 for query limit fetches all the entities from the datastore
	q, err := ufsds.ListQuery(ctx, RackLSEPrototypeKind, -1, pageToken, filterMap, keysOnly)
	if err != nil {
		return nil, "", err
	}
	var nextCur datastore.Cursor
	err = datastore.Run(ctx, q, func(ent *RackLSEPrototypeEntity, cb datastore.CursorCB) error {
		if keysOnly {
			rackLSEPrototype := &ufspb.RackLSEPrototype{
				Name: ent.ID,
			}
			res = append(res, rackLSEPrototype)
		} else {
			pm, err := ent.GetProto()
			if err != nil {
				logging.Errorf(ctx, "Failed to UnMarshal: %s", err)
				return nil
			}
			res = append(res, pm.(*ufspb.RackLSEPrototype))
		}
		if len(res) >= int(pageSize) {
			if nextCur, err = cb(); err != nil {
				return err
			}
			return datastore.Stop
		}
		return nil
	})
	if err != nil {
		logging.Errorf(ctx, "Failed to List RackLSEPrototype %s", err)
		return nil, "", status.Errorf(codes.Internal, ufsds.InternalError)
	}
	if nextCur != nil {
		nextPageToken = nextCur.String()
	}
	return
}

// DeleteRackLSEPrototype deletes the rackLSEPrototype in datastore
func DeleteRackLSEPrototype(ctx context.Context, id string) error {
	return ufsds.Delete(ctx, &ufspb.RackLSEPrototype{Name: id}, newRackLSEPrototypeEntity)
}

func putRackLSEPrototype(ctx context.Context, rackLSEPrototype *ufspb.RackLSEPrototype, update bool) (*ufspb.RackLSEPrototype, error) {
	rackLSEPrototype.UpdateTime = ptypes.TimestampNow()
	pm, err := ufsds.Put(ctx, rackLSEPrototype, newRackLSEPrototypeEntity, update)
	if err == nil {
		return pm.(*ufspb.RackLSEPrototype), err
	}
	return nil, err
}

// ImportRackLSEPrototypes creates or updates a batch of rack lse prototypes in datastore
func ImportRackLSEPrototypes(ctx context.Context, lps []*ufspb.RackLSEPrototype) (*ufsds.OpResults, error) {
	protos := make([]proto.Message, len(lps))
	utime := ptypes.TimestampNow()
	for i, m := range lps {
		m.UpdateTime = utime
		protos[i] = m
	}
	return ufsds.Insert(ctx, protos, newRackLSEPrototypeEntity, true, true)
}

// GetRackLSEPrototypeIndexedFieldName returns the index name
func GetRackLSEPrototypeIndexedFieldName(input string) (string, error) {
	var field string
	input = strings.TrimSpace(input)
	switch strings.ToLower(input) {
	case util.TagFilterName:
		field = "tags"
	default:
		return "", status.Errorf(codes.InvalidArgument, "Invalid field name %s - field name for rackprototype are tag", input)
	}
	return field, nil
}
