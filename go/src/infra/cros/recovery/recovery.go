// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package recovery provides ability to run recovery tasks against on the target units.
package recovery

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/luciexe/build"

	"infra/cros/recovery/config"
	"infra/cros/recovery/internal/engine"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/localtlw/localproxy"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/logger"
	"infra/cros/recovery/logger/metrics"
	"infra/cros/recovery/tlw"
	"infra/libs/skylab/buildbucket"
)

const (
	// Specify if we want to print the DUT info to the logs.
	// In some cases DUT info is too big and to avoid noise in the log you can block it.
	logDutInfo = true
)

// Run runs the recovery tasks against the provided unit.
// Process includes:
//   - Verification of input data.
//   - Set logger.
//   - Collect DUTs info.
//   - Load execution plan for required task with verification.
//   - Send DUTs info to inventory.
func Run(ctx context.Context, args *RunArgs) (rErr error) {
	if err := args.verify(); err != nil {
		return errors.Annotate(err, "run recovery: verify input").Err()
	}
	if args.Logger == nil {
		args.Logger = logger.NewLogger()
	}

	ctx = log.WithLogger(ctx, args.Logger)
	if !args.EnableRecovery {
		log.Infof(ctx, "Recovery actions is blocker by run arguments.")
	}
	log.Infof(ctx, "Run recovery for %q", args.UnitName)
	resources, err := retrieveResources(ctx, args)
	if err != nil {
		return errors.Annotate(err, "run recovery %q", args.UnitName).Err()
	}
	log.Infof(ctx, "Unit %q contains resources: %v", args.UnitName, resources)
	if args.Metrics == nil {
		log.Debugf(ctx, "run: metrics is nil")
	} else { // Guard against incorrectly setting up Karte client. See b:217746479 for details.
		log.Debugf(ctx, "run: metrics is non-nil")
		start := time.Now()
		// TODO(gregorynisbet): Create a helper function to make this more compact.
		defer (func() {
			// Keep this call up to date with NewMetric in execs.go.
			// Keep this up to date with runPlan in engine.go
			// Keep this up to date with runAction in engine.go
			action := &metrics.Action{
				ActionKind:     metrics.RunLibraryKind,
				StartTime:      start,
				StopTime:       time.Now(),
				SwarmingTaskID: args.SwarmingTaskID,
				BuildbucketID:  args.BuildbucketID,
				Hostname:       args.UnitName,
			}
			if rErr == nil {
				action.Status = metrics.ActionStatusSuccess
			} else {
				action.Status = metrics.ActionStatusFail
				action.FailReason = rErr.Error()
			}

			if mErr := args.Metrics.Create(ctx, action); mErr != nil {
				args.Logger.Errorf("Metrics error during teardown: %s", err)
			}
		})()
	}

	// Close all created local proxies.
	defer func() {
		localproxy.ClosePool()
	}()
	// Keep track of failure to run resources.
	// If one resource fail we still will try to run another one.
	var errs []error
	for ir, resource := range resources {
		if ir != 0 {
			log.Debugf(ctx, "Continue to the next resource.")
		}
		startTime := time.Now()
		err := runResource(ctx, resource, args)
		if err != nil {
			errs = append(errs, errors.Annotate(err, "run recovery %q", resource).Err())
		}
		// Create karte metric
		if createMetricErr := createTaskRunMetricsForResource(ctx, args, startTime, resource, err); createMetricErr != nil {
			args.Logger.Errorf("Create metric for resource: %q with error: %s", resource, createMetricErr)
		}
	}
	if len(errs) > 0 {
		return errors.Annotate(errors.MultiError(errs), "run recovery").Err()
	}
	return nil
}

