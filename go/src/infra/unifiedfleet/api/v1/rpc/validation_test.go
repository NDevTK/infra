package ufspb

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	chromeosLab "infra/unifiedfleet/api/v1/models/chromeos/lab"
)

func TestValidateHostnames(t *testing.T) {
	Convey("ValidateHostnames", t, func() {
		Convey("Different hostnames - successful path", func() {
			const h1, h2 = "h1", "h2"
			err := validateHostnames([]string{h1, h2}, "")
			So(err, ShouldBeNil)
		})
		Convey("Duplicated hostnames - returns error", func() {
			const h1, h2 = "h1", "h1"
			err := validateHostnames([]string{h1, h2}, "")
			So(err, ShouldNotBeNil)
		})
		Convey("Empty hostname - returns error", func() {
			const h1, h2 = "", "h1"
			err := validateHostnames([]string{h1, h2}, "")
			So(err, ShouldNotBeNil)
		})
		Convey("Nil input - successful path", func() {
			err := validateHostnames(nil, "")
			So(err, ShouldBeNil)
		})
	})
}

func TestValidateDutId(t *testing.T) {
	Convey("ValidateDutId", t, func() {
		Convey("ChromeOS device - successful path", func() {
			req := &UpdateDeviceRecoveryDataRequest{
				DeviceId:     "deviceId-1",
				ResourceType: UpdateDeviceRecoveryDataRequest_RESOURCE_TYPE_CHROMEOS_DEVICE,
				DeviceRecoveryData: &UpdateDeviceRecoveryDataRequest_Chromeos{
					Chromeos: &ChromeOsRecoveryData{
						DutState: &chromeosLab.DutState{
							Id: &chromeosLab.ChromeOSDeviceID{
								Value: "deviceId-1",
							},
						},
					},
				},
			}
			err := req.validateDutId()
			So(err, ShouldBeNil)
		})
		Convey("ChromeOS device - empty device Id - returns error", func() {
			req := &UpdateDeviceRecoveryDataRequest{
				DeviceId:     "",
				ResourceType: UpdateDeviceRecoveryDataRequest_RESOURCE_TYPE_CHROMEOS_DEVICE,
				DeviceRecoveryData: &UpdateDeviceRecoveryDataRequest_Chromeos{
					Chromeos: &ChromeOsRecoveryData{
						DutState: &chromeosLab.DutState{
							Id: &chromeosLab.ChromeOSDeviceID{
								Value: "deviceId-1",
							},
						},
					},
				},
			}
			err := req.validateDutId()
			So(err, ShouldNotBeNil)
		})
		Convey("ChromeOS device - invalid device Id - returns error", func() {
			req := &UpdateDeviceRecoveryDataRequest{
				DeviceId:     "",
				ResourceType: UpdateDeviceRecoveryDataRequest_RESOURCE_TYPE_CHROMEOS_DEVICE,
				DeviceRecoveryData: &UpdateDeviceRecoveryDataRequest_Chromeos{
					Chromeos: &ChromeOsRecoveryData{
						DutState: &chromeosLab.DutState{
							Id: &chromeosLab.ChromeOSDeviceID{
								Value: "deviceId-foo",
							},
						},
					},
				},
			}
			err := req.validateDutId()
			So(err, ShouldNotBeNil)
		})
		Convey("ChromeOS device - missing dut state - returns error", func() {
			req := &UpdateDeviceRecoveryDataRequest{
				DeviceId:     "",
				ResourceType: UpdateDeviceRecoveryDataRequest_RESOURCE_TYPE_CHROMEOS_DEVICE,
				DeviceRecoveryData: &UpdateDeviceRecoveryDataRequest_Chromeos{
					Chromeos: &ChromeOsRecoveryData{},
				},
			}
			err := req.validateDutId()
			So(err, ShouldNotBeNil)
		})
		Convey("ChromeOS device - mismatching device and dut stats Ids - returns error", func() {
			req := &UpdateDeviceRecoveryDataRequest{
				DeviceId:     "deviceId-1",
				ResourceType: UpdateDeviceRecoveryDataRequest_RESOURCE_TYPE_CHROMEOS_DEVICE,
				DeviceRecoveryData: &UpdateDeviceRecoveryDataRequest_Chromeos{
					Chromeos: &ChromeOsRecoveryData{
						DutState: &chromeosLab.DutState{
							Id: &chromeosLab.ChromeOSDeviceID{
								Value: "deviceId-2",
							},
						},
					},
				},
			}
			err := req.validateDutId()
			So(err, ShouldNotBeNil)
		})
		Convey("Attached device - successful path", func() {
			req := &UpdateDeviceRecoveryDataRequest{
				DeviceId:     "deviceId-1",
				ResourceType: UpdateDeviceRecoveryDataRequest_RESOURCE_TYPE_ATTACHED_DEVICE,
			}
			err := req.validateDutId()
			So(err, ShouldBeNil)
		})
		Convey("Attached device - invalid device Id - returns error", func() {
			req := &UpdateDeviceRecoveryDataRequest{
				DeviceId:     "deviceId-foo",
				ResourceType: UpdateDeviceRecoveryDataRequest_RESOURCE_TYPE_ATTACHED_DEVICE,
			}
			err := req.validateDutId()
			So(err, ShouldBeNil)
		})
	})
}
