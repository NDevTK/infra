// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configuration

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/gae/service/datastore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	ufspb "infra/unifiedfleet/api/v1/models"
)

// HwidDataKind is the datastore entity kind HwidData.
const HwidDataKind string = "HwidData"

// HwidDataEntity is a datastore entity that tracks a HwidData.
type HwidDataEntity struct {
	_kind    string `gae:"$kind,HwidData"`
	ID       string `gae:"$id"`
	HwidData []byte `gae:",noindex"`
	Updated  time.Time
}

// GetProto returns the unmarshaled HwidData.
func (e *HwidDataEntity) GetProto() (proto.Message, error) {
	p := &ufspb.HwidData{}
	if err := proto.Unmarshal(e.HwidData, p); err != nil {
		return nil, err
	}
	return p, nil
}

// GetDutLabelProto returns the unmarshaled HwidData.
// TODO (b/240771930): Remove this method once datastore data is overwritten.
func (e *HwidDataEntity) GetDutLabelProto() (proto.Message, error) {
	p := &ufspb.DutLabel{}
	if err := proto.Unmarshal(e.HwidData, p); err != nil {
		return nil, err
	}
	return p, nil
}

// UpdateHwidData updates HwidData in datastore.
func UpdateHwidData(ctx context.Context, d *ufspb.HwidData, hwid string) (*HwidDataEntity, error) {
	hwidData, err := proto.Marshal(d)
	if err != nil {
		return nil, errors.Annotate(err, "failed to marshal HwidData %s", d).Err()
	}

	if hwid == "" {
		return nil, status.Errorf(codes.InvalidArgument, "empty hwid")
	}

	entity := &HwidDataEntity{
		ID:       hwid,
		HwidData: hwidData,
		Updated:  time.Now().UTC(),
	}
	if err := datastore.Put(ctx, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

// UpdateLegacyHwidData updates HwidDataEntity with DutLabel as HwidData instead
// of HwidData proto in datastore. This is also the previous implementation of
// UpdateHwidData.
func UpdateLegacyHwidData(ctx context.Context, d *ufspb.DutLabel, hwid string) (*HwidDataEntity, error) {
	dutLabel, err := proto.Marshal(d)
	if err != nil {
		return nil, errors.Annotate(err, "failed to marshal DutLabel %s", d).Err()
	}

	if hwid == "" {
		return nil, status.Errorf(codes.InvalidArgument, "empty hwid")
	}

	entity := &HwidDataEntity{
		ID:       hwid,
		HwidData: dutLabel,
		Updated:  time.Now().UTC(),
	}
	if err := datastore.Put(ctx, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

// GetHwidData returns HwidData for the given hwid from datastore.
func GetHwidData(ctx context.Context, hwid string) (*HwidDataEntity, error) {
	entity := &HwidDataEntity{
		ID: hwid,
	}

	if err := datastore.Get(ctx, entity); err != nil {
		if datastore.IsErrNoSuchEntity(err) {
			errorMsg := fmt.Sprintf("Entity not found %+v", entity)
			return nil, status.Errorf(codes.NotFound, errorMsg)
		}
		return nil, err
	}
	return entity, nil
}

// ParseHwidData returns the HwidData proto based on the datastore entity.
//
// It parses a given HwidDataEntity into the ufspb.HwidData proto containing
// all DutLabels. No error is returned if the entity is nil.
func ParseHwidData(ent *HwidDataEntity) (*ufspb.HwidData, error) {
	if ent == nil {
		return nil, nil
	}

	// Try to get HwidData proto. Try DutLabel proto if fail.
	// TODO (b/240771930): Remove DutLabel conditional once datastore data is
	// overwritten.
	entData, err := ent.GetProto()
	if err != nil {
		return nil, err
	}
	hwidData, ok := entData.(*ufspb.HwidData)
	if !ok {
		return nil, errors.Reason("Failed to cast data to HwidData: %s", entData).Err()
	}

	// Unmarshaling into the wrong type of proto will not fail. So must check if
	// data actually has DutLabel. The correct proto type will always have a
	// DutLabel field after unmarshaling.
	if hwidData.GetDutLabel() == nil {
		data := &ufspb.HwidData{
			Hwid: ent.ID,
		}
		entData, err = ent.GetDutLabelProto()
		if err != nil {
			return nil, err
		}
		dutLabel, ok := entData.(*ufspb.DutLabel)
		if !ok {
			return nil, errors.Reason("Failed to cast data to DutLabel: %s", entData).Err()
		}
		data.DutLabel = dutLabel
		return SetHwidDataWithDutLabels(data), nil
	}
	return hwidData, nil
}

// SetHwidDataWithDutLabels sets the Sku and Variant of HwidData proto.
//
// It parses the DutLabels of a given HwidData entity and fills in the Sku and
// Variant values of the proto.
func SetHwidDataWithDutLabels(hwidData *ufspb.HwidData) *ufspb.HwidData {
	for _, l := range hwidData.GetDutLabel().GetLabels() {
		switch strings.ToLower(l.GetName()) {
		case "sku":
			hwidData.Sku = l.GetValue()
		case "variant":
			hwidData.Variant = l.GetValue()
		}
	}
	return hwidData
}
