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

	"go.chromium.org/luci/gae/impl/memory"
	"go.chromium.org/luci/gae/service/datastore"
)

func mockCreateAssetResourceRequest(assetId string, resourceId string, aliasName string) *proto.CreateAssetResourceRequest {
	return &proto.CreateAssetResourceRequest{
		AssetId:    assetId,
		ResourceId: resourceId,
		AliasName:  aliasName,
	}
}

func TestAssetResourceCreateWithValidData(t *testing.T) {
	t.Parallel()
	assetResourceRequest := mockCreateAssetResourceRequest("Test AsseId", "Test ResourceId", "Test AliasName")
	Convey("Create an asset_resource entity in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &AssetResourceHandler{}
		model, err := handler.Create(ctx, assetResourceRequest)
		So(err, ShouldBeNil)
		want := []string{assetResourceRequest.GetAssetId(), assetResourceRequest.GetResourceId(), assetResourceRequest.GetAliasName()}
		get := []string{model.GetAssetId(), model.GetResourceId(), model.GetAliasName()}
		So(get, ShouldResemble, want)
	})
}

func TestAssetResourceCreateWithInvalidAssetId(t *testing.T) {
	t.Parallel()
	assetResourceRequest := mockCreateAssetResourceRequest("", "Test ResourceId", "Test AliasName")
	Convey("Create an asset_resource entity in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &AssetResourceHandler{}
		_, err := handler.Create(ctx, assetResourceRequest)
		So(err, ShouldNotBeNil)
	})
}

func TestAssetResourceCreateWithInvalidResourceId(t *testing.T) {
	t.Parallel()
	assetResourceRequest := mockCreateAssetResourceRequest("Test AssetId", "", "Test AliasName")
	Convey("Create an asset_resource entity in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &AssetResourceHandler{}
		_, err := handler.Create(ctx, assetResourceRequest)
		So(err, ShouldNotBeNil)
	})
}

func TestAssetResourceCreateWithInvalidAliasName(t *testing.T) {
	t.Parallel()
	assetResourceRequest := mockCreateAssetResourceRequest("Test AssetId", "Test ResourceId", "")
	Convey("Create an asset_resource entity in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &AssetResourceHandler{}
		_, err := handler.Create(ctx, assetResourceRequest)
		So(err, ShouldNotBeNil)
	})
}

func TestGetAssetResourceWithValidData(t *testing.T) {
	assetResourceRequest := mockCreateAssetResourceRequest("Test AssetId", "Test ResourceId", "Test AliasName")
	Convey("Get an asset_resource based on id from datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &AssetResourceHandler{}
		entity, err := handler.Create(ctx, assetResourceRequest)
		So(err, ShouldBeNil)
		getRequest := &proto.GetAssetResourceRequest{
			AssetResourceId: entity.GetAssetResourceId(),
		}
		readEntity, err := handler.Get(ctx, getRequest)
		So(err, ShouldBeNil)

		want := []string{entity.GetAssetResourceId(), entity.GetAssetId(), entity.GetResourceId(), entity.GetAliasName()}
		get := []string{readEntity.GetAssetResourceId(), readEntity.GetAssetId(), readEntity.GetResourceId(), readEntity.GetAliasName()}
		So(get, ShouldResemble, want)
	})
}

func TestListAssetResources(t *testing.T) {
	assetResourceRequest1 := mockCreateAssetResourceRequest("Test AssetId1", "Test ResourceId1", "Test AliasName1")
	assetResourceRequest2 := mockCreateAssetResourceRequest("Test AssetId2", "Test ResourceId2", "Test AliasName2")
	Convey("Get an asset_resource based on id from datastore", t, func() {
		ctx := memory.Use(context.Background())
		datastore.GetTestable(ctx).Consistent(true)
		handler := &AssetResourceHandler{}
		_, err := handler.Create(ctx, assetResourceRequest1)
		So(err, ShouldBeNil)
		_, err = handler.Create(ctx, assetResourceRequest2)
		So(err, ShouldBeNil)
		// Verify
		response, err := handler.List(ctx, &proto.ListAssetResourcesRequest{})
		So(err, ShouldBeNil)
		So(response.GetAssetResources(), ShouldHaveLength, 2)
		asset_resources := response.GetAssetResources()
		want := []string{"Test AssetId1", "Test AssetId2"}
		get := []string{asset_resources[0].GetAssetId(), asset_resources[1].GetAssetId()}
		sort.Strings(get)
		So(get, ShouldResemble, want)
		want = []string{"Test ResourceId1", "Test ResourceId2"}
		get = []string{asset_resources[0].GetResourceId(), asset_resources[1].GetResourceId()}
		sort.Strings(get)
		So(get, ShouldResemble, want)
		want = []string{"Test AliasName1", "Test AliasName2"}
		get = []string{asset_resources[0].GetAliasName(), asset_resources[1].GetAliasName()}
		sort.Strings(get)
		So(get, ShouldResemble, want)
	})
}
