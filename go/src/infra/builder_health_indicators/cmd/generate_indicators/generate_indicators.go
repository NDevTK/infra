// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package generate_indicators

import (
	"fmt"
	"github.com/maruel/subcommands"
)

type generateRun struct {
	subcommands.CommandRunBase
}

func CmdGenerate() *subcommands.Command {
	return &subcommands.Command{
		UsageLine: `generate_indicators`,
		ShortDesc: "Generate builder health indicators for the previous day",
		LongDesc:  "Generate builder health indicators for the previous day",
		CommandRun: func() subcommands.CommandRun {
			r := &generateRun{}

			return r
		},
	}
}

func (r *generateRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	fmt.Println("hello world")

	return 0
}
