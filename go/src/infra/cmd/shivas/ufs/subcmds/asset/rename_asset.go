// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asset

import (
	"fmt"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/grpc/prpc"

	"infra/cmd/shivas/site"
	"infra/cmd/shivas/utils"
	"infra/cmdsupport/cmdlib"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
)

// RenameAssetCmd rename asset by given name.
var RenameAssetCmd = &subcommands.Command{
	UsageLine: "asset ...",
	ShortDesc: "Rename asset with new name",
	LongDesc: `Rename asset with new name.

Example:

shivas rename asset -name {oldName} -new-name {newName}

Renames the asset and prints the output in the user-specified format.

WARNING: Ensure that the asset is in required state after renaming it.
`,
	CommandRun: func() subcommands.CommandRun {
		c := &renameAsset{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.commonFlags.Register(&c.Flags)
		c.outputFlags.Register(&c.Flags)

		c.Flags.StringVar(&c.name, "name", "", "the name of the asset to rename")
		c.Flags.StringVar(&c.newName, "new-name", "", "the new name of the asset")
		return c
	},
}

type renameAsset struct {
	subcommands.CommandRunBase
	authFlags   authcli.Flags
	envFlags    site.EnvFlags
	commonFlags site.CommonFlags
	outputFlags site.OutputFlags

	name    string
	newName string
}

func (c *renameAsset) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *renameAsset) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if err := c.validateArgs(); err != nil {
		return err
	}
	ctx := cli.GetContext(a, c, env)
	ns, err := c.envFlags.Namespace()
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
	if _, err := utils.PrintExistingAsset(ctx, ic, c.name); err != nil {
		return err
	}
	_, err = ic.RenameAsset(ctx, &ufsAPI.RenameAssetRequest{
		Name:    ufsUtil.AddPrefix(ufsUtil.AssetCollection, c.name),
		NewName: ufsUtil.AddPrefix(ufsUtil.AssetCollection, c.newName),
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *renameAsset) validateArgs() error {
	if c.name == "" {
		return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\n'-name' is required")
	}
	if c.newName == "" {
		return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\n'-new-name' is required")
	}
	return nil
}