// createTaskRunMetricsForResource creates metric action for resource with reporting what is the tasking is running for it.
func createTaskRunMetricsForResource(ctx context.Context, args *RunArgs, startTime time.Time, resource string, runResourceErr error) error {
	if args.Metrics == nil {
		log.Debugf(ctx, "Create karte action for each resource: For resource %s: metrics is not provided.", resource)
		return nil
	}
	action := &metrics.Action{
		ActionKind:     fmt.Sprintf(metrics.PerResourceTaskKindGlob, args.TaskName),
		StartTime:      startTime,
		StopTime:       time.Now(),
		SwarmingTaskID: args.SwarmingTaskID,
		BuildbucketID:  args.BuildbucketID,
		Hostname:       resource,
		Status:         metrics.ActionStatusSuccess,
		FailReason:     "",
	}
	if runResourceErr != nil {
		action.Status = metrics.ActionStatusFail
		action.FailReason = runResourceErr.Error()
	}
	mErr := args.Metrics.Create(ctx, action)
	return errors.Annotate(mErr, "create task run metrics for resource %s", resource).Err()
}

// runResource run single resource.
func runResource(ctx context.Context, resource string, args *RunArgs) (rErr error) {
	log.Infof(ctx, "Resource %q: started", resource)
	if args.ShowSteps {
		var step *build.Step
		step, ctx = build.StartStep(ctx, fmt.Sprintf("Start %q for %q", args.TaskName, resource))
		defer func() { step.End(rErr) }()
		stepLogCloser := log.AddStepLog(ctx, args.Logger, step, "execution details")
		defer func() { stepLogCloser() }()
	}
	dut, err := readInventory(ctx, resource, args)
	if err != nil {
		return errors.Annotate(err, "run resource %q", resource).Err()
	}
	// Load Configuration.
	config, err := loadConfiguration(ctx, dut, args)
	if err != nil {
		return errors.Annotate(err, "run resource %q", args.UnitName).Err()
	}
	// In any case update inventory to update data back, even execution failed.
	var errs []error
	if err := runDUTPlans(ctx, dut, config, args); err != nil {
		errs = append(errs, err)
	}
	if err := updateInventory(ctx, dut, args); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return errors.Annotate(errors.MultiError(errs), "run recovery").Err()
	}
	return nil
}

// retrieveResources retrieves a list of target resources.
func retrieveResources(ctx context.Context, args *RunArgs) (resources []string, err error) {
	if args.ShowSteps {
		var step *build.Step
		step, ctx = build.StartStep(ctx, fmt.Sprintf("Retrieve resources for %s", args.UnitName))
		defer func() { step.End(err) }()
		stepLogCloser := log.AddStepLog(ctx, args.Logger, step, "execution details")
		defer func() { stepLogCloser() }()
	}
	if i, ok := args.Logger.(logger.LogIndenter); ok {
		i.Indent()
		defer func() { i.Dedent() }()
	}
	resources, err = args.Access.ListResourcesForUnit(ctx, args.UnitName)
	return resources, errors.Annotate(err, "retrieve resources").Err()
}

// loadConfiguration loads and verifies a configuration.
// If configuration is not provided by args then default is used.
func loadConfiguration(ctx context.Context, dut *tlw.Dut, args *RunArgs) (rc *config.Configuration, err error) {
	if args.ShowSteps {
		var step *build.Step
		step, ctx = build.StartStep(ctx, "Load configuration")
		defer func() { step.End(err) }()
		stepLogCloser := log.AddStepLog(ctx, args.Logger, step, "execution details")
		defer func() { stepLogCloser() }()
	}
	if i, ok := args.Logger.(logger.LogIndenter); ok {
		i.Indent()
		defer func() { i.Dedent() }()
	}
	cr := args.configReader
	if cr == nil {
		if args.TaskName == buildbucket.Custom {
			return nil, errors.Reason("load configuration: expected config to be provided for custom tasks").Err()
		}
		// Get default configuration if not provided.
		if c, err := defaultConfiguration(args.TaskName, dut.SetupType); err != nil {
			return nil, errors.Annotate(err, "load configuration").Err()

		} else if cv, err := config.Validate(ctx, c, execs.Exist); err != nil {
			return nil, errors.Annotate(err, "load configuration").Err()
		} else {
			log.Infof(ctx, "using config for task name %q", args.TaskName)
			return cv, nil
		}
	}
	if c, err := parseConfiguration(ctx, cr); err != nil {
		return nil, errors.Annotate(err, "load configuration").Err()
	} else {
		return c, nil
	}
}

