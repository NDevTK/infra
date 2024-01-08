// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/appengine/gaetesting"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/gae/service/datastore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	configpb "go.chromium.org/luci/swarming/proto/config"

	ufspb "infra/unifiedfleet/api/v1/models"
	"infra/unifiedfleet/app/config"
	"infra/unifiedfleet/app/external"
	"infra/unifiedfleet/app/model/inventory"
	"infra/unifiedfleet/app/model/registration"
)

var branchNumber uint32 = 0

// encTestingContext creates a testing context which mocks the logging and datastore services.
// Also loads a custom config, which will allow the loading of a dummy bot config file
func encTestingContext() context.Context {
	c := gaetesting.TestingContextWithAppID("dev~infra-unified-fleet-system")
	c = gologger.StdConfig.Use(c)
	c = logging.SetLevel(c, logging.Error)
	c = config.Use(c, &config.Config{})
	c = external.WithTestingContext(c)
	datastore.GetTestable(c).Consistent(true)
	return c
}

// Dummy UFS config with ownership config
func mockOwnershipConfig() *config.Config {
	return &config.Config{
		OwnershipConfig: &config.OwnershipConfig{
			GitilesHost: "test_gitiles",
			Project:     "test_project",
			Branch:      fmt.Sprintf("test_branch_%d", atomic.AddUint32(&branchNumber, 1)),
			SecurityConfig: []*config.OwnershipConfig_ConfigFile{
				{
					Name:       "test_name",
					RemotePath: "test_security_git_path",
				},
			},
		},
	}
}

// Dummy config for bots
func mockBotConfig(botRange string, pool string) *configpb.BotsCfg {
	return &configpb.BotsCfg{
		BotGroup: []*configpb.BotGroup{
			{
				BotId:      []string{botRange},
				Dimensions: []string{"pool:" + pool},
			},
		},
	}
}

// Dummy security config for bots
func mockSecurityConfig(botRange string, pool string, swarmingServerId string, customer string, securityLevel string, builders string) *ufspb.SecurityInfos {
	return &ufspb.SecurityInfos{
		Pools: []*ufspb.SecurityInfo{
			{
				Hosts:            []string{botRange},
				PoolName:         pool,
				SwarmingServerId: swarmingServerId,
				Customer:         customer,
				SecurityLevel:    securityLevel,
				Builders:         []string{builders},
			},
		},
	}
}

// Dummy ChromeBrowserMachine
func mockChromeBrowserMachine(id, name string) *ufspb.Machine {
	return &ufspb.Machine{
		Name: id,
		Device: &ufspb.Machine_ChromeBrowserMachine{
			ChromeBrowserMachine: &ufspb.ChromeBrowserMachine{
				Description: name,
			},
		},
	}
}

// Dummy MachineLSE
func mockMachineLSE(id string) *ufspb.MachineLSE {
	return &ufspb.MachineLSE{
		Name: id,
	}
}

// Tests the functionality for fetching the config file and importing the configs
// No t.Parallel(): The happy path test relies on comparing a global variable for sha1 hash testing.
// A race condition arises in which another test can modify this variable before the test finishes.
// In addition, the test suite hangs or panics when it reaches the timestamp deep equal assertion.
// TODO(b/265826661): Fix hang/panic with accessing test context and re-enable parallelization
func TestImportBotConfigs(t *testing.T) {
	ctx := encTestingContext()
	Convey("Import Bot Configs", t, func() {
		contextConfig := mockOwnershipConfig()
		ctx = config.Use(ctx, contextConfig)
		Convey("happy path", func() {
			resp, err := registration.CreateMachine(ctx, mockChromeBrowserMachine("test1-1", "test1"))
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			err = ImportBotConfigs(ctx)
			So(err, ShouldBeNil)

			resp, err = registration.GetMachine(ctx, "test1-1")
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp.Ownership, ShouldNotBeNil)

			botResp, err := inventory.GetOwnershipData(ctx, "test1-1")
			So(err, ShouldBeNil)
			So(botResp, ShouldNotBeNil)
			So(botResp.OwnershipData, ShouldNotBeNil)
			p, err := botResp.GetProto()
			So(err, ShouldBeNil)
			pm := p.(*ufspb.OwnershipData)
			So(pm, ShouldResembleProto, resp.Ownership)

			// Import Again, should not update the Asset
			err = ImportBotConfigs(ctx)
			So(err, ShouldBeNil)
			resp2, err := registration.GetMachine(ctx, "test1-1")
			So(resp2, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp2.Ownership, ShouldNotBeNil)
			So(resp2.GetUpdateTime(), ShouldResemble, resp.GetUpdateTime())
		})
	})
}

