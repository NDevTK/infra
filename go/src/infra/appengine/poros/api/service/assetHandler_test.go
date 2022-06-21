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
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"go.chromium.org/luci/gae/impl/memory"
	"go.chromium.org/luci/gae/service/datastore"
)

func mockCreateAssetRequest(name string, description string, asset_type string, assetResourcesToSave []*proto.AssetResourceModel) *proto.CreateAssetRequest {
	return &proto.CreateAssetRequest{
		Name:                 name,
		Description:          description,
		AssetType:            asset_type,
		AssetResourcesToSave: assetResourcesToSave,
	}
}

func mockAssetResource(assetResourceId string, assetId string, resourceId string, aliasName string) *proto.AssetResourceModel {
	return &proto.AssetResourceModel{
		AssetResourceId: assetResourceId,
		AssetId:         assetId,
		ResourceId:      resourceId,
		AliasName:       aliasName,
	}
}

func mockGetAssetConfigRequest(assetId string) *proto.GetAssetConfigurationRequest {
	return &proto.GetAssetConfigurationRequest{AssetId: assetId}
}

func mockGetHostConfigRequest(resourceIds []string) *proto.GetHostConfigurationRequest {
	return &proto.GetHostConfigurationRequest{ResourceIds: resourceIds}
}
func TestAssetCreateWithValidData(t *testing.T) {
	t.Parallel()
	ctx := memory.Use(context.Background())
	assetResourcesToSave := []*proto.AssetResourceModel{mockAssetResource("", "", "ResourceId", "Alias name")}
	assetRequest := mockCreateAssetRequest("Test Asset", "Test Asset description", "active_directory", assetResourcesToSave)
	Convey("Create an asset in datastore", t, func() {
		handler := &AssetHandler{}
		response, err := handler.Create(ctx, assetRequest)
		So(err, ShouldBeNil)
		want := []string{assetRequest.GetName(), assetRequest.GetDescription(), assetRequest.GetAssetType()}
		get := []string{response.GetAsset().GetName(), response.GetAsset().GetDescription(), response.GetAsset().GetAssetType()}
		So(get, ShouldResemble, want)
		So(response.GetAssetResources(), ShouldHaveLength, 1)
		want = []string{response.GetAsset().GetAssetId(), "ResourceId", "Alias name"}
		assetResource := response.GetAssetResources()[0]
		get = []string{assetResource.GetAssetId(), assetResource.GetResourceId(), assetResource.GetAliasName()}
		So(get, ShouldResemble, want)
	})
}

func TestAssetCreateWithInvalidName(t *testing.T) {
	t.Parallel()
	ctx := memory.Use(context.Background())
	assetRequest := mockCreateAssetRequest("", "Test Asset description", "active_directory", []*proto.AssetResourceModel{})
	Convey("Create an asset with invalid name in datastore", t, func() {
		handler := &AssetHandler{}
		_, err := handler.Create(ctx, assetRequest)
		So(err, ShouldNotBeNil)
	})
}

func TestAssetCreateWithInvalidDescription(t *testing.T) {
	t.Parallel()
	ctx := memory.Use(context.Background())
	assetRequest := mockCreateAssetRequest("Test Asset", "", "active_directory", []*proto.AssetResourceModel{})
	Convey("Create an asset with invalid description in datastore", t, func() {
		handler := &AssetHandler{}
		_, err := handler.Create(ctx, assetRequest)
		So(err, ShouldNotBeNil)
	})
}

func TestAssetCreateWithInvalidAssetType(t *testing.T) {
	t.Parallel()
	ctx := memory.Use(context.Background())
	assetRequest := mockCreateAssetRequest("Test Asset", "Test Asset description", "", []*proto.AssetResourceModel{})
	Convey("Create an asset with invalid asset_type in datastore", t, func() {
		handler := &AssetHandler{}
		_, err := handler.Create(ctx, assetRequest)
		So(err, ShouldNotBeNil)
	})
}

func TestAssetCreateWithInvalidAssetResource(t *testing.T) {
	t.Parallel()
	ctx := memory.Use(context.Background())
	assetRequest := mockCreateAssetRequest("Test Name", "Test Description", "active_directory",
		[]*proto.AssetResourceModel{mockAssetResource("", "", "", "")})
	Convey("Create an asset with invalid asset_resource in datastore", t, func() {
		handler := &AssetHandler{}
		_, err := handler.Create(ctx, assetRequest)
		So(err, ShouldNotBeNil)
	})
}

