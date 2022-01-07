// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dut

import (
	"context"
	"flag"
	"fmt"
	"infra/cmd/crosfleet/internal/buildbucket"
	"infra/cmd/crosfleet/internal/common"
	"infra/cmd/crosfleet/internal/flagx"
	crosfleetpb "infra/cmd/crosfleet/internal/proto"
	"infra/cmd/crosfleet/internal/site"
	"infra/cmdsupport/cmdlib"
	"strings"
	"sync"
	"time"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	swarmingapi "go.chromium.org/luci/common/api/swarming/swarming/v1"
	"go.chromium.org/luci/common/cli"
	luciflag "go.chromium.org/luci/common/flag"
)

const (
	// maxLeaseLengthMinutes is 24 hours in minutes.
	maxLeaseLengthMinutes = 24 * 60
	// Buildbucket priority for dut_leaser builds.
	dutLeaserBuildPriority = 15
	// leaseCmdName is the name of the `crosfleet dut lease` command.
	leaseCmdName = "lease"
	// Default DUT pool available for leasing from.
	defaultLeasesPool        = "DUT_POOL_QUOTA"
	maxLeaseReasonCharacters = 30
)

var lease = &subcommands.Command{
	UsageLine: fmt.Sprintf("%s [FLAGS...]", leaseCmdName),
	ShortDesc: "lease DUT for debugging",
	LongDesc: `Lease DUT for debugging.

DUTs can be leased by Swarming dimensions or by individual DUT hostname.
Leasing by dimensions is fastest, since the first available DUT matching the
requested dimensions is reserved. 'label-board' and 'label-model' dimensions can
be specified via the -board and -model flags, respectively; other Swarming
dimensions can be specified via the freeform -dim/-dims flags.

This command's behavior is subject to change without notice.
Do not build automation around this subcommand.`,
	CommandRun: func() subcommands.CommandRun {
		c := &leaseRun{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.printer.Register(&c.Flags)
		c.leaseFlags.register(&c.Flags)
		return c
	},
}

type leaseRun struct {
	subcommands.CommandRunBase
	leaseFlags
	authFlags authcli.Flags
	envFlags  common.EnvFlags
	printer   common.CLIPrinter
}

func (c *leaseRun) Run(a subcommands.Application, _ []string, env subcommands.Env) int {
	if err := c.innerRun(a, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *leaseRun) innerRun(a subcommands.Application, env subcommands.Env) error {
	if err := c.leaseFlags.validate(&c.Flags); err != nil {
		return err
	}
	// Extend board or models as necessary
	if len(c.boards) != len(c.models) {
		c.extendBoardOrModels()
	}
	c.printReceivedRequestsInfo()

	ctx := cli.GetContext(a, c, env)
	swarmingService, err := newSwarmingService(ctx, c.envFlags.Env().SwarmingService, &c.authFlags)
	if err != nil {
		return err
	}

	botDimsList, buildTagsList, err := botDimsAndBuildTags(ctx, swarmingService, c.leaseFlags)
	if err != nil {
		return err
	}

	// Verify provided DUT dimensions
	noMatchingDutFound := false
	c.printer.WriteTextStderr("Verifying the provided DUT dimensions...")
	for i, botDims := range botDimsList {
		duts, err := countBotsWithDims(ctx, swarmingService, botDims)
		if err != nil {
			return err
		}
		if duts.Count == 0 {
			c.printer.WriteTextStderr(fmt.Sprintf("(%v) No matching DUTs found; please double-check the provided DUT dimensions", i+1))
			noMatchingDutFound = true
		} else {
			c.printer.WriteTextStderr("(%v) Found %d DUT(s) (%d busy) matching the provided DUT dimensions", i+1, duts.Count, duts.Busy)
		}
	}

	// Should fail if any request has invalid dimension
	if noMatchingDutFound {
		return fmt.Errorf("please provide correct DUT dimensions for all requests")
	}

	leasesBBClient, err := buildbucket.NewClient(ctx, c.envFlags.Env().DUTLeaserBuilder, c.envFlags.Env().BuildbucketService, c.authFlags)
	if err != nil {
		return err
	}

	waitGroup := sync.WaitGroup{}
	mutex := sync.Mutex{}

	c.printer.WriteTextStderr("\nScheduling builds to lease DUTs...\n")
	if !c.exitEarly {
		c.printer.WriteTextStderr("Will wait to confirm %s completion and print leased DUT details...\n(To skip this step, pass the -exit-early flag on future DUT %s commands)", leaseCmdName, leaseCmdName)
	}
	for i, botDimensions := range botDimsList {
		buildTags := buildTagsList[i]
		buildProps := map[string]interface{}{
			"lease_length_minutes": c.durationMins,
		}
		reqNum := i + 1
		botDims := botDimensions

		waitGroup.Add(1)
		go func() {
			var leaseInfo crosfleetpb.LeaseInfo
			var err error
			leaseInfo.Build, err = leasesBBClient.ScheduleBuild(ctx, buildProps, botDims, buildTags, dutLeaserBuildPriority)
			if err != nil {
				c.printer.WriteTextStderr("(%v) Unable to schedule build to lease DUT: %v", reqNum, err)
				waitGroup.Done()
				return
			}

			c.printer.WriteTextStderr("(%v) Requesting %d minute lease at %s", reqNum, c.durationMins, leasesBBClient.BuildURL(leaseInfo.Build.Id))
			mutex.Lock()
			defer func() {
				waitGroup.Done()
				mutex.Unlock()
			}()
			if !c.exitEarly {
				leaseInfo.Build, err = leasesBBClient.WaitForBuildStepStart(ctx, leaseInfo.Build.Id, c.leaseStartStepName())
				if err != nil {
					c.printer.WriteTextStderr("(%v) Unable to confirm DUT lease through scheduled build: %v", reqNum, err)
					return
				}
				host := buildbucket.FindDimValInFinalDims("dut_name", leaseInfo.Build)
				endTime := time.Now().Add(time.Duration(c.durationMins) * time.Minute).Format(time.RFC822)
				c.printer.WriteTextStdout("\n(%v) Leased %s until %s\n", reqNum, host, endTime)
				ufsClient, err := newUFSClient(ctx, c.envFlags.Env().UFSService, &c.authFlags)
				if err != nil {
					// Don't fail the command here, since the DUT is already leased.
					c.printer.WriteTextStderr("Unable to contact UFS to print DUT info: %v", err)
					return
				}
				leaseInfo.DUT, err = getDutInfo(ctx, ufsClient, host)
				if err != nil {
					// Don't fail the command here, since the DUT is already leased.
					c.printer.WriteTextStderr("Unable to print DUT info: %v", err)
					return
				}
				c.printer.WriteTextStderr("%s\n", dutInfoAsBashVariables(leaseInfo.DUT))
			}
			c.printer.WriteJSONStdout(&leaseInfo)
		}()
	}
	waitGroup.Wait()

	return nil
}

// Print received requests info so that user can track using request number later
func (c *leaseRun) printReceivedRequestsInfo() {
	c.createRequestStrings()
	c.printer.WriteTextStderr("Requests received:")
	for _, req := range c.requestStrings {
		c.printer.WriteTextStderr(req)
	}
	// New line to separate rest of the section from this info section
	c.printer.WriteTextStderr("")
}

// botDimsAndBuildTags constructs bot dimensions and Buildbucket build tags for
// a dut_leaser build from the given lease flags and optional bot ID.
func botDimsAndBuildTags(ctx context.Context, swarmingService *swarmingapi.Service, leaseFlags leaseFlags) (dimList, tagList []map[string]string, err error) {
	dimList = []map[string]string{}
	tagList = []map[string]string{}
	if len(leaseFlags.hosts) > 0 {
		// Hostname-based lease.
		for _, host := range leaseFlags.hosts {
			correctedHostname := correctedHostname(host)
			id, err := hostnameToBotID(ctx, swarmingService, correctedHostname)
			if err != nil {
				return nil, nil, err
			}

			dimMap := make(map[string]string)
			tagMap := make(map[string]string)

			tagMap["lease-by"] = "host"
			tagMap["id"] = id
			dimMap["id"] = id

			tagMap[common.CrosfleetToolTag] = leaseCmdName
			tagMap["lease-reason"] = leaseFlags.reason
			tagMap["qs_account"] = "leases"

			dimList = append(dimList, dimMap)
			tagList = append(tagList, tagMap)
		}
	} else {
		// Swarming dimension-based lease.
		for i, board := range leaseFlags.boards {
			dimMap := make(map[string]string)
			tagMap := make(map[string]string)
			// Models should have the same length as of boards.
			// If they are provided in different lengths as input,
			// the remaining values should have been already extended with null strings.
			model := leaseFlags.models[i]

			dimMap["dut_state"] = "ready"
			dimMap["label-pool"] = defaultLeasesPool
			// Add user-added dimensions to both bot dimensions and build tags.
			for key, val := range leaseFlags.freeformDims {
				dimMap[key] = val
				tagMap[key] = val
			}
			if board != "" {
				tagMap["label-board"] = board
				dimMap["label-board"] = board
			}
			if model != "" {
				tagMap["label-model"] = model
				dimMap["label-model"] = model
			}

			// Add these metadata tags last to avoid being overwritten by freeform dims.
			tagMap[common.CrosfleetToolTag] = leaseCmdName
			tagMap["lease-reason"] = leaseFlags.reason
			tagMap["qs_account"] = "leases"

			dimList = append(dimList, dimMap)
			tagList = append(tagList, tagMap)
		}
	}
	return
}

// leaseFlags contains parameters for the "dut lease" subcommand.
type leaseFlags struct {
	durationMins   int64
	reason         string
	hosts          []string
	models         []string
	boards         []string
	freeformDims   map[string]string
	exitEarly      bool
	requestStrings []string
}

// Registers lease-specific flags.
func (c *leaseFlags) register(f *flag.FlagSet) {
	f.Int64Var(&c.durationMins, "minutes", 60, "Duration of lease in minutes.")
	f.StringVar(&c.reason, "reason", "", fmt.Sprintf("Optional reason for leasing (limit %d characters).", maxLeaseReasonCharacters))
	f.Var(luciflag.CommaList(&c.boards), "board", "Comma-separated list of 'label-board' Swarming dimension to lease DUTs by. Board and model placed in same list position in input will result in AND logic while requesting DUTs.")
	f.Var(luciflag.CommaList(&c.models), "model", "Comma-separated list of 'label-model' Swarming dimension to lease DUTs by.Board and model placed in same list position in input will result in AND logic while requesting DUTs.")
	f.Var(luciflag.CommaList(&c.hosts), "host", `Comma-separated list of hostnames of individual DUTs to lease. If leasing by hostname instead of other Swarming dimensions,
and the host DUT is running another task, the lease won't start until that task completes.
Mutually exclusive with -board/-model/-dim(s).`)
	f.Var(flagx.KeyVals(&c.freeformDims), "dim", "Freeform Swarming dimension to lease DUT by, in format key=val or key:val; may be specified multiple times.")
	f.Var(flagx.KeyVals(&c.freeformDims), "dims", "Comma-separated Swarming dimensions, in same format as -dim.")
	f.BoolVar(&c.exitEarly, "exit-early", false, `Exit command as soon as lease is scheduled. crosfleet will not notify on lease validation failure,
or print the hostname of the leased DUT.`)
}

func (c *leaseFlags) validate(f *flag.FlagSet) error {
	var errors []string
	if !c.hasEitherHostnameOrSwarmingDims() {
		errors = append(errors, "must specify DUT dimensions (-board/-model/-dim(s)) or DUT hostname (-host), but not both")
	} else if len(c.hosts) > 0 {
		c.checkHostNameDuplicates(&errors)
	}

	if c.durationMins <= 0 {
		errors = append(errors, "duration should be greater than 0")
	}
	if c.durationMins > maxLeaseLengthMinutes {
		errors = append(errors, fmt.Sprintf("duration cannot exceed %d minutes (%d hours)", maxLeaseLengthMinutes, maxLeaseLengthMinutes/60))
	}
	if len(c.reason) > maxLeaseReasonCharacters {
		errors = append(errors, fmt.Sprintf("reason cannot exceed %d characters", maxLeaseReasonCharacters))
	}

	if len(errors) > 0 {
		return cmdlib.NewUsageError(*f, strings.Join(errors, "\n"))
	}
	return nil
}

// hasOnePrimaryDim verifies that the lease flags contain either a DUT hostname
// or swarming dimensions (via -board/-model/-dim(s)), but not both.
func (c *leaseFlags) hasEitherHostnameOrSwarmingDims() bool {
	hasHostname := len(c.hosts) > 0
	hasSwarmingDims := len(c.boards) > 0 || len(c.models) > 0 || len(c.freeformDims) > 0
	return hasHostname != hasSwarmingDims
}

func (c *leaseRun) leaseStartStepName() string {
	hours := c.durationMins / 60
	mins := c.durationMins % 60
	return fmt.Sprintf("lease DUT for %d hr %d min", hours, mins)
}

// Extend board or models with null strings as necessary. They need to be of same length because,
// board and model in same position will result in AND logic while requesting DUT.
func (c *leaseFlags) extendBoardOrModels() {
	longer := &c.boards
	shorter := &c.models
	if len(c.boards) < len(c.models) {
		longer = &c.models
		shorter = &c.boards
	}
	newSlice := make([]string, len(*longer), cap(*longer))
	copy(newSlice, *shorter)
	*shorter = newSlice
}

// Check for duplicates in provided hostnames.
func (c *leaseFlags) checkHostNameDuplicates(errors *[]string) {
	visited := make(map[string]bool)
	for _, host := range c.hosts {
		_, ok := visited[host]
		if ok {
			*errors = append(*errors, fmt.Sprintf("duplicate host '%s' found in input", host))
		} else {
			visited[host] = true
		}
	}
}

// Create request strings to be used to print info.
func (c *leaseFlags) createRequestStrings() {
	if len(c.requestStrings) > 0 {
		return
	}

	if len(c.hosts) > 0 {
		for i, host := range c.hosts {
			c.requestStrings = append(c.requestStrings, fmt.Sprintf("(%v) Host name: %s", i+1, host))
		}
	} else {
		if len(c.boards) != len(c.models) {
			c.extendBoardOrModels()
		}
		for i, board := range c.boards {
			model := c.models[i]
			var reqString = fmt.Sprintf("(%v)", i+1)
			if board != "" {
				reqString += fmt.Sprintf(" Board: %s", board)
			}
			if model != "" {
				reqString += fmt.Sprintf(" Model: %s", model)
			}
			if len(c.freeformDims) > 0 {
				dimsString := ""
				for k, v := range c.freeformDims {
					dimsString += fmt.Sprintf("%s:%s ", k, v)
				}
				reqString += fmt.Sprintf(" Dims: [%s]", dimsString)
			}
			c.requestStrings = append(c.requestStrings, reqString)
		}
	}
}