// Tests the functionality for fetching the config file and importing the configs at a particular commit hash
func TestGetBotConfigsForCommitSh(t *testing.T) {
	ctx := encTestingContext()
	t.Parallel()
	Convey("Get Bot Configs For CommitSh", t, func() {
		contextConfig := mockOwnershipConfig()
		ctx = config.Use(ctx, contextConfig)
		Convey("happy path", func() {
			ownerships, err := GetBotConfigsForCommitSh(ctx, "5201756875e0405c5c44d0e6d97de653b0d6cfca")
			So(err, ShouldBeNil)
			So(ownerships, ShouldNotBeNil)
			So(len(ownerships), ShouldNotBeZeroValue)
		})
		Convey("unknown commitsh", func() {
			ownerships, err := GetBotConfigsForCommitSh(ctx, "blah")
			So(err, ShouldNotBeNil)
			So(ownerships, ShouldBeNil)
		})
		Convey("empty commitsh", func() {
			ownerships, err := GetBotConfigsForCommitSh(ctx, "")
			So(err, ShouldNotBeNil)
			So(ownerships, ShouldBeNil)
		})
	})
}

// Tests the functionality for reading config files and fetching git client
// No t.Parallel(): The happy path test relies on comparing a global variable for sha1 hash testing.
// A race condition arises in which another test can modify this variable before the test finishes.
func TestGetConfigAndGitClient(t *testing.T) {
	ctx := encTestingContext()
	Convey("Get Ownership Config and Git Client", t, func() {
		contextConfig := mockOwnershipConfig()
		ctx = config.Use(ctx, contextConfig)
		Convey("happy path", func() {
			ownershipConfig, gitClient, err := GetConfigAndGitClient(ctx)
			So(err, ShouldBeNil)
			So(ownershipConfig, ShouldNotBeNil)
			So(gitClient, ShouldNotBeNil)

			// Fetch Again, should not return ownership config
			ownershipConfig, gitClient, err = GetConfigAndGitClient(ctx)
			So(err, ShouldBeNil)
			So(ownershipConfig, ShouldBeNil)
			So(gitClient, ShouldBeNil)
		})
		Convey("No Ownership Config - Ownership not updated", func() {
			ctx = config.Use(ctx, &config.Config{})
			ownershipConfig, gitClient, err := GetConfigAndGitClient(ctx)
			So(err, ShouldNotBeNil)
			So(ownershipConfig, ShouldBeNil)
			So(gitClient, ShouldBeNil)
		})
	})
}

// Tests the functionality for getting all the bot security configs
func TestGetAllOwnershipConfigs(t *testing.T) {
	t.Parallel()
	ctx := encTestingContext()
	Convey("Get all ownership Configs", t, func() {
		contextConfig := mockOwnershipConfig()
		ctx = config.Use(ctx, contextConfig)
		ownershipConfig, gitClient, err := GetConfigAndGitClient(ctx)
		So(err, ShouldBeNil)
		Convey("happy path", func() {
			err = ImportSecurityConfig(ctx, ownershipConfig, gitClient)
			So(err, ShouldBeNil)
			ownerships, _, err := ListOwnershipConfigs(ctx, 20, "", "", false)
			So(err, ShouldBeNil)
			So(ownerships, ShouldNotBeNil)
			So(len(ownerships), ShouldEqual, 14)
		})
	})
}

