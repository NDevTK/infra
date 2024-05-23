// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package peripherals

import (
	"fmt"
	"strings"

	"github.com/maruel/subcommands"

	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/grpc/prpc"

	"infra/cmd/shivas/cmdhelp"
	"infra/cmd/shivas/site"
	"infra/cmd/shivas/utils"
	"infra/cmdsupport/cmdlib"
	rpc "infra/unifiedfleet/api/v1/rpc"
	"infra/unifiedfleet/app/util"
)

const (
	DefaultPasitHostCommand = "peripheral-pasit-host"
	errFileMissing          = "-f not provided"
	errIDMissing            = "host device does not have ID"
	errDuplicateID          = "host topology component has duplicate ID"
	errMissingDevice        = "host topology does not contain device with matching ID"
	errDUTNotInHost         = "dut is not included in host topology"
	errMissingChild         = "connection has no ChildID"
	errMissingParent        = "connection has no ParentID"
	errChildEqualsParent    = "child and parent IDs are equal"
	errNoDevices            = "host topology requires a minimum of two devices to be defined"
	errNoConnections        = "host topology requires at least one connection to be properly defined"
)

var (
	AddPasitHostCmd    = pasitHostCmd(actionAdd, DefaultPasitHostCommand)
	DeletePasitHostCmd = pasitHostCmd(actionDelete, DefaultPasitHostCommand)
)

// pasitHostCmd creates command for adding, removing, or replacing a DUTs pasit host topology
func pasitHostCmd(mode action, command string) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: fmt.Sprintf("%s -dut {DUT name}", command),
		ShortDesc: "Manage Testbed PASIT host",
		LongDesc:  cmdhelp.PasitHostLongDesc,
		CommandRun: func() subcommands.CommandRun {
			c := managePasitHostCmd{mode: mode}
			c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
			c.envFlags.Register(&c.Flags)
			c.commonFlags.Register(&c.Flags)

			c.Flags.StringVar(&c.dutName, "dut", "", "DUT name to update")
			c.Flags.StringVar(&c.hostFile, "f", "", "File path to json file containing serialized host proto")
			return &c
		},
	}
}

// managePasitHostCmd supports adding, replacing, or deleting a DUTs pasit host topology.
type managePasitHostCmd struct {
	subcommands.CommandRunBase
	authFlags   authcli.Flags
	envFlags    site.EnvFlags
	commonFlags site.CommonFlags

	dutName  string
	hostFile string
	hostObj  *labapi.PasitHost
	mode     action
}

// Run executed the PASIT host topology management subcommand.
func (c *managePasitHostCmd) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.run(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

// run implements the core logic for Run. It cleans up passed flags and validates them and updates the MachineLSE
func (c *managePasitHostCmd) run(a subcommands.Application, args []string, env subcommands.Env) error {
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

	if c.commonFlags.Verbose() {
		fmt.Println("New PASIT host: ", c.hostObj)
	}

	peripherals := lse.GetChromeosMachineLse().GetDeviceLse().GetDut().GetPeripherals()
	peripherals.PasitHost = c.hostObj

	_, err = client.UpdateMachineLSE(ctx, &rpc.UpdateMachineLSERequest{MachineLSE: lse})
	return err
}

// cleanAndValidateFlags returns an error with the result of all validations.
func (c *managePasitHostCmd) cleanAndValidateFlags() error {
	c.dutName = strings.TrimSpace(c.dutName)
	if len(c.dutName) == 0 {
		return errors.Reason(errDUTMissing).Err()
	}

	// If deleting, set host to nil and return.
	if c.mode == actionDelete {
		c.hostObj = nil
		return nil
	}

	// hostObj has already been defined (possibly by test) no need to read file.
	if c.hostObj != nil {
		return c.validateNewHost()
	}

	if c.hostFile == "" {
		return errors.Reason(errFileMissing).Err()
	}

	c.hostObj = &labapi.PasitHost{}
	if err := utils.ParseJSONFile(c.hostFile, c.hostObj); err != nil {
		return errors.Annotate(err, "json parse error").Err()
	}

	return c.validateNewHost()
}

// validateNewToplogy verifies that the requested host topology object does is valid.
func (c *managePasitHostCmd) validateNewHost() error {
	// Verify that the host is present in the host configuration.
	if !c.hasMatchingHost() {
		return errors.Reason(errDUTNotInHost).Err()
	}

	if len(c.hostObj.GetDevices()) < 2 {
		return errors.Reason(errNoDevices).Err()
	}
	if len(c.hostObj.GetConnections()) < 1 {
		return errors.Reason(errNoConnections).Err()
	}

	// Verify that all devices have IDs.
	ids, err := c.getIDs()
	if err != nil {
		return err
	}

	// Verify that all Child/Parent IDs in connections exist
	return c.checkIDsExists(ids)
}

// hasMatchingHost ensures that the dut to be updated is also included in the host topology.
func (c *managePasitHostCmd) hasMatchingHost() bool {
	for _, h := range c.hostObj.GetDevices() {
		if h.GetId() == c.dutName && h.Type == labapi.PasitHost_Device_DUT {
			return true
		}
	}
	return false
}

// getIDs gets all device IDs in the host topology and ensures there are no empty entries or duplicates.
func (c *managePasitHostCmd) getIDs() (map[string]bool, error) {
	ids := make(map[string]bool)
	for _, d := range c.hostObj.GetDevices() {
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

// checkIDsExist ensures that all connection child/parent IDs are found in the topology.
func (c *managePasitHostCmd) checkIDsExists(ids map[string]bool) error {
	for _, c := range c.hostObj.GetConnections() {
		if c.GetParentId() == "" {
			fmt.Println("Missing parent ID: ", c)
			return errors.Reason(errMissingParent).Err()
		}
		if c.GetChildId() == "" {
			fmt.Println("Missing child ID: ", c)
			return errors.Reason(errMissingChild).Err()
		}
		if strings.EqualFold(c.GetParentId(), c.GetChildId()) {
			fmt.Println("Child and parent IDs equal: ", c)
			return errors.Reason(errChildEqualsParent).Err()
		}
		if !ids[c.GetParentId()] {
			fmt.Println("Unknown ID: ", c.GetParentId())
			return errors.Reason(errMissingDevice).Err()
		}
		if !ids[c.GetChildId()] {
			fmt.Println("Unknown ID: ", c.GetChildId())
			return errors.Reason(errMissingDevice).Err()
		}
	}
	return nil
}
