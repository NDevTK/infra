// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"net/http"
	"time"

	"github.com/maruel/subcommands"

	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	prpc "go.chromium.org/luci/grpc/prpc"

	"infra/cmdsupport/cmdlib"
	fleetcostAPI "infra/cros/fleetcost/api/rpc"
	"infra/cros/fleetcost/internal/site"
)

var GetCostIndicatorCommand *subcommands.Command = &subcommands.Command{
	UsageLine: "get-ci [options...]",
	ShortDesc: "get a cost indicator",
	LongDesc:  "Get a cost indicator",
	CommandRun: func() subcommands.CommandRun {
		c := &getCostIndicatorCommand{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.authFlags.RegisterIDTokenFlags(&c.Flags)
		c.commonFlags.Register(&c.Flags)
		c.Flags.StringVar(&c.board, "board", "", "the board to search for")
		return c
	},
}

type getCostIndicatorCommand struct {
	subcommands.CommandRunBase
	authFlags   authcli.Flags
	commonFlags site.CommonFlags
	board       string
}

// Run is the main entrypoint to the get-ci.
func (c *getCostIndicatorCommand) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := cli.GetContext(a, c, env)
	if err := c.innerRun(ctx, a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *getCostIndicatorCommand) innerRun(ctx context.Context, a subcommands.Application, args []string, env subcommands.Env) error {
	host, err := c.commonFlags.Host()
	if err != nil {
		return err
	}
	var httpClient *http.Client
	if !c.commonFlags.HTTP() {
		var err error
		httpClient, err = getSecureClient(ctx, host, c.authFlags)
		if err != nil {
			return err
		}
	}
	prpcClient := &prpc.Client{
		C:    httpClient,
		Host: host,
		Options: &prpc.Options{
			Insecure:      c.commonFlags.HTTP(),
			PerRPCTimeout: 30 * time.Second,
		},
	}
	fleetCostClient := fleetcostAPI.NewFleetCostPRPCClient(prpcClient)
	resp, err := fleetCostClient.ListCostIndicators(
		ctx,
		&fleetcostAPI.ListCostIndicatorsRequest{
			Filter: &fleetcostAPI.ListCostIndicatorsFilter{
				Board: c.board,
			},
		},
	)
	if err != nil {
		return err
	}
	if _, err := showProto(a.GetOut(), resp); err != nil {
		return err
	}
	return nil
}
