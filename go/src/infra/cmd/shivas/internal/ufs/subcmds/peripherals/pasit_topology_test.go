// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package peripherals

import (
	"strings"
	"testing"

	lab "infra/unifiedfleet/api/v1/models/chromeos/lab"
)

var exampleTopology = &lab.PasitTopology{
	Hosts: []*lab.PasitTopology_Host{
		{
			Id: "chromeosX-rackX-rowY-hostN",
			Ports: []*lab.PasitTopology_Port{
				{
					Technology:         lab.PasitTopology_PERHIPHERAL_TECHNOLOGY_USBC,
					ConnectedComponent: "1912901",
				},
			},
		},
	},
	Switches: []*lab.PasitTopology_Switch{
		{
			Id:         "1912901",
			Technology: lab.PasitTopology_PERHIPHERAL_TECHNOLOGY_USBC,
			Ports: []*lab.PasitTopology_Port{
				{
					Technology:         lab.PasitTopology_PERHIPHERAL_TECHNOLOGY_USBC,
					ConnectedComponent: "dock_1",
				},
			},
		},
		{
			Id:         "2001901",
			Technology: lab.PasitTopology_PERHIPHERAL_TECHNOLOGY_USBA2,
			Ports:      []*lab.PasitTopology_Port{},
		},
		{
			Id:         "2007902",
			Technology: lab.PasitTopology_PERHIPHERAL_TECHNOLOGY_HDMI,
			Ports: []*lab.PasitTopology_Port{
				{
					Technology:         lab.PasitTopology_PERHIPHERAL_TECHNOLOGY_HDMI,
					ConnectedComponent: "monitor_1",
				},
			},
		},
		{
			Id:         "J45SW01",
			Technology: lab.PasitTopology_PERHIPHERAL_TECHNOLOGY_ETHERNET,
			Ports: []*lab.PasitTopology_Port{
				{
					Technology:         lab.PasitTopology_PERHIPHERAL_TECHNOLOGY_ETHERNET,
					ConnectedComponent: "network_1",
				},
			},
		},
	},
	Docks: []*lab.PasitTopology_Dock{
		{
			Id:    "dock_1",
			Model: "DOCK_XXYY",
			Ports: []*lab.PasitTopology_Port{
				{
					ConnectedComponent: "2001901",
					Technology:         lab.PasitTopology_PERHIPHERAL_TECHNOLOGY_USBA2,
				},
				{
					ConnectedComponent: "2007902",
					Technology:         lab.PasitTopology_PERHIPHERAL_TECHNOLOGY_HDMI,
				},
				{
					ConnectedComponent: "J45SW01",
					Technology:         lab.PasitTopology_PERHIPHERAL_TECHNOLOGY_ETHERNET,
				},
			},
		},
	},
	Monitors: []*lab.PasitTopology_Monitor{
		{
			Id:    "monitor_1",
			Model: "MONITOR_XXYY",
		},
	},
	Cameras: []*lab.PasitTopology_Camera{
		{
			Id:    "camera_1",
			Model: "CAMERA_XXYY",
		},
	},
	Networks: []*lab.PasitTopology_Network{
		{
			Id: "network_1",
		},
	},
}

func TestPasitCleanAndValidateFlags(t *testing.T) {
	// Test invalid flags
	errTests := []struct {
		cmd  *managePasitTopologyCmd
		want []string
	}{
		{
			cmd:  &managePasitTopologyCmd{},
			want: []string{errDUTMissing},
		},
		{
			cmd:  &managePasitTopologyCmd{dutName: "dut"},
			want: []string{errFileMissing},
		},
		{
			cmd: &managePasitTopologyCmd{
				dutName: "dut",
				topologyObj: &lab.PasitTopology{
					Hosts: []*lab.PasitTopology_Host{
						{
							Id: "dut",
						},
						{
							Id: "",
						},
					},
				},
			},

			want: []string{errIDMissing},
		},
		{
			want: []string{errDuplicateID},
			cmd: &managePasitTopologyCmd{
				dutName: "dut",
				topologyObj: &lab.PasitTopology{
					Hosts: []*lab.PasitTopology_Host{
						{
							Id: "dut",
						},
					},
					Switches: []*lab.PasitTopology_Switch{
						{
							Id: "dut",
						},
					},
				},
			},
		},
		{
			cmd: &managePasitTopologyCmd{
				dutName: "dut",
				topologyObj: &lab.PasitTopology{
					Hosts: []*lab.PasitTopology_Host{
						{
							Id: "id1",
						},
					},
				},
			},
			want: []string{errHostNotInTopology},
		},
		{
			cmd: &managePasitTopologyCmd{
				dutName: "dut",
				topologyObj: &lab.PasitTopology{
					Hosts: []*lab.PasitTopology_Host{
						{
							Id: "dut",
							Ports: []*lab.PasitTopology_Port{
								{
									Technology:         lab.PasitTopology_PERHIPHERAL_TECHNOLOGY_ETHERNET,
									ConnectedComponent: "id2",
								},
							},
						},
					},
				},
			},
			want: []string{errMissingComponent},
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
}
