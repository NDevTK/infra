// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configuration

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.chromium.org/luci/appengine/gaetesting"
	"go.chromium.org/luci/gae/service/datastore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"

	ufspb "infra/unifiedfleet/api/v1/models"
)

func mockHwidData() *ufspb.HwidData {
	return &ufspb.HwidData{
		Sku:      "test-sku",
		Variant:  "test-variant",
		Hwid:     "test-hwid",
		DutLabel: mockDutLabel(),
	}
}

func mockDutLabel() *ufspb.DutLabel {
	return &ufspb.DutLabel{
		PossibleLabels: []string{
			"test-possible-1",
			"test-possible-2",
		},
		Labels: []*ufspb.DutLabel_Label{
			{
				Name:  "test-label-1",
				Value: "test-value-1",
			},
			{
				Name:  "Sku",
				Value: "test-sku",
			},
			{
				Name:  "variant",
				Value: "test-variant",
			},
		},
	}
}

func TestUpdateHwidData(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	datastore.GetTestable(ctx).Consistent(true)

	t.Run("update non-existent HwidData", func(t *testing.T) {
		want := mockHwidData()
		got, err := UpdateHwidData(ctx, want, "test-hwid")
		if err != nil {
			t.Fatalf("UpdateHwidData failed: %s", err)
		}
		gotProto, err := got.GetProto()
		if err != nil {
			t.Fatalf("GetProto failed: %s", err)
		}
		if diff := cmp.Diff(want, gotProto, protocmp.Transform()); diff != "" {
			t.Errorf("UpdateHwidData returned unexpected diff (-want +got):\n%s", diff)
		}
	})

	t.Run("update existent HwidData - standard new proto to new proto", func(t *testing.T) {
		hd2Id := "test-hwid-2"
		hd2 := mockHwidData()

		hd2update := mockHwidData()
		hd2update.DutLabel.PossibleLabels = append(hd2update.DutLabel.PossibleLabels, "test-possible-3")

		// Insert hd2 into datastore
		_, err := UpdateHwidData(ctx, hd2, hd2Id)
		if err != nil {
			t.Fatalf("UpdateHwidData failed: %s", err)
		}

		// Update hd2
		got, err := UpdateHwidData(ctx, hd2update, hd2Id)
		if err != nil {
			t.Fatalf("UpdateHwidData failed: %s", err)
		}
		gotProto, err := got.GetProto()
		if err != nil {
			t.Fatalf("GetProto failed: %s", err)
		}
		if diff := cmp.Diff(hd2update, gotProto, protocmp.Transform()); diff != "" {
			t.Errorf("UpdateHwidData returned unexpected diff (-want +got):\n%s", diff)
		}
	})

	t.Run("update existent HwidData - legacy proto to new proto", func(t *testing.T) {
		hd3Id := "test-hwid-3"
		hd3 := mockDutLabel()
		_, err := UpdateLegacyHwidData(ctx, hd3, hd3Id)
		if err != nil {
			t.Fatalf("UpdateLegacyHwidData failed: %s", err)
		}

		hd3update := mockHwidData()
		hd3update.DutLabel.PossibleLabels = append(hd3update.DutLabel.PossibleLabels, "test-possible-3")

		// Update hd3
		got, err := UpdateHwidData(ctx, hd3update, hd3Id)
		if err != nil {
			t.Fatalf("UpdateHwidData failed: %s", err)
		}
		gotProto, err := got.GetProto()
		if err != nil {
			t.Fatalf("GetProto failed: %s", err)
		}
		if diff := cmp.Diff(hd3update, gotProto, protocmp.Transform()); diff != "" {
			t.Errorf("UpdateHwidData returned unexpected diff (-want +got):\n%s", diff)
		}
	})

	t.Run("update existent HwidData - new proto to legacy proto", func(t *testing.T) {
		// This is to prove backwards compatibility. UpdateLegacyHwidData holds
		// the old implementation of the UpdateHwidData method signature.
		h4Id := "test-hwid-4"
		h4 := mockHwidData()
		_, err := UpdateHwidData(ctx, h4, h4Id)
		if err != nil {
			t.Fatalf("UpdateHwidData failed: %s", err)
		}

		h4update := mockDutLabel()
		h4update.PossibleLabels = append(h4update.PossibleLabels, "test-possible-3")

		// Update h4
		got, err := UpdateLegacyHwidData(ctx, h4update, h4Id)
		if err != nil {
			t.Fatalf("UpdateLegacyHwidData failed: %s", err)
		}
		gotProto, err := got.GetDutLabelProto()
		if err != nil {
			t.Fatalf("GetDutLabelProto failed: %s", err)
		}
		if diff := cmp.Diff(h4update, gotProto, protocmp.Transform()); diff != "" {
			t.Errorf("UpdateHwidData returned unexpected diff (-want +got):\n%s", diff)
		}
	})

	t.Run("update HwidData with empty hwid", func(t *testing.T) {
		hd5 := mockHwidData()
		got, err := UpdateHwidData(ctx, hd5, "")
		if err == nil {
			t.Errorf("UpdateHwidData succeeded with empty hwid")
		}
		if c := status.Code(err); c != codes.InvalidArgument {
			t.Errorf("Unexpected error when calling UpdateHwidData: %s", err)
		}
		var hdNil *HwidDataEntity = nil
		if diff := cmp.Diff(hdNil, got); diff != "" {
			t.Errorf("UpdateHwidData returned unexpected diff (-want +got):\n%s", diff)
		}
	})
}

