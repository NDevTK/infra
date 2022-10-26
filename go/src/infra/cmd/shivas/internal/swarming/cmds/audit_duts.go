// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tasks

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/maruel/subcommands"

	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"

	"infra/cmd/shivas/site"
	"infra/libs/skylab/buildbucket"
	"infra/libs/skylab/worker"
	"infra/libs/swarming"
)

const dayInMinutes = 24 * 60

type auditRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
	envFlags  site.EnvFlags

	expirationMins      int
	runVerifyServoUSB   bool
	runVerifyDUTStorage bool
	runVerifyRpmConfig  bool

	actions string
	paris   bool
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
		c.Flags.IntVar(&c.expirationMins, "expiration-mins", 10, "The expiration minutes of the task request.")
		c.Flags.BoolVar(&c.paris, "paris", true, "Use PARIS rather than legacy flow.")
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

// innerRun runs paris. It validates the arguments and then hands control to legacy or paris as appropriate.
func (c *auditRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	ctx := cli.GetContext(a, c, env)
	if vErr := c.validateArgs(args); vErr != nil {
		return errors.Annotate(vErr, "audit dut").Err()
	}
	if c.paris {
		return errors.Annotate(c.innerRunParis(ctx, a, args, env), "audit dut").Err()
	}
	return errors.Annotate(c.innerRunLegacy(ctx, a, args, env), "audit dut").Err()
}

// innerRunParis runs audit for a paris task.
// We assume that the input parameters have been validated.
//
// Keep the behavior of this function consistent with innerRun.
func (c *auditRun) innerRunParis(ctx context.Context, a subcommands.Application, args []string, env subcommands.Env) error {
	e := c.envFlags.Env()
	creator, err := swarming.NewTaskCreator(ctx, &c.authFlags, e.SwarmingService)
	hc, err := buildbucket.NewHTTPClient(ctx, &c.authFlags)
	if err != nil {
		return errors.Annotate(err, "paris").Err()
	}
	bc, err := buildbucket.NewClient(ctx, hc, site.DefaultPRPCOptions, "chromeos", "labpack", "labpack")
	if err != nil {
		return errors.Annotate(err, "paris").Err()
	}
	sessionTag := fmt.Sprintf("admin-session:%s", uuid.New().String())
	taskNames, err := c.getTaskNames()
	if err != nil {
		return errors.Annotate(err, "paris").Err()
	}
	for _, host := range args {
		for _, taskName := range taskNames {
			creator.GenerateLogdogTaskCode()
			cmd := &worker.Command{TaskName: taskName}
			cmd.LogDogAnnotationURL = creator.LogdogURL()
			taskInfo, err := scheduleAuditBuilder(ctx, bc, e, taskName, host, sessionTag)
			if err != nil {
				fmt.Fprintf(a.GetErr(), "Skipping %q for %q because %s\n", taskName, host, err.Error())
			} else {
				fmt.Fprintf(a.GetErr(), "%s: %s: %s\n", host, taskName, taskInfo.TaskURL)
			}
		}
	}
	return nil
}

// innerRun is the main entrypoint for audit duts.
//
// Keep the behavior of this function consistent with innerRunWithParis.
func (c *auditRun) innerRunLegacy(ctx context.Context, a subcommands.Application, args []string, env subcommands.Env) (err error) {
	e := c.envFlags.Env()
	creator, err := swarming.NewTaskCreator(ctx, &c.authFlags, e.SwarmingService)
	if err != nil {
		return errors.Annotate(err, "audit dut").Err()
	}
	creator.LogdogService = e.LogdogService
	successMap := make(map[string]*swarming.TaskInfo)
	errorMap := make(map[string]error)
	for _, host := range args {
		cmd := &worker.Command{
			TaskName: "admin_audit",
			Actions:  c.actions,
		}
		creator.GenerateLogdogTaskCode()
		cmd.LogDogAnnotationURL = creator.LogdogURL()
		task, err := creator.LegacyAuditTask(ctx, e.SwarmingServiceAccount, host, c.expirationMins*60, cmd.Args(), cmd.LogDogAnnotationURL)
		if err != nil {
			errorMap[host] = err
		} else {
			successMap[host] = task
		}
	}
	creator.PrintResults(a.GetOut(), successMap, errorMap, true)
	return nil
}

func (c *auditRun) validateArgs(args []string) (err error) {
	if c.expirationMins >= dayInMinutes {
		return errors.Reason("validate args: expiration minutes (%d minutes) cannot exceed 1 day [%d minutes]", c.expirationMins, dayInMinutes).Err()
	}
	if len(args) == 0 {
		return errors.Reason("validate args: at least one host has to provided").Err()
	}
	c.actions, err = c.collectActions()
	if err != nil {
		return errors.Annotate(err, "validate args").Err()
	}
	return nil
}

// collectActions presents logic to generate actions string to run audit task.
//
// At least one action has to be specified.
func (c *auditRun) collectActions() (string, error) {
	var a []string
	if c.runVerifyDUTStorage {
		a = append(a, "verify-dut-storage")
	}
	if c.runVerifyServoUSB {
		a = append(a, "verify-servo-usb-drive")
	}
	if c.runVerifyRpmConfig {
		a = append(a, "verify-rpm-config")
	}
	if len(a) == 0 {
		return "", errors.Reason("collect actions: no actions was specified to run").Err()
	}
	return strings.Join(a, ","), nil
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
func scheduleAuditBuilder(ctx context.Context, bc buildbucket.Client, e site.Environment, taskName string, host string, adminSession string) (*swarming.TaskInfo, error) {
	tn, err := buildbucket.NormalizeTaskName(taskName)
	if err != nil {
		return nil, errors.Annotate(err, "schedule audit builder").Err()
	}
	v := buildbucket.CIPDProd
	p := &buildbucket.Params{
		BuilderProject: "",
		BuilderBucket:  "",
		BuilderName:    tn.BuilderName(),
		UnitName:       host,
		TaskName:       tn.String(),
		EnableRecovery: false,
		AdminService:   e.AdminService,
		// Note: UFS service is inventory service for fleet.
		InventoryService: e.UnifiedFleetService,
		NoStepper:        false,
		NoMetrics:        false,
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
	url, taskID, err := buildbucket.ScheduleTask(ctx, bc, v, p)
	if err != nil {
		return nil, err
	}
	taskInfo := &swarming.TaskInfo{
		// Use an ID format that makes it extremely obvious that we're dealing with a
		// buildbucket invocation number rather than a swarming task.
		ID:      fmt.Sprintf("buildbucket:%d", taskID),
		TaskURL: url,
	}
	return taskInfo, nil
}
