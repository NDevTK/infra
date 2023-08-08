// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package peripherals

import (
	"fmt"
	"infra/cmd/shivas/cmdhelp"
	"infra/cmd/shivas/site"
	"infra/cmd/shivas/utils"
	"infra/cmdsupport/cmdlib"
	lab "infra/unifiedfleet/api/v1/models/chromeos/lab"
	rpc "infra/unifiedfleet/api/v1/rpc"
	"infra/unifiedfleet/app/util"
	"strings"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/flag"
	"go.chromium.org/luci/grpc/prpc"
)

var (
	AddChameleonCmd     = chamCmd(actionAdd)
	ReplaceChameleonCmd = chamCmd(actionReplace)
	DeleteChameleonCmd  = chamCmd(actionDelete)
)

// chamCmd creates command for adding, removing, or completely replacing Chameleon on a DUT.
func chamCmd(mode action) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "chameleon -dut {DUT name} -hostname {chameleon hostname} ",
		ShortDesc: "Manage Chameleon connect to a DUT",
		LongDesc:  cmdhelp.ManageChameleonLongDesc,
		CommandRun: func() subcommands.CommandRun {
			c := manageChamCmd{mode: mode}
			c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
			c.envFlags.Register(&c.Flags)
			c.commonFlags.Register(&c.Flags)

			c.Flags.StringVar(&c.dutName, "dut", "", "DUT name to update")
			c.Flags.StringVar(&c.hostname, "hostname", "", "hostname for Chameleon")
			c.Flags.Var(flag.StringSlice(&c.types), "type", "type of chameleon, can be specified multiple times")

			c.Flags.StringVar(&c.rpmHostname, "rpm", "", "hostname for rpm connected to chameleon")
			c.Flags.StringVar(&c.rpmOutlet, "rpm-outlet", "", "outlet number of rpm connected to chameleon")
			c.Flags.BoolVar(&c.audioBoard, "audio-board", false, "audio board chameleon")
			return &c
		},
	}
}

// manageChamCmd supports adding, replacing, or deleting Cham.
type manageChamCmd struct {
	subcommands.CommandRunBase
	authFlags   authcli.Flags
	envFlags    site.EnvFlags
	commonFlags site.CommonFlags

	dutName        string
	hostname       string
	types          []string
	chameleonTypes []lab.ChameleonType
	typesMap       map[lab.ChameleonType]bool

	rpmHostname string
	rpmOutlet   string
	audioBoard  bool

	mode action
}

// Run executed the Chameleon management subcommand. It cleans up passed flags and validates them.
func (c *manageChamCmd) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.run(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

// run implements the core logic for Run.
func (c *manageChamCmd) run(a subcommands.Application, args []string, env subcommands.Env) error {
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
		return errors.Annotate(err, "DUT name is not a chromeOS machine").Err()
	}

	var (
		peripherals = lse.GetChromeosMachineLse().GetDeviceLse().GetDut().GetPeripherals()
		currentCham = peripherals.GetChameleon()
	)

	nb, err := c.newCham(currentCham)
	if err != nil {
		return err
	}
	if c.commonFlags.Verbose() {
		fmt.Println("New Chameleon list", nb)
	}

	peripherals.Chameleon = nb
	//TODO(b:258280356) Use UFS chameleon field mask to do atomic partial update for on dut
	_, err = client.UpdateMachineLSE(ctx, &rpc.UpdateMachineLSERequest{MachineLSE: lse})
	return err
}

// newCham returns a new list of Cham based on the action specified in c and current list.
func (c *manageChamCmd) newCham(current *lab.Chameleon) (*lab.Chameleon, error) {
	switch c.mode {
	case actionAdd:
		if current != nil && current.GetHostname() != "" {
			return nil, errors.Reason("There already exist chameleon %s", current.GetHostname()).Err()
		}
		return c.createChameleon(), nil
	case actionReplace:
		return c.createChameleon(), nil
	case actionDelete:
		if c.hostname != current.GetHostname() {
			return nil, errors.Reason("deleting hostname %s does not match existing one %s", c.hostname, current.GetHostname()).Err()
		}
		return nil, nil
	default:
		return nil, errors.Reason("unknown action %d", c.mode).Err()
	}
}

const (
	errEmptyType     = "empty type"
	errDuplicateType = "duplicate type specified"
)

// cleanAndValidateFlags returns an error with the result of all validations. It strips whitespaces
// around hostnames and removes empty ones.
func (c *manageChamCmd) cleanAndValidateFlags() error {
	var errStrs []string
	if len(c.dutName) == 0 {
		errStrs = append(errStrs, errDUTMissing)
	}
	c.hostname = strings.TrimSpace(c.hostname)
	if c.hostname == "" {
		errStrs = append(errStrs, errNoHostname)
	}

	if c.typesMap == nil {
		c.typesMap = map[lab.ChameleonType]bool{}
	}

	var types []lab.ChameleonType
	for _, ts := range c.types {
		ts = strings.ToUpper(strings.TrimSpace(ts))

		// Empty Type
		if len(ts) == 0 {
			if c.commonFlags.Verbose() {
				fmt.Println("Empty type specified")
			}
			errStrs = append(errStrs, errEmptyType)
			continue
		}

		if !isChameleonType(ts) {
			errStrs = append(errStrs, fmt.Sprintf("Invalid chameleon type specified %s", ts))
			continue
		}
		t := toChameleonType(ts)

		// Duplicate Type
		if c.typesMap[t] {
			if c.commonFlags.Verbose() {
				fmt.Println("Duplicate type specified:", ts)
			}
			errStrs = append(errStrs, fmt.Sprintf("%s: %s", errDuplicateType, ts))
			continue
		}
		c.typesMap[t] = true
		types = append(types, t)
	}
	c.chameleonTypes = types

	if (c.rpmHostname != "" && c.rpmOutlet == "") || (c.rpmHostname == "" && c.rpmOutlet != "") {
		errStrs = append(errStrs, fmt.Sprintf("Need both rpm and its outlet. %s:%s is invalid", c.rpmHostname, c.rpmOutlet))
	}
	if len(errStrs) == 0 {
		return nil
	}

	return cmdlib.NewQuietUsageError(c.Flags, fmt.Sprintf("Wrong usage!!\n%s", strings.Join(errStrs, "\n")))
}

// createChameleon creates a *lab.Chameleon object from manageChamCmd variables.
func (c *manageChamCmd) createChameleon() *lab.Chameleon {
	ret := &lab.Chameleon{
		Hostname:             c.hostname,
		ChameleonPeripherals: c.chameleonTypes,
		AudioBoard:           c.audioBoard,
	}
	if c.rpmHostname != "" {
		ret.Rpm = &lab.OSRPM{
			PowerunitName:   c.rpmHostname,
			PowerunitOutlet: c.rpmOutlet,
		}
	}
	return ret
}

func isChameleonType(s string) bool {
	if val, ok := lab.ChameleonType_value[fmt.Sprintf("CHAMELEON_TYPE_%s", s)]; ok {
		return val != 0
	}
	return false
}

func toChameleonType(s string) lab.ChameleonType {
	if isChameleonType(s) {
		return lab.ChameleonType(lab.ChameleonType_value[fmt.Sprintf("CHAMELEON_TYPE_%s", s)])
	}
	return lab.ChameleonType_CHAMELEON_TYPE_INVALID
}