func TestUpdateLegacyHwidData(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	datastore.GetTestable(ctx).Consistent(true)

	t.Run("update non-existent HwidData", func(t *testing.T) {
		want := mockDutLabel()
		got, err := UpdateLegacyHwidData(ctx, want, "test-hwid")
		if err != nil {
			t.Fatalf("UpdateLegacyHwidData failed: %s", err)
		}
		gotProto, err := got.GetDutLabelProto()
		if err != nil {
			t.Fatalf("GetDutLabelProto failed: %s", err)
		}
		if diff := cmp.Diff(want, gotProto, protocmp.Transform()); diff != "" {
			t.Errorf("UpdateLegacyHwidData returned unexpected diff (-want +got):\n%s", diff)
		}
	})

	t.Run("update existent HwidData", func(t *testing.T) {
		hd2Id := "test-hwid-2"
		hd2 := mockDutLabel()

		hd2update := mockDutLabel()
		hd2update.PossibleLabels = append(hd2update.PossibleLabels, "test-possible-3")

		// Insert hd2 into datastore
		_, err := UpdateLegacyHwidData(ctx, hd2, hd2Id)
		if err != nil {
			t.Fatalf("UpdateLegacyHwidData failed: %s", err)
		}

		// Update hd2
		got, err := UpdateLegacyHwidData(ctx, hd2update, hd2Id)
		if err != nil {
			t.Fatalf("UpdateLegacyHwidData failed: %s", err)
		}
		gotProto, err := got.GetDutLabelProto()
		if err != nil {
			t.Fatalf("GetDutLabelProto failed: %s", err)
		}
		if diff := cmp.Diff(hd2update, gotProto, protocmp.Transform()); diff != "" {
			t.Errorf("UpdateLegacyHwidData returned unexpected diff (-want +got):\n%s", diff)
		}
	})

	t.Run("update HwidData with empty hwid", func(t *testing.T) {
		hd3 := mockDutLabel()
		got, err := UpdateLegacyHwidData(ctx, hd3, "")
		if err == nil {
			t.Errorf("UpdateLegacyHwidData succeeded with empty hwid")
		}
		if c := status.Code(err); c != codes.InvalidArgument {
			t.Errorf("Unexpected error when calling UpdateLegacyHwidData: %s", err)
		}
		var hdNil *HwidDataEntity = nil
		if diff := cmp.Diff(hdNil, got); diff != "" {
			t.Errorf("UpdateLegacyHwidData returned unexpected diff (-want +got):\n%s", diff)
		}
	})
}

