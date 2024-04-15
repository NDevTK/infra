// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package controller

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/genproto/protobuf/field_mask"

	. "go.chromium.org/luci/common/testing/assertions"

	ufspb "infra/unifiedfleet/api/v1/models"
	. "infra/unifiedfleet/app/model/datastore"
	"infra/unifiedfleet/app/model/history"
)

func TestCreateDefaultWifi(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	Convey("CreateDefaultWifi", t, func() {
		Convey("Create new DefaultWifi - happy path", func() {
			wifi := &ufspb.DefaultWifi{Name: "zone_sfo36_os"}
			resp, err := CreateDefaultWifi(ctx, wifi)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp, ShouldResembleProto, wifi)
		})
		Convey("Create new DefaultWifi - already existing", func() {
			w1 := &ufspb.DefaultWifi{Name: "pool1"}
			_, _ = CreateDefaultWifi(ctx, w1)

			dup := &ufspb.DefaultWifi{Name: "pool1"}
			_, err := CreateDefaultWifi(ctx, dup)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "already exists")
		})
	})
}

func TestDeleteDefaultWifi(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	CreateDefaultWifi(ctx, &ufspb.DefaultWifi{Name: "pool"})
	Convey("DeleteDefaultWifi", t, func() {
		Convey("Delete DefaultWifi by existing ID - happy path", func() {
			err := DeleteDefaultWifi(ctx, "pool")
			So(err, ShouldBeNil)

			res, err := GetDefaultWifi(ctx, "pool")
			So(res, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})

		Convey("Delete DefaultWifi by non-existing ID", func() {
			err := DeleteDefaultWifi(ctx, "non-existing")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
	})
}

func TestUpdateDefaultWifi(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	Convey("UpdateDefaultWifi", t, func() {
		Convey("Update DefaultWifi for existing DefaultWifi - happy path", func() {
			CreateDefaultWifi(ctx, &ufspb.DefaultWifi{
				Name: "zone_sfo36_os",
				WifiSecret: &ufspb.Secret{
					ProjectId:  "p1",
					SecretName: "s1",
				},
			})
			w2 := &ufspb.DefaultWifi{
				Name: "zone_sfo36_os",
				WifiSecret: &ufspb.Secret{
					ProjectId:  "p1",
					SecretName: "s2",
				}}
			resp, err := UpdateDefaultWifi(ctx, w2, nil)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp, ShouldResembleProto, w2)
			changes, err := history.QueryChangesByPropertyName(ctx, "name", "defaultwifis/zone_sfo36_os")
			So(err, ShouldBeNil)
			So(changes, ShouldHaveLength, 2)
			So(changes[1].GetEventLabel(), ShouldEqual, "defaultwifi.secret.secret_name")
			So(changes[1].GetOldValue(), ShouldEqual, "s1")
			So(changes[1].GetNewValue(), ShouldEqual, "s2")
		})

		Convey("Update DefaultWifi for non-existing DefaultWifi", func() {
			resp, err := UpdateDefaultWifi(ctx, &ufspb.DefaultWifi{Name: "pool3"}, nil)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)

			changes, err := history.QueryChangesByPropertyName(ctx, "name", "defaultwifis/pool3")
			So(err, ShouldBeNil)
			So(changes, ShouldHaveLength, 0)
		})

		Convey("Update DefaultWifi for existing DefaultWifi with field mask - happy path", func() {
			w3 := &ufspb.DefaultWifi{
				Name: "zone_sfo36_os",
				WifiSecret: &ufspb.Secret{
					ProjectId:  "ppppp",
					SecretName: "s3",
				}}
			resp, _ := UpdateDefaultWifi(ctx, w3, &field_mask.FieldMask{Paths: []string{"wifi_secret.secret_name"}})
			So(resp, ShouldNotBeNil)
			So(resp.GetWifiSecret().GetProjectId(), ShouldEqual, "p1")
			So(resp.GetWifiSecret().GetSecretName(), ShouldEqual, "s3")
		})

		Convey("Update DefaultWifi for existing DefaultWifi with field mask - failure", func() {
			w4 := &ufspb.DefaultWifi{
				Name: "zone_sfo36_os",
				WifiSecret: &ufspb.Secret{
					ProjectId:  "p1",
					SecretName: "s4",
				}}
			resp, err := UpdateDefaultWifi(ctx, w4, &field_mask.FieldMask{Paths: []string{"wifi_secret.non-existing-field"}})
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})
	})
}
