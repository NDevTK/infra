// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package generate

import (
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth"
)

type generateRun struct {
	subcommands.CommandRunBase

	authOpt *auth.Options
}

func CmdGenerate(authOpt *auth.Options) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: `generate`,
		ShortDesc: "Generate builder health indicators for the previous day",
		LongDesc:  "Generate builder health indicators for the previous day",
		CommandRun: func() subcommands.CommandRun {
			r := &generateRun{authOpt: authOpt}

			return r
		},
	}
}

func (r *generateRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	println("hello world")

	return 0
}
