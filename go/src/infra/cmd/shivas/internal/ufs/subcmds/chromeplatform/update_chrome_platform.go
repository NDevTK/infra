// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package chromeplatform

import (
	"fmt"

	"github.com/maruel/subcommands"

	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/flag"
	"go.chromium.org/luci/grpc/prpc"

	"infra/cmd/shivas/cmdhelp"
	"infra/cmd/shivas/site"
	"infra/cmd/shivas/utils"
	"infra/cmdsupport/cmdlib"
	ufspb "infra/unifiedfleet/api/v1/models"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
)

// UpdateChromePlatformCmd update ChromePlatform by given name.
var UpdateChromePlatformCmd = &subcommands.Command{
	UsageLine: "platform",
	ShortDesc: "Update platform configuration for browser machine",
	LongDesc:  cmdhelp.UpdateChromePlatformLongDesc,
	CommandRun: func() subcommands.CommandRun {
		c := &updateChromePlatform{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.commonFlags.Register(&c.Flags)

		c.Flags.StringVar(&c.newSpecsFile, "f", "", cmdhelp.ChromePlatformFileText)
		c.Flags.BoolVar(&c.interactive, "i", false, "enable interactive mode for input")

		c.Flags.StringVar(&c.name, "name", "", "name of the platform")
		c.Flags.StringVar(&c.manufacturer, "manufacturer", "", "manufacturer name."+cmdhelp.ClearFieldHelpText)
		c.Flags.Var(flag.StringSlice(&c.tags), "tag", "Name(s) of tag(s). Can be specified multiple times. "+cmdhelp.ClearFieldHelpText)
		c.Flags.StringVar(&c.description, "desc", "", "description for the platform. "+cmdhelp.ClearFieldHelpText)
		return c
	},
}

type updateChromePlatform struct {
	subcommands.CommandRunBase
	authFlags   authcli.Flags
	envFlags    site.EnvFlags
	commonFlags site.CommonFlags

	newSpecsFile string
	interactive  bool

	name         string
	manufacturer string
	tags         []string
	description  string
}

func (c *updateChromePlatform) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *updateChromePlatform) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if err := c.validateArgs(); err != nil {
		return err
	}
	ctx := cli.GetContext(a, c, env)
	ns, err := c.envFlags.Namespace(nil, "")
	if err != nil {
		return err
	}
	ctx = utils.SetupContext(ctx, ns)
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
	var chromePlatform ufspb.ChromePlatform
	if c.interactive {
		utils.GetChromePlatformInteractiveInput(ctx, ic, &chromePlatform, true)
	} else {
		if c.newSpecsFile != "" {
			if err = utils.ParseJSONFile(c.newSpecsFile, &chromePlatform); err != nil {
				return err
			}
		} else {
			c.parseArgs(&chromePlatform)
		}
	}
	if err := utils.PrintExistingPlatform(ctx, ic, chromePlatform.Name); err != nil {
		return err
	}
	chromePlatform.Name = ufsUtil.AddPrefix(ufsUtil.ChromePlatformCollection, chromePlatform.Name)
	if !ufsUtil.ValidateTags(chromePlatform.Tags) {
		return fmt.Errorf(ufsAPI.InvalidTags)
	}
	res, err := ic.UpdateChromePlatform(ctx, &ufsAPI.UpdateChromePlatformRequest{
		ChromePlatform: &chromePlatform,
		UpdateMask: utils.GetUpdateMask(&c.Flags, map[string]string{
			"manufacturer": "manufacturer",
			"tag":          "tags",
			"desc":         "description",
		}),
	})
	if err != nil {
		return err
	}
	res.Name = ufsUtil.RemovePrefix(res.Name)
	fmt.Println("The chrome platform after update:")
	utils.PrintProtoJSON(res, !utils.NoEmitMode(false))
	fmt.Printf("Successfully updated the platform %s\n", res.Name)
	return nil
}

func (c *updateChromePlatform) parseArgs(chromePlatform *ufspb.ChromePlatform) {
	chromePlatform.Name = c.name
	if c.manufacturer == utils.ClearFieldValue {
		chromePlatform.Manufacturer = ""
	} else {
		chromePlatform.Manufacturer = c.manufacturer
	}
	if c.description == utils.ClearFieldValue {
		chromePlatform.Description = ""
	} else {
		chromePlatform.Description = c.description
	}
	if ufsUtil.ContainsAnyStrings(c.tags, utils.ClearFieldValue) {
		chromePlatform.Tags = nil
	} else {
		chromePlatform.Tags = c.tags
	}
}

func (c *updateChromePlatform) validateArgs() error {
	if c.newSpecsFile != "" && c.interactive {
		return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe interactive & JSON mode cannot be specified at the same time.")
	}
	if c.newSpecsFile != "" || c.interactive {
		if c.name != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe interactive/JSON mode is specified. '-name' cannot be specified at the same time.")
		}
		if c.manufacturer != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe interactive/JSON mode is specified. '-manufacturer' cannot be specified at the same time.")
		}
		if len(c.tags) > 0 {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe interactive/JSON mode is specified. '-tag' cannot be specified at the same time.")
		}
		if c.description != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe interactive/JSON mode is specified. '-description' cannot be specified at the same time.")
		}
	}
	if c.newSpecsFile == "" && !c.interactive {
		if c.name == "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\n'-name' is required, no mode ('-f' or '-i') is specified.")
		}
		if c.manufacturer == "" && c.tags == nil && c.description == "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nNothing to update. Please provide any field to update")
		}
	}
	return nil
}
