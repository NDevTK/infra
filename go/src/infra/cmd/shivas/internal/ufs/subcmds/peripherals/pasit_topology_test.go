// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package peripherals

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	lab "infra/unifiedfleet/api/v1/models/chromeos/lab"
)

var exampleTopology = &lab.PasitHost{
	Id: "pasit-host1",
	Devices: []*lab.PasitHost_Device{
		{
			Id:   "1912901",
			Type: lab.PasitHost_Device_SWITCH_FIXTURE,
		},
		{
			Id:   "2001901",
			Type: lab.PasitHost_Device_SWITCH_FIXTURE,
		},
		{
			Id:   "2007902",
			Type: lab.PasitHost_Device_SWITCH_FIXTURE,
		},
		{
			Id:   "J45SW01",
			Type: lab.PasitHost_Device_SWITCH_FIXTURE,
		},
		{
			Id:    "dock_1",
			Model: "DOCK_XXYY",
			Type:  lab.PasitHost_Device_DOCKING_STATION,
		},
		{
			Id:    "monitor_1",
			Model: "MONITOR_XXYY",
			Type:  lab.PasitHost_Device_MONITOR,
		},
		{
			Id:   "camera_1",
			Type: lab.PasitHost_Device_CAMERA,
		},
		{
			Id:   "network_1",
			Type: lab.PasitHost_Device_NETWORK,
		},
	},
	Connections: []*lab.PasitHost_Connection{
		{
			Technology: lab.PasitHost_Connection_USBC,
			Parent:     "chromeosX-rackX-rowY-hostN",
			Child:      "1912901",
		},
		{
			Technology: lab.PasitHost_Connection_USBC,
			Parent:     "1912901",
			Child:      "dock_1",
		},
		{
			Technology: lab.PasitHost_Connection_HDMI,
			Parent:     "2007902",
			Child:      "monitor_1",
		},
		{
			Technology: lab.PasitHost_Connection_ETHERNET,
			Parent:     "J45SW01",
			Child:      "network_1",
		},
		{
			Technology: lab.PasitHost_Connection_USBA2,
			Parent:     "dock_1",
			Child:      "2001901",
		},
		{
			Technology: lab.PasitHost_Connection_HDMI,
			Parent:     "dock_1",
			Child:      "2007902",
		},
		{
			Technology: lab.PasitHost_Connection_ETHERNET,
			Parent:     "dock_1",
			Child:      "J45SW01",
		},
	},
}

func TestPasitCleanAndValidateFlags(t *testing.T) {
	// Test invalid flags
	errTests := []struct {
		want []string
		cmd  *managePasitTopologyCmd
	}{
		{
			want: []string{errDUTMissing},
			cmd:  &managePasitTopologyCmd{},
		},
		{
			want: []string{errFileMissing},
			cmd:  &managePasitTopologyCmd{dutName: "dut"},
		},
		{
			want: []string{errIDMissing},
			cmd: &managePasitTopologyCmd{
				dutName: "dut",
				topologyObj: &lab.PasitHost{
					Devices: []*lab.PasitHost_Device{
						{
							Id: "id",
						},
						{
							Id: "",
						},
					},
				},
			},
		},
		{
			want: []string{errDuplicateID},
			cmd: &managePasitTopologyCmd{
				dutName: "dut",
				topologyObj: &lab.PasitHost{
					Devices: []*lab.PasitHost_Device{
						{
							Id: "id1",
						},
						{
							Id: "id2",
						},
						{
							Id: "id1",
						},
					},
				},
			},
		},
		{
			want: []string{errMissingComponent},
			cmd: &managePasitTopologyCmd{
				dutName: "dut",
				topologyObj: &lab.PasitHost{
					Devices: []*lab.PasitHost_Device{
						{
							Id: "id1",
						},
					},
					Connections: []*lab.PasitHost_Connection{
						{
							Parent: "id1",
							Child:  "id2",
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
	c := &managePasitTopologyCmd{
		dutName:     "chromeosX-rackX-rowY-hostN",
		mode:        actionAdd,
		topologyObj: exampleTopology,
	}
	if err := c.cleanAndValidateFlags(); err != nil {
		t.Errorf("cleanAndValidateFlags = %v; want nil", err)
	}
	b, _ := json.MarshalIndent(c.topologyObj, "", "  ")
	fmt.Println(string(b))
}
