// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"
	"net/http"
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
		c.commonFlags.Register(&c.Flags)
		return c
	},
}

type pingCommand struct {
	subcommands.CommandRunBase
	authFlags   authcli.Flags
	commonFlags site.CommonFlags
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

func (c *pingCommand) getSecureClient(ctx context.Context) (*http.Client, error) {
	authOptions, err := c.authFlags.Options()
	if err != nil {
		return nil, errors.Annotate(err, "ping").Err()
	}
	authenticator := auth.NewAuthenticator(ctx, auth.InteractiveLogin, authOptions)
	httpClient, err := authenticator.Client()
	if err != nil {
		return nil, errors.Annotate(err, "ping").Err()
	}
	return httpClient, nil
}

func (c *pingCommand) innerRun(ctx context.Context, a subcommands.Application, args []string, env subcommands.Env) error {
	var httpClient *http.Client
	if !c.commonFlags.HTTP() {
		var err error
		httpClient, err = c.getSecureClient(ctx)
		if err != nil {
			return err
		}
	}
	prpcClient := &prpc.Client{
		C:    httpClient,
		Host: "127.0.0.1:8800",
		Options: &prpc.Options{
			Insecure:      c.commonFlags.HTTP(),
			PerRPCTimeout: 30 * time.Second,
		},
	}
	fleetCostClient := fleetcostpb.NewFleetCostPRPCClient(prpcClient)
	resp, err := fleetCostClient.Ping(ctx, &fleetcostpb.PingRequest{})
	fmt.Printf("%#v\n", resp)
	return err
}
