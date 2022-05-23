// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package service

import (
	"context"
	"errors"
	"os"

	. "infra/appengine/poros/api/entities"
	proto "infra/appengine/poros/api/proto"

	"github.com/google/uuid"
	"go.chromium.org/luci/gae/service/datastore"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

type AssetHandler struct {
	proto.UnimplementedAssetServer
}

func toEntity(model *proto.AssetModel, ancestorKey *datastore.Key) *AssetEntity {
	if model != nil {
		return &AssetEntity{
			AssetId:     model.AssetId,
			Name:        model.Name,
			Description: model.Description,
			CreatedAt:   model.CreatedAt,
			CreatedBy:   model.CreatedBy,
			ModifiedAt:  model.ModifiedAt,
			ModifiedBy:  model.ModifiedBy,
			Parent:      ancestorKey,
		}
	}
	return nil
}

func toModel(entity *AssetEntity) *proto.AssetModel {
	if entity != nil {
		return &proto.AssetModel{
			AssetId:     entity.AssetId,
			Name:        entity.Name,
			Description: entity.Description,
			CreatedAt:   entity.CreatedAt,
			CreatedBy:   entity.CreatedBy,
			ModifiedAt:  entity.ModifiedAt,
			ModifiedBy:  entity.ModifiedBy,
		}
	}
	return nil
}

func validateEntity(entity *AssetEntity) error {
	// validate name & description
	if entity.Name == "" {
		return errors.New("name cannot be empty")
	}
	if entity.Description == "" {
		return errors.New("description cannot be empty")
	}
	return nil
}

// A query in transaction requires to have Ancestor filter, see
// https://cloud.google.com/appengine/docs/standard/python/datastore/query-restrictions#queries_inside_transactions_must_include_ancestor_filters
func fakeAncestorKey(ctx context.Context) *datastore.Key {
	return datastore.MakeKey(ctx, "AssetEntity", "key")
}

// Creates the given Asset.
func (e *AssetHandler) Create(ctx context.Context, req *proto.CreateAssetRequest) (*proto.AssetModel, error) {
	id := uuid.New().String()
	currentTime := timestamppb.Now()
	entity := &AssetEntity{
		AssetId:     id,
		Name:        req.GetName(),
		Description: req.GetDescription(),
		CreatedAt:   currentTime,
		Parent:      fakeAncestorKey(ctx),
	}
	if err := validateEntity(entity); err != nil {
		return nil, err
	}
	if err := datastore.Put(ctx, entity); err != nil {
		return nil, err
	}
	return toModel(entity), nil
}

// Retrieves a Asset for a given unique value.
func (e *AssetHandler) Get(ctx context.Context, req *proto.GetAssetRequest) (*proto.AssetModel, error) {
	entity, err := getById(ctx, req.GetAssetId())
	if err == nil {
		return toModel(entity), nil
	}
	return nil, err
}

// Update a single asset in Enterprise Asset.
func (e *AssetHandler) Update(ctx context.Context, req *proto.UpdateAssetRequest) (*proto.AssetModel, error) {
	id := req.GetAsset().GetAssetId()
	mask := req.GetUpdateMask()
	asset := &AssetEntity{}

	if mask == nil || len(mask.GetPaths()) == 0 || !mask.IsValid(req.GetAsset()) {
		return nil, errors.New("Update Mask can't be empty or invalid")
	}
	// In a transaction load asset, set fields based on field mask.
	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		asset, err := getById(ctx, id)
		if err != nil {
			return err
		}
		asset.ModifiedAt = timestamppb.Now()
		err = datastore.Put(ctx, id, &asset)
		return err
	}, nil)

	if err != nil {
		return toModel(asset), nil
	}
	return nil, err
}

// Deletes the given Asset.
func (e *AssetHandler) Delete(ctx context.Context, req *proto.DeleteAssetRequest) (*emptypb.Empty, error) {
	if err := datastore.Delete(ctx, &AssetEntity{
		AssetId: req.GetAssetId(),
		Parent:  fakeAncestorKey(ctx)}); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// Lists all Assets.
func (e *AssetHandler) List(ctx context.Context, in *proto.ListAssetsRequest) (*proto.ListAssetsResponse, error) {
	// TODO: crbug/1318606 - Implement Asset List functionality with filter,
	// orderby & paging.
	query := datastore.NewQuery("AssetEntity").Ancestor(fakeAncestorKey(ctx))
	var assetEntities []*AssetEntity
	res := &proto.ListAssetsResponse{}

	if err := datastore.GetAll(ctx, query, &assetEntities); err != nil {
		return nil, err
	}
	for _, asset := range assetEntities {
		res.Assets = append(res.Assets, toModel(asset))
	}
	return res, nil
}

func (c *AssetHandler) GetAssetConfiguration(ctx context.Context, in *proto.GetAssetConfigurationRequest) (*proto.GetAssetConfigurationResponse, error) {
	// TODO: crbug/1322948 - Static file used for implementing & testing the auth
	// in go binary, this method will be replaced by actual asset config from db
	data, err := os.ReadFile("./connector_test.asset.textpb")
	if err != nil {
		return nil, err
	}
	return &proto.GetAssetConfigurationResponse{Config: string(data)}, nil
}

func getById(ctx context.Context, id string) (*AssetEntity, error) {
	asset := &AssetEntity{AssetId: id, Parent: fakeAncestorKey(ctx)}
	if err := datastore.Get(ctx, asset); err != nil {
		return nil, err
	}
	return asset, nil
}
