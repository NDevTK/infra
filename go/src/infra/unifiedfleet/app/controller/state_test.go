// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"

	ufspb "infra/unifiedfleet/api/v1/models"
	"infra/unifiedfleet/app/model/history"
	"infra/unifiedfleet/app/model/inventory"
	"infra/unifiedfleet/app/model/state"
	"infra/unifiedfleet/app/util"
)

func TestGetState(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	bm1, _ := state.UpdateStateRecord(ctx, &ufspb.StateRecord{
		ResourceName: "machine/browser-machine-1",
	})
	state.UpdateStateRecord(ctx, &ufspb.StateRecord{
		ResourceName: "machine/os-machine-1",
		State:        ufspb.State_STATE_REGISTERED,
	})
	os2Registered, _ := state.UpdateStateRecord(ctx, &ufspb.StateRecord{
		ResourceName: "machine/os-machine-2",
		State:        ufspb.State_STATE_REGISTERED,
	})
	osCtx, _ := util.SetupDatastoreNamespace(ctx, util.OSNamespace)
	os1Serving, _ := state.UpdateStateRecord(osCtx, &ufspb.StateRecord{
		ResourceName: "machine/os-machine-1",
		State:        ufspb.State_STATE_SERVING,
	})
	os3Serving, _ := state.UpdateStateRecord(osCtx, &ufspb.StateRecord{
		ResourceName: "machine/os-machine-3",
		State:        ufspb.State_STATE_SERVING,
	})
	Convey("GetState", t, func() {
		Convey("GetState for a browser machine with default namespace context", func() {
			res, err := GetState(ctx, "machine/browser-machine-1")
			So(err, ShouldBeNil)
			So(res, ShouldResembleProto, bm1)
		})

		Convey("GetState for a os machine with default namespace context", func() {
			res, err := GetState(ctx, "machine/os-machine-1")
			So(err, ShouldBeNil)
			// TODO(eshwarn): change this check when fall back read is removed
			So(res, ShouldResembleProto, os1Serving)
			res, err = GetState(ctx, "machine/os-machine-2")
			So(err, ShouldBeNil)
			So(res, ShouldResembleProto, os2Registered)
			res, err = GetState(ctx, "machine/os-machine-3")
			So(err, ShouldBeNil)
			So(res, ShouldResembleProto, os3Serving)
		})

		Convey("GetState for a os machine with os namespace context", func() {
			res, err := GetState(osCtx, "machine/os-machine-1")
			So(err, ShouldBeNil)
			So(res, ShouldResembleProto, os1Serving)
			res, err = GetState(osCtx, "machine/os-machine-2")
			So(err, ShouldBeNil)
			// TODO(eshwarn): change this check when fall back read is removed
			So(res, ShouldResembleProto, os2Registered)
			res, err = GetState(osCtx, "machine/os-machine-3")
			So(err, ShouldBeNil)
			So(res, ShouldResembleProto, os3Serving)
		})
	})
}

