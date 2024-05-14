// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package peripherals

import (
	"fmt"
	"strings"

	"github.com/maruel/subcommands"

	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/grpc/prpc"

	"infra/cmd/shivas/cmdhelp"
	"infra/cmd/shivas/site"
	"infra/cmd/shivas/utils"
	"infra/cmdsupport/cmdlib"
	lab "infra/unifiedfleet/api/v1/models/chromeos/lab"
	rpc "infra/unifiedfleet/api/v1/rpc"
	"infra/unifiedfleet/app/util"
)

const (
	DefaultPasitTopologyCommand = "peripheral-pasit-topology"
	errFileMissing              = "-f not provided"
	errIDMissing                = "topology component does not have ID"
	errDuplicateID              = "topology component has duplicate ID"
	errMissingComponent         = "topology does not contain component with matching ID"
)

var (
	AddPasitTopologyCmd    = pasitTopologyCmd(actionAdd, DefaultPasitTopologyCommand)
	DeletePasitTopologyCmd = pasitTopologyCmd(actionDelete, DefaultPasitTopologyCommand)
)

// pasitTopologyCmd creates command for adding, removing, or replacing a pasit DUTs topology.
func pasitTopologyCmd(mode action, command string) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: fmt.Sprintf("%s -dut {DUT name}", command),
		ShortDesc: "Manage Testbed PASIT topology",
		LongDesc:  cmdhelp.PasitTopologyLongDesc,
		CommandRun: func() subcommands.CommandRun {
			c := managePasitTopologyCmd{mode: mode}
			c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
			c.envFlags.Register(&c.Flags)
			c.commonFlags.Register(&c.Flags)

			c.Flags.StringVar(&c.dutName, "dut", "", "DUT name to update")
			c.Flags.StringVar(&c.topologyFile, "f", "", "File path to json file containing serialized topology proto")
			return &c
		},
	}
}

// managePasitTopologyCmd supports adding, replacing, or deleting PASIT testbed topology.
type managePasitTopologyCmd struct {
	subcommands.CommandRunBase
	authFlags   authcli.Flags
	envFlags    site.EnvFlags
	commonFlags site.CommonFlags

	dutName      string
	topologyFile string
	topologyObj  *lab.PasitHost
	mode         action
}

// Run executed the PASIT topology management subcommand.
func (c *managePasitTopologyCmd) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.run(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

// run implements the core logic for Run. It cleans up passed flags and validates them and updates the MachineLSE
func (c *managePasitTopologyCmd) run(a subcommands.Application, args []string, env subcommands.Env) error {
	if err := c.cleanAndValidateFlags(); err != nil {
		return err
	}
	ctx := cli.GetContext(a, c, env)
	ctx = utils.SetupContext(ctx, util.OSNamespace)

	hc, err := cmdlib.NewHTTPClient(ctx, &c.authFlags)
	if err != nil {
		return err
	}
	e := c.envFlags.Env()
	if c.commonFlags.Verbose() {
		fmt.Printf("Using UFS service %s\n", e.UnifiedFleetService)
	}

	client := rpc.NewFleetPRPCClient(&prpc.Client{
		C:       hc,
		Host:    e.UnifiedFleetService,
		Options: site.DefaultPRPCOptions,
	})

	lse, err := client.GetMachineLSE(ctx, &rpc.GetMachineLSERequest{
		Name: util.AddPrefix(util.MachineLSECollection, c.dutName),
	})
	if err != nil {
		return err
	}
	if err := utils.IsDUT(lse); err != nil {
		return errors.Annotate(err, "not a dut").Err()
	}

	models := getModels(c.topologyObj)
	if c.commonFlags.Verbose() {
		fmt.Println("New PASIT topology: ", c.topologyObj)
		fmt.Println("New PASIT features: ", models)
	}

	peripherals := lse.GetChromeosMachineLse().GetDeviceLse().GetDut().GetPeripherals()
	// Configure all "PASIT" peripherals.
	peripherals.PasitHost = c.topologyObj
	peripherals.PasitFeatures = models

	_, err = client.UpdateMachineLSE(ctx, &rpc.UpdateMachineLSERequest{MachineLSE: lse})
	return err
}

// cleanAndValidateFlags returns an error with the result of all validations.
func (c *managePasitTopologyCmd) cleanAndValidateFlags() error {
	c.dutName = strings.TrimSpace(c.dutName)
	if len(c.dutName) == 0 {
		return errors.Reason(errDUTMissing).Err()
	}

	// If deleting, set topology to nil and return.
	if c.mode == actionDelete {
		c.topologyObj = nil
		return nil
	}

	// topologyObj has already been defined (possibly by test) no need to read file.
	if c.topologyObj != nil {
		return c.validateNewTopology()
	}

	if c.topologyFile == "" {
		return errors.Reason(errFileMissing).Err()
	}

	c.topologyObj = &lab.PasitHost{}
	if err := utils.ParseJSONFile(c.topologyFile, c.topologyObj); err != nil {
		return errors.Annotate(err, "json parse error").Err()
	}

	return c.validateNewTopology()
}

// validateNewToplogy verifies that the requested topology does not have issues.
func (c *managePasitTopologyCmd) validateNewTopology() error {
	ids, err := getIds(c.topologyObj)
	if err != nil {
		return err
	}

	// Add dutName into known Ids
	ids[c.dutName] = true

	return checkIdsExists(c.topologyObj, ids)
}

// getIds gets all device IDs in the topology and ensures there are no empty entries or duplicates.
func getIds(topology *lab.PasitHost) (map[string]bool, error) {
	ids := make(map[string]bool)
	for _, d := range topology.GetDevices() {
		id := d.GetId()
		if id == "" {
			fmt.Println("Missing ID: ", d)
			return nil, errors.Reason(errIDMissing).Err()
		}
		if ids[id] {
			fmt.Println("Duplicate ID: ", id)
			return nil, errors.Reason(errDuplicateID).Err()
		}
		ids[id] = true
	}
	return ids, nil
}

// checkIdsExist ensures that all IDs listed in "connected_components" are valid.
func checkIdsExists(topology *lab.PasitHost, ids map[string]bool) error {
	for _, c := range topology.GetConnections() {
		if !ids[c.GetParent()] {
			fmt.Println("Unknown ID: ", c.GetParent())
			return errors.Reason(errMissingComponent).Err()
		}
		if !ids[c.GetChild()] {
			fmt.Println("Unknown ID: ", c.GetChild())
			return errors.Reason(errMissingComponent).Err()
		}
	}
	return nil
}

// getModels gets all hardware models in the testbed so we know what features
// the testbed can support.
func getModels(topology *lab.PasitHost) []string {
	var models []string

	// Empty topology has no features.
	if topology == nil {
		return models
	}

	seen := make(map[string]bool)
	for _, d := range topology.GetDevices() {
		m := d.GetModel()
		if m != "" && !seen[m] {
			models = append(models, strings.ToUpper(m))
			seen[m] = true
		}
	}
	return models
}
