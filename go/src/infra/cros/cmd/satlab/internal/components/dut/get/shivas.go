// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package get

import (
	"infra/cmd/shivas/cmdhelp"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/flag"

	"infra/cros/cmd/satlab/internal/site"
)

// ShivasGetDUT contains the arguments that can be used to get DUTs.
// It is inherited from shivas.
type shivasGetDUT struct {
	subcommands.CommandRunBase

	authFlags   authcli.Flags
	envFlags    site.EnvFlags
	outputFlags site.OutputFlags
	commonFlags site.CommonFlags

	// Filters
	zones      []string
	racks      []string
	machines   []string
	prototypes []string
	tags       []string
	states     []string
	servos     []string
	servotypes []string
	switches   []string
	rpms       []string
	pools      []string

	pageSize          int
	keysOnly          bool
	wantHostInfoStore bool
}

// MakeDefaultShivasCommand produces an instance of getDUT with the same default values that would
// be used for shivas.
func makeDefaultShivasCommand() *getDUT {
	return &getDUT{}
}

// RegisterShivasFlags registers the flags inherited from shivas.
func registerShivasFlags(c *getDUT) {
	c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
	c.envFlags.Register(&c.Flags)
	c.outputFlags.Register(&c.Flags)
	c.commonFlags.Register(&c.Flags)

	c.Flags.IntVar(&c.pageSize, "n", 0, cmdhelp.ListPageSizeDesc)
	c.Flags.BoolVar(&c.keysOnly, "keys", false, cmdhelp.KeysOnlyText)

	c.Flags.Var(flag.StringSlice(&c.zones), "zone", "Name(s) of a zone to filter by. Can be specified multiple times."+cmdhelp.ZoneFilterHelpText)
	c.Flags.Var(flag.StringSlice(&c.racks), "rack", "Name(s) of a rack to filter by. Can be specified multiple times.")
	c.Flags.Var(flag.StringSlice(&c.machines), "machine", "Name(s) of a machine/asset to filter by. Can be specified multiple times.")
	c.Flags.Var(flag.StringSlice(&c.prototypes), "prototype", "Name(s) of a host prototype to filter by. Can be specified multiple times.")
	c.Flags.Var(flag.StringSlice(&c.tags), "tag", "Name(s) of a tag to filter by. Can be specified multiple times.")
	c.Flags.Var(flag.StringSlice(&c.states), "state", "Name(s) of a state to filter by. Can be specified multiple times."+cmdhelp.StateFilterHelpText)
	c.Flags.Var(flag.StringSlice(&c.servos), "servo", "Name(s) of a servo:port to filter by. Can be specified multiple times.")
	c.Flags.Var(flag.StringSlice(&c.servotypes), "servotype", "Name(s) of a servo type to filter by. Can be specified multiple times.")
	c.Flags.Var(flag.StringSlice(&c.switches), "switch", "Name(s) of a switch to filter by. Can be specified multiple times.")
	c.Flags.Var(flag.StringSlice(&c.rpms), "rpm", "Name(s) of a rpm to filter by. Can be specified multiple times.")
	c.Flags.Var(flag.StringSlice(&c.pools), "pools", "Name(s) of a tag to filter by. Can be specified multiple times.")
	c.Flags.BoolVar(&c.wantHostInfoStore, "host-info-store", false, "write host info store to stdout")
}
