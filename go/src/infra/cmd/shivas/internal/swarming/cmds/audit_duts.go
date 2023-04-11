// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tasks

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"

	"infra/cmd/shivas/site"
	"infra/libs/skylab/buildbucket"
	"infra/libs/skylab/common/heuristics"
)

type auditRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
	envFlags  site.EnvFlags

	runVerifyServoUSB   bool
	runVerifyDUTStorage bool
	runVerifyRpmConfig  bool

	latestVersion bool
}

// AuditDutsCmd contains audit-duts command specification
var AuditDutsCmd = &subcommands.Command{
	UsageLine: "audit-duts",
	ShortDesc: "Audit the DUT by name",
	LongDesc: `Audit the DUT by name.
	./shivas audit-duts -action1 -action2 <dut_name1> ...
	Schedule a swarming Audit task with required actions to the DUT to verify it.`,
	CommandRun: func() subcommands.CommandRun {
		c := &auditRun{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.Flags.BoolVar(&c.runVerifyServoUSB, "servo-usb", false, "Run the verifier for Servo USB drive.")
		c.Flags.BoolVar(&c.runVerifyDUTStorage, "dut-storage", false, "Run the verifier for DUT storage.")
		c.Flags.BoolVar(&c.runVerifyRpmConfig, "rpm-config", false, "Run the verifier to check and cache mac address of DUT NIC to Servo.")
		c.Flags.BoolVar(&c.latestVersion, "latest", false, "Use latest version of CIPD when scheduling. By default use prod.")
		return c
	},
}

// Run represent runner for reserve command
func (c *auditRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		fmt.Fprintf(a.GetErr(), "%s: %s\n", a.GetName(), err)
		return 1
	}
	return 0
}

// innerRun is the main entrypoint for audit duts.
// We assume that the input parameters have been validated.
func (c *auditRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	ctx := cli.GetContext(a, c, env)
	if len(args) == 0 {
		return errors.Reason("audit dut: at least one host has to provided").Err()
	}
	taskNames, err := c.getTaskNames()
	if err != nil {
		return errors.Annotate(err, "audit dut").Err()
	}
	e := c.envFlags.Env()
	hc, err := buildbucket.NewHTTPClient(ctx, &c.authFlags)
	if err != nil {
		return errors.Annotate(err, "audit dut").Err()
	}
	bc, err := buildbucket.NewClient(ctx, hc, site.DefaultPRPCOptions, "chromeos", "labpack", "labpack")
	if err != nil {
		return errors.Annotate(err, "audit dut").Err()
	}
	sessionTag := fmt.Sprintf("admin-session:%s", uuid.New().String())
	for _, host := range args {
		host = heuristics.NormalizeBotNameToDeviceName(host)
		for _, taskName := range taskNames {
			taskURL, err := scheduleAuditBuilder(ctx, bc, e, taskName, host, c.latestVersion, sessionTag)
			if err != nil {
				fmt.Fprintf(a.GetErr(), "Skipping %q for %q because %s\n", taskName, host, err.Error())
			} else {
				fmt.Fprintf(a.GetErr(), "%s: %s: %s\n", host, taskName, taskURL)
			}
		}
	}
	return nil
}

// getTaskNames gets the names of the Paris tasks that are going to be executed, one at a time
// in order to perform the audit tasks in question.
func (c *auditRun) getTaskNames() ([]string, error) {
	var a []string
	if c.runVerifyDUTStorage {
		a = append(a, buildbucket.AuditStorage.String())
	}
	if c.runVerifyServoUSB {
		a = append(a, buildbucket.AuditUSB.String())
	}
	if c.runVerifyRpmConfig {
		a = append(a, buildbucket.AuditRPM.String())
	}
	if len(a) == 0 {
		return nil, errors.Reason("get task names: no actions was specified to run").Err()
	}
	return a, nil
}

// scheduleAuditBuilder schedules a labpack Buildbucket builder/recipe with the necessary arguments to run repair.
func scheduleAuditBuilder(ctx context.Context, bc buildbucket.Client, e site.Environment, taskName string, host string, latestVersion bool, adminSession string) (string, error) {
	tn, err := buildbucket.NormalizeTaskName(taskName)
	if err != nil {
		return "", errors.Annotate(err, "schedule audit builder").Err()
	}
	v := buildbucket.CIPDProd
	if latestVersion {
		v = buildbucket.CIPDLatest
	}
	p := &buildbucket.Params{
		BuilderName:    tn.BuilderName(),
		UnitName:       host,
		TaskName:       tn.String(),
		EnableRecovery: false,
		AdminService:   e.AdminService,
		// Note: UFS service is inventory service for fleet.
		InventoryService: e.UnifiedFleetService,
		UpdateInventory:  true,
		// Note: Scheduled tasks are not expected custom configuration.
		Configuration: "",
		ExtraTags: []string{
			adminSession,
			fmt.Sprintf("task:%s", taskName),
			parisClientTag,
			fmt.Sprintf("version:%s", v),
			"qs_account:unmanaged_p0",
		},
	}
	url, _, err := buildbucket.ScheduleTask(ctx, bc, v, p)
	return url, errors.Annotate(err, "schedule audit builder").Err()
}
