// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package service

import (
	"context"
	"sort"
	"testing"

	proto "infra/appengine/poros/api/proto"

	"go.chromium.org/luci/gae/service/datastore"

	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/luci/gae/impl/memory"
)

func mockCreateResourceRequest(name string, description string, Type string, image string) *proto.CreateResourceRequest {
	return &proto.CreateResourceRequest{
		Name:        name,
		Description: description,
		Type:        Type,
		Image:       image,
	}
}

func TestResourceCreateWithValidData(t *testing.T) {
	t.Parallel()
	resourceRequest := mockCreateResourceRequest("Test Resource", "Test Resource description", "machine", "image-1")
	Convey("Create a resource in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &ResourceHandler{}
		model, err := handler.Create(ctx, resourceRequest)
		So(err, ShouldBeNil)
		want := []string{resourceRequest.GetName(), resourceRequest.GetDescription(), resourceRequest.GetType(), resourceRequest.GetImage()}
		get := []string{model.GetName(), model.GetDescription(), model.GetType(), model.GetImage()}
		So(get, ShouldResemble, want)
	})
}

func TestResourceCreateWithInvalidName(t *testing.T) {
	t.Parallel()
	resourceRequest := mockCreateResourceRequest("", "Test Resource description", "machine", "image-1")
	Convey("Create a resource with invalid name in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &ResourceHandler{}
		_, err := handler.Create(ctx, resourceRequest)
		So(err, ShouldNotBeNil)
	})
}

func TestResourceCreateWithInvalidDescription(t *testing.T) {
	t.Parallel()
	resourceRequest := mockCreateResourceRequest("Test Resource", "", "machine", "image-1")
	Convey("Create a resource with invalid description in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &ResourceHandler{}
		_, err := handler.Create(ctx, resourceRequest)
		So(err, ShouldNotBeNil)
	})
}

func TestResourceCreateWithInvalidType(t *testing.T) {
	t.Parallel()
	resourceRequest := mockCreateResourceRequest("Test Resource", "Test Resource Description", "", "image-1")
	Convey("Create a resource with invalid type in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &ResourceHandler{}
		_, err := handler.Create(ctx, resourceRequest)
		So(err, ShouldNotBeNil)
	})
}

func TestResourceCreateOfTypeMachineWithInvalidImage(t *testing.T) {
	t.Parallel()
	resourceRequest := mockCreateResourceRequest("Test Resource", "Test Resource Description", "machine", "")
	Convey("Create a resource with invalid image in datastore if type if machine", t, func() {
		ctx := memory.Use(context.Background())
		handler := &ResourceHandler{}
		_, err := handler.Create(ctx, resourceRequest)
		So(err, ShouldNotBeNil)
	})
}

func TestGetResourceWithValidData(t *testing.T) {
	resourceRequest := mockCreateResourceRequest("Test Resource", "Test Resource description", "machine", "image-1")
	Convey("Get a resource based on id from datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &ResourceHandler{}
		entity, err := handler.Create(ctx, resourceRequest)
		So(err, ShouldBeNil)
		getRequest := &proto.GetResourceRequest{
			ResourceId: entity.GetResourceId(),
		}
		readEntity, err := handler.Get(ctx, getRequest)
		So(err, ShouldBeNil)

		want := []string{entity.GetName(), entity.GetDescription(), entity.GetType(), entity.GetImage()}
		get := []string{readEntity.GetName(), readEntity.GetDescription(), readEntity.GetType(), readEntity.GetImage()}
		So(get, ShouldResemble, want)
	})
}

func TestListResources(t *testing.T) {
	t.Parallel()

	resourceRequest1 := mockCreateResourceRequest("Test Resource1", "Test Resource description", "machine", "image-1")
	resourceRequest2 := mockCreateResourceRequest("Test Resource2", "Test Resource description2", "machine", "image-5")

	Convey("Get all resources from datastore", t, func() {
		ctx := memory.Use(context.Background())
		datastore.GetTestable(ctx).Consistent(true)
		handler := &ResourceHandler{}
		_, err := handler.Create(ctx, resourceRequest1)
		So(err, ShouldBeNil)
		_, err = handler.Create(ctx, resourceRequest2)
		So(err, ShouldBeNil)
		// Verify
		response, err := handler.List(ctx, &proto.ListResourcesRequest{})
		So(err, ShouldBeNil)
		So(response.GetResources(), ShouldHaveLength, 2)
		resources := response.GetResources()
		want := []string{"Test Resource1", "Test Resource2"}
		get := []string{resources[0].GetName(), resources[1].GetName()}
		sort.Strings(get)
		So(get, ShouldResemble, want)
	})
}