// ParsedDefaultConfiguration returns parsed default configuration for requested task and setup.
func ParsedDefaultConfiguration(ctx context.Context, tn buildbucket.TaskName, ds tlw.DUTSetupType) (*config.Configuration, error) {
	if c, err := defaultConfiguration(tn, ds); err != nil {
		return nil, errors.Annotate(err, "parse default configuration").Err()
	} else if cv, err := config.Validate(ctx, c, execs.Exist); err != nil {
		return nil, errors.Annotate(err, "parse default configuration").Err()
	} else {
		return cv, nil
	}
}

// parseConfiguration parses configuration to configuration proto instance.
func parseConfiguration(ctx context.Context, cr io.Reader) (*config.Configuration, error) {
	if c, err := config.Load(ctx, cr, execs.Exist); err != nil {
		return c, errors.Annotate(err, "parse configuration").Err()
	} else if len(c.GetPlans()) == 0 {
		return nil, errors.Reason("load configuration: no plans provided by configuration").Err()
	} else {
		return c, nil
	}
}

// defaultConfiguration provides configuration based on type of setup and task name.
func defaultConfiguration(tn buildbucket.TaskName, ds tlw.DUTSetupType) (*config.Configuration, error) {
	switch tn {
	case buildbucket.Recovery:
		switch ds {
		case tlw.DUTSetupTypeCros:
			return config.CrosRepairConfig(), nil
		case tlw.DUTSetupTypeLabstation:
			return config.LabstationRepairConfig(), nil
		case tlw.DUTSetupTypeAndroid:
			return config.AndroidRepairConfig(), nil
		case tlw.DUTSetupTypeCrosVM:
			return config.CrosVMSuccessConfig(), nil
		default:
			return nil, errors.Reason("Setup type: %q is not supported for task: %q!", ds, tn).Err()
		}
	case buildbucket.Deploy:
		switch ds {
		case tlw.DUTSetupTypeCros:
			return config.CrosDeployConfig(), nil
		case tlw.DUTSetupTypeLabstation:
			return config.LabstationDeployConfig(), nil
		case tlw.DUTSetupTypeAndroid:
			return config.AndroidDeployConfig(), nil
		default:
			return nil, errors.Reason("Setup type: %q is not supported for task: %q!", ds, tn).Err()
		}
	case buildbucket.AuditRPM:
		switch ds {
		case tlw.DUTSetupTypeCros:
			return config.CrosAuditRPMConfig(), nil
		case tlw.DUTSetupTypeCrosVM:
			return config.CrosVMSuccessConfig(), nil
		default:
			return nil, errors.Reason("setup type: %q is not supported for task: %q!", ds, tn).Err()
		}
	case buildbucket.AuditStorage:
		switch ds {
		case tlw.DUTSetupTypeCros:
			return config.CrosAuditStorageConfig(), nil
		case tlw.DUTSetupTypeCrosVM:
			return config.CrosVMSuccessConfig(), nil
		default:
			return nil, errors.Reason("setup type: %q is not supported for task: %q!", ds, tn).Err()
		}
	case buildbucket.AuditUSB:
		switch ds {
		case tlw.DUTSetupTypeCros:
			return config.CrosAuditUSBConfig(), nil
		case tlw.DUTSetupTypeCrosVM:
			return config.CrosVMSuccessConfig(), nil
		default:
			return nil, errors.Reason("setup type: %q is not supported for task: %q!", ds, tn).Err()
		}
	case buildbucket.Custom:
		return nil, errors.Reason("Setup type: %q does not have default configuration for custom tasks", ds).Err()
	default:
		return nil, errors.Reason("TaskName: %q is not supported..", tn).Err()
	}
}

