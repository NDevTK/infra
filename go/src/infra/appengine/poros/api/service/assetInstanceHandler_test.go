// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package service

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	. "infra/appengine/poros/api/entities"
	proto "infra/appengine/poros/api/proto"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	"go.chromium.org/luci/gae/impl/memory"
	"go.chromium.org/luci/gae/service/datastore"
	"go.chromium.org/luci/server/auth"
)

func mockCreateAssetInstanceRequest(assetId string, statusValue int32) *proto.CreateAssetInstanceRequest {
	return &proto.CreateAssetInstanceRequest{
		AssetId: assetId,
		Status:  proto.DeploymentStatus_name[statusValue],
	}
}

func mockTriggerDeploymentRequest(entityType string, entityId string) *proto.TriggerDeploymentRequest {
	return &proto.TriggerDeploymentRequest{
		EntityType: entityType,
		EntityId:   entityId,
	}
}

func mockFetchLogsRequest(assetInstanceId string) *proto.FetchLogsRequest {
	return &proto.FetchLogsRequest{
		AssetInstanceId: assetInstanceId,
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
		So(model.CreatedAt.AsTime().Add(time.Hour*24*7), ShouldEqual, model.DeleteAt.AsTime())
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
		timestamp := time.Now().UTC()
		entity.DeleteAt = timestamppb.New(timestamp)

		updateRequest := &proto.UpdateAssetInstanceRequest{
			AssetInstance: entity,
			UpdateMask:    &fieldmaskpb.FieldMask{Paths: []string{"asset_id", "status", "delete_at"}},
		}
		updatedEntity, err := handler.Update(ctx, updateRequest)
		So(err, ShouldBeNil)
		want := []string{"Test AssetId Updated", proto.DeploymentStatus_name[1]}
		get := []string{updatedEntity.GetAssetId(), updatedEntity.GetStatus()}
		So(get, ShouldResemble, want)
		So(timestamp.Format(time.UnixDate), ShouldEqual, updatedEntity.DeleteAt.AsTime().Format(time.UnixDate))

		// Retrieve the updated AssetInstance and make sure that the values were correctly updated
		getRequest := &proto.GetAssetInstanceRequest{
			AssetInstanceId: entity.GetAssetInstanceId(),
		}
		readEntity, err := handler.Get(ctx, getRequest)
		want = []string{"Test AssetId Updated", proto.DeploymentStatus_name[1]}
		get = []string{readEntity.GetAssetId(), readEntity.GetStatus()}
		So(get, ShouldResemble, want)
		So(timestamp.Format(time.UnixDate), ShouldEqual, readEntity.DeleteAt.AsTime().Format(time.UnixDate))
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

func TestTriggerDeployment_TypeAssetInstance(t *testing.T) {
	t.Parallel()
	createRequest := mockCreateAssetInstanceRequest("Test AssetId", 0)
	Convey("Test TriggerDeployment function for Type AssetInstance", t, func() {
		ctx := memory.Use(context.Background())
		handler := &AssetInstanceHandler{}
		entity, _ := handler.Create(ctx, createRequest)
		So(entity.ProjectId, ShouldBeEmpty)
		So(entity.Status, ShouldEqual, "STATUS_PENDING")
		triggerRequest := mockTriggerDeploymentRequest("AssetInstance", entity.AssetInstanceId)
		triggerDeploymentResponse, err := handler.TriggerDeployment(ctx, triggerRequest)
		So(err, ShouldBeNil)

		assetInstanceEntity, _ := getAssetInstanceById(ctx, entity.AssetInstanceId)
		So(assetInstanceEntity.ProjectId, ShouldNotBeEmpty)
		So(assetInstanceEntity.Status, ShouldEqual, "STATUS_RUNNING")

		want := []string{"Test AssetId", assetInstanceEntity.AssetInstanceId}
		get := []string{triggerDeploymentResponse.AssetId, triggerDeploymentResponse.AssetInstanceId}
		So(get, ShouldResemble, want)
		So(triggerDeploymentResponse.ProjectId, ShouldNotBeEmpty)
		So(triggerDeploymentResponse.ProjectPrefix, ShouldNotBeEmpty)
	})
}

func TestTriggerDeployment_TypeAsset(t *testing.T) {
	t.Parallel()
	assetResourcesToSave := []*proto.AssetResourceModel{mockAssetResource("", "", "ResourceId", "Alias name")}
	assetRequest := mockCreateAssetRequest("Test Asset", "Test Asset description", "active_directory", assetResourcesToSave)
	Convey("Test TriggerDeployment function for Type Asset", t, func() {
		ctx := memory.Use(context.Background())
		datastore.GetTestable(ctx).Consistent(true)
		err := createDefaultResources(ctx)
		So(err, ShouldBeNil)
		handler := &AssetHandler{}
		asset, err := handler.Create(ctx, assetRequest)
		So(err, ShouldBeNil)

		assetInstanceHandler := &AssetInstanceHandler{}
		triggerRequest := mockTriggerDeploymentRequest("Asset", asset.Asset.AssetId)
		triggerDeploymentResponse, err := assetInstanceHandler.TriggerDeployment(ctx, triggerRequest)
		So(err, ShouldBeNil)

		assetInstance, err := getAssetInstanceById(ctx, triggerDeploymentResponse.AssetInstanceId)
		So(err, ShouldBeNil)

		So(assetInstance.ProjectId, ShouldNotBeEmpty)
		So(assetInstance.Status, ShouldEqual, "STATUS_RUNNING")

		want := []string{asset.Asset.AssetId, assetInstance.AssetInstanceId}
		get := []string{triggerDeploymentResponse.AssetId, triggerDeploymentResponse.AssetInstanceId}
		So(get, ShouldResemble, want)
		So(triggerDeploymentResponse.ProjectId, ShouldNotBeEmpty)
		So(triggerDeploymentResponse.ProjectPrefix, ShouldNotBeEmpty)
	})
}

func TestDeploymentProject(t *testing.T) {
	t.Parallel()
	projectList := gcpProjectList()
	createData := [][]string{
		{"Test AssetId1", "STATUS_PENDING", ""},
		{"Test AssetId2", "STATUS_RUNNING", projectList[0]},
		{"Test AssetId3", "STATUS_COMPLETED", projectList[1]},
	}

	Convey("Select a project for deployment", t, func() {
		ctx := memory.Use(context.Background())
		datastore.GetTestable(ctx).Consistent(true)
		for _, data := range createData {
			id := uuid.New().String()
			entity := &AssetInstanceEntity{
				AssetInstanceId: id,
				AssetId:         data[0],
				Status:          data[1],
				ProjectId:       data[2],
				CreatedBy:       auth.CurrentUser(ctx).Email,
				CreatedAt:       time.Now().UTC(),
			}
			err := datastore.Put(ctx, entity)
			So(err, ShouldBeNil)
		}

		project, err := deploymentProject(ctx)
		So(err, ShouldBeNil)

		So(project, ShouldEqual, projectList[2])
	})
}

func TestDeploymentProject_NoAvailableProject(t *testing.T) {
	t.Parallel()
	projectList := gcpProjectList()
	Convey("Should throw an error since all projects are deployed at the moment", t, func() {
		ctx := memory.Use(context.Background())
		datastore.GetTestable(ctx).Consistent(true)
		for i, project := range projectList {
			id := uuid.New().String()
			entity := &AssetInstanceEntity{
				AssetInstanceId: id,
				AssetId:         fmt.Sprintf("Test AssetId %v", i),
				Status:          "STATUS_RUNNING",
				ProjectId:       project,
				CreatedBy:       auth.CurrentUser(ctx).Email,
				CreatedAt:       time.Now().UTC(),
			}
			err := datastore.Put(ctx, entity)
			So(err, ShouldBeNil)
		}

		project, err := deploymentProject(ctx)
		So(project, ShouldEqual, "")
		So(err.Error(), ShouldEqual, "No Projects available at the moment")
	})
}

func TestFetchLogs(t *testing.T) {
	t.Parallel()
	Convey("Fetch Asset Logs", t, func() {
		id := uuid.New().String()
		timestamp := time.Now().UTC()
		entity := &AssetInstanceEntity{
			AssetInstanceId: id,
			AssetId:         "Test Asset Id",
			Status:          "STATUS_RUNNING",
			Logs:            "My Test Logs",
			CreatedBy:       "test@test.com",
			CreatedAt:       timestamp,
			DeleteAt:        timestamp.Add(time.Hour * 24 * 7),
		}

		ctx := memory.Use(context.Background())
		err := datastore.Put(ctx, entity)
		So(err, ShouldBeNil)

		request := mockFetchLogsRequest(id)
		handler := &AssetInstanceHandler{}
		logs, err := handler.FetchLogs(ctx, request)
		So(err, ShouldBeNil)
		So(logs.Logs, ShouldEqual, "My Test Logs")
	})
}
