// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package service

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"time"

	. "infra/appengine/poros/api/entities"
	proto "infra/appengine/poros/api/proto"

	"github.com/google/uuid"
	"go.chromium.org/luci/gae/service/datastore"
	"go.chromium.org/luci/server/auth"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

type ResourceHandler struct {
	proto.UnimplementedResourceServer
}

func toResourceEntity(model *proto.ResourceModel) *ResourceEntity {
	if model != nil {
		return &ResourceEntity{
			ResourceId:  model.ResourceId,
			Name:        model.Name,
			Description: model.Description,
			Type:        model.Type,
			Image:       model.Image,
			CreatedAt:   model.CreatedAt.AsTime(),
			CreatedBy:   model.CreatedBy,
			ModifiedAt:  model.ModifiedAt.AsTime(),
			ModifiedBy:  model.ModifiedBy,
		}
	}
	return nil
}

func toResourceModel(entity *ResourceEntity) *proto.ResourceModel {
	if entity != nil {
		return &proto.ResourceModel{
			ResourceId:  entity.ResourceId,
			Name:        entity.Name,
			Description: entity.Description,
			Type:        entity.Type,
			Image:       entity.Image,
			CreatedAt:   timestamppb.New(entity.CreatedAt),
			CreatedBy:   entity.CreatedBy,
			ModifiedAt:  timestamppb.New(entity.ModifiedAt),
			ModifiedBy:  entity.ModifiedBy,
		}
	}
	return nil
}

func validateResourceEntity(entity *ResourceEntity) error {
	// validate name, description, type, Image
	if entity.Name == "" {
		return errors.New("name cannot be empty")
	}
	if entity.Description == "" {
		return errors.New("description cannot be empty")
	}
	if entity.Type == "" {
		return errors.New("type cannot be empty")
	}
	if entity.Type == "machine" && entity.Image == "" {
		return errors.New("VM Image needs to be specified")
	}
	return nil
}

// Creates the given Resource.
func (e *ResourceHandler) Create(ctx context.Context, req *proto.CreateResourceRequest) (*proto.ResourceModel, error) {
	id := uuid.New().String()
	entity := &ResourceEntity{
		ResourceId:  id,
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Type:        req.GetType(),
		Image:       req.GetImage(),
		CreatedBy:   auth.CurrentUser(ctx).Email,
		CreatedAt:   time.Now().UTC(),
	}
	if err := validateResourceEntity(entity); err != nil {
		return nil, err
	}
	if err := datastore.Put(ctx, entity); err != nil {
		return nil, err
	}
	return toResourceModel(entity), nil
}

// Retrieves a Resource for a given unique value.
func (e *ResourceHandler) Get(ctx context.Context, req *proto.GetResourceRequest) (*proto.ResourceModel, error) {
	entity, err := getResourceById(ctx, req.GetResourceId())
	if err == nil {
		return toResourceModel(entity), nil
	}
	return nil, err
}

// Updates a single resource.
func (e *ResourceHandler) Update(ctx context.Context, req *proto.UpdateResourceRequest) (*proto.ResourceModel, error) {
	id := req.GetResource().GetResourceId()
	mask := req.GetUpdateMask()
	resource := &ResourceEntity{}

	if mask == nil || len(mask.GetPaths()) == 0 || !mask.IsValid(req.GetResource()) {
		return nil, errors.New("Update Mask can't be empty or invalid")
	}
	// In a transaction load resource, set fields based on field mask.
	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		resource, err := getResourceById(ctx, id)
		if err != nil {
			return err
		}

		// Set updated values for fields specified in Update Mask
		for _, field := range mask.GetPaths() {
			newValue := reflect.ValueOf(req.GetResource()).Elem().FieldByName(strings.Title(field))
			reflect.ValueOf(resource).Elem().FieldByName(strings.Title(field)).Set(newValue)
		}

		resource.ModifiedBy = auth.CurrentUser(ctx).Email
		resource.ModifiedAt = time.Now().UTC()

		if err := validateResourceEntity(resource); err != nil {
			return err
		}

		err = datastore.Put(ctx, resource)
		return err
	}, nil)

	if err == nil {
		return toResourceModel(resource), nil
	}
	return nil, err
}

// Deletes the given Resource.
func (e *ResourceHandler) Delete(ctx context.Context, req *proto.DeleteResourceRequest) (*emptypb.Empty, error) {
	if err := datastore.Delete(ctx, &ResourceEntity{
		ResourceId: req.GetResourceId()}); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// Lists all Resources.
func (e *ResourceHandler) List(ctx context.Context, in *proto.ListResourcesRequest) (*proto.ListResourcesResponse, error) {
	query := datastore.NewQuery("ResourceEntity")
	var resourceEntities []*ResourceEntity
	res := &proto.ListResourcesResponse{}
	if err := datastore.GetAll(ctx, query, &resourceEntities); err != nil {
		return nil, err
	}
	for _, resource := range resourceEntities {
		res.Resources = append(res.Resources, toResourceModel(resource))
	}
	return res, nil
}

// Helper functions
func getResourceById(ctx context.Context, id string) (*ResourceEntity, error) {
	resource := &ResourceEntity{ResourceId: id}
	if err := datastore.Get(ctx, resource); err != nil {
		return nil, err
	}
	return resource, nil
}