// readInventory reads single resource info from inventory.
func readInventory(ctx context.Context, resource string, args *RunArgs) (dut *tlw.Dut, err error) {
	if args.ShowSteps {
		step, _ := build.StartStep(ctx, "Read inventory")
		defer func() { step.End(err) }()
		stepLogCloser := log.AddStepLog(ctx, args.Logger, step, "execution details")
		defer func() { stepLogCloser() }()
	}
	if i, ok := args.Logger.(logger.LogIndenter); ok {
		i.Indent()
		defer func() { i.Dedent() }()
	}
	defer func() {
		if r := recover(); r != nil {
			log.Debugf(ctx, "Read resource received panic!")
			err = errors.Reason("read resource panic: %v", r).Err()
		}
	}()
	dut, err = args.Access.GetDut(ctx, resource)
	if err != nil {
		return nil, errors.Annotate(err, "read inventory %q", resource).Err()
	}
	if logDutInfo {
		logDUTInfo(ctx, resource, dut, "DUT info from inventory")
	}
	return dut, nil
}

// updateInventory updates updated DUT info back to inventory.
//
// Skip update if not enabled.
func updateInventory(ctx context.Context, dut *tlw.Dut, args *RunArgs) (rErr error) {
	if args.ShowSteps {
		step, _ := build.StartStep(ctx, "Update inventory")
		defer func() { step.End(rErr) }()
		stepLogCloser := log.AddStepLog(ctx, args.Logger, step, "execution details")
		defer func() { stepLogCloser() }()
	}
	if i, ok := args.Logger.(logger.LogIndenter); ok {
		i.Indent()
		defer func() { i.Dedent() }()
	}
	if logDutInfo {
		logDUTInfo(ctx, dut.Name, dut, "updated DUT info")
	}
	if args.EnableUpdateInventory {
		log.Infof(ctx, "Update inventory %q: starting...", dut.Name)
		// Update DUT info in inventory in any case. When fail and when it passed
		if err := args.Access.UpdateDut(ctx, dut); err != nil {
			return errors.Annotate(err, "update inventory").Err()
		}
		log.Infof(ctx, "Update inventory %q: successful.", dut.Name)
	} else {
		log.Infof(ctx, "Update inventory %q: disabled.", dut.Name)
	}
	return nil
}

func logDUTInfo(ctx context.Context, resource string, dut *tlw.Dut, msg string) {
	s, err := json.MarshalIndent(dut, "", "\t")
	if err != nil {
		log.Debugf(ctx, "Resource %q: %s. Fail to print DUT info. Error: %s", resource, msg, err)
	} else {
		log.Infof(ctx, "Resource %q: %s \n%s", resource, msg, s)
	}
}