func TestAssetUpdateWithValidData(t *testing.T) {
	t.Parallel()
	assetResourcesToSave := []*proto.AssetResourceModel{mockAssetResource("", "", "ResourceId", "Alias name")}
	assetResourcesToDelete := []*proto.AssetResourceModel{}
	assetRequest := mockCreateAssetRequest("Test Asset", "Test Asset description", "active_directory", assetResourcesToSave)
	Convey("Update an asset with valid data in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &AssetHandler{}
		createAssetesponse, err := handler.Create(ctx, assetRequest)
		So(err, ShouldBeNil)

		// Update asset with some new value and the operation should not throw any error
		entity := createAssetesponse.GetAsset()
		entity.Name = "Test Asset Name Updated"
		entity.Description = "Test Asset Description Updated"
		entity.AssetType = "active_directory_updated"
		assetResourcesToSave[0].ResourceId = "ResourceId Updated"
		assetResourcesToSave[0].AliasName = "Alias Name Updated"

		updateRequest := &proto.UpdateAssetRequest{
			Asset:                   entity,
			AssetUpdateMask:         &fieldmaskpb.FieldMask{Paths: []string{"name", "description", "asset_type"}},
			AssetResourcesToSave:    assetResourcesToSave,
			AssetResourceUpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"resource_id", "alias_name"}},
			AssetResourcesToDelete:  assetResourcesToDelete,
		}
		updateAssetResponse, err := handler.Update(ctx, updateRequest)
		So(err, ShouldBeNil)

		// Retrieve the updated asset and make sure that the values were correctly updated
		getRequest := &proto.GetAssetRequest{
			AssetId: entity.GetAssetId(),
		}
		readEntity, err := handler.Get(ctx, getRequest)
		want := []string{"Test Asset Name Updated", "Test Asset Description Updated", "active_directory_updated"}
		get := []string{readEntity.GetName(), readEntity.GetDescription(), readEntity.GetAssetType()}
		So(get, ShouldResemble, want)

		//Retrieve the updated asset_resource to make sure the update goes through
		assetResourceHanlder := &AssetResourceHandler{}
		req := &proto.GetAssetResourceRequest{AssetResourceId: updateAssetResponse.GetAssetResources()[0].GetAssetResourceId()}

		readAssetResource, err := assetResourceHanlder.Get(ctx, req)
		want = []string{"ResourceId Updated", "Alias Name Updated"}
		get = []string{readAssetResource.GetResourceId(), readAssetResource.GetAliasName()}
		So(get, ShouldResemble, want)
	})
}

func TestAssetUpdateWithInvalidName(t *testing.T) {
	t.Parallel()
	assetRequest := mockCreateAssetRequest("Test Asset Name", "Test Asset description", "active_directory", []*proto.AssetResourceModel{})
	Convey("Update an asset with invalid name in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &AssetHandler{}
		response, err := handler.Create(ctx, assetRequest)
		So(err, ShouldBeNil)
		entity := response.GetAsset()
		entity.Name = ""
		entity.Description = "Test Asset Description"
		entity.AssetType = "active_directory"

		updateRequest := &proto.UpdateAssetRequest{
			Asset:                   entity,
			AssetUpdateMask:         &fieldmaskpb.FieldMask{Paths: []string{"name", "description", "asset_type"}},
			AssetResourceUpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"resource_id", "alias_name"}},
			AssetResourcesToSave:    []*proto.AssetResourceModel{},
			AssetResourcesToDelete:  []*proto.AssetResourceModel{},
		}
		_, err = handler.Update(ctx, updateRequest)
		// should not save the asset as name is empty
		So(err, ShouldNotBeNil)
	})
}

func TestAssetUpdateWithInvalidDescription(t *testing.T) {
	t.Parallel()
	assetRequest := mockCreateAssetRequest("Test Asset Name", "Test Asset description", "active_directory", []*proto.AssetResourceModel{})
	Convey("Update an asset with invalid name in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &AssetHandler{}
		response, err := handler.Create(ctx, assetRequest)
		So(err, ShouldBeNil)
		entity := response.GetAsset()
		entity.Name = "Test Asset Name"
		entity.Description = ""
		entity.AssetType = "active_directory"

		updateRequest := &proto.UpdateAssetRequest{
			Asset:                   entity,
			AssetUpdateMask:         &fieldmaskpb.FieldMask{Paths: []string{"name", "description", "asset_type"}},
			AssetResourceUpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"resource_id", "alias_name"}},
			AssetResourcesToSave:    []*proto.AssetResourceModel{},
			AssetResourcesToDelete:  []*proto.AssetResourceModel{},
		}
		_, err = handler.Update(ctx, updateRequest)
		// should not save the asset as name is empty
		So(err, ShouldNotBeNil)
	})
}

