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
	"go.chromium.org/luci/common/errors"
	prpc "go.chromium.org/luci/grpc/prpc"

	"infra/cmdsupport/cmdlib"
	fleetcostAPI "infra/cros/fleetcost/api/rpc"
	"infra/cros/fleetcost/internal/site"
)

// PersistToBigqueryCommand persists to BigQuery.
var PersistToBigqueryCommand *subcommands.Command = &subcommands.Command{
	UsageLine: "persist-to-bigquery [options...]",
	ShortDesc: "Persist to bigquery",
	LongDesc:  "Persist to bigquery",
	CommandRun: func() subcommands.CommandRun {
		c := &persistToBigqueryCommand{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.authFlags.RegisterIDTokenFlags(&c.Flags)
		c.commonFlags.Register(&c.Flags)
		return c
	},
}

type persistToBigqueryCommand struct {
	subcommands.CommandRunBase
	authFlags   authcli.Flags
	commonFlags site.CommonFlags
}

// Run is the main entrypoint for calling the PersistToBigqueryCommand RPC.
func (c *persistToBigqueryCommand) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := cli.GetContext(a, c, env)
	if err := c.innerRun(ctx, a); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *persistToBigqueryCommand) innerRun(ctx context.Context, a subcommands.Application) error {
	host, err := c.commonFlags.Host()
	if err != nil {
		return errors.Annotate(err, "persist to bigquery").Err()
	}
	var httpClient *http.Client
	if !c.commonFlags.HTTP() {
		var err error
		httpClient, err = getSecureClient(ctx, host, c.authFlags)
		if err != nil {
			return errors.Annotate(err, "persist to bigquery").Err()
		}
	}
	prpcClient := &prpc.Client{
		C:    httpClient,
		Host: host,
		Options: &prpc.Options{
			Insecure:      c.commonFlags.HTTP(),
			PerRPCTimeout: 10 * time.Minute,
		},
	}
	fleetCostClient := fleetcostAPI.NewFleetCostPRPCClient(prpcClient)
	request := &fleetcostAPI.PersistToBigqueryRequest{
		Readonly: true,
	}
	resp, err := fleetCostClient.PersistToBigquery(ctx, request)
	if err != nil {
		return errors.Annotate(err, "persist to bigquery").Err()
	}
	_, err = showProto(a.GetOut(), resp)
	return errors.Annotate(err, "persist to bigquery").Err()
}
