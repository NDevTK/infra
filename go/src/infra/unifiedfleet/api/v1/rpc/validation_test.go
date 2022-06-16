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
	dutId := "deviceId-1"
	req := &UpdateDeviceRecoveryDataRequest{
		DutState: &chromeosLab.DutState{
			Id: &chromeosLab.ChromeOSDeviceID{
				Value: dutId,
			},
		},
	}
	req.DeviceId = dutId
	if err := req.validateDutId(); err != nil {
		t.Errorf("validateDutId(%v) with DeviceId failed %v", req, err)
	}
	req.DeviceId = ""
	req.ChromeosDeviceId = dutId
	if err := req.validateDutId(); err != nil {
		t.Errorf("validateDutId(%v) with ChromeosDeviceId failed %v", req, err)
	}
}
