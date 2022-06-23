// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package service

import (
	"context"
	"sort"
	"testing"

	proto "infra/appengine/poros/api/proto"

	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"go.chromium.org/luci/gae/impl/memory"
	"go.chromium.org/luci/gae/service/datastore"
)

func mockCreateAssetInstanceRequest(assetId string, statusValue int32) *proto.CreateAssetInstanceRequest {
	return &proto.CreateAssetInstanceRequest{
		AssetId: assetId,
		Status:  proto.DeploymentStatus_name[statusValue],
	}
}

func TestAssetInstanceCreateWithValidData(t *testing.T) {
	t.Parallel()
	request := mockCreateAssetInstanceRequest("Test AssetId", 0)
	Convey("Create an AssetInstance in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &AssetInstanceHandler{}
		model, err := handler.Create(ctx, request)
		So(err, ShouldBeNil)
		want := []string{request.GetAssetId(), request.GetStatus()}
		get := []string{model.GetAssetId(), model.GetStatus()}
		So(get, ShouldResemble, want)
	})
}

func TestAssetInstanceCreateWithInvalidAssetId(t *testing.T) {
	t.Parallel()
	request := mockCreateAssetInstanceRequest("", 0)
	Convey("Create an AssetInstance with invalid assetId in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &AssetInstanceHandler{}
		_, err := handler.Create(ctx, request)
		So(err, ShouldNotBeNil)
	})
}

func TestAssetInstanceCreateWithInvalidStatus(t *testing.T) {
	t.Parallel()
	request := mockCreateAssetInstanceRequest("Test Asset Id", -1)
	Convey("Create an AssetInstance with invalid status in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &AssetInstanceHandler{}
		_, err := handler.Create(ctx, request)
		So(err, ShouldNotBeNil)
	})
}

func TestAssetInstanceUpdateWithValidData(t *testing.T) {
	t.Parallel()
	createRequest := mockCreateAssetInstanceRequest("Test AssetId", 0)
	Convey("Update an AssetInstance with valid data in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &AssetInstanceHandler{}
		entity, err := handler.Create(ctx, createRequest)
		So(err, ShouldBeNil)

		// Update AssetInstance with some new value and the operation should not throw any error
		entity.AssetId = "Test AssetId Updated"
		entity.Status = proto.DeploymentStatus(1).String()

		updateRequest := &proto.UpdateAssetInstanceRequest{
			AssetInstance: entity,
			UpdateMask:    &fieldmaskpb.FieldMask{Paths: []string{"asset_id", "status"}},
		}
		_, err = handler.Update(ctx, updateRequest)
		So(err, ShouldBeNil)

		// Retrieve the updated AssetInstance and make sure that the values were correctly updated
		getRequest := &proto.GetAssetInstanceRequest{
			AssetInstanceId: entity.GetAssetInstanceId(),
		}
		readEntity, err := handler.Get(ctx, getRequest)
		want := []string{"Test AssetId Updated", proto.DeploymentStatus_name[1]}
		get := []string{readEntity.GetAssetId(), readEntity.GetStatus()}
		So(get, ShouldResemble, want)
	})
}

func TestAssetInstanceUpdateWithInvalidAssetId(t *testing.T) {
	t.Parallel()
	createRequest := mockCreateAssetInstanceRequest("Test AssetId", 0)
	Convey("Update an AssetInstance with invalid assetId in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &AssetInstanceHandler{}
		entity, err := handler.Create(ctx, createRequest)
		So(err, ShouldBeNil)
		entity.AssetId = ""
		entity.Status = proto.DeploymentStatus(1).String()

		updateRequest := &proto.UpdateAssetInstanceRequest{
			AssetInstance: entity,
			UpdateMask:    &fieldmaskpb.FieldMask{Paths: []string{"asset_id", "status"}},
		}
		_, err = handler.Update(ctx, updateRequest)
		// should not save the AssetInstance as AssetId is empty
		So(err, ShouldNotBeNil)
	})
}

func TestAssetInstanceUpdateWithInvalidStatus(t *testing.T) {
	t.Parallel()
	createRequest := mockCreateAssetInstanceRequest("Test AssetId", 0)
	Convey("Update an AssetInstance with invalid deployment status in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &AssetInstanceHandler{}
		entity, err := handler.Create(ctx, createRequest)
		So(err, ShouldBeNil)
		entity.AssetId = "Test AssetId Updated"
		entity.Status = proto.DeploymentStatus(-1).String()

		updateRequest := &proto.UpdateAssetInstanceRequest{
			AssetInstance: entity,
			UpdateMask:    &fieldmaskpb.FieldMask{Paths: []string{"asset_id", "status"}},
		}
		readEntity, err := handler.Update(ctx, updateRequest)
		// should not save the AssetInstance as DeploymentStatus is invalid
		So(readEntity, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})
}

func TestGetAssetInstanceWithValidData(t *testing.T) {
	createRequest := mockCreateAssetInstanceRequest("Test RAssetId", 0)
	Convey("Get a AssetInstance based on id from datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &AssetInstanceHandler{}
		entity, err := handler.Create(ctx, createRequest)
		So(err, ShouldBeNil)
		getRequest := &proto.GetAssetInstanceRequest{
			AssetInstanceId: entity.GetAssetInstanceId(),
		}
		readEntity, err := handler.Get(ctx, getRequest)
		So(err, ShouldBeNil)

		want := []string{entity.GetAssetId(), entity.GetStatus()}
		get := []string{readEntity.GetAssetId(), readEntity.GetStatus()}
		So(get, ShouldResemble, want)
	})
}

func TestListAssetInstance(t *testing.T) {
	t.Parallel()
	createRequest1 := mockCreateAssetInstanceRequest("Test AssetId1", 0)
	createRequest2 := mockCreateAssetInstanceRequest("Test AssetId2", 1)
	Convey("Get all AssetInstances from datastore", t, func() {
		ctx := memory.Use(context.Background())
		datastore.GetTestable(ctx).Consistent(true)
		handler := &AssetInstanceHandler{}
		_, err := handler.Create(ctx, createRequest1)
		So(err, ShouldBeNil)
		_, err = handler.Create(ctx, createRequest2)
		So(err, ShouldBeNil)
		// Verify
		response, err := handler.List(ctx, &proto.ListAssetInstancesRequest{})
		So(err, ShouldBeNil)
		So(response.GetAssetInstances(), ShouldHaveLength, 2)
		asset_instances := response.GetAssetInstances()
		want := []string{"Test AssetId1", "Test AssetId2"}
		get := []string{asset_instances[0].GetAssetId(), asset_instances[1].GetAssetId()}
		sort.Strings(get)
		So(get, ShouldResemble, want)
		want = []string{"STATUS_PENDING", "STATUS_RUNNING"}
		get = []string{asset_instances[0].GetStatus(), asset_instances[1].GetStatus()}
		sort.Strings(get)
		So(get, ShouldResemble, want)
	})
}
