// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dumper

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/appengine/gaetesting"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/gae/service/datastore"
	"google.golang.org/protobuf/proto"

	ufspb "infra/unifiedfleet/api/v1/models"
	apibq "infra/unifiedfleet/api/v1/models/bigquery"
	"infra/unifiedfleet/app/model/configuration"
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

// Tests for the getAllHwidData method.
func TestGetAllHwidData(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("dev~infra-unified-fleet-system")
	ctx = gologger.StdConfig.Use(ctx)
	ctx = logging.SetLevel(ctx, logging.Error)
	datastore.GetTestable(ctx).Consistent(true)

	bqMsgs := make([]proto.Message, 0, 4)
	for i := 0; i < 4; i++ {
		hdId := fmt.Sprintf("test-hwid-%d", i)
		hd := mockHwidData()
		resp, err := configuration.UpdateHwidData(ctx, hd, hdId)
		if err != nil {
			t.Fatalf("UpdateHwidData failed: %s", err)
		}
		respProto, err := resp.GetProto()
		if err != nil {
			t.Fatalf("GetProto failed: %s", err)
		}
		bqMsgs = append(bqMsgs, &apibq.HwidDataRow{
			HwidData: respProto.(*ufspb.HwidData),
		})
	}

	Convey("getAllHwidData", t, func() {
		Convey("happy path", func() {
			resp, err := getAllHwidData(ctx)
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, bqMsgs)
		})
	})
}
