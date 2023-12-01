// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"time"

	"github.com/maruel/subcommands"

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"
	prpc "go.chromium.org/luci/grpc/prpc"

	"infra/cmdsupport/cmdlib"
	fleetcostpb "infra/cros/fleetcost/api"
	"infra/cros/fleetcost/internal/site"
)

// PingCommand pings the service.
var PingCommand *subcommands.Command = &subcommands.Command{
	UsageLine: "ping [options...]",
	ShortDesc: "ping a fleet cost instance",
	LongDesc:  "Ping a fleet cost instance",
	CommandRun: func() subcommands.CommandRun {
		c := &pingCommand{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		return c
	},
}

type pingCommand struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
}

// Run is the main entrypoint to the ping.
func (c *pingCommand) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := cli.GetContext(a, c, env)
	if err := c.innerRun(ctx, a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *pingCommand) innerRun(ctx context.Context, a subcommands.Application, args []string, env subcommands.Env) error {
	authOptions, err := c.authFlags.Options()
	if err != nil {
		return errors.Annotate(err, "ping").Err()
	}
	authenticator := auth.NewAuthenticator(ctx, auth.InteractiveLogin, authOptions)
	httpClient, err := authenticator.Client()
	if err != nil {
		return errors.Annotate(err, "ping").Err()
	}
	prpcClient := &prpc.Client{
		C:    httpClient,
		Host: "127.0.0.1:8800",
		Options: &prpc.Options{
			PerRPCTimeout: 30 * time.Second,
		},
	}
	fleetCostClient := fleetcostpb.NewFleetCostPRPCClient(prpcClient)
	_, err = fleetCostClient.Ping(ctx, &fleetcostpb.PingRequest{})
	return err
}
