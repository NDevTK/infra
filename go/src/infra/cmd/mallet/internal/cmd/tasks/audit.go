// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tasks

import (
	b64 "encoding/base64"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"

	"infra/cmd/mallet/internal/site"
	"infra/cmdsupport/cmdlib"
	"infra/cros/recovery/tasknames"
	"infra/libs/skylab/buildbucket"
	"infra/libs/skylab/buildbucket/labpack"
	"infra/libs/skylab/swarming"
)

// Audit subcommand: Recovering the devices.
var Audit = &subcommands.Command{
	UsageLine: "audit -type (usb|rpm|storage) [-no-steps] [-update-inv]",
	ShortDesc: "Audit the DUT",
	LongDesc:  "Audit the DUT.",
	CommandRun: func() subcommands.CommandRun {
		c := &auditRun{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.Flags.StringVar(&c.configFile, "config", "", "Path to the custom json config file.")
		c.Flags.StringVar(&c.auditType, "type", "", `Type of audit task: valid choices are "usb", "rpm", "storage".`)
		c.Flags.BoolVar(&c.noStepper, "no-stepper", false, "Block steper from using. This will prevent by using steps and you can only see logs.")
		c.Flags.BoolVar(&c.updateInv, "update-inv", false, "Update result to inventory. By default no.")
		c.Flags.BoolVar(&c.latest, "latest", false, "Use latest version of CIPD when scheduling. By default no.")
		c.Flags.StringVar(&c.adminSession, "admin-session", "", "Admin session used to group created tasks. By default generated.")
		return c
	},
}

// auditRun is the audit command.
type auditRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
	envFlags  site.EnvFlags

	noStepper    bool
	configFile   string
	auditType    string
	updateInv    bool
	latest       bool
	adminSession string
}

// names maps readable names provided at the command line to task names.
var names = map[string]tasknames.TaskName{
	"usb":     tasknames.AuditUSB,
	"storage": tasknames.AuditStorage,
	"rpm":     tasknames.AuditRPM,
}

// getTaskName gets the name of a task.
func (c auditRun) getTaskName() (string, error) {
	out, ok := names[c.auditType]
	if !ok {
		return "", errors.Reason(`get task name: unrecognized task name %q, try "-type rpm"`, c.auditType).Err()
	}
	return string(out), nil
}

// Run runs the audit task.
func (c *auditRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *auditRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	ctx := cli.GetContext(a, c, env)
	hc, err := buildbucket.NewHTTPClient(ctx, &c.authFlags)
	if err != nil {
		return errors.Annotate(err, "audit run").Err()
	}
	bc, err := buildbucket.NewClient2(ctx, hc, site.DefaultPRPCOptions, site.BBProject, site.MalletBucket, site.MalletBuilder)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return errors.Reason("create audit task: unit is not specified").Err()
	}
	// Admin session used to created common tag across created tasks.
	if c.adminSession == "" {
		c.adminSession = uuid.New().String()
	}
	sessionTag := fmt.Sprintf("admin-session:%s", c.adminSession)
	e := c.envFlags.Env()
	for _, unit := range args {
		var configuration string
		if c.configFile != "" {
			b, err := os.ReadFile(c.configFile)
			if err != nil {
				return errors.Annotate(err, "create audit task: open configuration file").Err()
			}
			configuration = b64.StdEncoding.EncodeToString(b)
		}

		v := labpack.CIPDProd
		if c.latest {
			v = labpack.CIPDLatest
		}
		task, err := c.getTaskName()
		if err != nil {
			return errors.Annotate(err, "create audit task").Err()
		}
		_, taskID, err := labpack.ScheduleTask(
			ctx,
			bc,
			v,
			&labpack.Params{
				UnitName:         unit,
				TaskName:         task,
				AdminService:     e.AdminService,
				InventoryService: e.UFSService,
				UpdateInventory:  c.updateInv,
				NoStepper:        c.noStepper,
				// TODO(gregorynisbet): send our metrics to the dev karte instance instead of dropping them.
				NoMetrics:     true,
				Configuration: configuration,
				ExtraTags: []string{
					sessionTag,
					fmt.Sprintf("task:%s", task),
					site.ClientTag,
					fmt.Sprintf("version:%s", v),
				},
			},
		)
		if err != nil {
			return errors.Annotate(err, "create audit task").Err()
		}
		fmt.Fprintf(a.GetOut(), "Created audit task for %s: %s\n", unit, bc.BuildURL(taskID))
	}
	fmt.Fprintf(a.GetOut(), "Created tasks: %s\n", swarming.TaskListURLForTags(e.SwarmingService, []string{sessionTag}))
	return nil
}