// Tests the functionality for importing security configs from the config files
func TestImportSecurityConfig(t *testing.T) {
	t.Parallel()
	Convey("Import Security Config", t, func() {
		contextConfig := mockOwnershipConfig()
		Convey("happy path", func() {
			ctx := encTestingContext()
			ctx = config.Use(ctx, contextConfig)
			ownershipConfig, gitClient, err := GetConfigAndGitClient(ctx)
			So(err, ShouldBeNil)
			resp, err := registration.CreateMachine(ctx, mockChromeBrowserMachine("test1-1", "test1"))
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			err = ImportSecurityConfig(ctx, ownershipConfig, gitClient)
			So(err, ShouldBeNil)

			resp, err = registration.GetMachine(ctx, "test1-1")
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp.Ownership, ShouldNotBeNil)

			botResp, err := inventory.GetOwnershipData(ctx, "test1-1")
			So(err, ShouldBeNil)
			So(botResp, ShouldNotBeNil)
			So(botResp.OwnershipData, ShouldNotBeNil)
			p, err := botResp.GetProto()
			So(err, ShouldBeNil)
			pm := p.(*ufspb.OwnershipData)
			So(pm, ShouldResembleProto, resp.Ownership)
			So(pm.GetPools(), ShouldResemble, []string{"abc"})

			// Update ownership and Import Again, should update the ownership to the original value
			pm.Pools = []string{"dummy"}
			registration.UpdateMachineOwnership(ctx, "test1-1", pm)
			entity, err := inventory.PutOwnershipData(ctx, pm, "test1-1", inventory.AssetTypeMachine)
			So(err, ShouldBeNil)
			p, err = entity.GetProto()
			So(err, ShouldBeNil)
			pm = p.(*ufspb.OwnershipData)
			So(pm.Pools, ShouldNotResemble, resp.Ownership.Pools)

			err = ImportSecurityConfig(ctx, ownershipConfig, gitClient)
			So(err, ShouldBeNil)
			resp2, err := registration.GetMachine(ctx, "test1-1")

			So(resp2, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp2.Ownership, ShouldNotBeNil)
			So(resp2.Ownership, ShouldResembleProto, resp.Ownership)
		})
		Convey("happy path - Bot ID Prefix", func() {
			ctx := encTestingContext()
			ctx = config.Use(ctx, contextConfig)
			ownershipConfig, gitClient, err := GetConfigAndGitClient(ctx)
			So(err, ShouldBeNil)
			resp, err := registration.CreateMachine(ctx, mockChromeBrowserMachine("testing-1", "testing1"))
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			resp, err = registration.CreateMachine(ctx, mockChromeBrowserMachine("tester-1", "tester1"))
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			err = ImportSecurityConfig(ctx, ownershipConfig, gitClient)
			So(err, ShouldBeNil)

			resp, err = registration.GetMachine(ctx, "testing-1")
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp.Ownership, ShouldNotBeNil)

			resp, err = registration.GetMachine(ctx, "tester-1")
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp.Ownership, ShouldBeNil)
		})
		Convey("happy path - Bot ID Prefix for MachineLSE", func() {
			ctx := encTestingContext()
			ctx = config.Use(ctx, contextConfig)
			ownershipConfig, gitClient, err := GetConfigAndGitClient(ctx)
			So(err, ShouldBeNil)
			resp, err := inventory.CreateMachineLSE(ctx, mockMachineLSE("testLSE1"))
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			err = ImportSecurityConfig(ctx, ownershipConfig, gitClient)
			So(err, ShouldBeNil)

			resp, err = inventory.GetMachineLSE(ctx, "testLSE1")
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp.Ownership, ShouldNotBeNil)

			// Import Again, should not update the Asset
			err = ImportSecurityConfig(ctx, ownershipConfig, gitClient)
			So(err, ShouldBeNil)
			resp2, err := inventory.GetMachineLSE(ctx, "testLSE1")
			So(resp2, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp2.Ownership, ShouldNotBeNil)
			So(resp2.GetUpdateTime(), ShouldResemble, resp.GetUpdateTime())
		})
		Convey("happy path - Bot ID Prefix for VM", func() {
			ctx := encTestingContext()
			ctx = config.Use(ctx, contextConfig)
			ownershipConfig, gitClient, err := GetConfigAndGitClient(ctx)
			So(err, ShouldBeNil)
			vm1 := &ufspb.VM{
				Name: "vm-1",
			}
			_, err = inventory.BatchUpdateVMs(ctx, []*ufspb.VM{vm1})
			So(err, ShouldBeNil)

			err = ImportSecurityConfig(ctx, ownershipConfig, gitClient)
			So(err, ShouldBeNil)

			resp, err := inventory.GetVM(ctx, "vm-1")
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp.Ownership, ShouldNotBeNil)

			// Import Again, should not update the Asset
			err = ImportSecurityConfig(ctx, ownershipConfig, gitClient)
			So(err, ShouldBeNil)
			resp2, err := inventory.GetVM(ctx, "vm-1")
			So(resp2, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp2.Ownership, ShouldNotBeNil)
			So(resp2.GetUpdateTime(), ShouldResemble, resp.GetUpdateTime())
		})
	})
}