// runDUTPlans executes single DUT against task's plans.
func runDUTPlans(ctx context.Context, dut *tlw.Dut, c *config.Configuration, args *RunArgs) error {
	if i, ok := args.Logger.(logger.LogIndenter); ok {
		i.Indent()
		defer func() { i.Dedent() }()
	}
	log.Infof(ctx, "Run DUT %q: starting...", dut.Name)
	planNames := c.GetPlanNames()
	log.Debugf(ctx, "Run DUT %q plans: will use %s.", dut.Name, planNames)
	for _, planName := range planNames {
		if _, ok := c.GetPlans()[planName]; !ok {
			return errors.Reason("run dut %q plans: plan %q not found in configuration", dut.Name, planName).Err()
		}
	}
	// Creating one run argument for each resource.
	execArgs := &execs.RunArgs{
		DUT:            dut,
		Access:         args.Access,
		EnableRecovery: args.EnableRecovery,
		Logger:         args.Logger,
		ShowSteps:      args.ShowSteps,
		Metrics:        args.Metrics,
		SwarmingTaskID: args.SwarmingTaskID,
		BuildbucketID:  args.BuildbucketID,
		LogRoot:        args.LogRoot,
		// JumpHost: -- We explicitly do NOT pass the jump host to execs directly.
	}
	// As port 22 to connect to the lab is closed and there is work around to
	// create proxy for local execution. Creating proxy for all resources used
	// for this devices. We need created all of them at the beginning as one
	// plan can have access to current resource or another one.
	// Always has to be empty for merge code
	if jumpHostForLocalProxy := args.DevJumpHost; jumpHostForLocalProxy != "" {
		for _, planName := range planNames {
			resources := collectResourcesForPlan(planName, execArgs.DUT)
			for _, resource := range resources {
				if sh := execArgs.DUT.GetChromeos().GetServo(); sh.GetName() == resource && sh.GetContainerName() != "" {
					continue
				}
				if associatedHostname := execArgs.DUT.GetAndroid().GetAssociatedHostname(); associatedHostname != "" {
					resource = associatedHostname
				}
				if err := localproxy.RegHost(ctx, resource, jumpHostForLocalProxy); err != nil {
					return errors.Annotate(err, "run plans: create proxy for %q", resource).Err()
				}
			}
		}
	}
	defer func() {
		// Always try to run closing plan as the end of the configuration.
		plan, ok := c.GetPlans()[config.PlanClosing]
		if !ok {
			log.Infof(ctx, "Run plans: plan %q not found in configuration.", config.PlanClosing)
		} else {
			// Closing plan always allowed to fail.
			plan.AllowFail = true
			if err := runSinglePlan(ctx, config.PlanClosing, plan, execArgs); err != nil {
				log.Debugf(ctx, "Run plans: plan %q for %q finished with error: %s", config.PlanClosing, dut.Name, err)
			} else {
				log.Debugf(ctx, "Run plans: plan %q for %q finished successfully", config.PlanClosing, dut.Name)
			}
		}
	}()
	for _, planName := range planNames {
		if planName == config.PlanClosing {
			// The closing plan is always run as last one.
			continue
		}
		plan, ok := c.GetPlans()[planName]
		if !ok {
			return errors.Reason("run plans: plan %q: not found in configuration", planName).Err()
		}
		if err := runSinglePlan(ctx, planName, plan, execArgs); err != nil {
			return errors.Annotate(err, "run plans").Err()
		}
	}
	log.Infof(ctx, "Run DUT %q plans: finished successfully.", dut.Name)
	return nil
}

// runSinglePlan run single plan for all resources associated with plan.
func runSinglePlan(ctx context.Context, planName string, plan *config.Plan, execArgs *execs.RunArgs) error {
	log.Infof(ctx, "------====================-----")
	log.Infof(ctx, "Run plan %q: starting...", planName)
	log.Infof(ctx, "------====================-----")
	resources := collectResourcesForPlan(planName, execArgs.DUT)
	if len(resources) == 0 {
		log.Infof(ctx, "Run plan %q: no resources found.", planName)
		return nil
	}
	for _, resource := range resources {
		if err := runDUTPlanPerResource(ctx, resource, planName, plan, execArgs); err != nil {
			log.Infof(ctx, "Run %q plan for %s: finished with error: %s.", planName, resource, err)
			if plan.GetAllowFail() {
				log.Debugf(ctx, "Run plan %q for %q: ignore error as allowed to fail.", planName, resource)
			} else {
				return errors.Annotate(err, "run plan %q", planName).Err()
			}
		}
	}
	return nil
}

// runDUTPlanPerResource runs a plan against the single resource of the DUT.
func runDUTPlanPerResource(ctx context.Context, resource, planName string, plan *config.Plan, execArgs *execs.RunArgs) (rErr error) {
	execArgs.ResourceName = resource
	err := engine.Run(ctx, planName, plan, execArgs)
	return errors.Annotate(err, "run plan %q for %q", planName, execArgs.ResourceName).Err()
}

