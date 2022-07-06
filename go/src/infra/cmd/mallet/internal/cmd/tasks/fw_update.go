// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tasks

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"

	"infra/cmd/mallet/internal/site"
	"infra/cmdsupport/cmdlib"
	"infra/cros/recovery/config"
	"infra/cros/recovery/tasknames"
	"infra/libs/skylab/buildbucket"
	"infra/libs/skylab/buildbucket/labpack"
)

// Recovery subcommand: Deep repair task
var DeepRepair = &subcommands.Command{
	UsageLine: "deep-repair",
	ShortDesc: "Schedule deep repair task.",
	CommandRun: func() subcommands.CommandRun {
		c := &fwUpdateRun{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.Flags.StringVar(&c.gbbFlag, "flag", "0x18", "GBB flag")
		return c
	},
}

type fwUpdateRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
	envFlags  site.EnvFlags

	gbbFlag string
}

func (c *fwUpdateRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *fwUpdateRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	ctx := cli.GetContext(a, c, env)
	hc, err := buildbucket.NewHTTPClient(ctx, &c.authFlags)
	if err != nil {
		return errors.Annotate(err, "deep repair").Err()
	}
	bc, err := buildbucket.NewClient2(ctx, hc, site.DefaultPRPCOptions, site.BBProject, site.MalletBucket, site.MalletBuilder)
	if err != nil {
		return errors.Annotate(err, "deep repair").Err()
	}
	if len(args) == 0 {
		return errors.Reason("deep repair: unit is not specified").Err()
	}
	v := labpack.CIPDProd
	for _, unit := range args {
		e := c.envFlags.Env()
		configuration := b64.StdEncoding.EncodeToString(c.createPlan())
		_, taskID, err := labpack.ScheduleTask(
			ctx,
			bc,
			v,
			&labpack.Params{
				UnitName:         unit,
				TaskName:         string(tasknames.Custom),
				AdminService:     e.AdminService,
				InventoryService: e.UFSService,
				NoMetrics:        true,
				Configuration:    configuration,
				// We do not update as this is just manual action.
				UpdateInventory: false,
				ExtraTags: []string{
					"task:download_to_usb",
					site.ClientTag,
					fmt.Sprintf("version:%s", v),
				},
			},
		)
		if err != nil {
			fmt.Fprintf(a.GetErr(), "Created task for %q fail: %s\n", unit, err)
		} else {
			fmt.Fprintf(a.GetOut(), "Created task for %q: %s\n", unit, bc.BuildURL(taskID))
		}
	}
	return nil
}

func (c *fwUpdateRun) createPlan() []byte {
	rc := config.DeepRepairConfig()
	b, err := json.Marshal(rc)
	if err != nil {
		log.Fatalf("Failed to create JSON config: %v", err)
	}
	return b
}
