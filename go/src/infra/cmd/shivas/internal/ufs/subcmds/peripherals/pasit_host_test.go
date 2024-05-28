// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package peripherals

import (
	"strings"
	"testing"

	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
)

var exampleHost = &labapi.PasitHost{
	Hostname: "pasit-host1",
	Devices: []*labapi.PasitHost_Device{
		{
			Id:   "1912901",
			Type: labapi.PasitHost_Device_SWITCH_FIXTURE,
		},
		{
			Id:   "2001901",
			Type: labapi.PasitHost_Device_SWITCH_FIXTURE,
		},
		{
			Id:   "2007902",
			Type: labapi.PasitHost_Device_SWITCH_FIXTURE,
		},
		{
			Id:   "J45SW01",
			Type: labapi.PasitHost_Device_SWITCH_FIXTURE,
		},
		{
			Id:    "dock_1",
			Model: "DOCK_XXYY",
			Type:  labapi.PasitHost_Device_DOCKING_STATION,
		},
		{
			Id:    "monitor_1",
			Model: "MONITOR_XXYY",
			Type:  labapi.PasitHost_Device_MONITOR,
		},
		{
			Id:   "camera_1",
			Type: labapi.PasitHost_Device_CAMERA,
		},
		{
			Id:   "network_1",
			Type: labapi.PasitHost_Device_NETWORK,
		},
		{
			Id:   "chromeosX-rackX-rowY-hostN",
			Type: labapi.PasitHost_Device_DUT,
		},
	},
	Connections: []*labapi.PasitHost_Connection{
		{
			Type:     "USBC",
			ParentId: "chromeosX-rackX-rowY-hostN",
			ChildId:  "1912901",
		},
		{
			Type:     "USBC",
			ParentId: "1912901",
			ChildId:  "dock_1",
		},
		{
			Type:     "HDMI",
			ParentId: "2007902",
			ChildId:  "monitor_1",
		},
		{
			Type:     "ETHERNET",
			ParentId: "J45SW01",
			ChildId:  "network_1",
		},
		{
			Type:     "USBA",
			ParentId: "dock_1",
			ChildId:  "2001901",
		},
		{
			Type:     "HDMI",
			ParentId: "dock_1",
			ChildId:  "2007902",
		},
		{
			Type:     "ETHERNET",
			ParentId: "dock_1",
			ChildId:  "J45SW01",
		},
	},
}

func TestPasitCleanAndValidateFlags(t *testing.T) {
	// Test invalid flags
	errTests := []struct {
		want []string
		cmd  *managePasitHostCmd
	}{
		{
			want: []string{errDUTMissing},
			cmd:  &managePasitHostCmd{},
		},
		{
			want: []string{errFileMissing},
			cmd:  &managePasitHostCmd{dutName: "dut"},
		},
		{
			want: []string{errDUTNotInHost},
			cmd:  &managePasitHostCmd{dutName: "dut", hostObj: &labapi.PasitHost{}},
		},
		{
			want: []string{errIDMissing},
			cmd: &managePasitHostCmd{
				dutName: "dut",
				hostObj: &labapi.PasitHost{
					Devices: []*labapi.PasitHost_Device{
						{
							Id:   "dut",
							Type: labapi.PasitHost_Device_DUT,
						},
						{
							Id: "",
						},
						{
							Id: "id2",
						},
					},
					Connections: []*labapi.PasitHost_Connection{
						{
							ParentId: "dut",
							ChildId:  "id2",
						},
					},
				},
			},
		},
		{
			want: []string{errDuplicateID},
			cmd: &managePasitHostCmd{
				dutName: "dut",
				hostObj: &labapi.PasitHost{
					Devices: []*labapi.PasitHost_Device{
						{
							Id: "id1",
						},
						{
							Id:   "dut",
							Type: labapi.PasitHost_Device_DUT,
						},
						{
							Id: "id1",
						},
					},
					Connections: []*labapi.PasitHost_Connection{
						{
							ParentId: "dut",
							ChildId:  "id2",
						},
					},
				},
			},
		},
		{
			want: []string{errMissingDevice},
			cmd: &managePasitHostCmd{
				dutName: "dut",
				hostObj: &labapi.PasitHost{
					Devices: []*labapi.PasitHost_Device{
						{
							Id:   "dut",
							Type: labapi.PasitHost_Device_DUT,
						},
						{
							Id: "id1",
						},
					},
					Connections: []*labapi.PasitHost_Connection{
						{
							ParentId: "id1",
							ChildId:  "id2",
						},
					},
				},
			},
		},
		{
			want: []string{errMissingChild},
			cmd: &managePasitHostCmd{
				dutName: "dut",
				hostObj: &labapi.PasitHost{
					Devices: []*labapi.PasitHost_Device{
						{
							Id:   "dut",
							Type: labapi.PasitHost_Device_DUT,
						},
						{
							Id: "id1",
						},
					},
					Connections: []*labapi.PasitHost_Connection{
						{
							ParentId: "id1",
							ChildId:  "",
						},
					},
				},
			},
		},
		{
			want: []string{errMissingParent},
			cmd: &managePasitHostCmd{
				dutName: "dut",
				hostObj: &labapi.PasitHost{
					Devices: []*labapi.PasitHost_Device{
						{
							Id:   "dut",
							Type: labapi.PasitHost_Device_DUT,
						},
						{
							Id: "id1",
						},
					},
					Connections: []*labapi.PasitHost_Connection{
						{
							ParentId: "",
							ChildId:  "id2",
						},
					},
				},
			},
		},
		{
			want: []string{errChildEqualsParent},
			cmd: &managePasitHostCmd{
				dutName: "dut",
				hostObj: &labapi.PasitHost{
					Devices: []*labapi.PasitHost_Device{
						{
							Id:   "dut",
							Type: labapi.PasitHost_Device_DUT,
						},
						{
							Id: "id1",
						},
					},
					Connections: []*labapi.PasitHost_Connection{
						{
							ParentId: "id1",
							ChildId:  "id1",
						},
					},
				},
			},
		},
		{
			want: []string{errNoDevices},
			cmd: &managePasitHostCmd{
				dutName: "dut",
				hostObj: &labapi.PasitHost{
					Devices: []*labapi.PasitHost_Device{
						{
							Id:   "dut",
							Type: labapi.PasitHost_Device_DUT,
						},
					},
				},
			},
		},
		{
			want: []string{errNoConnections},
			cmd: &managePasitHostCmd{
				dutName: "dut",
				hostObj: &labapi.PasitHost{
					Devices: []*labapi.PasitHost_Device{
						{
							Id:   "dut",
							Type: labapi.PasitHost_Device_DUT,
						},
						{
							Id: "id1",
						},
					},
				},
			},
		},
	}

	for _, tt := range errTests {
		err := tt.cmd.cleanAndValidateFlags()
		if err == nil {
			t.Errorf("cleanAndValidateFlags = nil; want errors: %v", tt.want)
			continue
		}
		for _, errStr := range tt.want {
			if !strings.Contains(err.Error(), errStr) {
				t.Errorf("cleanAndValidateFlags = %q; want err %q included", err, errStr)
			}
		}
	}

	// Test valid flags with hostname cleanup
	c := &managePasitHostCmd{
		dutName: "chromeosX-rackX-rowY-hostN",
		mode:    actionAdd,
		hostObj: exampleHost,
	}
	if err := c.cleanAndValidateFlags(); err != nil {
		t.Errorf("cleanAndValidateFlags = %v; want nil", err)
	}
}