// collectResourcesForPlan collect resource names for supported plan.
// Mostly we have one resource per plan but in some cases we can have more
// resources and then we will run the same plan for each resource.
func collectResourcesForPlan(planName string, dut *tlw.Dut) []string {
	switch planName {
	case config.PlanCrOS, config.PlanAndroid, config.PlanClosing:
		if dut.Name != "" {
			return []string{dut.Name}
		}
	case config.PlanServo:
		if s := dut.GetChromeos().GetServo(); s != nil {
			return []string{s.GetName()}
		}
	case config.PlanBluetoothPeer:
		var resources []string
		for _, bp := range dut.GetChromeos().GetBluetoothPeers() {
			resources = append(resources, bp.GetName())
		}
		return resources
	case config.PlanChameleon:
		if c := dut.GetChromeos().GetChameleon(); c.GetName() != "" {
			return []string{c.GetName()}
		}
	case config.PlanWifiRouter:
		var resources []string
		for _, router := range dut.GetChromeos().GetWifiRouters() {
			resources = append(resources, router.GetName())
		}
		return resources
	}
	return nil
}

// RunArgs holds input arguments for recovery process.
//
// Keep this type up to date with internal/execs/execs.go:RunArgs .
// Also update recovery.go:runDUTPlans .
type RunArgs struct {
	// Access to the lab TLW layer.
	Access tlw.Access
	// UnitName represents some device setup against which running some tests or task in the system.
	// The unit can be represented as a single DUT or group of the DUTs registered in inventory as single unit.
	UnitName string
	// Provide access to read custom plans outside recovery. The default plans with actions will be ignored.
	configReader io.Reader
	// Logger prints message to the logs.
	Logger logger.Logger
	// Option to use steps.
	ShowSteps bool
	// Metrics is the metrics sink and event search API.
	Metrics metrics.Metrics
	// TaskName used to drive the recovery process.
	TaskName buildbucket.TaskName
	// EnableRecovery tells if recovery actions are enabled.
	EnableRecovery bool
	// EnableUpdateInventory tells if update inventory after finishing the plans is enabled.
	EnableUpdateInventory bool
	// SwarmingTaskID is the ID of the swarming task.
	SwarmingTaskID string
	// BuildbucketID is the ID of the buildbucket build
	BuildbucketID string
	// LogRoot is an absolute path to a directory.
	// All logs produced by actions or verifiers must be deposited there.
	LogRoot string
	// JumpHost is the host to use as a SSH proxy between ones dev environment and the lab,
	// if necessary. An empty JumpHost means do not use a jump host.
	DevJumpHost string
}

// UseConfigBase64 attaches a base64 encoded string as a config
// reader.
func (a *RunArgs) UseConfigBase64(blob string) error {
	if a == nil || blob == "" {
		return nil
	}
	dc, err := base64.StdEncoding.DecodeString(blob)
	if err != nil {
		return errors.Annotate(err, "original input %q", blob).Err()
	}
	a.configReader = bytes.NewReader(dc)
	return nil
}

// UseConfigFile attaches a config file to the current recovery object.
// We successfully do nothing when the path is empty.
func (a *RunArgs) UseConfigFile(path string) error {
	if path == "" || a == nil {
		return nil
	}
	cr, oErr := os.Open(path)
	a.configReader = cr
	return errors.Annotate(oErr, "use config file").Err()
}

// verify verifies input arguments.
func (a *RunArgs) verify() error {
	switch {
	case a == nil:
		return errors.Reason("is empty").Err()
	case a.UnitName == "":
		return errors.Reason("unit name is not provided").Err()
	case a.Access == nil:
		return errors.Reason("access point is not provided").Err()
	case a.LogRoot == "":
		// TODO(gregorynisbet): Upgrade this to a real error.
		fmt.Fprintf(os.Stderr, "%s\n", "log root cannot be empty!\n")
	}
	fmt.Fprintf(os.Stderr, "log root is %q\n", a.LogRoot)
	return nil
}