// Tests the functionality for parsing and storing bot security configs in Datastore
func TestParseSecurityConfig(t *testing.T) {
	t.Parallel()
	ctx := encTestingContext()
	Convey("Parse Security Config", t, func() {
		contextConfig := mockOwnershipConfig()
		ctx = config.Use(ctx, contextConfig)
		Convey("happy path", func() {
			resp, err := registration.CreateMachine(ctx, mockChromeBrowserMachine("test1-1", "test1"))
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			ParseSecurityConfig(ctx, mockSecurityConfig("test{1,2}-1", "abc", "testSwarming", "customer", "trusted", "builder"))

			resp, err = registration.GetMachine(ctx, "test1-1")
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp.Ownership, ShouldNotBeNil)
			So(resp.Ownership.Pools, ShouldContain, "abc")
			So(resp.Ownership.SwarmingInstance, ShouldEqual, "testSwarming")
			So(resp.Ownership.Customer, ShouldEqual, "customer")
			So(resp.Ownership.SecurityLevel, ShouldEqual, "trusted")
			So(resp.Ownership.Builders, ShouldResemble, []string{"builder"})
		})
		Convey("Does not update non existent bots", func() {
			ParseSecurityConfig(ctx, mockSecurityConfig("test{2,3}-1", "abc", "testSwarming", "customer", "trusted", "builder"))

			resp, err := registration.GetMachine(ctx, "test2-1")
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "NotFound")
		})
		Convey("Cleans up stale configs", func() {
			ownership := &ufspb.OwnershipData{PoolName: "pool1"}
			resp, err := registration.CreateMachine(ctx, mockChromeBrowserMachine("test10-1", "test1"))
			So(resp, ShouldNotBeNil)
			So(resp.Ownership, ShouldBeNil)
			So(err, ShouldBeNil)

			resp, err = registration.UpdateMachineOwnership(ctx, "test10-1", ownership)
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp.Ownership, ShouldNotBeNil)

			_, err = inventory.PutOwnershipData(ctx, ownership, "test10-1", inventory.AssetTypeMachine)
			So(err, ShouldBeNil)

			ParseSecurityConfig(ctx, mockSecurityConfig("test{5,6}-1", "abc", "testSwarming", "customer", "trusted", "builder"))

			configResp, err := inventory.GetOwnershipData(ctx, "test10-1")
			So(configResp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "NotFound")
			resp, err = registration.GetMachine(ctx, "test10-1")
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp.Ownership, ShouldBeNil)
		})
	})
}

