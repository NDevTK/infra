// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package schedulingunit

import (
	"fmt"
	"strings"

	"github.com/maruel/subcommands"

	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/flag"
	"go.chromium.org/luci/grpc/prpc"

	"infra/cmd/shivas/cmdhelp"
	"infra/cmd/shivas/site"
	"infra/cmd/shivas/utils"
	"infra/cmdsupport/cmdlib"
	"infra/libs/skylab/common/heuristics"
	ufspb "infra/unifiedfleet/api/v1/models"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
)

// defaultSchedulingUnitExposeType is the default label expose type for the SchedulingUnit.
const defaultSchedulingUnitExposeType = "default"

// AddSchedulingUnitCmd add SchedulingUnit to the system.
var AddSchedulingUnitCmd = &subcommands.Command{
	UsageLine: "schedulingunit",
	ShortDesc: "Add SchedulingUnit",
	LongDesc:  cmdhelp.AddSchedulingUnitLongDesc,
	CommandRun: func() subcommands.CommandRun {
		c := &addSchedulingUnit{
			pools: []string{},
			duts:  []string{},
			tags:  []string{},
		}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.commonFlags.Register(&c.Flags)

		c.Flags.StringVar(&c.newSpecsFile, "f", "", cmdhelp.SchedulingUnitFileText)

		c.Flags.StringVar(&c.name, "name", "", "name of the SchedulingUnit")
		c.Flags.Var(utils.CSVString(&c.pools), "pools", "comma separated pools")
		c.Flags.Var(utils.CSVString(&c.duts), "duts", "comma separated DUTs")
		c.Flags.Var(flag.StringSlice(&c.tags), "tag", "Name(s) of tag(s). Can be specified multiple times.")
		c.Flags.StringVar(&c.schedulingUnitType, "type", "all", "Type of SchedulingUnit. "+cmdhelp.SchedulingUnitTypesHelpText)
		c.Flags.StringVar(&c.description, "desc", "", "description for the SchedulingUnit")
		c.Flags.StringVar(&c.primaryDut, "primary-dut", "", "primary dut hostname")
		c.Flags.StringVar(&c.exposeType, "expose-type", defaultSchedulingUnitExposeType, "type of labels to expose to scheduling unit"+cmdhelp.SchedulingUnitExposeTypesHelpText)
		c.Flags.BoolVar(&c.wificell, "wificell", false, "adding this flag will specify if the scheduling unit is hosted in a wificell.")
		c.Flags.StringVar(&c.carrier, "carrier", "", "adding this flag will specify a carrier for scheduling units in cellular testbeds.")
		return c
	},
}

type addSchedulingUnit struct {
	subcommands.CommandRunBase
	authFlags   authcli.Flags
	envFlags    site.EnvFlags
	commonFlags site.CommonFlags

	newSpecsFile string

	name               string
	pools              []string
	duts               []string
	tags               []string
	schedulingUnitType string
	description        string
	primaryDut         string
	exposeType         string
	carrier            string
	wificell           bool
}

var mcsvFields = []string{
	"name",
	"duts",
	"pools",
	"type",
	"tags",
	"desc",
}

func (c *addSchedulingUnit) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *addSchedulingUnit) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if err := c.validateArgs(); err != nil {
		return err
	}
	ctx := cli.GetContext(a, c, env)
	ctx = utils.SetupContext(ctx, ufsUtil.OSNamespace)
	hc, err := cmdlib.NewHTTPClient(ctx, &c.authFlags)
	if err != nil {
		return err
	}
	e := c.envFlags.Env()
	if c.commonFlags.Verbose() {
		fmt.Printf("Using UFS service %s\n", e.UnifiedFleetService)
	}
	ic := ufsAPI.NewFleetPRPCClient(&prpc.Client{
		C:       hc,
		Host:    e.UnifiedFleetService,
		Options: site.DefaultPRPCOptions,
	})
	var su ufspb.SchedulingUnit
	var schedulingUnits []*ufspb.SchedulingUnit
	if c.newSpecsFile != "" {
		if utils.IsCSVFile(c.newSpecsFile) {
			schedulingUnits, err = c.parseMCSV()
			if err != nil {
				return err
			}
		} else {
			if err = utils.ParseJSONFile(c.newSpecsFile, &su); err != nil {
				return err
			}
			schedulingUnits = append(schedulingUnits, &su)
		}
	} else {
		c.parseArgs(&su)
		schedulingUnits = append(schedulingUnits, &su)
	}
	for _, r := range schedulingUnits {
		res, err := ic.CreateSchedulingUnit(ctx, &ufsAPI.CreateSchedulingUnitRequest{
			SchedulingUnit:   r,
			SchedulingUnitId: r.GetName(),
		})
		if err != nil {
			fmt.Printf("Failed to add SchedulingUnit %s. %s\n", r.GetName(), err)
			continue
		}
		res.Name = ufsUtil.RemovePrefix(res.Name)
		utils.PrintProtoJSON(res, !utils.NoEmitMode(false))
		fmt.Printf("Successfully added the SchedulingUnit %s\n", res.Name)
	}
	return nil
}

