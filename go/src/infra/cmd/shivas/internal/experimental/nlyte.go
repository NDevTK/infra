// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package experimental

import (
	"context"
	"infra/cmd/shivas/site"
	"infra/cmdsupport/cmdlib"
	ufspb "infra/unifiedfleet/api/v1/models"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
)

// DumpNlyteCmd dumps updated entries from Nlyte to UFS.
var DumpNlyteCmd = &subcommands.Command{
	UsageLine: "nlyte ...",
	ShortDesc: "Dump nlyte updates",
	LongDesc: `Dump nlyte updates from the provided json file.
Example:
shivas nlyte -f testing.json`,
	CommandRun: func() subcommands.CommandRun {
		c := &dumpNlyte{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.commonFlags.Register(&c.Flags)

		c.Flags.StringVar(&c.updatedEntryFile, "f", "", "Path to a file containing AssetAndHosts specification in JSON format.")

		c.outputFlags.Register(&c.Flags)
		return c
	},
}

type dumpNlyte struct {
	subcommands.CommandRunBase
	authFlags   authcli.Flags
	envFlags    site.EnvFlags
	commonFlags site.CommonFlags

	updatedEntryFile string

	outputFlags site.OutputFlags
}

func (c *dumpNlyte) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *dumpNlyte) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	return nil
}

// addAssetToUFS attempts to add given asset to UFS. Returns updated asset and error if any
func (c *dumpNlyte) addAssetToUFS(ctx context.Context, ic ufsAPI.FleetClient, req *ufsAPI.CreateAssetRequest) (*ufspb.Asset, error) {
	if req.Asset == nil {
		return nil, cmdlib.NewQuietUsageError(c.Flags, "Failed to add asset: Invalid input, Missing asset to add")
	}
	if req.Asset.Location == nil {
		return nil, cmdlib.NewQuietUsageError(c.Flags, "Failed to add asset %s: Invalid input, Missing any location information", req.Asset.GetName())
	}
	if req.Asset.Location.Rack == "" {
		return nil, cmdlib.NewQuietUsageError(c.Flags, "Failed to add asset %s: Invalid input, Missing rack", req.Asset.GetName())
	}
	if req.Asset.Location.Zone == ufspb.Zone_ZONE_UNSPECIFIED {
		return nil, cmdlib.NewQuietUsageError(c.Flags, "Failed to add asset %s: Invalid zone", req.Asset.GetName())
	}
	ufsAsset, err := ic.CreateAsset(ctx, req)
	if ufsAsset != nil {
		// Remove the prefix from the asset returned by UFS
		ufsAsset.Name = ufsUtil.RemovePrefix(ufsAsset.Name)
	}
	return ufsAsset, err
}
