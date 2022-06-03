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
	"google.golang.org/protobuf/types/known/fieldmaskpb"

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

func TestResourceUpdateWithValidData(t *testing.T) {
	t.Parallel()
	resourceRequest := mockCreateResourceRequest("Test Resource Name", "Test Resource description", "machine", "image-1")
	Convey("Update a resource with valid data in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &ResourceHandler{}
		entity, err := handler.Create(ctx, resourceRequest)
		So(err, ShouldBeNil)

		// Update resource with some new value and the operation should not throw any error
		entity.Name = "Test Resource Name Updated"
		entity.Description = "Test Resource description Updated"
		entity.Type = "machine"
		entity.Image = "image-2"

		updateRequest := &proto.UpdateResourceRequest{
			Resource:   entity,
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name", "description", "type", "image"}},
		}
		_, err = handler.Update(ctx, updateRequest)
		So(err, ShouldBeNil)

		// Retrieve the updated resource and make sure that the values were correctly updated
		getRequest := &proto.GetResourceRequest{
			ResourceId: entity.GetResourceId(),
		}
		readEntity, err := handler.Get(ctx, getRequest)
		want := []string{"Test Resource Name Updated", "Test Resource description Updated", "machine", "image-2"}
		get := []string{readEntity.GetName(), readEntity.GetDescription(), readEntity.GetType(), readEntity.GetImage()}
		So(get, ShouldResemble, want)
	})
}

func TestResourceUpdateWithInvalidName(t *testing.T) {
	t.Parallel()
	resourceRequest := mockCreateResourceRequest("Test Resource Name", "Test Resource description", "machine", "image-1")
	Convey("Update a resource with invalid name in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &ResourceHandler{}
		entity, err := handler.Create(ctx, resourceRequest)
		So(err, ShouldBeNil)
		entity.Name = ""
		entity.Description = "Test Resource description Updated"
		entity.Type = "machine"
		entity.Image = "image-1"

		updateRequest := &proto.UpdateResourceRequest{
			Resource:   entity,
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name", "description", "type", "image"}},
		}
		_, err = handler.Update(ctx, updateRequest)
		// should not save the resource as name is empty
		So(err, ShouldNotBeNil)
	})
}

func TestResourceUpdateWithInvalidDescription(t *testing.T) {
	t.Parallel()
	resourceRequest := mockCreateResourceRequest("Test Resource Name", "Test Resource description", "machine", "image-1")
	Convey("Update a resource with invalid descriprion in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &ResourceHandler{}
		entity, err := handler.Create(ctx, resourceRequest)
		So(err, ShouldBeNil)
		entity.Name = "Test Resource Name Updated"
		entity.Description = ""
		entity.Type = "machine"
		entity.Image = "image-1"

		updateRequest := &proto.UpdateResourceRequest{
			Resource:   entity,
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name", "description", "type", "image"}},
		}
		_, err = handler.Update(ctx, updateRequest)
		// should not save the resource as description is empty
		So(err, ShouldNotBeNil)
	})
}

func TestResourceUpdateWithInvalidType(t *testing.T) {
	t.Parallel()
	resourceRequest := mockCreateResourceRequest("Test Resource Name", "Test Resource description", "machine", "image-1")
	Convey("Update a resource with invalid type in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &ResourceHandler{}
		entity, err := handler.Create(ctx, resourceRequest)
		So(err, ShouldBeNil)
		entity.Name = "Test Resource Name Updated"
		entity.Description = "Test Resource description"
		entity.Type = ""
		entity.Image = "image-1"

		updateRequest := &proto.UpdateResourceRequest{
			Resource:   entity,
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name", "description", "type", "image"}},
		}
		_, err = handler.Update(ctx, updateRequest)
		// should not save the resource as type is missing
		So(err, ShouldNotBeNil)
	})
}

func TestResourceUpdateOfTypeMachineWithInvalidImage(t *testing.T) {
	t.Parallel()
	resourceRequest := mockCreateResourceRequest("Test Resource Name", "Test Resource description", "machine", "image-1")
	Convey("Update a resource with invalid image in datastore if type if machine", t, func() {
		ctx := memory.Use(context.Background())
		handler := &ResourceHandler{}
		entity, err := handler.Create(ctx, resourceRequest)
		So(err, ShouldBeNil)
		entity.Name = "Test Resource Name Updated"
		entity.Description = "Test Resource description"
		entity.Type = "machine"
		entity.Image = ""

		updateRequest := &proto.UpdateResourceRequest{
			Resource:   entity,
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name", "description", "type", "image"}},
		}
		_, err = handler.Update(ctx, updateRequest)
		// should not save the resource with type machine as image is missing
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
