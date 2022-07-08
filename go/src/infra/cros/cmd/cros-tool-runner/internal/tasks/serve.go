// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tasks

import (
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"
	"infra/cros/cmd/cros-tool-runner/internal/v2/server"
)

type runServeCmd struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
	port      int
}

func Serve(authOpts auth.Options) *subcommands.Command {
	const serveDesc = `serve,

	Tool used to start cros-tool-runner v2 services.

	Example:
	cros-tool-runner serve
	`

	c := &runServeCmd{}
	return &subcommands.Command{
		UsageLine: "serve",
		ShortDesc: "serve starts CTRv2 services",
		LongDesc:  serveDesc,
		CommandRun: func() subcommands.CommandRun {
			c.authFlags.Register(&c.Flags, authOpts)
			c.Flags.IntVar(&c.port, "port", 8082, "port number server listens to")
			return c
		},
	}
}

// Run executes the tool.
func (c *runServeCmd) Run(subcommands.Application, []string, subcommands.Env) int {
	return server.StartServer(c.port)
}