func TestAssetUpdateWithInvalidAssetType(t *testing.T) {
	t.Parallel()
	assetRequest := mockCreateAssetRequest("Test Asset Name", "Test Asset description", "active_directory", []*proto.AssetResourceModel{})
	Convey("Update an asset with invalid name in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &AssetHandler{}
		response, err := handler.Create(ctx, assetRequest)
		So(err, ShouldBeNil)
		entity := response.GetAsset()
		entity.Name = "Test Asset Name"
		entity.Description = "Test Asset description"
		entity.AssetType = ""

		updateRequest := &proto.UpdateAssetRequest{
			Asset:                   entity,
			AssetUpdateMask:         &fieldmaskpb.FieldMask{Paths: []string{"name", "description", "asset_type"}},
			AssetResourceUpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"resource_id", "alias_name"}},
			AssetResourcesToSave:    []*proto.AssetResourceModel{},
			AssetResourcesToDelete:  []*proto.AssetResourceModel{},
		}
		_, err = handler.Update(ctx, updateRequest)
		// should not save the asset as name is empty
		So(err, ShouldNotBeNil)
	})
}

func TestAssetUpdateWithInvalidAssetResource(t *testing.T) {
	t.Parallel()
	assetResourcesToSave := []*proto.AssetResourceModel{mockAssetResource("", "", "ResourceId", "Alias name")}
	assetResourcesToDelete := []*proto.AssetResourceModel{}
	assetRequest := mockCreateAssetRequest("Test Asset Name", "Test Asset description", "active_directory", assetResourcesToSave)
	Convey("Update an asset with invalid name in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &AssetHandler{}
		response, err := handler.Create(ctx, assetRequest)
		So(err, ShouldBeNil)
		assetResourcesToSave[0].ResourceId = ""

		updateRequest := &proto.UpdateAssetRequest{
			Asset:                   response.GetAsset(),
			AssetUpdateMask:         &fieldmaskpb.FieldMask{Paths: []string{"name", "description", "asset_type"}},
			AssetResourceUpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"resource_id", "alias_name"}},
			AssetResourcesToSave:    assetResourcesToSave,
			AssetResourcesToDelete:  assetResourcesToDelete,
		}
		_, err = handler.Update(ctx, updateRequest)
		// should not save the asset as the asset_resource is invalid
		So(err, ShouldNotBeNil)
	})
}

func TestGetAssetWithValidData(t *testing.T) {
	ctx := memory.Use(context.Background())
	assetRequest := mockCreateAssetRequest("Test Asset", "Test Asset description", "active_directory", []*proto.AssetResourceModel{})
	Convey("Get an assets based on id from datastore", t, func() {
		handler := &AssetHandler{}
		response, err := handler.Create(ctx, assetRequest)
		So(err, ShouldBeNil)
		getRequest := &proto.GetAssetRequest{
			AssetId: response.GetAsset().GetAssetId(),
		}
		readEntity, err := handler.Get(ctx, getRequest)
		So(err, ShouldBeNil)

		want := []string{response.GetAsset().GetName(), response.GetAsset().GetDescription(), response.GetAsset().GetAssetType()}
		get := []string{readEntity.GetName(), readEntity.GetDescription(), readEntity.GetAssetType()}
		So(get, ShouldResemble, want)
	})
}

func TestListAssets(t *testing.T) {
	t.Parallel()
	ctx := memory.Use(context.Background())
	assetRequest1 := mockCreateAssetRequest("Test Asset1", "Test Asset description", "active_directory", []*proto.AssetResourceModel{})
	assetRequest2 := mockCreateAssetRequest("Test Asset2", "Test Asset description", "active_directory", []*proto.AssetResourceModel{})
	Convey("Get all assets from datastore", t, func() {
		handler := &AssetHandler{}
		_, err := handler.Create(ctx, assetRequest1)
		So(err, ShouldBeNil)
		_, err = handler.Create(ctx, assetRequest2)
		So(err, ShouldBeNil)
		// Verify
		response, err := handler.List(ctx, &proto.ListAssetsRequest{})
		So(err, ShouldBeNil)
		So(response.GetAssets(), ShouldHaveLength, 2)
		assets := response.GetAssets()
		want := []string{"Test Asset1", "Test Asset2"}
		get := []string{assets[0].GetName(), assets[1].GetName()}
		sort.Strings(get)
		So(get, ShouldResemble, want)
	})
}

