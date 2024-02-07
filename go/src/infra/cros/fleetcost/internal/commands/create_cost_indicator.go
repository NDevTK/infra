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

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"
	prpc "go.chromium.org/luci/grpc/prpc"

	"infra/cmdsupport/cmdlib"
	fleetcostpb "infra/cros/fleetcost/api"
	"infra/cros/fleetcost/internal/site"
)

var CreateCostIndicatorCommand *subcommands.Command = &subcommands.Command{
	UsageLine: "create-ci [options...]",
	ShortDesc: "create a cost indicator",
	LongDesc:  "Create a cost indicator",
	CommandRun: func() subcommands.CommandRun {
		c := &createCostIndicatorCommand{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.authFlags.RegisterIDTokenFlags(&c.Flags)
		c.commonFlags.Register(&c.Flags)
		c.Flags.StringVar(&c.name, "name", "", "name of cost indicator")
		return c
	},
}

type createCostIndicatorCommand struct {
	subcommands.CommandRunBase
	authFlags   authcli.Flags
	commonFlags site.CommonFlags
	name        string
}

// Run is the main entrypoint to the create-ci.
func (c *createCostIndicatorCommand) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := cli.GetContext(a, c, env)
	if err := c.innerRun(ctx, a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *createCostIndicatorCommand) getSecureClient(ctx context.Context, host string) (*http.Client, error) {
	authOptions, err := c.authFlags.Options()
	if err != nil {
		return nil, errors.Annotate(err, "create-ci").Err()
	}
	if authOptions.UseIDTokens && authOptions.Audience == "" {
		authOptions.Audience = "https://" + host
	}
	authenticator := auth.NewAuthenticator(ctx, auth.InteractiveLogin, authOptions)
	httpClient, err := authenticator.Client()
	if err != nil {
		return nil, errors.Annotate(err, "create-ci").Err()
	}
	return httpClient, nil
}

func (c *createCostIndicatorCommand) innerRun(ctx context.Context, a subcommands.Application, args []string, env subcommands.Env) error {
	host, err := c.commonFlags.Host()
	if err != nil {
		return err
	}
	var httpClient *http.Client
	if !c.commonFlags.HTTP() {
		var err error
		httpClient, err = c.getSecureClient(ctx, host)
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
	resp, err := fleetCostClient.CreateCostIndicator(ctx, &fleetcostpb.CreateCostIndicatorRequest{CostIndicator: &fleetcostpb.CostIndicator{
		Name: c.name,
	}})
	fmt.Printf("%#v\n", resp)
	return err
}
