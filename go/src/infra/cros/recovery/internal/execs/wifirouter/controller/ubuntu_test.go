// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"testing"
)

func TestUbuntuRouterController_parseNetworkControllerName(t *testing.T) {
	tests := []struct {
		name        string
		lscpiOutput string
		want        string
		wantErr     bool
	}{
		{
			"no output",
			"",
			"",
			true,
		},
		{
			"matching output",
			`00:00.0 Host bridge: Intel Corporation 11th Gen Core Processor Host Bridge/DRAM Registers (rev 01)
00:02.0 VGA compatible controller: Intel Corporation TigerLake-LP GT2 [Iris Xe Graphics] (rev 01)
00:06.0 PCI bridge: Intel Corporation 11th Gen Core Processor PCIe Controller (rev 01)
00:07.0 PCI bridge: Intel Corporation Tiger Lake-LP Thunderbolt 4 PCI Express Root Port #1 (rev 01)
00:07.2 PCI bridge: Intel Corporation Tiger Lake-LP Thunderbolt 4 PCI Express Root Port #2 (rev 01)
00:08.0 System peripheral: Intel Corporation GNA Scoring Accelerator module (rev 01)
00:0d.0 USB controller: Intel Corporation Tiger Lake-LP Thunderbolt 4 USB Controller (rev 01)
00:0d.2 USB controller: Intel Corporation Tiger Lake-LP Thunderbolt 4 NHI #0 (rev 01)
00:0d.3 USB controller: Intel Corporation Tiger Lake-LP Thunderbolt 4 NHI #1 (rev 01)
00:14.0 USB controller: Intel Corporation Tiger Lake-LP USB 3.2 Gen 2x1 xHCI Host Controller (rev 20)
00:14.2 RAM memory: Intel Corporation Tiger Lake-LP Shared SRAM (rev 20)
00:15.0 Serial bus controller: Intel Corporation Tiger Lake-LP Serial IO I2C Controller #0 (rev 20)
00:15.1 Serial bus controller: Intel Corporation Tiger Lake-LP Serial IO I2C Controller #1 (rev 20)
00:16.0 Communication controller: Intel Corporation Tiger Lake-LP Management Engine Interface (rev 20)
00:17.0 SATA controller: Intel Corporation Tiger Lake-LP SATA Controller (rev 20)
00:1d.0 PCI bridge: Intel Corporation Tiger Lake-LP PCI Express Root Port #9 (rev 20)
00:1d.1 PCI bridge: Intel Corporation Tiger Lake-LP PCI Express Root Port #10 (rev 20)
00:1f.0 ISA bridge: Intel Corporation Tiger Lake-LP LPC Controller (rev 20)
00:1f.3 Audio device: Intel Corporation Tiger Lake-LP Smart Sound Technology Audio Controller (rev 20)
00:1f.4 SMBus: Intel Corporation Tiger Lake-LP SMBus Controller (rev 20)
00:1f.5 Serial bus controller: Intel Corporation Tiger Lake-LP SPI Controller (rev 20)
01:00.0 Network controller: Intel Corporation Device 272b (rev 1a)
58:00.0 Network controller: Intel Corporation Device 272b (rev 1a)
59:00.0 Ethernet controller: Intel Corporation Ethernet Controller I225-LM (rev 03)
`,
			"Intel Corporation Device 272b",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &UbuntuRouterController{}
			got, err := c.parseNetworkControllerName(tt.lscpiOutput)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseNetworkControllerName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseNetworkControllerName() got = %v, want %v", got, tt.want)
			}
		})
	}
}
