// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tasks

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os/user"

	"github.com/google/uuid"
	"github.com/maruel/subcommands"

	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"

	"infra/cmd/shivas/site"
	"infra/cmd/shivas/utils"
	"infra/cros/recovery/config"
	"infra/libs/skylab/buildbucket"
)

type reserveDuts struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
	envFlags  site.EnvFlags

	comment        string
	session        string
	expirationMins int
	// Configuration for reserve task.
	config string
}

// ReserveDutsCmd contains reserve-dut command specification
var ReserveDutsCmd = &subcommands.Command{
	UsageLine: "reserve-duts [-comment {comment}] [-session {admin-session}] [-expiration-mins 120] {HOST...}",
	ShortDesc: "Reserve the DUT by name",
	LongDesc: `Reserve the DUT by name.
	./shivas reserve <dut_name>
	Schedule a swarming Reserve task to the DUT to set the state to RESERVED to prevent scheduling tasks and tests to the DUT.
	Reserved DUT does not have expiration time and can be changed by scheduling any admin task on it.`,
	CommandRun: func() subcommands.CommandRun {
		c := &reserveDuts{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.Flags.IntVar(&c.expirationMins, "expiration-mins", 120, "The expiration minutes of the repair request.")
		c.Flags.StringVar(&c.comment, "comment", "", "The comment for reserved devices.")
		c.Flags.StringVar(&c.session, "session", "", "The admin session to group the tasks.")
		return c
	},
}

// Run represent runner for reserve command
func (c *reserveDuts) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		fmt.Fprintf(a.GetErr(), "%s: %s\n", a.GetName(), err)
		return 1
	}
	return 0
}

func (c *reserveDuts) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if len(args) == 0 {
		return errors.Reason("at least one hostname has to be provided").Err()
	}
	if c.comment == "" {
		user, err := user.Current()
		if err == nil && user != nil {
			c.comment = fmt.Sprintf("Reserved by %s", user.Username)
		}
	}
	if err := c.initConfig(); err != nil {
		return err
	}
	ctx := cli.GetContext(a, c, env)
	e := c.envFlags.Env()
	hc, err := buildbucket.NewHTTPClient(ctx, &c.authFlags)
	if err != nil {
		return err
	}
	bc, err := buildbucket.NewClient(ctx, hc, site.DefaultPRPCOptions, "chromeos", "labpack", "labpack")
	if err != nil {
		return err
	}
	if c.session == "" {
		c.session = uuid.New().String()
	}
	c.session = fmt.Sprintf("admin-session:%s", c.session)
	for _, host := range args {
		// TODO(crbug/1128496): update state directly in the UFS without creating the swarming task
		if url, _, err := c.scheduleReserveBuilder(ctx, bc, e, host); err != nil {
			fmt.Fprintf(a.GetErr(), "%s: fail with %s\n", host, err)
		} else {
			fmt.Fprintf(a.GetErr(), "%s: %s\n", host, url)
		}
	}
	utils.PrintTasksBatchLink(a.GetErr(), e.SwarmingService, c.session)
	return nil
}

// scheduleReserveBuilder schedules a labpack Buildbucket builder/recipe with the necessary arguments to run reserve.
func (c *reserveDuts) scheduleReserveBuilder(ctx context.Context, bc buildbucket.Client, e site.Environment, host string) (string, int64, error) {
	// TODO(b/229896419): refactor to hide labpack.Params struct.
	v := buildbucket.CIPDProd
	p := &buildbucket.Params{
		UnitName:     host,
		TaskName:     string(buildbucket.Custom),
		AdminService: e.AdminService,
		// NOTE: We use the UFS service, not the Inventory service here.
		InventoryService: e.UnifiedFleetService,
		NoStepper:        false,
		NoMetrics:        false,
		UpdateInventory:  true,
		Configuration:    c.config,
		ExtraTags: []string{
			c.session,
			"task:reserve",
			parisClientTag,
			fmt.Sprintf("version:%s", v),
			fmt.Sprintf("comment:%s", c.comment),
		},
	}
	url, taskID, err := buildbucket.ScheduleTask(ctx, bc, v, p)
	return url, taskID, errors.Annotate(err, "scheduleReserveBuilder").Err()
}

// initConfig initializes config used for scheduling reserve tasks.
func (c *reserveDuts) initConfig() error {
	rc := config.ReserveDutConfig()
	jsonByte, err := json.Marshal(rc)
	if err != nil {
		return errors.Annotate(err, "initConfig json err:").Err()
	}
	c.config = base64.StdEncoding.EncodeToString(jsonByte)
	return nil
}