func TestAssetConfigWithValidDetails(t *testing.T) {
	t.Parallel()

	Convey("Test Generated Asset Configuration with valid data", t, func() {
		ctx := memory.Use(context.Background())
		datastore.GetTestable(ctx).Consistent(true)

		asset, assetResource, resource, err := generateAssetAndResources(ctx)
		So(err, ShouldBeNil)

		handler := &AssetHandler{}
		mockRequest := mockGetAssetConfigRequest(asset.AssetId)
		res, err := handler.GetAssetConfiguration(ctx, mockRequest)
		So(err, ShouldBeNil)

		assetConfig := &proto.AssetConfiguration{}
		err = protojson.Unmarshal([]byte(res.Config), assetConfig)
		So(err, ShouldBeNil)

		So(assetConfig.AssetId, ShouldEqual, asset.AssetId)

		So(len(assetConfig.Resources), ShouldEqual, 1)

		So(assetConfig.Resources[0].ResourceId, ShouldEqual, resource.ResourceId)
		So(assetConfig.Resources[0].Description, ShouldEqual, resource.Description)
		So(assetConfig.Resources[0].OperatingSystem, ShouldEqual, resource.OperatingSystem)
		So(assetConfig.Resources[0].AliasName, ShouldEqual, assetResource.AliasName)
		So(assetConfig.Resources[0].ResourceName, ShouldEqual, resource.Name)
		So(assetConfig.Resources[0].ResourceType, ShouldEqual, resource.Type)
	})
}

func TestHostConfigWithValidDetails(t *testing.T) {
	t.Parallel()

	Convey("Test Generated Host Configuration with valid data", t, func() {
		ctx := memory.Use(context.Background())
		datastore.GetTestable(ctx).Consistent(true)

		_, _, resource, err := generateAssetAndResources(ctx)
		So(err, ShouldBeNil)

		handler := &AssetHandler{}
		mockRequest := mockGetHostConfigRequest([]string{resource.ResourceId})
		res, err := handler.GetHostConfiguration(ctx, mockRequest)
		So(err, ShouldBeNil)

		hostConfig := &proto.HostConfiguration{}
		err = protojson.Unmarshal([]byte(res.Config), hostConfig)
		So(err, ShouldBeNil)

		So(len(hostConfig.Resources), ShouldEqual, 1)

		So(hostConfig.Resources[0].ResourceId, ShouldEqual, resource.ResourceId)
		So(hostConfig.Resources[0].ResourceName, ShouldEqual, resource.Name)
		So(hostConfig.Resources[0].ResourceImage, ShouldEqual, resource.Image)
	})
}

func generateAssetAndResources(ctx context.Context) (*proto.AssetModel, *proto.AssetResourceModel, *proto.ResourceModel, error) {
	resHandler := &ResourceHandler{}
	resourceRequest := mockCreateResourceRequest("win2016mock", "mock windows machine", "machine", "windows_machine", "win-image-1")
	resource, err := resHandler.Create(ctx, resourceRequest)
	if err != nil {
		return nil, nil, nil, err
	}

	assetHandler := &AssetHandler{}
	var assetsToSave []*proto.AssetResourceModel
	assetRequest := mockCreateAssetRequest("Test Asset Name", "Test Asset description", "active_directory", assetsToSave)
	asset, err := assetHandler.Create(ctx, assetRequest)
	if err != nil {
		return nil, nil, nil, err
	}

	assetResHandler := &AssetResourceHandler{}
	assetResourceRequest := mockCreateAssetResourceRequest(asset.Asset.AssetId, resource.ResourceId, "win-1")
	assetResource, err := assetResHandler.Create(ctx, assetResourceRequest)
	_, err = getByAssetResourceId(ctx, assetResource.AssetResourceId)
	if err != nil {
		return nil, nil, nil, err
	}

	return asset.Asset, assetResource, resource, nil
}
