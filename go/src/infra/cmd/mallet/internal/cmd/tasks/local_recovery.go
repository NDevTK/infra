// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tasks

import (
	"context"
	"log"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/grpc/prpc"
	"google.golang.org/grpc/metadata"

	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
	commonFlags "infra/cmd/mallet/internal/cmd/cmdlib"
	"infra/cmd/mallet/internal/site"
	"infra/cmdsupport/cmdlib"
	"infra/cros/recovery"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
)

// LocalRecovery subcommand: Running verify/recovery against the DUT from local environment.
var LocalRecovery = &subcommands.Command{
	UsageLine: "local-recovery UNIT_NAME",
	ShortDesc: "run recovery from local env.",
	LongDesc: `Run recovery against a DUT from local environment.

For now only running in testing mode.`,
	CommandRun: func() subcommands.CommandRun {
		c := &localRecoveryRun{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.CommonFlags.Register(&c.Flags)
		c.envFlags.Register(&c.Flags)
		return c
	},
}

type localRecoveryRun struct {
	subcommands.CommandRunBase
	commonFlags.CommonFlags
	authFlags authcli.Flags
	envFlags  site.EnvFlags
}

// Run initiates execution of local recovery.
func (c *localRecoveryRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

// innerRun executes internal logic of control.
func (c *localRecoveryRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	var unit string
	if len(args) > 1 {
		return errors.New("does not support more than one unit per request")
	} else if len(args) == 1 {
		unit = args[0]
	}
	if unit == "" {
		return errors.New("unit is not specified")
	}
	ctx := cli.GetContext(a, c, env)
	ctx = setupContextNamespace(ctx, ufsUtil.OSNamespace)
	hc, err := cmdlib.NewHTTPClient(ctx, &c.authFlags)
	if err != nil {
		return errors.Annotate(err, "local recovery: create http client").Err()
	}
	e := c.envFlags.Env()
	if c.Verbose() {
		log.Printf("Init TLW with inventory: %s and csa: %s sources", e.UFSService, e.AdminService)
	}
	ic := ufsAPI.NewFleetPRPCClient(&prpc.Client{
		C:       hc,
		Host:    e.UFSService,
		Options: site.UFSPRPCOptions,
	})
	csac := fleet.NewInventoryPRPCClient(
		&prpc.Client{
			C:       hc,
			Host:    e.AdminService,
			Options: site.DefaultPRPCOptions,
		},
	)
	access, err := recovery.NewLocalTLWAccess(ic, csac)
	if err != nil {
		return errors.Annotate(err, "local recovery: create tlw access").Err()
	}
	defer access.Close()
	in := &recovery.Input{
		UnitName: unit,
		Access:   access,
	}
	if err = recovery.Run(ctx, in); err != nil {
		return errors.Annotate(err, "local recovery").Err()
	}
	log.Printf("Task on %q has completed successfully", unit)
	return nil
}

// setupContextNamespace sets namespace to the context for UFS client.
func setupContextNamespace(ctx context.Context, namespace string) context.Context {
	md := metadata.Pairs(ufsUtil.Namespace, namespace)
	return metadata.NewOutgoingContext(ctx, md)
}
