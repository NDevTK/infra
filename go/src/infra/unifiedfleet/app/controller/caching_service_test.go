// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
	"google.golang.org/genproto/protobuf/field_mask"

	ufspb "infra/unifiedfleet/api/v1/models"
	"infra/unifiedfleet/app/model/caching"
	. "infra/unifiedfleet/app/model/datastore"
	"infra/unifiedfleet/app/model/history"
	"infra/unifiedfleet/app/model/state"
)

func mockCachingService(name string) *ufspb.CachingService {
	return &ufspb.CachingService{
		Name: name,
	}
}

func TestCreateCachingService(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	Convey("CreateCachingService", t, func() {
		Convey("Create new CachingService - happy path", func() {
			cs := mockCachingService("127.0.0.1")
			cs.State = ufspb.State_STATE_SERVING
			resp, err := CreateCachingService(ctx, cs)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp, ShouldResembleProto, cs)

			s, err := state.GetStateRecord(ctx, "cachingservices/127.0.0.1")
			So(err, ShouldBeNil)
			So(s.GetState(), ShouldEqual, ufspb.State_STATE_SERVING)

			changes, err := history.QueryChangesByPropertyName(ctx, "name", "cachingservices/127.0.0.1")
			So(err, ShouldBeNil)
			So(changes, ShouldHaveLength, 1)
			So(changes[0].GetOldValue(), ShouldEqual, LifeCycleRegistration)
			So(changes[0].GetNewValue(), ShouldEqual, LifeCycleRegistration)
			So(changes[0].GetEventLabel(), ShouldEqual, "cachingservice")

			msgs, err := history.QuerySnapshotMsgByPropertyName(ctx, "resource_name", "cachingservices/127.0.0.1")
			So(err, ShouldBeNil)
			So(msgs, ShouldHaveLength, 1)
			So(msgs[0].Delete, ShouldBeFalse)
		})

		Convey("Create new CachingService - already existing", func() {
			cs1 := mockCachingService("128.0.0.1")
			caching.CreateCachingService(ctx, cs1)

			cs2 := mockCachingService("128.0.0.1")
			_, err := CreateCachingService(ctx, cs2)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "already exists")

			changes, err := history.QueryChangesByPropertyName(ctx, "name", "cachingservices/128.0.0.1")
			So(err, ShouldBeNil)
			So(changes, ShouldHaveLength, 0)

			msgs, err := history.QuerySnapshotMsgByPropertyName(ctx, "resource_name", "cachingservices/128.0.0.1")
			So(err, ShouldBeNil)
			So(msgs, ShouldHaveLength, 0)
		})
	})
}

func TestUpdateCachingService(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	Convey("UpdateCachingService", t, func() {
		Convey("Update CachingService for existing CachingService - happy path", func() {
			cs1 := mockCachingService("127.0.0.1")
			cs1.Port = 43560
			caching.CreateCachingService(ctx, cs1)

			cs2 := mockCachingService("127.0.0.1")
			cs2.Port = 25653
			resp, _ := UpdateCachingService(ctx, cs2, nil)
			So(resp, ShouldNotBeNil)
			So(resp, ShouldResembleProto, cs2)

			changes, err := history.QueryChangesByPropertyName(ctx, "name", "cachingservices/127.0.0.1")
			So(err, ShouldBeNil)
			So(changes, ShouldHaveLength, 1)
			So(changes[0].GetEventLabel(), ShouldEqual, "cachingservice.port")
			So(changes[0].GetOldValue(), ShouldEqual, "43560")
			So(changes[0].GetNewValue(), ShouldEqual, "25653")

			msgs, err := history.QuerySnapshotMsgByPropertyName(ctx, "resource_name", "cachingservices/127.0.0.1")
			So(err, ShouldBeNil)
			So(msgs, ShouldHaveLength, 1)
			So(msgs[0].Delete, ShouldBeFalse)
		})

		Convey("Update CachingService for non-existing CachingService", func() {
			cs := mockCachingService("128.0.0.1")
			resp, err := UpdateCachingService(ctx, cs, nil)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)

			changes, err := history.QueryChangesByPropertyName(ctx, "name", "cachingservices/128.0.0.1")
			So(err, ShouldBeNil)
			So(changes, ShouldHaveLength, 0)
		})

		Convey("Update CachingService for existing CachingService - partial update state", func() {
			cs1 := mockCachingService("129.0.0.1")
			cs1.Port = 10101
			cs1.PrimaryNode = "0.0.0.0"
			cs1.State = ufspb.State_STATE_SERVING
			caching.CreateCachingService(ctx, cs1)

			cs2 := mockCachingService("129.0.0.1")
			cs2.State = ufspb.State_STATE_DISABLED
			resp, _ := UpdateCachingService(ctx, cs2, &field_mask.FieldMask{Paths: []string{"state"}})
			So(resp, ShouldNotBeNil)
			So(resp.GetName(), ShouldEqual, cs2.GetName())
			So(resp.GetPort(), ShouldEqual, 10101)
			So(resp.GetPrimaryNode(), ShouldEqual, "0.0.0.0")
			So(resp.GetState(), ShouldEqual, ufspb.State_STATE_DISABLED)

			s, err := state.GetStateRecord(ctx, "cachingservices/129.0.0.1")
			So(err, ShouldBeNil)
			So(s.GetState(), ShouldEqual, ufspb.State_STATE_DISABLED)

			changes, err := history.QueryChangesByPropertyName(ctx, "name", "cachingservices/129.0.0.1")
			So(err, ShouldBeNil)
			So(changes, ShouldHaveLength, 1)
			So(changes[0].GetEventLabel(), ShouldEqual, "cachingservice.state")
			So(changes[0].GetOldValue(), ShouldEqual, ufspb.State_STATE_SERVING.String())
			So(changes[0].GetNewValue(), ShouldEqual, ufspb.State_STATE_DISABLED.String())

			msgs, err := history.QuerySnapshotMsgByPropertyName(ctx, "resource_name", "cachingservices/129.0.0.1")
			So(err, ShouldBeNil)
			So(msgs, ShouldHaveLength, 1)
			So(msgs[0].Delete, ShouldBeFalse)
		})
	})
}

