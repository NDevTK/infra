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

func mockCreateResourceRequest(name string, description string, Type string, operatingSystem string, imageProject string, imageFamily string) *proto.CreateResourceRequest {
	return &proto.CreateResourceRequest{
		Name:            name,
		Description:     description,
		Type:            Type,
		OperatingSystem: operatingSystem,
		ImageProject:    imageProject,
		ImageFamily:     imageFamily,
	}
}

func TestResourceCreateWithValidData(t *testing.T) {
	t.Parallel()
	resourceRequest := mockCreateResourceRequest("Test Resource", "Test Resource description", "ad_joined_machine", "windows_machine", "image-project", "image-family")
	Convey("Create a resource in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &ResourceHandler{}
		model, err := handler.Create(ctx, resourceRequest)
		So(err, ShouldBeNil)
		want := []string{resourceRequest.GetName(), resourceRequest.GetDescription(), resourceRequest.GetType(), resourceRequest.GetImageProject(), resourceRequest.GetImageFamily()}
		get := []string{model.GetName(), model.GetDescription(), model.GetType(), model.GetImageProject(), model.GetImageFamily()}
		So(get, ShouldResemble, want)
		So(model.Deleted, ShouldEqual, false)
	})
}

func TestResourceCreateWithInvalidName(t *testing.T) {
	t.Parallel()
	resourceRequest := mockCreateResourceRequest("", "Test Resource description", "ad_joined_machine", "windows_machine", "image-project", "image-family")
	Convey("Create a resource with invalid name in datastore", t, func() {
		ctx := memory.Use(context.Background())
		datastore.GetTestable(ctx).Consistent(true)
		handler := &ResourceHandler{}
		_, err := handler.Create(ctx, resourceRequest)
		So(err, ShouldNotBeNil)

		resourceRequest := mockCreateResourceRequest("Test Resource Name", "Test Resource description", "ad_joined_machine", "windows_machine", "image-project", "image-family")
		entity, err := handler.Create(ctx, resourceRequest)
		So(err, ShouldBeNil)
		So(entity.Name, ShouldEqual, "Test Resource Name")

		// Creating another resource with duplicate name; the Create operation should fail
		_, err = handler.Create(ctx, resourceRequest)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "Resource name must be unique")
	})
}

func TestResourceCreateWithInvalidDescription(t *testing.T) {
	t.Parallel()
	resourceRequest := mockCreateResourceRequest("Test Resource", "", "ad_joined_machine", "windows_machine", "image-project", "image-family")
	Convey("Create a resource with invalid description in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &ResourceHandler{}
		_, err := handler.Create(ctx, resourceRequest)
		So(err, ShouldNotBeNil)
	})
}

func TestResourceCreateWithInvalidType(t *testing.T) {
	t.Parallel()
	resourceRequest := mockCreateResourceRequest("Test Resource", "Test Resource Description", "", "windows_machine", "image-project", "image-family")
	Convey("Create a resource with invalid type in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &ResourceHandler{}
		_, err := handler.Create(ctx, resourceRequest)
		So(err, ShouldNotBeNil)
	})
}

func TestResourceCreateOfTypeMachineWithInvalidOperatingSystem(t *testing.T) {
	t.Parallel()
	resourceRequest := mockCreateResourceRequest("Test Resource", "Test Resource Description", "ad_joined_machine", "", "image-project", "image-family")
	Convey("Create a resource with invalid operating system in datastore if type if machine", t, func() {
		ctx := memory.Use(context.Background())
		handler := &ResourceHandler{}
		_, err := handler.Create(ctx, resourceRequest)
		So(err, ShouldNotBeNil)
	})
}

func TestResourceCreateOfTypeMachineWithInvalidImageProject(t *testing.T) {
	t.Parallel()
	resourceRequest := mockCreateResourceRequest("Test Resource", "Test Resource Description", "ad_joined_machine", "windows_machine", "", "image-family")
	Convey("Create a resource with invalid image project in datastore if type is machine", t, func() {
		ctx := memory.Use(context.Background())
		handler := &ResourceHandler{}
		_, err := handler.Create(ctx, resourceRequest)
		So(err, ShouldNotBeNil)
	})
}

