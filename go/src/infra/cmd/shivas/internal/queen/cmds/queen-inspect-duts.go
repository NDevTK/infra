// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package queen

import (
	"bufio"
	"fmt"
	"text/tabwriter"
	"text/template"

	"github.com/maruel/subcommands"

	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/grpc/prpc"

	"infra/appengine/drone-queen/api"
	"infra/cmd/shivas/site"
	"infra/cmdsupport/cmdlib"
)

// InspectDuts subcommand: Inspect drone queen DUT info.
var InspectDuts = &subcommands.Command{
	UsageLine: "queen-inspect-duts",
	ShortDesc: "inspect drone queen DUT info",
	LongDesc: `Inspect drone queen DUT info.

This command is for developer inspection and debugging of drone queen state.
Do not use this command as part of scripts or pipelines.
This command is unstable.

You must be in the respective inspectors group to use this.`,
	CommandRun: func() subcommands.CommandRun {
		c := &inspectDutsRun{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		return c
	},
}

type inspectDutsRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
	envFlags  site.EnvFlags
}

func (c *inspectDutsRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, errors.Annotate(err, "queen-inspect-duts").Err())
		return 1
	}
	return 0
}

func (c *inspectDutsRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	ctx := cli.GetContext(a, c, env)
	hc, err := cmdlib.NewHTTPClient(ctx, &c.authFlags)
	if err != nil {
		return err
	}
	e := c.envFlags.Env()
	ic := api.NewInspectPRPCClient(&prpc.Client{
		C:       hc,
		Host:    e.QueenService,
		Options: site.DefaultPRPCOptions,
	})

	res, err := ic.ListDuts(ctx, &api.ListDutsRequest{})
	if err != nil {
		return err
	}

	bw := bufio.NewWriter(a.GetOut())
	defer bw.Flush()
	tw := tabwriter.NewWriter(bw, 0, 2, 2, ' ', 0)
	defer tw.Flush()
	fmt.Fprintf(tw, "DUT\tHive\tDrone\tDraining\t\n")
	t := template.Must(template.New("output").Parse("{{range .}}{{.GetId}}\t{{.GetHive}}\t{{.GetAssignedDrone}}\t{{.GetDraining}} \n{{end}}"))
	t.Execute(tw, res.GetDuts())
	return nil
}
