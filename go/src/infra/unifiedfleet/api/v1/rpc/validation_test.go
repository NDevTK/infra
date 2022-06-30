package ufspb

import (
	"testing"

	chromeosLab "infra/unifiedfleet/api/v1/models/chromeos/lab"
)

func TestValidateHostnames(t *testing.T) {
	tt := []struct {
		in     []string
		wantOK bool
	}{
		{[]string{"h1", "h2"}, true},
		{[]string{"h1", "h1"}, false},
		{[]string{"", "h1"}, false},
		{nil, true},
	}

	for _, test := range tt {
		err := validateHostnames(test.in, "")
		if test.wantOK && err != nil {
			t.Errorf("validateHostnames(%v) failed %v", test.in, err)
		}
		if !test.wantOK && err == nil {
			t.Errorf("validateHostnames(%v) succeeded but want failure", test.in)
		}
	}
}

func TestValidateDutId(t *testing.T) {
	reqCases := [2]*UpdateDeviceRecoveryDataRequest{
		{
			ChromeosDeviceId: "deviceId-1",
			DutState: &chromeosLab.DutState{
				Id: &chromeosLab.ChromeOSDeviceID{
					Value: "deviceId-1",
				},
			},
		},
		{
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
		},
	}
	for _, req := range reqCases {
		if err := req.validateDutId(); err != nil {
			t.Errorf("validateDutId(%v) has failed %v", req, err)
		}
	}
}