// Tests the functionality for parsing botId strings
func TestParseBotIds(t *testing.T) {
	t.Parallel()
	Convey("Parse ENC Bot Config", t, func() {
		Convey("Parse comma separated and ranges", func() {
			ids := parseBotIds("mac{9,10..11,12}-483")
			So(ids, ShouldResemble, []string{"mac9-483", "mac10-483", "mac11-483", "mac12-483"})
		})
		Convey("Parse multiple ranges", func() {
			ids := parseBotIds("mac{9,10..11,18..20}-483")
			So(ids, ShouldResemble, []string{"mac9-483", "mac10-483", "mac11-483", "mac18-483", "mac19-483", "mac20-483"})
		})
		Convey("Parse invalid range - ignores invalid range", func() {
			ids := parseBotIds("mac{9,10..11,22..20}-483")
			So(ids, ShouldResemble, []string{"mac9-483", "mac10-483", "mac11-483"})
		})
		Convey("Parse mal formed range - ignores malformed range", func() {
			ids := parseBotIds("mac{9,10..11,..20}-483")
			So(ids, ShouldResemble, []string{"mac9-483", "mac10-483", "mac11-483"})
		})
		Convey("Parse non digit characters in range - ignores", func() {
			ids := parseBotIds("mac{9,10,11..a}-483")
			So(ids, ShouldResemble, []string{"mac9-483", "mac10-483"})
		})
	})
}

// Tests the functionality for getting ownership data for a machine/vm/machineLSE
func TestGetOwnershipData(t *testing.T) {
	t.Parallel()
	Convey("GetOwnership Data", t, func() {
		contextConfig := mockOwnershipConfig()
		Convey("happy path - machine", func() {
			ctx := encTestingContext()
			ctx = config.Use(ctx, contextConfig)
			resp, err := registration.CreateMachine(ctx, mockChromeBrowserMachine("testing-1", "testing1"))
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			err = ImportBotConfigs(ctx)
			So(err, ShouldBeNil)
			ownership, err := GetOwnershipData(ctx, "testing-1")

			So(err, ShouldBeNil)
			So(ownership, ShouldNotBeNil)
			So(ownership.Pools, ShouldContain, "test")
			So(ownership.SwarmingInstance, ShouldEqual, "testSwarming")
		})
		Convey("happy path - vm", func() {
			ctx := encTestingContext()
			ctx = config.Use(ctx, contextConfig)
			vm1 := &ufspb.VM{
				Name: "vm-1",
			}
			_, err := inventory.BatchUpdateVMs(ctx, []*ufspb.VM{vm1})
			So(err, ShouldBeNil)

			err = ImportBotConfigs(ctx)
			So(err, ShouldBeNil)
			ownership, err := GetOwnershipData(ctx, "vm-1")

			So(ownership, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(ownership.Pools, ShouldContain, "test")
			So(ownership.SwarmingInstance, ShouldEqual, "testSwarming")
		})
		Convey("happy path - machineLSE", func() {
			ctx := encTestingContext()
			ctx = config.Use(ctx, contextConfig)
			resp, err := inventory.CreateMachineLSE(ctx, mockMachineLSE("testLSE1"))
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			err = ImportBotConfigs(ctx)
			So(err, ShouldBeNil)
			ownership, err := GetOwnershipData(ctx, "testLSE1")

			So(ownership, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(ownership.Pools, ShouldContain, "test")
			So(ownership.SwarmingInstance, ShouldEqual, "testSwarming")
		})
		Convey("missing host in inventory", func() {
			ctx := encTestingContext()
			ctx = config.Use(ctx, contextConfig)
			ParseSecurityConfig(ctx, mockSecurityConfig("test{2,3}-1", "abc", "testSwarming", "trusted", "customer", "builder"))
			ownership, err := GetOwnershipData(ctx, "blah4-1")
			s, _ := status.FromError(err)

			So(ownership, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(s.Code(), ShouldEqual, codes.NotFound)
		})
	})
}
