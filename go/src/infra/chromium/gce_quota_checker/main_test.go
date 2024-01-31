// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/encoding/prototext"

	gceproviderpb "go.chromium.org/luci/gce/api/config/v1"
)

func generateVMConfig(project string, zone string, amountMax int32, network string) *gceproviderpb.Configs {
	attributes := &gceproviderpb.VM{
		Project: project,
		NetworkInterface: []*gceproviderpb.NetworkInterface{
			{Network: network},
		},
		Zone: zone,
	}
	amount := &gceproviderpb.Amount{
		Max: amountMax,
	}
	config := &gceproviderpb.Config{
		Attributes: attributes,
		Amount:     amount,
	}
	configs := &gceproviderpb.Configs{
		Vms: []*gceproviderpb.Config{config},
	}
	return configs
}

func writeConfigs(tmpDir string, configs ...*gceproviderpb.Configs) []string {
	var configPaths []string
	for i, config := range configs {
		blob, err := prototext.Marshal(config)
		So(err, ShouldBeNil)
		configPath := path.Join(tmpDir, fmt.Sprintf("config%d.cfg", i))
		err = os.WriteFile(configPath, blob, 0666)
		So(err, ShouldBeNil)
		configPaths = append(configPaths, configPath)
	}
	return configPaths

}

// initMaps creates and returns all the quota maps, with empty quota entries
// created for all possible regions and networks. This would normally be done
// for us via the get*Quotas() functions, but those aren't covered in tests
// here, so need to do this ourselves.
func initMaps(possibleRegions []string, possibleNetworks []string) (map[string]*regionQuotas, map[string]*quotaVals) {
	quotasPerRegion := make(map[string]*regionQuotas)
	quotasPerNetwork := make(map[string]*quotaVals)
	for _, region := range possibleRegions {
		quotasPerRegion[region] = &regionQuotas{localSSDPerFamilyQuota: make(map[string]*quotaVals)}
	}
	for _, network := range possibleNetworks {
		quotasPerNetwork[network] = &quotaVals{}
	}
	return quotasPerRegion, quotasPerNetwork
}

func TestParseCfgFiles(t *testing.T) {
	t.Parallel()

	possibleRegions := []string{"us-east1", "us-west1", "us-central1"}
	possibleNetworks := []string{"networkA", "networkB"}

	Convey("test multiple projects", t, func() {
		quotasPerRegion, quotasPerNetwork := initMaps(possibleRegions, possibleNetworks)
		region, zone := possibleRegions[0], fmt.Sprintf("%s-a", possibleRegions[0])
		configs := []*gceproviderpb.Configs{
			generateVMConfig("projectA", zone, 1, possibleNetworks[0]),
			generateVMConfig("projectB", zone, 10, possibleNetworks[0]),
		}
		configPaths := writeConfigs(t.TempDir(), configs...)

		parseCfgFiles("projectA", configPaths, possibleRegions, quotasPerRegion, quotasPerNetwork)

		So(quotasPerRegion[region].instancesQuota.used, ShouldEqual, 1)
	})

	Convey("test multiple zones and regions", t, func() {
		quotasPerRegion, quotasPerNetwork := initMaps(possibleRegions, possibleNetworks)
		region1, region2 := possibleRegions[0], possibleRegions[1]
		zone1a, zone1b := fmt.Sprintf("%s-a", region1), fmt.Sprintf("%s-b", region1)
		zone2a, zone2b := fmt.Sprintf("%s-a", region2), fmt.Sprintf("%s-b", region2)
		configs := []*gceproviderpb.Configs{
			generateVMConfig("project", zone1a, 1, possibleNetworks[0]),
			generateVMConfig("project", zone1b, 2, possibleNetworks[0]),
			generateVMConfig("project", zone2a, 10, possibleNetworks[0]),
			generateVMConfig("project", zone2b, 20, possibleNetworks[0]),
		}
		configPaths := writeConfigs(t.TempDir(), configs...)

		parseCfgFiles("project", configPaths, possibleRegions, quotasPerRegion, quotasPerNetwork)

		So(quotasPerRegion[region1].instancesQuota.used, ShouldEqual, 3)
		So(quotasPerRegion[region2].instancesQuota.used, ShouldEqual, 30)
	})

	Convey("test networks", t, func() {
		quotasPerRegion, quotasPerNetwork := initMaps(possibleRegions, possibleNetworks)
		zone := fmt.Sprintf("%s-a", possibleRegions[0])
		network1, network2 := possibleNetworks[0], possibleNetworks[1]
		configs := []*gceproviderpb.Configs{
			generateVMConfig("project", zone, 1, network1),
			generateVMConfig("project", zone, 1, network1),
			generateVMConfig("project", zone, 1, network1),
			generateVMConfig("project", zone, 1, network2),
		}
		configPaths := writeConfigs(t.TempDir(), configs...)

		parseCfgFiles("project", configPaths, possibleRegions, quotasPerRegion, quotasPerNetwork)

		So(quotasPerNetwork[network1].used, ShouldEqual, 3)
		So(quotasPerNetwork[network2].used, ShouldEqual, 1)
	})

	// TODO: Add test coverage for IP address

	// TODO: Add test coverage for core count

	// TODO: Add test coverage for HDD and SSD quota
}
