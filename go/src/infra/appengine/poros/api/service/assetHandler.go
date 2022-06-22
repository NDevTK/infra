// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package service

import (
	"context"
	"encoding/json"
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

type AssetHandler struct {
	proto.UnimplementedAssetServer
}

func toEntity(model *proto.AssetModel, ancestorKey *datastore.Key) *AssetEntity {
	if model != nil {
		return &AssetEntity{
			AssetId:     model.AssetId,
			Name:        model.Name,
			AssetType:   model.AssetType,
			Description: model.Description,
			CreatedAt:   model.CreatedAt.AsTime(),
			CreatedBy:   model.CreatedBy,
			ModifiedAt:  model.ModifiedAt.AsTime(),
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
			AssetType:   entity.AssetType,
			Description: entity.Description,
			CreatedAt:   timestamppb.New(entity.CreatedAt),
			CreatedBy:   entity.CreatedBy,
			ModifiedAt:  timestamppb.New(entity.ModifiedAt),
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
	if entity.AssetType == "" {
		return errors.New("type cannot be empty")
	}
	return nil
}

// A query in transaction requires to have Ancestor filter, see
// https://cloud.google.com/appengine/docs/standard/python/datastore/query-restrictions#queries_inside_transactions_must_include_ancestor_filters
func fakeAncestorKey(ctx context.Context) *datastore.Key {
	return datastore.MakeKey(ctx, "AssetEntity", "key")
}

// Creates the given Asset.
func (e *AssetHandler) Create(ctx context.Context, req *proto.CreateAssetRequest) (*proto.CreateAssetResponse, error) {
	id := uuid.New().String()
	entity := &AssetEntity{
		AssetId:     id,
		Name:        req.GetName(),
		Description: req.GetDescription(),
		AssetType:   req.GetAssetType(),
		CreatedBy:   auth.CurrentUser(ctx).Email,
		CreatedAt:   time.Now().UTC(),
		Parent:      fakeAncestorKey(ctx),
	}
	response := &proto.CreateAssetResponse{}

	var defaultResourcesToSave []*proto.AssetResourceModel
	defaultResourcesToSave, err := defaultResources(ctx, entity.AssetType)
	if err != nil {
		return nil, err
	}

	err = datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		if err := validateEntity(entity); err != nil {
			return err
		}
		if err := datastore.Put(ctx, entity); err != nil {
			return err
		}
		response.Asset = toModel(entity)
		assetResourcesToSave := req.GetAssetResourcesToSave()
		for _, defaultResource := range defaultResourcesToSave {
			assetResourcesToSave = append(assetResourcesToSave, defaultResource)
		}

		for _, assetResourceModel := range assetResourcesToSave {
			assetResourceModel.AssetResourceId = uuid.New().String()
			assetResourceModel.AssetId = id
			assetResourceEntity := toAssetResourceEntity(assetResourceModel)
			if err := validateAssetResourceEntity(assetResourceEntity); err != nil {
				return err
			}
			if err := datastore.Put(ctx, assetResourceEntity); err != nil {
				return err
			}
			response.AssetResources = append(response.AssetResources, assetResourceModel)
		}
		// If no error occurs, return nil
		return nil
	}, nil)

	if err != nil {
		return nil, err
	}
	return response, nil
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
func (e *AssetHandler) Update(ctx context.Context, req *proto.UpdateAssetRequest) (*proto.UpdateAssetResponse, error) {
	assetId := req.GetAsset().GetAssetId()
	assetmask := req.GetAssetUpdateMask()
	response := &proto.UpdateAssetResponse{}

	if assetmask == nil || len(assetmask.GetPaths()) == 0 || !assetmask.IsValid(req.GetAsset()) {
		return nil, errors.New("Update Mask for Asset can't be empty or invalid")
	}
	// In a transaction load asset, set fields based on field mask.
	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		asset, err := getById(ctx, assetId)
		if err != nil {
			return err
		}
		// Set updated values for fields specified in Update Mask
		for _, field := range assetmask.GetPaths() {
			newValue := reflect.ValueOf(req.GetAsset()).Elem().FieldByName(snakeToPascalCase(field))
			reflect.ValueOf(asset).Elem().FieldByName(snakeToPascalCase(field)).Set(newValue)
		}

		asset.ModifiedBy = auth.CurrentUser(ctx).Email
		asset.ModifiedAt = time.Now().UTC()

		if err := validateEntity(asset); err != nil {
			return err
		}

		err = datastore.Put(ctx, asset)
		if err != nil {
			return err
		}

		response.Asset = toModel(asset)
		assetResourcesToSave := req.GetAssetResourcesToSave()
		for _, assetResourceModel := range assetResourcesToSave {
			if assetResourceModel.GetAssetResourceId() == "" {
				assetResourceModel.AssetResourceId = uuid.New().String()
				assetResourceModel.AssetId = assetId
				assetResourceEntity := toAssetResourceEntity(assetResourceModel)
				if err := validateAssetResourceEntity(assetResourceEntity); err != nil {
					return err
				}
				if err := datastore.Put(ctx, assetResourceEntity); err != nil {
					return err
				}
				response.AssetResources = append(response.AssetResources, assetResourceModel)
			} else {
				assetResourceId := assetResourceModel.GetAssetResourceId()
				assetResourcemask := req.GetAssetResourceUpdateMask()

				if assetResourcemask == nil || len(assetResourcemask.GetPaths()) == 0 || !assetResourcemask.IsValid(assetResourceModel) {
					return errors.New("Update Mask for AssetResource can't be empty or invalid")
				}
				assetResource, err := getByAssetResourceId(ctx, assetResourceId)
				if err != nil {
					return err
				}
				for _, field := range assetResourcemask.GetPaths() {
					newValue := reflect.ValueOf(assetResourceModel).Elem().FieldByName(snakeToPascalCase(field))
					reflect.ValueOf(assetResource).Elem().FieldByName(snakeToPascalCase(field)).Set(newValue)
				}

				assetResource.ModifiedBy = auth.CurrentUser(ctx).Email
				assetResource.ModifiedAt = time.Now().UTC()

				if err := validateAssetResourceEntity(assetResource); err != nil {
					return err
				}
				if err := datastore.Put(ctx, assetResource); err != nil {
					return err
				}
				response.AssetResources = append(response.AssetResources, assetResourceModel)
			}
		}

		assetResourcesToDelete := req.GetAssetResourcesToDelete()
		for _, assetResourceModel := range assetResourcesToDelete {
			if err := datastore.Delete(ctx, &AssetResourceEntity{
				AssetResourceId: assetResourceModel.GetAssetResourceId()}); err != nil {
				return err
			}
		}
		// If no error occurs, return nil
		return nil
	}, nil)

	if err != nil {
		return nil, err
	}

	return response, nil

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

// Given the asset Id get all the information required to generate asset configuration
// file in the go-binary for system testing
func (c *AssetHandler) GetAssetConfiguration(ctx context.Context, in *proto.GetAssetConfigurationRequest) (*proto.GetAssetConfigurationResponse, error) {
	assetId := in.GetAssetId()
	asset, err := getById(ctx, assetId)
	if err != nil {
		return nil, err
	}
	res := &proto.AssetConfiguration{AssetId: assetId, AssetType: asset.AssetType}

	query := datastore.NewQuery("AssetResourceEntity").Eq("AssetId", assetId)
	var assetResourceEntites []*AssetResourceEntity
	if err := datastore.GetAll(ctx, query, &assetResourceEntites); err != nil {
		return nil, err
	}

	for _, assetResource := range assetResourceEntites {
		resource, err := getResourceById(ctx, assetResource.ResourceId)
		if err != nil {
			return nil, err
		}

		res.Resources = append(
			res.Resources,
			&proto.AssetConfigurationResource{
				ResourceId:      resource.ResourceId,
				Description:     resource.Description,
				OperatingSystem: resource.OperatingSystem,
				AliasName:       assetResource.AliasName,
				ResourceType:    resource.Type,
				ResourceName:    resource.Name})
	}
	jsonBytes, _ := json.MarshalIndent(res, "", "    ")

	return &proto.GetAssetConfigurationResponse{Config: string(jsonBytes)}, nil
}

// Given the resource Ids get all the information required to generate host configuration
// file in the go-binary for system testing
func (c *AssetHandler) GetHostConfiguration(ctx context.Context, in *proto.GetHostConfigurationRequest) (*proto.GetHostConfigurationResponse, error) {
	resourceIds := in.ResourceIds

	res := &proto.HostConfiguration{}

	for _, resourceId := range resourceIds {
		resource, err := getResourceById(ctx, resourceId)
		if err != nil {
			return nil, err
		}

		res.Resources = append(res.Resources, &proto.HostConfigurationResource{ResourceId: resource.ResourceId, ResourceName: resource.Name, ResourceImage: resource.Image})
	}
	jsonBytes, _ := json.MarshalIndent(res, "", "    ")

	return &proto.GetHostConfigurationResponse{Config: string(jsonBytes)}, nil
}

// Based on different asset types there may be additional resources that must be created by default
// necessary to generate the aasset & host configuration files in the go-binary for system tests
func defaultResources(ctx context.Context, assetType string) ([]*proto.AssetResourceModel, error) {
	var resourceData [][]string
	switch assetType {
	case "active_directory":
		resourceData = [][]string{
			{"network", "primary"},
			{"ad_domain", "foo.example"},
			{"domain_controller_machine", "win2008r2"},
			{"user", "Joe"},
		}
		break
	default:
		resourceData = [][]string{}
	}

	var defaultResourcesToSave []*proto.AssetResourceModel
	for _, data := range resourceData {
		query := datastore.NewQuery("ResourceEntity").Eq("Type", data[0]).Limit(1)
		var resources []*ResourceEntity
		if err := datastore.GetAll(ctx, query, &resources); err != nil {
			return nil, err
		}

		defaultResourcesToSave = append(defaultResourcesToSave,
			&proto.AssetResourceModel{ResourceId: resources[0].ResourceId, AliasName: data[1]})
	}

	return defaultResourcesToSave, nil
}

func getById(ctx context.Context, id string) (*AssetEntity, error) {
	asset := &AssetEntity{AssetId: id, Parent: fakeAncestorKey(ctx)}
	if err := datastore.Get(ctx, asset); err != nil {
		return nil, err
	}
	return asset, nil
}