func TestGetCachingService(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	cs, _ := caching.CreateCachingService(ctx, &ufspb.CachingService{
		Name: "127.0.0.1",
	})
	Convey("GetCachingService", t, func() {
		Convey("Get CachingService by existing ID - happy path", func() {
			resp, _ := GetCachingService(ctx, "127.0.0.1")
			So(resp, ShouldNotBeNil)
			So(resp, ShouldResembleProto, cs)
		})

		Convey("Get CachingService by non-existing ID", func() {
			_, err := GetCachingService(ctx, "128.0.0.1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
	})
}

func TestDeleteCachingService(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	caching.CreateCachingService(ctx, &ufspb.CachingService{
		Name: "127.0.0.1",
	})
	Convey("DeleteCachingService", t, func() {
		Convey("Delete CachingService by existing ID - happy path", func() {
			err := DeleteCachingService(ctx, "127.0.0.1")
			So(err, ShouldBeNil)

			res, err := caching.GetCachingService(ctx, "127.0.0.1")
			So(res, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)

			_, err = state.GetStateRecord(ctx, "cachingservices/127.0.0.1")
			So(err.Error(), ShouldContainSubstring, NotFound)

			changes, err := history.QueryChangesByPropertyName(ctx, "name", "cachingservices/127.0.0.1")
			So(err, ShouldBeNil)
			So(changes, ShouldHaveLength, 1)
			So(changes[0].GetOldValue(), ShouldEqual, LifeCycleRetire)
			So(changes[0].GetNewValue(), ShouldEqual, LifeCycleRetire)
			So(changes[0].GetEventLabel(), ShouldEqual, "cachingservice")

			msgs, err := history.QuerySnapshotMsgByPropertyName(ctx, "resource_name", "cachingservices/127.0.0.1")
			So(err, ShouldBeNil)
			So(msgs, ShouldHaveLength, 1)
			So(msgs[0].Delete, ShouldBeTrue)
		})

		Convey("Delete CachingService by non-existing ID", func() {
			err := DeleteCachingService(ctx, "128.0.0.1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
	})
}

func TestListCachingServices(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	cachingServicesWithState := make([]*ufspb.CachingService, 0, 2)
	cachingServices := make([]*ufspb.CachingService, 0, 4)
	for i := 0; i < 4; i++ {
		cs := mockCachingService(fmt.Sprintf("cs-%d", i))
		if i%2 == 0 {
			cs.State = ufspb.State_STATE_SERVING
		}
		resp, _ := caching.CreateCachingService(ctx, cs)
		if i%2 == 0 {
			cachingServicesWithState = append(cachingServicesWithState, resp)
		}
		cachingServices = append(cachingServices, resp)
	}
	Convey("ListCachingServices", t, func() {
		Convey("List CachingServices - filter invalid - error", func() {
			_, _, err := ListCachingServices(ctx, 5, "", "invalid=mx-1", false)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Invalid field name invalid")
		})

		Convey("List CachingServices - filter switch - happy path", func() {
			resp, _, _ := ListCachingServices(ctx, 5, "", "state=serving", false)
			So(resp, ShouldNotBeNil)
			So(resp, ShouldResembleProto, cachingServicesWithState)
		})

		Convey("ListCachingServices - Full listing - happy path", func() {
			resp, _, _ := ListCachingServices(ctx, 5, "", "", false)
			So(resp, ShouldNotBeNil)
			So(resp, ShouldResembleProto, cachingServices)
		})
	})
}