func TestResourceCreateOfTypeMachineWithInvalidImageFamily(t *testing.T) {
	t.Parallel()
	resourceRequest := mockCreateResourceRequest("Test Resource", "Test Resource Description", "ad_joined_machine", "windows_machine", "image-project", "")
	Convey("Create a resource with invalid image family in datastore if type is machine", t, func() {
		ctx := memory.Use(context.Background())
		handler := &ResourceHandler{}
		_, err := handler.Create(ctx, resourceRequest)
		So(err, ShouldNotBeNil)
	})
}

func TestResourceUpdateWithValidData(t *testing.T) {
	t.Parallel()
	resourceRequest := mockCreateResourceRequest("Test Resource Name", "Test Resource description", "ad_joined_machine", "windows_machine", "image-project", "image-family")
	Convey("Update a resource with valid data in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &ResourceHandler{}
		entity, err := handler.Create(ctx, resourceRequest)
		So(err, ShouldBeNil)

		// Update resource with some new value and the operation should not throw any error
		entity.Name = "Test Resource Name Updated"
		entity.Description = "Test Resource description Updated"
		entity.Type = "ad_joined_machine"
		entity.OperatingSystem = "linux_system"
		entity.ImageProject = "image-project-updated"
		entity.ImageFamily = "image-family-updated"

		updateRequest := &proto.UpdateResourceRequest{
			Resource:   entity,
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name", "description", "type", "operating_system", "image_project", "image_family"}},
		}
		updatedEntity, err := handler.Update(ctx, updateRequest)
		So(err, ShouldBeNil)
		want := []string{"Test Resource Name Updated", "Test Resource description Updated", "ad_joined_machine", "linux_system", "image-project-updated", "image-family-updated"}
		get := []string{updatedEntity.GetName(), updatedEntity.GetDescription(), updatedEntity.GetType(), updatedEntity.GetOperatingSystem(), updatedEntity.GetImageProject(), updatedEntity.GetImageFamily()}
		So(get, ShouldResemble, want)

		// Retrieve the updated resource and make sure that the values were correctly updated
		getRequest := &proto.GetResourceRequest{
			ResourceId: entity.GetResourceId(),
		}
		readEntity, err := handler.Get(ctx, getRequest)
		want = []string{"Test Resource Name Updated", "Test Resource description Updated", "ad_joined_machine", "linux_system", "image-project-updated", "image-family-updated"}
		get = []string{readEntity.GetName(), readEntity.GetDescription(), readEntity.GetType(), readEntity.GetOperatingSystem(), readEntity.GetImageProject(), readEntity.GetImageFamily()}
		So(get, ShouldResemble, want)
	})
}

func TestResourceUpdateWithInvalidName(t *testing.T) {
	t.Parallel()
	resourceRequest := mockCreateResourceRequest("Test Resource Name", "Test Resource description", "ad_joined_machine", "windows_system", "image-project", "image-family")
	Convey("Update a resource with invalid name in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &ResourceHandler{}
		entity, err := handler.Create(ctx, resourceRequest)
		So(err, ShouldBeNil)
		entity.Name = ""
		entity.Description = "Test Resource description Updated"
		entity.Type = "ad_joined_machine"
		entity.OperatingSystem = "windows_system"
		entity.ImageProject = "image-project"
		entity.ImageFamily = "image-family"

		updateRequest := &proto.UpdateResourceRequest{
			Resource:   entity,
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name", "description", "type", "operating_system", "image_project", "image_family"}},
		}
		_, err = handler.Update(ctx, updateRequest)
		// should not save the resource as name is empty
		So(err, ShouldNotBeNil)
	})
}

