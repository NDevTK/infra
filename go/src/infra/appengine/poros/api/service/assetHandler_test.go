// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package service

import (
	"context"
	"testing"

	proto "infra/appengine/poros/api/proto"

	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/luci/gae/impl/memory"
)

func mockCreateAssetRequest(name string, description string) *proto.CreateAssetRequest {
	return &proto.CreateAssetRequest{
		Name:        name,
		Description: description,
	}
}

func TestAssetCreateWithValidData(t *testing.T) {
	t.Parallel()
	ctx := memory.Use(context.Background())
	assetRequest := mockCreateAssetRequest("Test Asset", "Test Asset description")
	Convey("Create an asset in datastore", t, func() {
		handler := &AssetHandler{}
		model, err := handler.Create(ctx, assetRequest)
		So(err, ShouldBeNil)
		want := []string{assetRequest.GetName(), assetRequest.GetDescription()}
		get := []string{model.GetName(), model.GetDescription()}
		So(get, ShouldResemble, want)
	})
}

func TestAssetCreateWithInvalidName(t *testing.T) {
	t.Parallel()
	ctx := memory.Use(context.Background())
	assetRequest := mockCreateAssetRequest("", "Test Asset description")
	Convey("Create an asset with invalid name in datastore", t, func() {
		handler := &AssetHandler{}
		_, err := handler.Create(ctx, assetRequest)
		So(err, ShouldNotBeNil)
	})
}

func TestAssetCreateWithInvalidDescription(t *testing.T) {
	t.Parallel()
	ctx := memory.Use(context.Background())
	assetRequest := mockCreateAssetRequest("Test Asset", "")
	Convey("Create an asset with invalid description in datastore", t, func() {
		handler := &AssetHandler{}
		_, err := handler.Create(ctx, assetRequest)
		So(err, ShouldNotBeNil)
	})
}

func TestGetAssetWithValidData(t *testing.T) {
	ctx := memory.Use(context.Background())
	assetRequest := mockCreateAssetRequest("Test Asset", "Test Asset description")
	Convey("Get an assets based on id from datastore", t, func() {
		handler := &AssetHandler{}
		entity, err := handler.Create(ctx, assetRequest)
		So(err, ShouldBeNil)
		getRequest := &proto.GetAssetRequest{
			AssetId: entity.GetAssetId(),
		}
		readEntity, err := handler.Get(ctx, getRequest)
		So(err, ShouldBeNil)

		want := []string{entity.GetName(), entity.GetDescription()}
		get := []string{readEntity.GetName(), readEntity.GetDescription()}
		So(get, ShouldResemble, want)
	})
}

func TestListAssets(t *testing.T) {
	t.Parallel()
	ctx := memory.Use(context.Background())
	assetRequest1 := mockCreateAssetRequest("Test Asset1", "Test Asset description")
	assetRequest2 := mockCreateAssetRequest("Test Asset2", "Test Asset description")
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
		get := []string{assets[1].GetName(), assets[0].GetName()}
		So(get, ShouldResemble, want)
	})
}
