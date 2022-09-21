package controller

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/appengine/gaetesting"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"

	"infra/unifiedfleet/app/config"
	"infra/unifiedfleet/app/external"

	"go.chromium.org/luci/gae/service/datastore"
)

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
func TestImportENCBotConfig(t *testing.T) {
	t.Parallel()
	ctx := encTestingContext()
	Convey("Is Valid Public Chromium Test", t, func() {
		Convey("happy path", func() {
			err := ImportENCBotConfig(ctx)
			So(err, ShouldBeNil)
		})
	})
}