func TestUpdateState(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	// os namespace context
	osCtx, _ := util.SetupDatastoreNamespace(ctx, util.OSNamespace)
	Convey("UpdateState", t, func() {
		Convey("UpdateState for machine only in os namespace", func() {
			// creating in os namespace
			state.UpdateStateRecord(osCtx, &ufspb.StateRecord{
				ResourceName: "machines/os-machine-1",
				State:        ufspb.State_STATE_SERVING,
			})

			sr := &ufspb.StateRecord{
				ResourceName: "machines/os-machine-1",
				State:        ufspb.State_STATE_NEEDS_REPAIR,
			}
			res, err := UpdateState(osCtx, sr)
			So(err, ShouldBeNil)
			So(res, ShouldResembleProto, sr)

			res, err = state.GetStateRecord(osCtx, "machines/os-machine-1")
			So(err, ShouldBeNil)
			So(res.GetResourceName(), ShouldEqual, "machines/os-machine-1")
			So(res.GetState(), ShouldEqual, ufspb.State_STATE_NEEDS_REPAIR)

			changes, err := history.QueryChangesByPropertyName(osCtx, "name", "states/machines/os-machine-1")
			So(err, ShouldBeNil)
			So(changes, ShouldHaveLength, 1)
			So(changes[0].GetName(), ShouldEqual, "states/machines/os-machine-1")
			So(changes[0].GetOldValue(), ShouldEqual, ufspb.State_STATE_SERVING.String())
			So(changes[0].GetNewValue(), ShouldEqual, ufspb.State_STATE_NEEDS_REPAIR.String())
			So(changes[0].GetEventLabel(), ShouldEqual, "state_record.state")
			msgs, err := history.QuerySnapshotMsgByPropertyName(osCtx, "resource_name", "states/machines/os-machine-1")
			So(err, ShouldBeNil)
			So(msgs, ShouldHaveLength, 1)
			So(msgs[0].Delete, ShouldBeFalse)
		})

		Convey("UpdateState for non-existing machine in os namespace", func() {
			sr := &ufspb.StateRecord{
				ResourceName: "machines/os-machine-2",
				State:        ufspb.State_STATE_NEEDS_REPAIR,
			}
			res, err := UpdateState(osCtx, sr)
			So(err, ShouldBeNil)
			So(res, ShouldResembleProto, sr)

			res, err = state.GetStateRecord(osCtx, "machines/os-machine-2")
			So(err, ShouldBeNil)
			So(res.GetResourceName(), ShouldEqual, "machines/os-machine-2")
			So(res.GetState(), ShouldEqual, ufspb.State_STATE_NEEDS_REPAIR)

			changes, err := history.QueryChangesByPropertyName(osCtx, "name", "states/machines/os-machine-2")
			So(err, ShouldBeNil)
			So(changes, ShouldHaveLength, 1)
			So(changes[0].GetName(), ShouldEqual, "states/machines/os-machine-2")
			So(changes[0].GetOldValue(), ShouldEqual, ufspb.State_STATE_UNSPECIFIED.String())
			So(changes[0].GetNewValue(), ShouldEqual, ufspb.State_STATE_NEEDS_REPAIR.String())
			So(changes[0].GetEventLabel(), ShouldEqual, "state_record.state")
			msgs, err := history.QuerySnapshotMsgByPropertyName(osCtx, "resource_name", "states/machines/os-machine-2")
			So(err, ShouldBeNil)
			So(msgs, ShouldHaveLength, 1)
			So(msgs[0].Delete, ShouldBeFalse)
		})
		Convey("UpdateState for machine lse lacking perms", func() {
			// user has no realm permissions
			noPermsCtx := withAuthorizedNoPermsUser(osCtx)
			state.UpdateStateRecord(noPermsCtx, &ufspb.StateRecord{
				ResourceName: "machinelses/os-machine-3",
				State:        ufspb.State_STATE_SERVING,
			})
			inventory.CreateMachineLSE(noPermsCtx, &ufspb.MachineLSE{
				Name:  "os-machine-3",
				Realm: util.AtlLabAdminRealm,
			})

			sr := &ufspb.StateRecord{
				ResourceName: "machineslses/os-machine-3",
				State:        ufspb.State_STATE_NEEDS_REPAIR,
			}
			res, err := UpdateState(noPermsCtx, sr)
			So(res, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "PermissionDenied")

			changes, err := history.QueryChangesByPropertyName(osCtx, "name", "states/machines/os-machine-3")
			So(err, ShouldBeNil)
			So(changes, ShouldHaveLength, 0)
			msgs, err := history.QuerySnapshotMsgByPropertyName(osCtx, "resource_name", "states/machines/os-machine-3")
			So(err, ShouldBeNil)
			So(msgs, ShouldHaveLength, 0)
		})
		Convey("UpdateState for machine lse with perms", func() {
			// user has correct realm perms
			atlPermsCtx := withAuthorizedAtlUser(osCtx)
			state.UpdateStateRecord(atlPermsCtx, &ufspb.StateRecord{
				ResourceName: "machinelses/os-machine-4",
				State:        ufspb.State_STATE_SERVING,
			})
			inventory.CreateMachineLSE(atlPermsCtx, &ufspb.MachineLSE{
				Name:  "os-machine-4",
				Realm: util.AtlLabAdminRealm,
			})

			sr := &ufspb.StateRecord{
				ResourceName: "machinelses/os-machine-4",
				State:        ufspb.State_STATE_NEEDS_REPAIR,
			}
			res, err := UpdateState(atlPermsCtx, sr)
			So(err, ShouldBeNil)
			So(res, ShouldResembleProto, sr)

			res, err = state.GetStateRecord(atlPermsCtx, "machinelses/os-machine-4")
			So(err, ShouldBeNil)
			So(res.GetResourceName(), ShouldEqual, "machinelses/os-machine-4")
			So(res.GetState(), ShouldEqual, ufspb.State_STATE_NEEDS_REPAIR)

			changes, err := history.QueryChangesByPropertyName(atlPermsCtx, "name", "states/machinelses/os-machine-4")
			So(err, ShouldBeNil)
			So(changes, ShouldHaveLength, 1)
			So(changes[0].GetName(), ShouldEqual, "states/machinelses/os-machine-4")
			So(changes[0].GetOldValue(), ShouldEqual, ufspb.State_STATE_SERVING.String())
			So(changes[0].GetNewValue(), ShouldEqual, ufspb.State_STATE_NEEDS_REPAIR.String())
			So(changes[0].GetEventLabel(), ShouldEqual, "state_record.state")
			msgs, err := history.QuerySnapshotMsgByPropertyName(atlPermsCtx, "resource_name", "states/machinelses/os-machine-4")
			So(err, ShouldBeNil)
			So(msgs, ShouldHaveLength, 1)
			So(msgs[0].Delete, ShouldBeFalse)
		})
	})
}
