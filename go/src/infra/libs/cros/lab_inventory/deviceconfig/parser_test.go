// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package deviceconfig

import (
	"testing"

	"github.com/golang/protobuf/proto"
	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/chromiumos/config/go/payload"
	"go.chromium.org/chromiumos/infra/proto/go/device"
)

var testConfig = `
partners: <>
components: <>
programs: <>
designs: <
  value: <
    id: <
      value: "Galaxy"
    >
    program_id: <
      value: "Galaxy"
    >
    name: "Galaxy"
    configs: <
      id: <
        value: "Galaxy:1"
      >
      hardware_topology: <
        form_factor: <
          id: "CLAMSHELL"
          type: FORM_FACTOR
          description: <
            key: "EN"
            value: "Device can only rotate 200 degrees"
          >
          hardware_feature: <
            form_factor: <
              form_factor: CLAMSHELL
            >
          >
        >
      >
      hardware_features: <
        fw_config: <>
        form_factor: <
          form_factor: CLAMSHELL
        >
      >
    >
  >
>
device_brands: <>
software_configs: <
  design_config_id: <
    value: "Galaxy:1"
  >
  id_scan_config: <
    smbios_name_match: "Galaxy"
    firmware_sku: 1
  >
>
`

func TestParseConfigBundle(t *testing.T) {
	Convey("Test config bundle parsing", t, func() {
		var payload payload.ConfigBundle
		err := proto.UnmarshalText(testConfig, &payload)
		So(err, ShouldBeNil)
		Convey("Happy path", func() {
			dcs := parseConfigBundle(payload)
			So(dcs, ShouldHaveLength, 1)
			So(dcs[0].GetId().GetPlatformId().GetValue(), ShouldEqual, "Galaxy")
			So(dcs[0].GetId().GetModelId().GetValue(), ShouldEqual, "Galaxy")
			So(dcs[0].GetId().GetVariantId().GetValue(), ShouldEqual, "1")
			So(dcs[0].GetFormFactor(), ShouldEqual, device.Config_FORM_FACTOR_CLAMSHELL)
		})
	})
}
