// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tasks

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"

	"infra/cmd/mallet/internal/site"
	"infra/cmdsupport/cmdlib"
	"infra/cros/recovery/config"
	"infra/libs/skylab/buildbucket"
	"infra/libs/skylab/common/heuristics"
	"infra/libs/skylab/swarming"
)

// Recovery subcommand: Erase mrc cache task
var EraseMRCCache = &subcommands.Command{
	UsageLine: "erase-mrc-cache",
	ShortDesc: "Schedule task to erase mrc cache of DUT(s) via their servo.",
	CommandRun: func() subcommands.CommandRun {
		c := &EraseMRCCacheRun{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.Flags.BoolVar(&c.latest, "latest", false, "Use latest version of CIPD when scheduling. By default no.")
		c.Flags.StringVar(&c.adminSession, "admin-session", "", "Admin session used to group created tasks. By default generated.")
		return c
	},
}

type EraseMRCCacheRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
	envFlags  site.EnvFlags

	latest       bool
	adminSession string
}

func (c *EraseMRCCacheRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *EraseMRCCacheRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	ctx := cli.GetContext(a, c, env)
	hc, err := buildbucket.NewHTTPClient(ctx, &c.authFlags)
	if err != nil {
		return errors.Annotate(err, "erase MRC cache").Err()
	}
	bc, err := buildbucket.NewClient(ctx, hc, site.DefaultPRPCOptions, site.BBProject, site.MalletBucket, site.MalletBuilder)
	if err != nil {
		return errors.Annotate(err, "erase MRC cache").Err()
	}
	if len(args) == 0 {
		return errors.Reason("erase MRC cache: unit is not specified").Err()
	}
	v := buildbucket.CIPDProd
	if c.latest {
		v = buildbucket.CIPDLatest
	}
	// Admin session used to created common tag across created tasks.
	if c.adminSession == "" {
		c.adminSession = uuid.New().String()
	}
	sessionTag := fmt.Sprintf("admin-session:%s", c.adminSession)
	for _, unit := range args {
		unit = heuristics.NormalizeBotNameToDeviceName(unit)
		e := c.envFlags.Env()
		configuration := b64.StdEncoding.EncodeToString(c.createPlan())
		url, _, err := buildbucket.ScheduleTask(
			ctx,
			bc,
			v,
			&buildbucket.Params{
				UnitName:         unit,
				TaskName:         string(buildbucket.Custom),
				AdminService:     e.AdminService,
				InventoryService: e.UFSService,
				NoMetrics:        true,
				Configuration:    configuration,
				// We do not update inventory as this is an ad-hoc task.
				UpdateInventory: false,
				ExtraTags: []string{
					"task:erase_mrc_cache",
					site.ClientTag,
					fmt.Sprintf("version:%s", v),
					sessionTag,
				},
			},
		)
		if err != nil {
			fmt.Fprintf(a.GetErr(), "Created task for %q fail: %s\n", unit, err)
		} else {
			fmt.Fprintf(a.GetOut(), "Created task for %q: %s\n", unit, url)
		}
	}
	// For run with more than one DUTs we provide a grouped tasks link for user to track all of them.
	if len(args) > 1 {
		fmt.Fprintf(a.GetOut(), "Created tasks: %s\n", swarming.TaskListURLForTags(c.envFlags.Env().SwarmingService, []string{sessionTag}))
	}
	return nil
}

func (c *EraseMRCCacheRun) createPlan() []byte {
	rc := config.EraseMRCCacheConfig()
	b, err := json.Marshal(rc)
	if err != nil {
		log.Fatalf("Failed to create JSON config: %v", err)
	}
	return b
}
