// Copyright 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package experimental

import (
	"fmt"

	"github.com/maruel/subcommands"

	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/errors"

	"infra/cmd/shivas/site"
)

// GetDutsForLabstationCmd gets the DUTs associated with a labstation.
var GetDutsForLabstationCmd = &subcommands.Command{
	UsageLine: "get-duts-for-labstation [duts]",
	ShortDesc: "get the DUTs attached to a labstation",
	LongDesc: `get duts for labstation gets the DUTs for a labstation.

	./shivas get-duts-for-labstation [duts]`,
	CommandRun: func() subcommands.CommandRun {
		c := &getDutsForLabstationRun{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		return c
	},
}

type getDutsForLabstationRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
	envFlags  site.EnvFlags
}

func (c *getDutsForLabstationRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		fmt.Fprintf(a.GetErr(), "%s: %s\n", a.GetName(), err)
		return 1
	}
	return 0
}

func (c *getDutsForLabstationRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) (err error) {
	return errors.New("not yet implemented")
}
