// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package registration

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/appengine/gaetesting"
	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/gae/service/datastore"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/auth/authtest"
	"google.golang.org/protobuf/testing/protocmp"

	ufspb "infra/unifiedfleet/api/v1/models"
	ufsutil "infra/unifiedfleet/app/util"
)

func mockAsset(name, model, host string, assettype ufspb.AssetType, zone ufspb.Zone) *ufspb.Asset {
	return &ufspb.Asset{
		Name:  name,
		Type:  assettype,
		Model: model,
		Location: &ufspb.Location{
			BarcodeName: host,
			Zone:        zone,
		},
	}
}

func assertAssetEqual(a, b *ufspb.Asset) {
	So(cmp.Equal(a, b, protocmp.Transform()), ShouldEqual, true)
}

func TestCreateAsset(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	datastore.GetTestable(ctx).Consistent(true)
	asset1 := mockAsset("C001001", "krane", "cros4-row3-rack5-host4", ufspb.AssetType_DUT, ufspb.Zone_ZONE_CHROMEOS4)
	asset2 := mockAsset("C001001", "Servo V4", "cros5-row3-rack5-host4", ufspb.AssetType_SERVO, ufspb.Zone_ZONE_CHROMEOS5)
	asset3 := mockAsset("", "eve", "cros6-row3-rack5-host4", ufspb.AssetType_DUT, ufspb.Zone_ZONE_CHROMEOS6)
	asset4 := mockAsset("C002002", "eve", "cros7-row3-rack5-host4", ufspb.AssetType_UNDEFINED, ufspb.Zone_ZONE_CHROMEOS7)
	asset6 := mockAsset("C004004", "eve", "cros9-row3-rack5-host4", ufspb.AssetType_DUT, ufspb.Zone_ZONE_UNSPECIFIED)
	asset6.Location = nil
	Convey("CreateAsset", t, func() {
		Convey("Create new asset", func() {
			resp, err := CreateAsset(ctx, asset1)
			So(err, ShouldBeNil)
			assertAssetEqual(asset1, resp)
		})
		Convey("Create existing asset", func() {
			_, err := CreateAsset(ctx, asset2)
			So(err, ShouldNotBeNil)
		})
		Convey("Create asset with invalid name", func() {
			_, err := CreateAsset(ctx, asset3)
			So(err, ShouldNotBeNil)
		})
		Convey("Create asset without type", func() {
			_, err := CreateAsset(ctx, asset4)
			So(err, ShouldNotBeNil)
		})
		Convey("Create asset without location", func() {
			_, err := CreateAsset(ctx, asset6)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestUpdateAsset(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	datastore.GetTestable(ctx).Consistent(true)
	asset1 := mockAsset("C001001", "krane", "cros4-row3-rack5-host4", ufspb.AssetType_DUT, ufspb.Zone_ZONE_CHROMEOS4)
	asset2 := mockAsset("C001001", "Servo V3", "cros6-row3-rack5-host4", ufspb.AssetType_SERVO, ufspb.Zone_ZONE_CHROMEOS6)
	asset3 := mockAsset("C002002", "Whizz", "cros6-row3-rack5-host4", ufspb.AssetType_LABSTATION, ufspb.Zone_ZONE_CHROMEOS6)
	asset4 := mockAsset("", "Whizz-Labstation", "cros6-row3-rack5-host4", ufspb.AssetType_UNDEFINED, ufspb.Zone_ZONE_CHROMEOS6)
	Convey("UpdateAsset", t, func() {
		Convey("Update existing Asset", func() {
			resp, err := CreateAsset(ctx, asset1)
			So(err, ShouldBeNil)
			assertAssetEqual(asset1, resp)
			resp, err = UpdateAsset(ctx, asset2)
			So(err, ShouldBeNil)
			assertAssetEqual(asset2, resp)
		})
		Convey("Update non-existinent asset", func() {
			resp, err := UpdateAsset(ctx, asset3)
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
		})
		Convey("Update asset with invalid name", func() {
			resp, err := UpdateAsset(ctx, asset4)
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
		})
	})
}

func TestGetAsset(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	asset1 := mockAsset("C001001", "krane", "cros4-row3-rack5-host4", ufspb.AssetType_DUT, ufspb.Zone_ZONE_CHROMEOS4)
	Convey("GetAsset", t, func() {
		Convey("Get asset by existing name", func() {
			resp, err := CreateAsset(ctx, asset1)
			So(err, ShouldBeNil)
			assertAssetEqual(resp, asset1)
			resp, err = GetAsset(ctx, asset1.GetName())
			So(err, ShouldBeNil)
			assertAssetEqual(resp, asset1)
		})
		Convey("Get asset by non-existent name", func() {
			resp, err := GetAsset(ctx, "C009009")
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
		})
		Convey("Get asset by invalid name", func() {
			resp, err := GetAsset(ctx, "")
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
		})
	})
}

func TestDeleteAsset(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	asset1 := mockAsset("C001001", "krane", "cros4-row3-rack5-host4", ufspb.AssetType_DUT, ufspb.Zone_ZONE_CHROMEOS4)
	Convey("DeleteAsset", t, func() {
		Convey("Delete asset by existing name", func() {
			resp, cerr := CreateAsset(ctx, asset1)
			So(cerr, ShouldBeNil)
			assertAssetEqual(resp, asset1)
			err := DeleteAsset(ctx, asset1.GetName())
			So(err, ShouldBeNil)
			res, err := GetAsset(ctx, asset1.GetName())
			So(res, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})
		Convey("Delete asset by non-existing name", func() {
			err := DeleteAsset(ctx, "C000000")
			So(err, ShouldNotBeNil)
		})
		Convey("Delete asset - invalid name", func() {
			err := DeleteAsset(ctx, "")
			So(err, ShouldNotBeNil)
		})
	})
}

func TestListAssets(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	assets := make([]*ufspb.Asset, 0, 10)
	for i := 0; i < 10; i++ {
		asset := mockAsset(fmt.Sprintf("C00000%d", i), "eve", fmt.Sprintf("cros4-row3-rack5-host%d", i), ufspb.AssetType_DUT, ufspb.Zone_ZONE_CHROMEOS4)
		resp, _ := CreateAsset(ctx, asset)
		assets = append(assets, resp)
	}
	Convey("ListAssets", t, func() {
		Convey("List assets - page_token invalid", func() {
			resp, nextPageToken, err := ListAssets(ctx, 5, "abc", nil, false)
			So(resp, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(err, ShouldNotBeNil)
		})

		Convey("List assets - Full listing with no pagination", func() {
			resp, nextPageToken, err := ListAssets(ctx, 10, "", nil, false)
			So(resp, ShouldNotBeNil)
			So(nextPageToken, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, assets)
		})

		Convey("List assets - listing with pagination", func() {
			resp, nextPageToken, err := ListAssets(ctx, 3, "", nil, false)
			So(resp, ShouldNotBeNil)
			So(nextPageToken, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, assets[:3])

			resp, _, err = ListAssets(ctx, 7, nextPageToken, nil, false)
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, assets[3:])
		})
	})
}

func TestListAssetsACL(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	assets := make([]*ufspb.Asset, 0, 20)
	for i := 0; i < 10; i++ {
		asset := mockAsset(fmt.Sprintf("C00000%d", i), "eve", fmt.Sprintf("chromeos4-row3-rack5-host%d", i), ufspb.AssetType_DUT, ufspb.Zone_ZONE_CHROMEOS4)
		resp, _ := CreateAsset(ctx, asset)
		assets = append(assets, resp)
	}
	for i := 0; i < 10; i++ {
		asset := mockAsset(fmt.Sprintf("C00001%d", i), "eve", fmt.Sprintf("chromeos5-row3-rack5-host%d", i), ufspb.AssetType_DUT, ufspb.Zone_ZONE_CHROMEOS5)
		resp, _ := CreateAsset(ctx, asset)
		assets = append(assets, resp)
	}
	// User bat has permissions in atl lab
	ctxBat := auth.WithState(ctx, &authtest.FakeState{
		Identity: "user:bat@man.com",
		IdentityPermissions: []authtest.RealmPermission{
			{
				Realm:      ufsutil.AtlLabAdminRealm,
				Permission: ufsutil.RegistrationsList,
			},
		},
	})
	// User spider has permissions in both atl and acs lab
	ctxSpider := auth.WithState(ctx, &authtest.FakeState{
		Identity: "user:spider@man.com",
		IdentityPermissions: []authtest.RealmPermission{
			{
				Realm:      ufsutil.AtlLabAdminRealm,
				Permission: ufsutil.RegistrationsList,
			},
			{
				Realm:      ufsutil.AcsLabAdminRealm,
				Permission: ufsutil.RegistrationsList,
			},
		},
	})
	// User mermaid has no permissions
	ctxMermaid := auth.WithState(ctx, &authtest.FakeState{
		Identity: "user:mermaid@man.com",
	})
	// Anonymous User has no permissions
	ctx = auth.WithState(ctx, &authtest.FakeState{})
	Convey("ListAssetsACL", t, func() {
		Convey("List assets - AnonymousUser", func() {
			// User anonymous has an error
			resp, nextPageToken, err := ListAssetsACL(ctx, 100, "", nil, false)
			So(err, ShouldBeNil)
			So(resp, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
		})
		Convey("List assets - filter on realm", func() {
			// User bat cannot filter on realm
			resp, nextPageToken, err := ListAssetsACL(ctxBat, 3, "", map[string][]interface{}{"realm": {"test"}}, false)
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
		})
		Convey("List assets - happy path, no permissions", func() {
			// User mermaid has no permissions, sees nothing
			resp, nextPageToken, err := ListAssetsACL(ctxMermaid, 100, "", nil, false)
			So(err, ShouldBeNil)
			So(resp, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
		})
		Convey("List assets - happy path, single realm", func() {
			// User bat has permissions only for ATL and sees only those assets
			resp, nextPageToken, err := ListAssetsACL(ctxBat, 3, "", nil, false)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp, ShouldResembleProto, assets[:3])
			So(nextPageToken, ShouldNotBeEmpty)

			resp, nextPageToken, err = ListAssetsACL(ctxBat, 100, nextPageToken, nil, false)
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, assets[3:10])
			So(nextPageToken, ShouldBeEmpty)
		})
		Convey("List assets - happy path, two realms", func() {
			// User spider has permissions for both ATL and ACS. sees all the assets
			resp, nextPageToken, err := ListAssetsACL(ctxSpider, 3, "", nil, false)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp, ShouldResembleProto, assets[:3])
			So(nextPageToken, ShouldNotBeEmpty)

			resp, nextPageToken, err = ListAssetsACL(ctxSpider, 7, nextPageToken, nil, false)
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, assets[3:10])
			So(nextPageToken, ShouldNotBeEmpty)

			resp, nextPageToken, err = ListAssetsACL(ctxSpider, 100, nextPageToken, nil, false)
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, assets[10:])
			So(nextPageToken, ShouldBeEmpty)
		})
		Convey("List assets - happy path, two realms, zone filter", func() {
			// User spider has permissions for both ATL and ACS. But filters for chromeos4 so sees only those assets
			resp, nextPageToken, err := ListAssetsACL(ctxSpider, 3, "", map[string][]interface{}{"zone": {"ZONE_CHROMEOS4"}}, false)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp, ShouldResembleProto, assets[:3])
			So(nextPageToken, ShouldNotBeEmpty)

			resp, nextPageToken, err = ListAssetsACL(ctxSpider, 100, nextPageToken, map[string][]interface{}{"zone": {"ZONE_CHROMEOS4"}}, false)
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, assets[3:10])
			So(nextPageToken, ShouldBeEmpty)
		})
		Convey("List assets - happy path, one realms, wrong zone filter", func() {
			// User bat has permissions for ATL. Attempts to filter on zone chromeos5
			resp, nextPageToken, err := ListAssetsACL(ctxBat, 100, "", map[string][]interface{}{"zone": {"ZONE_CHROMEOS5"}}, false)
			So(err, ShouldBeNil)
			So(resp, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
		})
	})
}

func TestBatchUpdateAssets(t *testing.T) {
	t.Parallel()
	Convey("BatchUpdateAssets", t, func() {
		ctx := gaetesting.TestingContextWithAppID("go-test")
		datastore.GetTestable(ctx).Consistent(true)
		assets := make([]*ufspb.Asset, 0, 4)
		for i := 0; i < 4; i++ {
			asset := mockAsset(fmt.Sprintf("C0000%d0", i), "eve", fmt.Sprintf("cros4-row3-rack5-host%d", i), ufspb.AssetType_DUT, ufspb.Zone_ZONE_CHROMEOS4)
			resp, err := CreateAsset(ctx, asset)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, asset)
			asset.Model = "krane"
			assets = append(assets, resp)
		}
		Convey("BatchUpdate all assets", func() {
			resp, err := BatchUpdateAssets(ctx, assets)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, assets)
		})
		Convey("BatchUpdate existing and invalid assets", func() {
			asset := mockAsset("", "krane", "cros4-row3-rack5-host4", ufspb.AssetType_DUT, ufspb.Zone_ZONE_CHROMEOS4)
			assets = append(assets, asset)
			resp, err := BatchUpdateAssets(ctx, assets)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestGetAllAssets(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	datastore.GetTestable(ctx).Consistent(true)
	Convey("GetAllAssets", t, func() {
		Convey("GetAllAssets - Emtpy database", func() {
			resp, err := GetAllAssets(ctx)
			So(len(resp), ShouldEqual, 0)
			So(err, ShouldBeNil)
		})
		Convey("GetAllAssets - non-empty database", func() {
			assets := make([]*ufspb.Asset, 0, 10)
			for i := 0; i < 10; i++ {
				asset := mockAsset(fmt.Sprintf("C000300%d", i), "eve", fmt.Sprintf("cros6-row7-rack5-host%d", i), ufspb.AssetType_DUT, ufspb.Zone_ZONE_CHROMEOS4)
				resp, err := CreateAsset(ctx, asset)
				So(err, ShouldBeNil)
				assets = append(assets, resp)
			}
			resp, err := GetAllAssets(ctx)
			So(len(resp), ShouldEqual, 10)
			So(len(assets), ShouldEqual, 10)
			So(assets, ShouldResembleProto, resp)
			So(err, ShouldBeNil)
		})
	})
}

func TestQueryAssetByPropertyName(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	datastore.GetTestable(ctx).Consistent(true)
	// Add some assets to the database to test
	assets := []*ufspb.Asset{
		{
			Name:  "A1",
			Type:  ufspb.AssetType_DUT,
			Model: "test",
			Location: &ufspb.Location{
				Zone: ufspb.Zone_ZONE_CHROMEOS6,
				Rack: "chromeos6-row1-rack1",
			},
			Info: &ufspb.AssetInfo{
				Model:       "test",
				BuildTarget: "notest",
				Phase:       "14",
			},
			Tags: []string{"dut", "no-battery"},
		},
		{
			Name:  "B1",
			Type:  ufspb.AssetType_LABSTATION,
			Model: "labtest",
			Location: &ufspb.Location{
				Zone: ufspb.Zone_ZONE_CHROMEOS6,
				Rack: "chromeos6-row1-rack1",
			},
			Info: &ufspb.AssetInfo{
				Model:       "labtest",
				BuildTarget: "labnotest",
				Phase:       "15",
			},
			Tags: []string{"labstation", "pending decommission"},
		},
		{
			Name: "S1",
			Type: ufspb.AssetType_SERVO,
			Location: &ufspb.Location{
				Zone: ufspb.Zone_ZONE_CHROMEOS6,
				Rack: "chromeos6-row1-rack1",
			},
			Tags: []string{"servo v4", "stable"},
		},
		{
			Name:  "A2",
			Type:  ufspb.AssetType_DUT,
			Model: "test",
			Location: &ufspb.Location{
				Zone: ufspb.Zone_ZONE_CHROMEOS2,
				Rack: "chromeos2-row2-rack1",
			},
			Info: &ufspb.AssetInfo{
				Model:       "test",
				BuildTarget: "notest",
				Phase:       "14",
			},
			Tags: []string{"dut", "no-battery"},
		},
		{
			Name:  "B2",
			Type:  ufspb.AssetType_LABSTATION,
			Model: "labtest",
			Location: &ufspb.Location{
				Zone: ufspb.Zone_ZONE_CHROMEOS2,
				Rack: "chromeos2-row2-rack1",
			},
			Info: &ufspb.AssetInfo{
				Model:       "labtest",
				BuildTarget: "labnotest",
				Phase:       "15",
			},
			Tags: []string{"labstation", "pending decommission"},
		},
		{
			Name: "S2",
			Type: ufspb.AssetType_SERVO,
			Location: &ufspb.Location{
				Zone: ufspb.Zone_ZONE_CHROMEOS2,
				Rack: "chromeo2-row2-rack1",
			},
			Tags: []string{"servo v4", "stable"},
		},
	}
	for _, asset := range assets {
		CreateAsset(ctx, asset)
	}
	Convey("QueryAssetByPropertyName", t, func() {
		Convey("QueryAssetByPropertyName - NotFound", func() {
			resp, err := QueryAssetByPropertyName(ctx, "rack", "chromeosπ-rowk-rackh", false)
			So(err, ShouldBeNil)
			So(resp, ShouldBeNil)
		})
		Convey("QueryAssetByPropertyName - type; keys only", func() {
			resp, err := QueryAssetByPropertyName(ctx, "type", "LABSTATION", true)
			So(err, ShouldBeNil)
			So(resp, ShouldHaveLength, 2)
			So(resp[0].Name, ShouldEqual, "B1")
			So(resp[1].Name, ShouldEqual, "B2")
		})
		Convey("QueryAssetByPropertyName - zone; keys only", func() {
			resp, err := QueryAssetByPropertyName(ctx, "zone", "ZONE_CHROMEOS6", true)
			So(err, ShouldBeNil)
			So(resp, ShouldHaveLength, 3)
			So(resp[0].Name, ShouldEqual, "A1")
			So(resp[1].Name, ShouldEqual, "B1")
			So(resp[2].Name, ShouldEqual, "S1")
		})
		Convey("QueryAssetByPropertyName - phase; keys only", func() {
			resp, err := QueryAssetByPropertyName(ctx, "phase", "14", true)
			So(err, ShouldBeNil)
			So(resp, ShouldHaveLength, 2)
			So(resp[0].Name, ShouldEqual, "A1")
			So(resp[1].Name, ShouldEqual, "A2")
		})
		Convey("QueryAssetByPropertyName - tags; keys only", func() {
			resp, err := QueryAssetByPropertyName(ctx, "tags", "servo v4", true)
			So(err, ShouldBeNil)
			So(resp, ShouldHaveLength, 2)
			So(resp[0].Name, ShouldEqual, "S1")
			So(resp[1].Name, ShouldEqual, "S2")
		})
		Convey("QueryAssetByPropertyName - model; keys only", func() {
			resp, err := QueryAssetByPropertyName(ctx, "model", "test", true)
			So(err, ShouldBeNil)
			So(resp, ShouldHaveLength, 2)
			So(resp[0].Name, ShouldEqual, "A1")
			So(resp[1].Name, ShouldEqual, "A2")
		})
		Convey("QueryAssetByPropertyName - build_target; keys only", func() {
			resp, err := QueryAssetByPropertyName(ctx, "build_target", "labnotest", true)
			So(err, ShouldBeNil)
			So(resp, ShouldHaveLength, 2)
			So(resp[0].Name, ShouldEqual, "B1")
			So(resp[1].Name, ShouldEqual, "B2")
		})
		Convey("QueryAssetByPropertyName - type", func() {
			resp, err := QueryAssetByPropertyName(ctx, "type", "LABSTATION", false)
			So(err, ShouldBeNil)
			So(resp, ShouldHaveLength, 2)
			So(resp[0].Name, ShouldEqual, "B1")
			So(resp[1].Name, ShouldEqual, "B2")
		})
		Convey("QueryAssetByPropertyName - zone", func() {
			resp, err := QueryAssetByPropertyName(ctx, "zone", "ZONE_CHROMEOS6", false)
			So(err, ShouldBeNil)
			So(resp, ShouldHaveLength, 3)
			So(resp[0], ShouldResembleProto, assets[0]) // DUT A1
			So(resp[1], ShouldResembleProto, assets[1]) // Labstation B1
			So(resp[2], ShouldResembleProto, assets[2]) // Servo S1
		})
		Convey("QueryAssetByPropertyName - phase", func() {
			resp, err := QueryAssetByPropertyName(ctx, "phase", "14", false)
			So(err, ShouldBeNil)
			So(resp, ShouldHaveLength, 2)
			So(resp[0], ShouldResembleProto, assets[0]) // DUT A1
			So(resp[1], ShouldResembleProto, assets[3]) // DUT A2
		})
		Convey("QueryAssetByPropertyName - tags", func() {
			resp, err := QueryAssetByPropertyName(ctx, "tags", "servo v4", false)
			So(err, ShouldBeNil)
			So(resp, ShouldHaveLength, 2)
			So(resp[0], ShouldResembleProto, assets[2]) // Servo S1
			So(resp[1], ShouldResembleProto, assets[5]) // Servo S2
		})
		Convey("QueryAssetByPropertyName - model", func() {
			resp, err := QueryAssetByPropertyName(ctx, "model", "test", false)
			So(err, ShouldBeNil)
			So(resp, ShouldHaveLength, 2)
			So(resp[0], ShouldResembleProto, assets[0]) // DUT A1
			So(resp[1], ShouldResembleProto, assets[3]) // DUT A2
		})
		Convey("QueryAssetByPropertyName - build_target", func() {
			resp, err := QueryAssetByPropertyName(ctx, "build_target", "labnotest", false)
			So(err, ShouldBeNil)
			So(resp, ShouldHaveLength, 2)
			So(resp[0], ShouldResembleProto, assets[1]) // Labstation B1
			So(resp[1], ShouldResembleProto, assets[4]) // Labstation B2
		})
	})
}
