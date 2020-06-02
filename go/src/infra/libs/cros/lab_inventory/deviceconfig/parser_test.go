// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package deviceconfig

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/chromiumos/config/go/payload"
	"go.chromium.org/chromiumos/infra/proto/go/device"
)

func TestParseConfigBundle(t *testing.T) {
	Convey("Test config bundle parsing", t, func() {
		var payload payload.ConfigBundle
		unmarshaller := &jsonpb.Unmarshaler{AllowUnknownFields: false}
		b, err := ioutil.ReadFile("dc_v2.jsonproto")
		So(err, ShouldBeNil)
		err = unmarshaller.Unmarshal(bytes.NewReader(b), &payload)
		So(err, ShouldBeNil)
		Convey("Happy path", func() {
			dcs := parseConfigBundle(payload)
			So(dcs, ShouldHaveLength, 4)
			for _, dc := range dcs {
				So(dc.GetId().GetPlatformId().GetValue(), ShouldEqual, "Puff")
				So(dc.GetId().GetModelId().GetValue(), ShouldEqual, "Puff")
				sku := dc.GetId().GetVariantId().GetValue()
				So(sku, ShouldBeIn, []string{"0", "1", "2", "2147483647"})
				So(dc.GetFormFactor(), ShouldEqual, device.Config_FORM_FACTOR_CLAMSHELL)
				So(dc.GetHardwareFeatures()[0], ShouldEqual, device.Config_HARDWARE_FEATURE_INTERNAL_DISPLAY)
			}
		})
	})
}