func TestResourceUpdateWithDuplicateName(t *testing.T) {
	t.Parallel()
	createRequest1 := mockCreateResourceRequest("Test Resource Name1", "Test Resource description", "ad_joined_machine", "windows_system", "image-project", "image-family")
	createRequest2 := mockCreateResourceRequest("Test Resource Name2", "Test Resource description", "ad_joined_machine", "windows_system", "image-project", "image-family")
	Convey("Update a resource with duplicate name in datastore", t, func() {
		ctx := memory.Use(context.Background())
		datastore.GetTestable(ctx).Consistent(true)
		handler := &ResourceHandler{}
		entity, err := handler.Create(ctx, createRequest1)
		So(err, ShouldBeNil)
		_, err = handler.Create(ctx, createRequest2)
		So(err, ShouldBeNil)
		entity.Name = "Test Resource Name2"

		updateRequest := &proto.UpdateResourceRequest{
			Resource:   entity,
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name"}},
		}
		_, err = handler.Update(ctx, updateRequest)
		// should not save the resource as name is not unique
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "Resource name must be unique")
	})
}

func TestResourceUpdateWithInvalidDescription(t *testing.T) {
	t.Parallel()
	resourceRequest := mockCreateResourceRequest("Test Resource Name", "Test Resource description", "ad_joined_machine", "windows_system", "image-project", "image-family")
	Convey("Update a resource with invalid descriprion in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &ResourceHandler{}
		entity, err := handler.Create(ctx, resourceRequest)
		So(err, ShouldBeNil)
		entity.Name = "Test Resource Name Updated"
		entity.Description = ""
		entity.Type = "ad_joined_machine"
		entity.OperatingSystem = "windows_system"
		entity.ImageProject = "image-project"
		entity.ImageFamily = "image-family"

		updateRequest := &proto.UpdateResourceRequest{
			Resource:   entity,
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name", "description", "type", "operating_system", "image-project", "image-family"}},
		}
		_, err = handler.Update(ctx, updateRequest)
		// should not save the resource as description is empty
		So(err, ShouldNotBeNil)
	})
}

func TestResourceUpdateWithInvalidType(t *testing.T) {
	t.Parallel()
	resourceRequest := mockCreateResourceRequest("Test Resource Name", "Test Resource description", "ad_joined_machine", "windows_system", "image-project", "image-family")
	Convey("Update a resource with invalid type in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &ResourceHandler{}
		entity, err := handler.Create(ctx, resourceRequest)
		So(err, ShouldBeNil)
		entity.Name = "Test Resource Name Updated"
		entity.Description = "Test Resource description"
		entity.Type = ""
		entity.OperatingSystem = "windows_system"
		entity.ImageProject = "image-project"
		entity.ImageFamily = "image-family"

		updateRequest := &proto.UpdateResourceRequest{
			Resource:   entity,
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name", "description", "type", "operating_system", "image-project", "image-family"}},
		}
		_, err = handler.Update(ctx, updateRequest)
		// should not save the resource as type is missing
		So(err, ShouldNotBeNil)
	})
}

func TestResourceUpdateOfTypeMachineWithInvalidOperatingSystem(t *testing.T) {
	t.Parallel()
	resourceRequest := mockCreateResourceRequest("Test Resource Name", "Test Resource description", "ad_joined_machine", "windows_system", "image-project", "image-family")
	Convey("Update a resource with invalid operating system in datastore if type is machine", t, func() {
		ctx := memory.Use(context.Background())
		handler := &ResourceHandler{}
		entity, err := handler.Create(ctx, resourceRequest)
		So(err, ShouldBeNil)
		entity.Name = "Test Resource Name Updated"
		entity.Description = "Test Resource description"
		entity.Type = "ad_joined_machine"
		entity.OperatingSystem = ""
		entity.ImageProject = "image-project"
		entity.ImageFamily = "image-family"

		updateRequest := &proto.UpdateResourceRequest{
			Resource:   entity,
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name", "description", "type", "operating_system", "image-project", "image-family"}},
		}
		_, err = handler.Update(ctx, updateRequest)
		// should not save the resource with type machine as operating system is missing
		So(err, ShouldNotBeNil)
	})
}

