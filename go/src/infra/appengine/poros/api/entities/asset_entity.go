// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package entities

import (
	"go.chromium.org/luci/gae/service/datastore"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

type AssetEntity struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Unique identifier of the asset
	AssetId string `protobuf:"bytes,1,opt,name=asset_id,json=assetId,proto3" json:"asset_id,omitempty" gae:"$id"`
	// Name of the asset
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	// Description of the asset
	Description string `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	// User who created the record.
	CreatedBy string `protobuf:"bytes,4,opt,name=created_by,json=createdBy,proto3" json:"created_by,omitempty"`
	// Timestamp for the creation of the record.
	CreatedAt *timestamppb.Timestamp `protobuf:"bytes,5,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	// Timestamp for the last update of the record.
	ModifiedAt *timestamppb.Timestamp `protobuf:"bytes,6,opt,name=modified_at,json=modifiedAt,proto3" json:"modified_at,omitempty"`
	// User who modified the record.
	ModifiedBy string `protobuf:"bytes,7,opt,name=modified_by,json=modifiedBy,proto3" json:"modified_by,omitempty"`

	Parent *datastore.Key `gae:"$parent"`
}