func (c *addSchedulingUnit) parseArgs(su *ufspb.SchedulingUnit) {
	su.Name = c.name
	su.Pools = c.pools
	su.MachineLSEs = c.duts
	su.Tags = c.tags
	su.Type = ufsUtil.ToSchedulingUnitType(c.schedulingUnitType)
	su.Description = c.description
	su.PrimaryDut = c.primaryDut
	val := ufspb.SchedulingUnit_ExposeType_value[c.exposeType]
	su.ExposeType = ufspb.SchedulingUnit_ExposeType(val)
	su.Wificell = c.wificell
	su.Carrier = c.carrier
}

func (c *addSchedulingUnit) validateArgs() error {
	if c.newSpecsFile != "" {
		if c.name != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe file mode is specified. '-name' cannot be specified at the same time.")
		}
		if len(c.pools) != 0 {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe file mode is specified. '-pools' cannot be specified at the same time.")
		}
		if len(c.duts) != 0 {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe file mode is specified. '-duts' cannot be specified at the same time.")
		}
		if len(c.tags) != 0 {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe file mode is specified. '-tags' cannot be specified at the same time.")
		}
		if c.description != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe file mode is specified. '-description' cannot be specified at the same time.")
		}
		if c.primaryDut != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe file mode is specified. '-primary-dut' cannot be specified at the same time.")
		}
		if c.exposeType != defaultSchedulingUnitExposeType {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe file mode is specified. '-expose-type' cannot be specified at the same time.")
		}
	}
	if c.newSpecsFile == "" {
		if c.name == "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\n'-name' is required, no mode ('-f') is specified.")
		}
		if c.schedulingUnitType != "" && !ufsUtil.IsSchedulingUnitType(c.schedulingUnitType) {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\n%s is not a valid SchedulingUnitType name, please check help info for '-type'.", c.schedulingUnitType)
		}
		if c.exposeType != "" {
			c.exposeType = strings.ToUpper(c.exposeType)
			if val, ok := ufspb.SchedulingUnit_ExposeType_value[c.exposeType]; !ok || val == 0 {
				return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\n%s is not valid exposeType name, please check help info for 'expose-type'.", c.exposeType)
			}
		}
	}
	return nil
}

// parseMCSV parses the MCSV file and returns SchedulingUnit requests.
func (c *addSchedulingUnit) parseMCSV() ([]*ufspb.SchedulingUnit, error) {
	records, err := utils.ParseMCSVFile(c.newSpecsFile)
	if err != nil {
		return nil, err
	}
	var schedulingUnits []*ufspb.SchedulingUnit
	for i, rec := range records {
		// if i is 0, determine whether this is a header.
		if i == 0 && heuristics.LooksLikeHeader(rec) {
			if err := utils.ValidateSameStringArray(mcsvFields, rec); err != nil {
				return nil, err
			}
			continue
		}
		su := &ufspb.SchedulingUnit{}
		for i := range mcsvFields {
			name := mcsvFields[i]
			value := rec[i]
			switch name {
			case "name":
				su.Name = value
			case "pools":
				su.Pools = strings.Fields(value)
			case "duts":
				su.MachineLSEs = strings.Fields(value)
			case "tags":
				su.Tags = strings.Fields(value)
			case "type":
				if !ufsUtil.IsSchedulingUnitType(value) {
					return nil, fmt.Errorf("Error in line %d.\n%s is not a valid SchedulingUnitType. %s", i, value, cmdhelp.SchedulingUnitTypesHelpText)
				}
				su.Type = ufsUtil.ToSchedulingUnitType(value)
			case "desc":
				su.Description = value
			default:
				return nil, fmt.Errorf("Error in line %d.\nUnknown field: %s", i, name)
			}
		}
		schedulingUnits = append(schedulingUnits, su)
	}
	return schedulingUnits, nil
}
