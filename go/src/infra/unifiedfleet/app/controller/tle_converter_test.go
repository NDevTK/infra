// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"bytes"
	"os"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/google/go-cmp/cmp"
	"go.chromium.org/chromiumos/config/go/test/api"

	"infra/libs/fleet/boxster/swarming"
	ufspb "infra/unifiedfleet/api/v1/models"
	chromeosLab "infra/unifiedfleet/api/v1/models/chromeos/lab"
	"infra/unifiedfleet/app/external"
)

func parseDutAttribute(t *testing.T, protoText string) api.DutAttribute {
	var da api.DutAttribute
	if err := jsonpb.UnmarshalString(protoText, &da); err != nil {
		t.Fatalf("Error unmarshalling example text: %s", err)
	}
	return da
}

func mockMachineLSEWithLabConfigs(name string) *ufspb.MachineLSE {
	return &ufspb.MachineLSE{
		Name:     name,
		Hostname: name,
		Lse: &ufspb.MachineLSE_ChromeosMachineLse{
			ChromeosMachineLse: &ufspb.ChromeOSMachineLSE{
				ChromeosLse: &ufspb.ChromeOSMachineLSE_DeviceLse{
					DeviceLse: &ufspb.ChromeOSDeviceLSE{
						Device: &ufspb.ChromeOSDeviceLSE_Dut{
							Dut: &chromeosLab.DeviceUnderTest{
								Hostname: name,
								Peripherals: &chromeosLab.Peripherals{
									Audio: &chromeosLab.Audio{
										AudioBox: true,
									},
									Carrier: "test-carrier",
								},
								Licenses: []*chromeosLab.License{
									{
										Type:       chromeosLab.LicenseType_LICENSE_TYPE_WINDOWS_10_PRO,
										Identifier: "test-license",
									},
									{
										Type:       chromeosLab.LicenseType_LICENSE_TYPE_MS_OFFICE_STANDARD,
										Identifier: "test-license-2",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func TestConvert(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	ctx = external.WithTestingContext(ctx)
	ctx = useTestingCfg(ctx)

	t.Run("convert lab config label - happy path; single boolean value", func(t *testing.T) {
		dutMachinelse := mockMachineLSEWithLabConfigs("lse-1")
		dutState := mockDutState("dutstate-id-1", "dutstate-hostname-1")
		daText := `{
			"id": {
				"value": "peripheral-audio-box"
			},
			"aliases": [
				"label-audio_box"
			],
			"tleSource": {}
		}`
		da := parseDutAttribute(t, daText)
		want := swarming.Dimensions{
			"label-audio_box":      {"true"},
			"peripheral-audio-box": {"true"},
		}
		got, err := Convert(ctx, &da, nil, dutMachinelse, dutState)
		if err != nil {
			t.Fatalf("Convert failed: %s", err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Convert returned unexpected diff (-want +got):\n%s", diff)
		}
	})

	t.Run("convert lab config label - happy path; array of values", func(t *testing.T) {
		dutMachinelse := mockMachineLSEWithLabConfigs("lse-1")
		dutState := mockDutState("dutstate-id-1", "dutstate-hostname-1")
		daText := `{
			"id": {
				"value": "misc-license"
			},
			"aliases": [
				"label-license"
			],
			"tleSource": {}
		}`
		da := parseDutAttribute(t, daText)
		want := swarming.Dimensions{
			"label-license": {"LICENSE_TYPE_WINDOWS_10_PRO", "LICENSE_TYPE_MS_OFFICE_STANDARD"},
			"misc-license":  {"LICENSE_TYPE_WINDOWS_10_PRO", "LICENSE_TYPE_MS_OFFICE_STANDARD"},
		}
		got, err := Convert(ctx, &da, nil, dutMachinelse, dutState)
		if err != nil {
			t.Fatalf("Convert failed: %s", err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Convert returned unexpected diff (-want +got):\n%s", diff)
		}
	})

	t.Run("convert dut state label - happy path; single value", func(t *testing.T) {
		dutState := mockDutState("dutstate-id-1", "dutstate-hostname-1")
		dutState.WorkingBluetoothBtpeer = 10
		daText := `{
			"id": {
				"value": "peripheral-num-btpeer"
			},
			"aliases": [
				"label-working_bluetooth_btpeer"
			],
			"tleSource": {}
		}`
		da := parseDutAttribute(t, daText)
		want := swarming.Dimensions{
			"label-working_bluetooth_btpeer": {"10"},
			"peripheral-num-btpeer":          {"10"},
		}
		got, err := Convert(ctx, &da, nil, nil, dutState)
		if err != nil {
			t.Fatalf("Convert failed: %s", err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Convert returned unexpected diff (-want +got):\n%s", diff)
		}
	})
}

func TestStandardConverter(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	ctx = external.WithTestingContext(ctx)
	ctx = useTestingCfg(ctx)

	t.Run("truncate label prefix - happy path", func(t *testing.T) {
		dutMachinelse := mockMachineLSEWithLabConfigs("lse-1")
		dutState := mockDutState("dutstate-id-1", "dutstate-hostname-1")
		dutState.ServoUsbState = chromeosLab.HardwareState_HARDWARE_NORMAL
		daText := `{
			"id": {
				"value": "peripheral-servo-usb-state"
			},
			"aliases": [
				"label-servo_usb_state"
			],
			"tleSource": {}
		}`
		da := parseDutAttribute(t, daText)
		want := swarming.Dimensions{
			"label-servo_usb_state":      {"NORMAL"},
			"peripheral-servo-usb-state": {"NORMAL"},
		}
		got, err := Convert(ctx, &da, nil, dutMachinelse, dutState)
		if err != nil {
			t.Fatalf("Convert failed: %s", err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Convert returned unexpected diff (-want +got):\n%s", diff)
		}
	})

	t.Run("append prefix to label - happy path", func(t *testing.T) {
		dutMachinelse := mockMachineLSEWithLabConfigs("lse-1")
		dutState := mockDutState("dutstate-id-1", "dutstate-hostname-1")

		daText := `{
			"id": {
				"value": "peripheral-carrier"
			},
			"aliases": [
				"label-carrier"
			],
			"tleSource": {}
		}`
		da := parseDutAttribute(t, daText)
		want := swarming.Dimensions{
			"label-carrier":      {"CARRIER_test-carrier"},
			"peripheral-carrier": {"CARRIER_test-carrier"},
		}
		got, err := Convert(ctx, &da, nil, dutMachinelse, dutState)
		if err != nil {
			t.Fatalf("Convert failed: %s", err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Convert returned unexpected diff (-want +got):\n%s", diff)
		}
	})
}

func TestExistenceConverter(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	ctx = external.WithTestingContext(ctx)
	ctx = useTestingCfg(ctx)

	t.Run("existence check - no servo exists in MachineLSE", func(t *testing.T) {
		dutLse := mockMachineLSEWithLabConfigs("lse-1")
		dutState := mockDutState("dutstate-no-servo", "dutstate-hostname-no-servo")
		dutState.Servo = chromeosLab.PeripheralState_NOT_CONNECTED
		daText := `{
      "id": {
        "value": "peripheral-servo"
      },
      "aliases": [
        "label-servo"
      ],
      "tleSource": {}
    }`
		da := parseDutAttribute(t, daText)
		want := swarming.Dimensions{
			"label-servo":      {"false"},
			"peripheral-servo": {"false"},
		}
		got, err := Convert(ctx, &da, nil, dutLse, dutState)
		if err != nil {
			t.Fatalf("Convert failed: %s", err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Convert returned unexpected diff (-want +got):\n%s", diff)
		}
	})

	t.Run("existence check - servo exists in MachineLSE", func(t *testing.T) {
		dutLse := mockMachineLSEWithLabConfigs("lse-1")
		dutState := mockDutState("dutstate-with-servo", "dutstate-hostname-with-servo")
		dutState.Servo = chromeosLab.PeripheralState_WORKING
		daText := `{
      "id": {
        "value": "peripheral-servo"
      },
      "aliases": [
        "label-servo"
      ],
      "tleSource": {}
    }`
		da := parseDutAttribute(t, daText)
		want := swarming.Dimensions{
			"label-servo":      {"true"},
			"peripheral-servo": {"true"},
		}
		got, err := Convert(ctx, &da, nil, dutLse, dutState)
		if err != nil {
			t.Fatalf("Convert failed: %s", err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Convert returned unexpected diff (-want +got):\n%s", diff)
		}
	})
}

// Basic test to test integrity and parseability of tle_sources.jsonproto
func TestTleSourcesJsonproto(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	ctx = external.WithTestingContext(ctx)
	ctx = useTestingCfg(ctx)

	t.Run("read and parse file into proto", func(t *testing.T) {
		mapFile, err := os.ReadFile("tle_sources.jsonproto")
		if err != nil {
			t.Fatalf("Failed to read tle_sources.jsonproto: %s", err)
		}

		var tleMappings ufspb.TleSources
		err = jsonpb.Unmarshal(bytes.NewBuffer(mapFile), &tleMappings)
		if err != nil {
			t.Fatalf("Failed to unmarshal file into TleSources proto: %s", err)
		}
	})
}
