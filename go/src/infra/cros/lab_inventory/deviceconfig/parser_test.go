// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package deviceconfig

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/chromiumos/config/go/payload"
	"go.chromium.org/chromiumos/infra/proto/go/device"
)

func TestParseConfigBundle(t *testing.T) {
	Convey("Test config bundle parsing", t, func() {
		var payloads payload.ConfigBundleList
		unmarshaller := &jsonpb.Unmarshaler{AllowUnknownFields: false}
		// Refer to https://chromium.googlesource.com/chromiumos/config/+/refs/heads/master/test/project/fake/fake/config.star for unittest check
		b, err := ioutil.ReadFile("test_device_config_v2.jsonproto")
		So(err, ShouldBeNil)
		buf, err := fixFieldMaskForConfigBundleList([]byte(b))
		So(err, ShouldBeNil)
		err = unmarshaller.Unmarshal(bytes.NewBuffer(buf), &payloads)
		So(err, ShouldBeNil)
		Convey("Happy path", func() {
			So(payloads.GetValues(), ShouldHaveLength, 1)
			dcs := parseConfigBundle(payloads.GetValues()[0])
			// 6 sku-less device configs & 7 real device configs
			So(dcs, ShouldHaveLength, 13)
			for _, dc := range dcs {
				So(dc.GetId().GetPlatformId().GetValue(), ShouldEqual, "FAKE_PROGRAM")
				modelWithSku := fmt.Sprintf("%s:%s", dc.GetId().GetModelId().GetValue(), dc.GetId().GetVariantId().GetValue())
				switch modelWithSku {
				case "FAKE-REF-DESIGN:", "PROJECT-A:", "PROJECT-B:", "PROJECT-C:", "PROJECT-WL:", "PROJECT-U:":
					// These are sku-less device config, every config entry is nil by default
					So(dc.GetFormFactor(), ShouldEqual, device.Config_FORM_FACTOR_UNSPECIFIED)
					So(dc.GetPower(), ShouldEqual, device.Config_POWER_SUPPLY_UNSPECIFIED)
					So(dc.GetHardwareFeatures(), ShouldBeNil)
					So(dc.GetStorage(), ShouldEqual, device.Config_STORAGE_UNSPECIFIED)
					So(dc.GetCpu(), ShouldEqual, device.Config_ARCHITECTURE_UNDEFINED)
				case "FAKE-REF-DESIGN:2147483647":
					So(dc.GetFormFactor(), ShouldEqual, device.Config_FORM_FACTOR_CLAMSHELL)
					So(dc.GetPower(), ShouldEqual, device.Config_POWER_SUPPLY_BATTERY)
					So(dc.GetHardwareFeatures(), ShouldResemble, []device.Config_HardwareFeature{
						device.Config_HARDWARE_FEATURE_BLUETOOTH,
						device.Config_HARDWARE_FEATURE_INTERNAL_DISPLAY,
						device.Config_HARDWARE_FEATURE_STYLUS,
						device.Config_HARDWARE_FEATURE_TOUCHPAD,
						device.Config_HARDWARE_FEATURE_TOUCHSCREEN,
						device.Config_HARDWARE_FEATURE_DETACHABLE_KEYBOARD,
					})
					So(dc.GetStorage(), ShouldEqual, device.Config_STORAGE_MMC)
					So(dc.GetCpu(), ShouldEqual, device.Config_ARCHITECTURE_UNDEFINED)
				case "FAKE-REF-DESIGN:0":
					fallthrough
				case "FAKE-REF-DESIGN:2":
					So(dc.GetFormFactor(), ShouldEqual, device.Config_FORM_FACTOR_CLAMSHELL)
					So(dc.GetPower(), ShouldEqual, device.Config_POWER_SUPPLY_BATTERY)
					So(dc.GetPower(), ShouldEqual, device.Config_POWER_SUPPLY_BATTERY)
					So(dc.GetHardwareFeatures(), ShouldResemble, []device.Config_HardwareFeature{
						device.Config_HARDWARE_FEATURE_BLUETOOTH,
						device.Config_HARDWARE_FEATURE_INTERNAL_DISPLAY,
						device.Config_HARDWARE_FEATURE_STYLUS,
						device.Config_HARDWARE_FEATURE_TOUCHPAD,
						device.Config_HARDWARE_FEATURE_TOUCHSCREEN,
						device.Config_HARDWARE_FEATURE_DETACHABLE_KEYBOARD,
					})
					So(dc.GetStorage(), ShouldEqual, device.Config_STORAGE_SSD)
					So(dc.GetCpu(), ShouldEqual, device.Config_ARCHITECTURE_UNDEFINED)
					So(dc.GetEc(), ShouldEqual, device.Config_EC_CHROME)
				case "PROJECT-A:32":
					So(dc.GetFormFactor(), ShouldEqual, device.Config_FORM_FACTOR_CONVERTIBLE)
					So(dc.GetPower(), ShouldEqual, device.Config_POWER_SUPPLY_BATTERY)
					So(dc.GetHardwareFeatures(), ShouldResemble, []device.Config_HardwareFeature{
						device.Config_HARDWARE_FEATURE_BLUETOOTH,
						device.Config_HARDWARE_FEATURE_INTERNAL_DISPLAY,
						device.Config_HARDWARE_FEATURE_STYLUS,
						device.Config_HARDWARE_FEATURE_TOUCHPAD,
						device.Config_HARDWARE_FEATURE_TOUCHSCREEN,
						device.Config_HARDWARE_FEATURE_DETACHABLE_KEYBOARD,
					})
					So(dc.GetStorage(), ShouldEqual, device.Config_STORAGE_MMC)
					So(dc.GetCpu(), ShouldEqual, device.Config_ARCHITECTURE_UNDEFINED)
				case "PROJECT-B:33":
					So(dc.GetFormFactor(), ShouldEqual, device.Config_FORM_FACTOR_CONVERTIBLE)
					So(dc.GetPower(), ShouldEqual, device.Config_POWER_SUPPLY_BATTERY)
					So(dc.GetHardwareFeatures(), ShouldResemble, []device.Config_HardwareFeature{
						device.Config_HARDWARE_FEATURE_BLUETOOTH,
						device.Config_HARDWARE_FEATURE_INTERNAL_DISPLAY,
						device.Config_HARDWARE_FEATURE_STYLUS,
						device.Config_HARDWARE_FEATURE_TOUCHPAD,
						device.Config_HARDWARE_FEATURE_TOUCHSCREEN,
						device.Config_HARDWARE_FEATURE_DETACHABLE_KEYBOARD,
					})
					So(dc.GetStorage(), ShouldEqual, device.Config_STORAGE_MMC)
					So(dc.GetCpu(), ShouldEqual, device.Config_ARCHITECTURE_UNDEFINED)
				case "PROJECT-C:34":
					So(dc.GetFormFactor(), ShouldEqual, device.Config_FORM_FACTOR_CLAMSHELL)
					So(dc.GetPower(), ShouldEqual, device.Config_POWER_SUPPLY_BATTERY)
					So(dc.GetHardwareFeatures(), ShouldResemble, []device.Config_HardwareFeature{
						device.Config_HARDWARE_FEATURE_BLUETOOTH,
						device.Config_HARDWARE_FEATURE_INTERNAL_DISPLAY,
						device.Config_HARDWARE_FEATURE_STYLUS,
						device.Config_HARDWARE_FEATURE_TOUCHPAD,
						device.Config_HARDWARE_FEATURE_TOUCHSCREEN,
						device.Config_HARDWARE_FEATURE_DETACHABLE_KEYBOARD,
					})
					So(dc.GetStorage(), ShouldEqual, device.Config_STORAGE_NVME)
					So(dc.GetCpu(), ShouldEqual, device.Config_ARCHITECTURE_UNDEFINED)
				case "PROJECT-WL:64":
					So(dc.GetFormFactor(), ShouldEqual, device.Config_FORM_FACTOR_CHROMEBIT)
					So(dc.GetPower(), ShouldEqual, device.Config_POWER_SUPPLY_AC_ONLY)
					So(dc.GetHardwareFeatures(), ShouldResemble, []device.Config_HardwareFeature{
						device.Config_HARDWARE_FEATURE_BLUETOOTH,
						device.Config_HARDWARE_FEATURE_INTERNAL_DISPLAY,
						device.Config_HARDWARE_FEATURE_TOUCHPAD,
						device.Config_HARDWARE_FEATURE_TOUCHSCREEN,
						device.Config_HARDWARE_FEATURE_DETACHABLE_KEYBOARD,
					})
					So(dc.GetStorage(), ShouldEqual, device.Config_STORAGE_NVME)
					So(dc.GetCpu(), ShouldEqual, device.Config_ARCHITECTURE_UNDEFINED)
				case "PROJECT-U:75":
					So(dc.GetFormFactor(), ShouldEqual, device.Config_FORM_FACTOR_CONVERTIBLE)
					So(dc.GetPower(), ShouldEqual, device.Config_POWER_SUPPLY_BATTERY)
					So(dc.GetHardwareFeatures(), ShouldResemble, []device.Config_HardwareFeature{
						device.Config_HARDWARE_FEATURE_BLUETOOTH,
						device.Config_HARDWARE_FEATURE_INTERNAL_DISPLAY,
						device.Config_HARDWARE_FEATURE_STYLUS,
						device.Config_HARDWARE_FEATURE_TOUCHPAD,
						device.Config_HARDWARE_FEATURE_TOUCHSCREEN,
						device.Config_HARDWARE_FEATURE_DETACHABLE_KEYBOARD,
					})
					So(dc.GetStorage(), ShouldEqual, device.Config_STORAGE_UFS)
					So(dc.GetCpu(), ShouldEqual, device.Config_ARCHITECTURE_UNDEFINED)
				default:
					t.Errorf("Invalid model:sku: %s", modelWithSku)
				}
			}
		})
	})
}
