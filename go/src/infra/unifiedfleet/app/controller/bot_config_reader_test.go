package controller

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/appengine/gaetesting"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
	"go.chromium.org/luci/gae/service/datastore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	ufspb "infra/unifiedfleet/api/v1/models"

	configpb "go.chromium.org/luci/swarming/proto/config"

	"infra/unifiedfleet/app/config"
	"infra/unifiedfleet/app/external"
	"infra/unifiedfleet/app/model/inventory"
	"infra/unifiedfleet/app/model/registration"
)

// encTestingContext creates a testing context which mocks the logging and datastore services.
// Also loads a custom config, which will allow the loading of a dummy bot config file
func encTestingContext() context.Context {
	c := gaetesting.TestingContextWithAppID("dev~infra-unified-fleet-system")
	c = gologger.StdConfig.Use(c)
	c = logging.SetLevel(c, logging.Error)
	c = config.Use(c, &config.Config{
		OwnershipConfig: &config.OwnershipConfig{
			GitilesHost: "test_gitiles",
			Project:     "test_project",
			Branch:      "test_branch",
			EncConfig: []*config.OwnershipConfig_ENCConfigFile{
				{
					Name:       "test_name",
					RemotePath: "test_enc_git_path",
				},
			},
		},
	})
	c = external.WithTestingContext(c)
	datastore.GetTestable(c).Consistent(true)
	return c
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

// Tests the functionality for importing bot configs from the config files
func TestImportENCBotConfig(t *testing.T) {
	t.Parallel()
	ctx := encTestingContext()
	Convey("Import ENC Bot Config", t, func() {
		Convey("happy path", func() {
			err := ImportENCBotConfig(ctx)
			So(err, ShouldBeNil)
		})
	})
}

// Tests the functionality for parsing and storing bot configs in Datastore
func TestParseBotConfig(t *testing.T) {
	t.Parallel()
	ctx := encTestingContext()
	Convey("Parse ENC Bot Config", t, func() {
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
	ctx := encTestingContext()
	Convey("GetOwnership Data", t, func() {
		Convey("happy path - machine", func() {
			resp, err := registration.CreateMachine(ctx, mockChromeBrowserMachine("test1-1", "test1"))
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			ParseBotConfig(ctx, mockBotConfig("test{1,2}-1", "abc"), "testSwarming")
			ownership, err := GetOwnershipData(ctx, "test1-1")

			So(ownership, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(ownership.PoolName, ShouldEqual, "abc")
			So(ownership.SwarmingInstance, ShouldEqual, "testSwarming")
		})
		Convey("happy path - vm", func() {
			resp, err := inventory.BatchUpdateVMs(ctx, []*ufspb.VM{{
				Name: "test2-1",
			}})
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			ParseBotConfig(ctx, mockBotConfig("test{1,2}-1", "abc"), "testSwarming")
			ownership, err := GetOwnershipData(ctx, "test2-1")

			So(ownership, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(ownership.PoolName, ShouldEqual, "abc")
			So(ownership.SwarmingInstance, ShouldEqual, "testSwarming")
		})
		Convey("happy path - machineLSE", func() {
			resp, err := inventory.CreateMachineLSE(ctx, &ufspb.MachineLSE{
				Name: "test3-1",
			})
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			ParseBotConfig(ctx, mockBotConfig("test{1,2,3}-1", "abc"), "testSwarming")
			ownership, err := GetOwnershipData(ctx, "test3-1")

			So(ownership, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(ownership.PoolName, ShouldEqual, "abc")
			So(ownership.SwarmingInstance, ShouldEqual, "testSwarming")
		})
		Convey("missing host in inventory", func() {
			ParseBotConfig(ctx, mockBotConfig("test{4}-1", "abc"), "testSwarming")
			ownership, err := GetOwnershipData(ctx, "test4-1")
			s, _ := status.FromError(err)

			So(ownership, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(s.Code(), ShouldEqual, codes.NotFound)
		})
	})
}
