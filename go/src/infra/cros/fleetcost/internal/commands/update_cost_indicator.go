// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/maruel/subcommands"

	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	prpc "go.chromium.org/luci/grpc/prpc"

	"infra/cmdsupport/cmdlib"
	fleetcostpb "infra/cros/fleetcost/api"
	"infra/cros/fleetcost/internal/site"
)

var UpdateCostIndicatorCommand *subcommands.Command = &subcommands.Command{
	UsageLine: "update-ci [options...]",
	ShortDesc: "update a cost indicator",
	LongDesc:  "Update a cost indicator",
	CommandRun: func() subcommands.CommandRun {
		c := &updateCostIndicatorCommand{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.authFlags.RegisterIDTokenFlags(&c.Flags)
		c.commonFlags.Register(&c.Flags)
		return c
	},
}

type updateCostIndicatorCommand struct {
	subcommands.CommandRunBase
	authFlags   authcli.Flags
	commonFlags site.CommonFlags
}

// Run is the main entrypoint to update-ci.
func (c *updateCostIndicatorCommand) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := cli.GetContext(a, c, env)
	if err := c.innerRun(ctx, a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *updateCostIndicatorCommand) innerRun(ctx context.Context, a subcommands.Application, args []string, env subcommands.Env) error {
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
	fleetCostClient := fleetcostpb.NewFleetCostPRPCClient(prpcClient)
	resp, err := fleetCostClient.ListCostIndicators(ctx, &fleetcostpb.ListCostIndicatorsRequest{})
	fmt.Printf("%#v\n", resp)
	return err
}
