// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package service

import (
	"context"
	"errors"
	"reflect"
	"time"

	. "infra/appengine/poros/api/entities"
	proto "infra/appengine/poros/api/proto"

	"github.com/google/uuid"
	"go.chromium.org/luci/gae/service/datastore"
	"go.chromium.org/luci/server/auth"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

type AssetInstanceHandler struct {
	proto.UnimplementedAssetInstanceServer
}

func toAssetInstanceEntity(model *proto.AssetInstanceModel) *AssetInstanceEntity {
	if model != nil {
		return &AssetInstanceEntity{
			AssetInstanceId: model.AssetInstanceId,
			AssetId:         model.AssetId,
			Status:          model.Status,
			CreatedAt:       model.CreatedAt.AsTime(),
			CreatedBy:       model.CreatedBy,
			ModifiedAt:      model.ModifiedAt.AsTime(),
			ModifiedBy:      model.ModifiedBy,
		}
	}
	return nil
}

func toAssetIntanceModel(entity *AssetInstanceEntity) *proto.AssetInstanceModel {
	if entity != nil {
		return &proto.AssetInstanceModel{
			AssetInstanceId: entity.AssetInstanceId,
			AssetId:         entity.AssetId,
			Status:          entity.Status,
			CreatedAt:       timestamppb.New(entity.CreatedAt),
			CreatedBy:       entity.CreatedBy,
			ModifiedAt:      timestamppb.New(entity.ModifiedAt),
			ModifiedBy:      entity.ModifiedBy,
		}
	}
	return nil
}

func validateAssetInstanceEntity(entity *AssetInstanceEntity) error {
	// validate AssetId and Status
	if entity.AssetId == "" {
		return errors.New("assetId cannot be empty")
	}
	if _, validStatus := proto.DeploymentStatus_value[entity.Status]; !validStatus {
		return errors.New("status is invalid")
	}
	return nil
}

// Creates the given AssetInstance.
func (e *AssetInstanceHandler) Create(ctx context.Context, req *proto.CreateAssetInstanceRequest) (*proto.AssetInstanceModel, error) {
	id := uuid.New().String()
	entity := &AssetInstanceEntity{
		AssetInstanceId: id,
		AssetId:         req.GetAssetId(),
		Status:          req.GetStatus(),
		CreatedBy:       auth.CurrentUser(ctx).Email,
		CreatedAt:       time.Now().UTC(),
	}
	if err := validateAssetInstanceEntity(entity); err != nil {
		return nil, err
	}
	if err := datastore.Put(ctx, entity); err != nil {
		return nil, err
	}
	return toAssetIntanceModel(entity), nil
}

// Retrieves an AssetInstance for a given unique value.
func (e *AssetInstanceHandler) Get(ctx context.Context, req *proto.GetAssetInstanceRequest) (*proto.AssetInstanceModel, error) {
	entity, err := getAssetInstanceById(ctx, req.GetAssetInstanceId())
	if err == nil {
		return toAssetIntanceModel(entity), nil
	}
	return nil, err
}

// Updates a single AssetInstance.
func (e *AssetInstanceHandler) Update(ctx context.Context, req *proto.UpdateAssetInstanceRequest) (*proto.AssetInstanceModel, error) {
	id := req.GetAssetInstance().GetAssetInstanceId()
	mask := req.GetUpdateMask()
	asset_instance := &AssetInstanceEntity{}

	if mask == nil || len(mask.GetPaths()) == 0 || !mask.IsValid(req.GetAssetInstance()) {
		return nil, errors.New("Update Mask can't be empty or invalid")
	}
	// In a transaction load AssetInstance, set fields based on field mask.
	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		asset_instance, err := getAssetInstanceById(ctx, id)
		if err != nil {
			return err
		}

		// Set updated values for fields specified in Update Mask
		for _, field := range mask.GetPaths() {
			newValue := reflect.ValueOf(req.GetAssetInstance()).Elem().FieldByName(snakeToPascalCase(field))
			reflect.ValueOf(asset_instance).Elem().FieldByName(snakeToPascalCase(field)).Set(newValue)
		}
		asset_instance.ModifiedBy = auth.CurrentUser(ctx).Email
		asset_instance.ModifiedAt = time.Now().UTC()

		if err := validateAssetInstanceEntity(asset_instance); err != nil {
			return err
		}

		err = datastore.Put(ctx, asset_instance)
		return err
	}, nil)

	if err == nil {
		return toAssetIntanceModel(asset_instance), nil
	}
	return nil, err
}

// Deletes the given AssetInstance.
func (e *AssetInstanceHandler) Delete(ctx context.Context, req *proto.DeleteAssetInstanceRequest) (*emptypb.Empty, error) {
	if err := datastore.Delete(ctx, &AssetInstanceEntity{
		AssetInstanceId: req.GetAssetInstanceId()}); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// Lists all AssetInstance.
func (e *AssetInstanceHandler) List(ctx context.Context, in *proto.ListAssetInstancesRequest) (*proto.ListAssetInstancesResponse, error) {
	query := datastore.NewQuery("AssetInstanceEntity")
	var asset_instances []*AssetInstanceEntity
	res := &proto.ListAssetInstancesResponse{}
	if err := datastore.GetAll(ctx, query, &asset_instances); err != nil {
		return nil, err
	}
	for _, asset_instance := range asset_instances {
		res.AssetInstances = append(res.AssetInstances, toAssetIntanceModel(asset_instance))
	}
	return res, nil
}

func getAssetInstanceById(ctx context.Context, id string) (*AssetInstanceEntity, error) {
	asset_instance := &AssetInstanceEntity{AssetInstanceId: id}
	if err := datastore.Get(ctx, asset_instance); err != nil {
		return nil, err
	}
	return asset_instance, nil
}