func TestGetHwidData(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	datastore.GetTestable(ctx).Consistent(true)

	t.Run("get HwidData by existing ID", func(t *testing.T) {
		id := "test-hwid"
		want := mockHwidData()
		_, err := UpdateHwidData(ctx, want, id)
		if err != nil {
			t.Fatalf("UpdateHwidData failed: %s", err)
		}

		got, err := GetHwidData(ctx, id)
		if err != nil {
			t.Fatalf("GetHwidData failed: %s", err)
		}
		gotProto, err := got.GetProto()
		if err != nil {
			t.Fatalf("GetProto failed: %s", err)
		}
		if diff := cmp.Diff(want, gotProto, protocmp.Transform()); diff != "" {
			t.Errorf("GetHwidData returned unexpected diff (-want +got):\n%s", diff)
		}
	})

	t.Run("get HwidData by non-existent ID", func(t *testing.T) {
		id := "test-hwid-2"
		_, err := GetHwidData(ctx, id)
		if err == nil {
			t.Errorf("GetHwidData succeeded with non-existent ID: %s", id)
		}
		if c := status.Code(err); c != codes.NotFound {
			t.Errorf("Unexpected error when calling GetHwidData: %s", err)
		}
	})
}

func TestParseHwidData(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	datastore.GetTestable(ctx).Consistent(true)

	id := "test-hwid"
	_, err := UpdateHwidData(ctx, mockHwidData(), id)
	if err != nil {
		t.Fatalf("UpdateHwidData failed: %s", err)
	}

	t.Run("parse nil HwidDataEntity", func(t *testing.T) {
		var want *ufspb.HwidData = nil
		got, err := ParseHwidData(nil)
		if err != nil {
			t.Fatalf("ParseHwidData failed: %s", err)
		}
		if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
			t.Errorf("ParseHwidData returned unexpected diff (-want +got):\n%s", diff)
		}
	})

	t.Run("parse hwid data from HwidDataEntity", func(t *testing.T) {
		want := mockHwidData()
		ent, err := GetHwidData(ctx, id)
		if err != nil {
			t.Fatalf("GetHwidData failed: %s", err)
		}
		got, err := ParseHwidData(ent)
		if err != nil {
			t.Fatalf("ParseHwidData failed: %s", err)
		}
		if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
			t.Errorf("ParseHwidData returned unexpected diff (-want +got):\n%s", diff)
		}
	})

	t.Run("update HwidDataEntity with DutLabel", func(t *testing.T) {
		id := "test-dutlabel-hwid"
		_, err := UpdateLegacyHwidData(ctx, mockDutLabel(), id)
		if err != nil {
			t.Fatalf("UpdateLegacyHwidData failed: %s", err)
		}

		want := &ufspb.HwidData{
			Sku:     "test-sku",
			Variant: "test-variant",
			Hwid:    "test-dutlabel-hwid",
			DutLabel: &ufspb.DutLabel{
				PossibleLabels: []string{
					"test-possible-1",
					"test-possible-2",
				},
				Labels: []*ufspb.DutLabel_Label{
					{
						Name:  "test-label-1",
						Value: "test-value-1",
					},
					{
						Name:  "Sku",
						Value: "test-sku",
					},
					{
						Name:  "variant",
						Value: "test-variant",
					},
				},
			},
		}
		ent, err := GetHwidData(ctx, id)
		if err != nil {
			t.Fatalf("GetHwidData failed: %s", err)
		}
		got, err := ParseHwidData(ent)
		if err != nil {
			t.Fatalf("ParseHwidData failed: %s", err)
		}
		if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
			t.Errorf("ParseHwidData returned unexpected diff (-want +got):\n%s", diff)
		}
	})
}