func TestResourceUpdateOfTypeMachineWithInvalidImageProject(t *testing.T) {
	t.Parallel()
	resourceRequest := mockCreateResourceRequest("Test Resource Name", "Test Resource description", "ad_joined_machine", "windows_system", "image-project", "image-family")
	Convey("Update a resource with invalid image project in datastore if type is machine", t, func() {
		ctx := memory.Use(context.Background())
		handler := &ResourceHandler{}
		entity, err := handler.Create(ctx, resourceRequest)
		So(err, ShouldBeNil)
		entity.Name = "Test Resource Name Updated"
		entity.Description = "Test Resource description"
		entity.Type = "ad_joined_machine"
		entity.OperatingSystem = "windows_machine"
		entity.ImageProject = ""
		entity.ImageFamily = "image-family"

		updateRequest := &proto.UpdateResourceRequest{
			Resource:   entity,
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name", "description", "type", "operating_system", "image_project", "image_family"}},
		}
		_, err = handler.Update(ctx, updateRequest)
		// should not save the resource with type machine as image project is missing
		So(err, ShouldNotBeNil)
	})
}

func TestResourceUpdateOfTypeMachineWithInvalidImageFamily(t *testing.T) {
	t.Parallel()
	resourceRequest := mockCreateResourceRequest("Test Resource Name", "Test Resource description", "ad_joined_machine", "windows_system", "image-project", "image-family")
	Convey("Update a resource with invalid image family in datastore if type is machine", t, func() {
		ctx := memory.Use(context.Background())
		handler := &ResourceHandler{}
		entity, err := handler.Create(ctx, resourceRequest)
		So(err, ShouldBeNil)
		entity.Name = "Test Resource Name Updated"
		entity.Description = "Test Resource description"
		entity.Type = "ad_joined_machine"
		entity.OperatingSystem = "windows_machine"
		entity.ImageProject = "image-project"
		entity.ImageFamily = ""

		updateRequest := &proto.UpdateResourceRequest{
			Resource:   entity,
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name", "description", "type", "operating_system", "image_project", "image_family"}},
		}
		_, err = handler.Update(ctx, updateRequest)
		// should not save the resource with type machine as image family is missing
		So(err, ShouldNotBeNil)
	})
}

func TestGetResourceWithValidData(t *testing.T) {
	resourceRequest := mockCreateResourceRequest("Test Resource", "Test Resource description", "ad_joined_machine", "windows_system", "image-project", "image-family")
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

		want := []string{entity.GetName(), entity.GetDescription(), entity.GetType(), entity.GetOperatingSystem(), entity.GetImageProject(), entity.GetImageFamily()}
		get := []string{readEntity.GetName(), readEntity.GetDescription(), readEntity.GetType(), readEntity.GetOperatingSystem(), readEntity.GetImageProject(), readEntity.GetImageFamily()}
		So(get, ShouldResemble, want)
	})
}

func TestListResources(t *testing.T) {
	t.Parallel()

	resourceRequest1 := mockCreateResourceRequest("Test Resource1", "Test Resource description", "ad_joined_machine", "windows_system", "image-project-1", "image-family-1")
	resourceRequest2 := mockCreateResourceRequest("Test Resource2", "Test Resource description2", "ad_joined_machine", "linux_system", "image-project-2", "image-family-2")

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

func TestResourceDeleteWithValidData(t *testing.T) {
	t.Parallel()
	createRequest := mockCreateResourceRequest("Test Resource", "Test Resource description", "ad_joined_machine", "windows_machine", "image-project", "image-family")
	Convey("Create a resource in datastore", t, func() {
		ctx := memory.Use(context.Background())
		handler := &ResourceHandler{}
		model, err := handler.Create(ctx, createRequest)
		So(err, ShouldBeNil)
		deleteRequest := &proto.DeleteResourceRequest{ResourceId: model.ResourceId}
		_, err = handler.Delete(ctx, deleteRequest)
		So(err, ShouldBeNil)
		getRequest := &proto.GetResourceRequest{ResourceId: model.ResourceId}
		readEntity, err := handler.Get(ctx, getRequest)
		So(err, ShouldBeNil)
		So(readEntity.Deleted, ShouldEqual, true)
	})
}
