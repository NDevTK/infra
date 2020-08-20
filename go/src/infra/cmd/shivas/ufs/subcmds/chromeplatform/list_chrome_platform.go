// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package chromeplatform

import (
	"context"
	"fmt"
	"strings"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/grpc/prpc"

	"infra/cmd/shivas/cmdhelp"
	"infra/cmd/shivas/site"
	"infra/cmd/shivas/utils"
	"infra/cmdsupport/cmdlib"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
)

// ListChromePlatformCmd list all Platforms.
var ListChromePlatformCmd = &subcommands.Command{
	UsageLine: "platform",
	ShortDesc: "List all platforms",
	LongDesc:  cmdhelp.ListChromePlatformLongDesc,
	CommandRun: func() subcommands.CommandRun {
		c := &listChromePlatform{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.outputFlags.Register(&c.Flags)
		c.Flags.IntVar(&c.pageSize, "n", 0, cmdhelp.ListPageSizeDesc)
		c.Flags.StringVar(&c.filter, "filter", "", cmdhelp.ChromePlatformFilterHelp)
		c.Flags.BoolVar(&c.keysOnly, "keys", false, cmdhelp.KeysOnlyText)
		return c
	},
}

type listChromePlatform struct {
	subcommands.CommandRunBase
	authFlags   authcli.Flags
	envFlags    site.EnvFlags
	commonFlags site.CommonFlags
	outputFlags site.OutputFlags
	pageSize    int
	filter      string
	keysOnly    bool
}

func (c *listChromePlatform) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *listChromePlatform) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if err := c.validateArgs(); err != nil {
		return err
	}
	ctx := cli.GetContext(a, c, env)
	hc, err := cmdlib.NewHTTPClient(ctx, &c.authFlags)
	if err != nil {
		return err
	}
	e := c.envFlags.Env()
	if c.commonFlags.Verbose() {
		fmt.Printf("Using UnifiedFleet service %s\n", e.UnifiedFleetService)
	}
	ic := ufsAPI.NewFleetPRPCClient(&prpc.Client{
		C:       hc,
		Host:    e.UnifiedFleetService,
		Options: site.DefaultPRPCOptions,
	})
	if c.outputFlags.JSON() {
		return utils.PrintListJSONFormat(ctx, ic, printChromePlatforms, true, int32(c.pageSize), c.filter, c.keysOnly, !c.outputFlags.NoEmit())
	}
	return utils.PrintListTableFormat(ctx, ic, printChromePlatforms, false, int32(c.pageSize), c.filter, c.keysOnly, utils.ChromePlatformTitle, c.outputFlags.Tsv())
}

func printChromePlatforms(ctx context.Context, ic ufsAPI.FleetClient, json bool, pageSize int32, pageToken, filter string, keysOnly, tsv, emit bool) (string, error) {
	req := &ufsAPI.ListChromePlatformsRequest{
		PageSize:  pageSize,
		PageToken: pageToken,
		Filter:    filter,
		KeysOnly:  keysOnly,
	}
	res, err := ic.ListChromePlatforms(ctx, req)
	if err != nil {
		return "", err
	}
	if json {
		utils.PrintChromePlatformsJSON(res.ChromePlatforms, emit)
	} else if tsv {
		utils.PrintTSVPlatforms(res.ChromePlatforms, keysOnly)
	} else {
		utils.PrintChromePlatforms(res.ChromePlatforms, keysOnly)
	}
	return res.GetNextPageToken(), nil
}

func (c *listChromePlatform) validateArgs() error {
	if c.filter != "" {
		filter := fmt.Sprintf(strings.Replace(c.filter, " ", "", -1))
		if !ufsAPI.FilterRegex.MatchString(filter) {
			return cmdlib.NewUsageError(c.Flags, ufsAPI.InvalidFilterFormat)
		}
	}
	return nil
}
