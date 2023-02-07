// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package asset

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/golang/protobuf/proto"
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

// DeleteAssetCmd delete a asset on a machine.
var DeleteAssetCmd = &subcommands.Command{
	UsageLine: "asset {assetname}...",
	ShortDesc: "Delete an asset(Chromebook, Servo, Labstation)",
	LongDesc: `Delete an asset.

Example:
shivas delete asset {assetname}

shivas delete asset {assetname1} {assetname2}

shivas delete asset -scan

Deletes the Asset(s).`,
	CommandRun: func() subcommands.CommandRun {
		c := &deleteAsset{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.Flags.BoolVar(&c.scan, "scan", false, "Use barcode scanner to delete multiple assets.")
		c.Flags.BoolVar(&c.skipYes, "yes", true, "Skip yes option by saying yes.")
		return c
	},
}

type deleteAsset struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
	envFlags  site.EnvFlags
	scan      bool
	skipYes   bool
}

func (c *deleteAsset) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *deleteAsset) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if err := c.validateArgs(); err != nil {
		return err
	}
	ctx := cli.GetContext(a, c, env)
	ns, err := c.getNamespace()
	if err != nil {
		return err
	}
	ctx = utils.SetupContext(ctx, ns)
	hc, err := cmdlib.NewHTTPClient(ctx, &c.authFlags)
	if err != nil {
		return err
	}
	e := c.envFlags.Env()
	ic := ufsAPI.NewFleetPRPCClient(&prpc.Client{
		C:       hc,
		Host:    e.UnifiedFleetService,
		Options: site.DefaultPRPCOptions,
	})
	if c.scan {
		c.scanAndDelete(ctx, ic, a.GetOut(), os.Stdin)
		return nil
	}
	if !c.skipYes {
		prompt := utils.CLIPrompt(a.GetOut(), os.Stdin, false)
		if prompt != nil && !prompt(fmt.Sprintf("Are you sure you want to delete the asset(s): %s", args)) {
			return nil
		}
	}
	assets := utils.ConcurrentGet(ctx, ic, args, c.getSingle)
	fmt.Fprintln(a.GetOut(), "\nAsset(s) before deletion:")
	utils.PrintAssetsJSON(assets, true)
	pass, fail := utils.ConcurrentDelete(ctx, ic, args, c.deleteSingle)
	fmt.Fprintln(a.GetOut(), "\nSuccessfully deleted Asset(s):")
	fmt.Fprintln(a.GetOut(), pass)
	fmt.Fprintln(a.GetOut(), "\nFailed to delete Asset(s):")
	fmt.Fprintln(a.GetOut(), fail)
	return nil
}

// getNamespace returns the namespace used to call UFS with appropriate
// validation and default behavior. It is primarily separated from the main
// function for testing purposes
func (c *deleteAsset) getNamespace() (string, error) {
	return c.envFlags.Namespace(site.OSLikeNamespaces, ufsUtil.OSNamespace)
}

func (c *deleteAsset) getSingle(ctx context.Context, ic ufsAPI.FleetClient, name string) (proto.Message, error) {
	return ic.GetAsset(ctx, &ufsAPI.GetAssetRequest{
		Name: ufsUtil.AddPrefix(ufsUtil.AssetCollection, name),
	})
}

func (c *deleteAsset) deleteSingle(ctx context.Context, ic ufsAPI.FleetClient, name string) error {
	_, err := ic.DeleteAsset(ctx, &ufsAPI.DeleteAssetRequest{
		Name: ufsUtil.AddPrefix(ufsUtil.AssetCollection, name),
	})
	return err
}

func (c *deleteAsset) validateArgs() error {
	if !c.scan && c.Flags.NArg() == 0 {
		return cmdlib.NewUsageError(c.Flags, "Please provide the name(s) of the asset to delete.")
	}
	return nil
}

// scanAndDelete is intended to the used with a scanner for deleting assets.
func (c *deleteAsset) scanAndDelete(ctx context.Context, ic ufsAPI.FleetClient, w io.Writer, r io.Reader) {
	scanner := bufio.NewScanner(r)
	fmt.Fprintf(w, "Connect the barcode scanner to your device.\n")
	fmt.Fprintf(w, "Scan asset tag to delete: ")
	for scanner.Scan() {
		asset := scanner.Text()
		err := c.deleteSingle(ctx, ic, asset)
		if err != nil {
			fmt.Fprintf(w, "Failed to delete %s: %s\n", asset, err.Error())
		} else {
			fmt.Fprintf(w, "Successfully deleted %s \n", asset)
		}
		fmt.Fprintf(w, "\nScan asset tag to delete: ")
	}
}
