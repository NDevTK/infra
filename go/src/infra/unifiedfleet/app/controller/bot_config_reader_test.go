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
			EncConfig: []*config.OwnershipConfig_ConfigFile{
				{
					Name:       "test_name",
					RemotePath: "test_enc_git_path",
				},
			},
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

// Tests the functionality for importing bot configs from the config files
func TestImportENCBotConfig(t *testing.T) {
	t.Parallel()
	Convey("Import ENC Bot Config", t, func() {
		contextConfig := mockOwnershipConfig()
		Convey("happy path", func() {
			ctx := encTestingContext()
			ctx = config.Use(ctx, contextConfig)
			ownershipConfig, gitClient, err := GetConfigAndGitClient(ctx)
			So(err, ShouldBeNil)

			resp, err := registration.CreateMachine(ctx, mockChromeBrowserMachine("test1-1", "test1"))
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			err = ImportENCBotConfig(ctx, ownershipConfig, gitClient)
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
			err = ImportENCBotConfig(ctx, ownershipConfig, gitClient)
			So(err, ShouldBeNil)
			resp2, err := registration.GetMachine(ctx, "test1-1")
			So(resp2, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp2.Ownership, ShouldNotBeNil)
			So(resp2.GetUpdateTime(), ShouldResemble, resp.GetUpdateTime())
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

			err = ImportENCBotConfig(ctx, ownershipConfig, gitClient)
			So(err, ShouldBeNil)

			resp, err = registration.GetMachine(ctx, "testing-1")
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp.Ownership, ShouldNotBeNil)

			botResp, err := inventory.GetOwnershipData(ctx, "testing")
			So(err, ShouldBeNil)
			So(botResp, ShouldNotBeNil)
			So(botResp.OwnershipData, ShouldNotBeNil)
			p, err := botResp.GetProto()
			So(err, ShouldBeNil)
			pm := p.(*ufspb.OwnershipData)
			So(pm, ShouldResembleProto, resp.Ownership)

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

			err = ImportENCBotConfig(ctx, ownershipConfig, gitClient)
			So(err, ShouldBeNil)

			resp, err = inventory.GetMachineLSE(ctx, "testLSE1")
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp.Ownership, ShouldNotBeNil)

			// Import Again, should not update the Asset
			err = ImportENCBotConfig(ctx, ownershipConfig, gitClient)
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

			err = ImportENCBotConfig(ctx, ownershipConfig, gitClient)
			So(err, ShouldBeNil)

			resp, err := inventory.GetVM(ctx, "vm-1")
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp.Ownership, ShouldNotBeNil)

			// Import Again, should not update the Asset
			err = ImportENCBotConfig(ctx, ownershipConfig, gitClient)
			So(err, ShouldBeNil)
			resp2, err := inventory.GetVM(ctx, "vm-1")
			So(resp2, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp2.Ownership, ShouldNotBeNil)
			So(resp2.GetUpdateTime(), ShouldResemble, resp.GetUpdateTime())
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
			err = ImportENCBotConfig(ctx, ownershipConfig, gitClient)
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

// Tests the functionality for parsing and storing bot ownership configs in Datastore
func TestParseBotConfig(t *testing.T) {
	t.Parallel()
	ctx := encTestingContext()
	Convey("Parse ENC Bot Config", t, func() {
		contextConfig := mockOwnershipConfig()
		ctx = config.Use(ctx, contextConfig)
		Convey("happy path", func() {
			resp, err := registration.CreateMachine(ctx, mockChromeBrowserMachine("test1-1", "test1"))
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			ParseBotConfig(ctx, mockBotConfig("test{1,2}-1", "abc"), "testSwarming")

			resp, err = registration.GetMachine(ctx, "test1-1")
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp.Ownership, ShouldNotBeNil)
			So(resp.Ownership.PoolName, ShouldEqual, "abc")
			So(resp.Ownership.SwarmingInstance, ShouldEqual, "testSwarming")
		})
		Convey("Does not update non existent bots", func() {
			ParseBotConfig(ctx, mockBotConfig("test{2,3}-1", "abc"), "testSwarming")

			resp, err := registration.GetMachine(ctx, "test2-1")
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "NotFound")
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
			So(resp.Ownership.PoolName, ShouldEqual, "abc")
			So(resp.Ownership.SwarmingInstance, ShouldEqual, "testSwarming")
			So(resp.Ownership.Customer, ShouldEqual, "customer")
			So(resp.Ownership.SecurityLevel, ShouldEqual, "trusted")
			So(resp.Ownership.Builders, ShouldResemble, []string{"builder"})

			// Import bot configs, should not remove security fields
			ParseBotConfig(ctx, mockBotConfig("test{1,2}-1", "abc"), "testSwarming")
			resp, err = registration.GetMachine(ctx, "test1-1")
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp.Ownership, ShouldNotBeNil)
			So(resp.Ownership.PoolName, ShouldEqual, "abc")
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
			So(ownership.PoolName, ShouldEqual, "test")
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
			So(ownership.PoolName, ShouldEqual, "test")
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
			So(ownership.PoolName, ShouldEqual, "test")
			So(ownership.SwarmingInstance, ShouldEqual, "testSwarming")
		})
		Convey("missing host in inventory", func() {
			ctx := encTestingContext()
			ctx = config.Use(ctx, contextConfig)
			ParseBotConfig(ctx, mockBotConfig("test{4}-1", "abc"), "testSwarming")
			ownership, err := GetOwnershipData(ctx, "blah4-1")
			s, _ := status.FromError(err)

			So(ownership, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(s.Code(), ShouldEqual, codes.NotFound)
		})
	})
}
