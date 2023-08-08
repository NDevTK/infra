// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package meta

import (
	"context"
	"fmt"
	"time"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"

	"infra/cmd/shivas/site"
	"infra/libs/cipd"
)

// Version subcommand: Version shivas.
var Version = &subcommands.Command{
	UsageLine: "version",
	ShortDesc: "print shivas version",
	LongDesc:  "Print shivas version.",
	CommandRun: func() subcommands.CommandRun {
		c := &versionRun{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)

		c.Flags.BoolVar(&c.short, "short", false, "if printing a short version of shivas")
		return c
	},
}

type versionRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags

	short bool
}

func (c *versionRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		fmt.Fprintf(a.GetErr(), "%s: %s\n", a.GetName(), err)
		return 1
	}
	return 0
}

func (c *versionRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if c.short {
		fmt.Printf("shivas %s\n", site.VersionNumber)
		return nil
	}

	p, err := cipd.FindPackage("shivas", site.CipdInstalledPath)
	if err != nil {
		return err
	}
	ctx := context.Background()
	d, err := cipd.DescribePackage(ctx, p.Package, p.Pin.InstanceID)
	if err != nil {
		return err
	}

	fmt.Printf("shivas CLI tool: v%s+%s\n", site.VersionNumber, time.Time(d.RegisteredTs).Format("20060102150405"))
	fmt.Printf("CIPD Package:\t%s\n", p.Package)
	fmt.Printf("CIPD Version:\t%s\n", p.Pin.InstanceID)
	fmt.Printf("CIPD Updated:\t%s\n", d.RegisteredTs)
	fmt.Printf("CIPD Tracking:\t%s\n", p.Tracking)

	return nil
}
