// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package inventory

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/gae/service/datastore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	ufspb "infra/unifiedfleet/api/v1/models"
)

const (
	// OwnershipDataKind is the datastore entity kind OwnershipData.
	OwnershipDataKind string = "OwnershipData"

	// AssetType Machine
	AssetTypeMachine string = "Machine"

	// AssetType MachineLSE
	AssetTypeMachineLSE string = "MachineLSE"

	// AssetType VM
	AssetTypeVM string = "VM"
)

// OwnershipDataEntity is a datastore entity that tracks a OwnershipData.
type OwnershipDataEntity struct {
	_kind         string `gae:"$kind,Ownership"`
	Name          string `gae:"$id"`
	AssetType     string `gae:"AssetType"`
	OwnershipData []byte `gae:",noindex"`
}

// GetProto returns the unmarshaled OwnershipData.
func (e *OwnershipDataEntity) GetProto() (proto.Message, error) {
	var p ufspb.OwnershipData
	if err := proto.Unmarshal(e.OwnershipData, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// PutOwnershipData adds a bot name and its corresponding ownership in the datastore.
func PutOwnershipData(ctx context.Context, ownership *ufspb.OwnershipData, name string, assetType string) (*OwnershipDataEntity, error) {
	if name == "" {
		return nil, status.Errorf(codes.Internal, "Empty bot name")
	}
	ownershipBytes, err := proto.Marshal(ownership)
	if err != nil {
		return nil, errors.Annotate(err, "fail to marshal Ownership data %s", ownership).Err()
	}
	entity := &OwnershipDataEntity{
		Name:          name,
		OwnershipData: ownershipBytes,
		AssetType:     assetType,
	}
	if err := datastore.Put(ctx, entity); err != nil {
		logging.Errorf(ctx, "Failed to put ownership in datastore : %s - %s", name, err)
		return nil, err
	}
	return entity, nil
}

// GetOwnershipData returns OwnershipData for the given bot name from datastore.
func GetOwnershipData(ctx context.Context, name string) (*OwnershipDataEntity, error) {
	entity := &OwnershipDataEntity{
		Name: name,
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
