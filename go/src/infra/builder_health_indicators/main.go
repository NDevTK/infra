// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"os"

	"infra/builder_health_indicators/indicators_pb"

	//"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/flag/fixflagpos"
	"go.chromium.org/luci/common/logging/gologger"
)

type generateRun struct {
	subcommands.CommandRunBase
	Flags  flag.FlagSet
	logCfg gologger.LoggerConfig

	luciexe bool
	day     string
}

func main() {
	r := &generateRun{}
	r.logCfg = gologger.LoggerConfig{Out: os.Stderr}
	r.Flags.BoolVar(&r.luciexe, "luciexe", false, "Tells us to act as a luciexe app")
	r.Flags.Parse(os.Args[1:])

	cliApp := &cli.Application{
		Name:  "builder-health-indicators",
		Title: "Builder Health Indicators track Chromium builders' long term health",
		Commands: []*subcommands.Command{
			{
				UsageLine: `generate_indicators`,
				ShortDesc: "Generate builder health indicators for the previous day",
				LongDesc:  "Generate builder health indicators for the previous day",
				CommandRun: func() subcommands.CommandRun {
					r.Flags.StringVar(&r.day, "day", "", "The day to generate for in ISO 8601 (YYYY-MM-DD)")

					return r
				},
			},
		},
	}

	if !r.luciexe {
		// Command line invocation
		os.Exit(subcommands.Run(cliApp, fixflagpos.FixSubcommands(os.Args[1:])))
	}

	// LUCIExe invocation
	request := indicators_pb.BBGenerateRequest{}

	build.Main(&request, nil, nil, func(ctx context.Context, userArgs []string, state *build.State) error {
		r := &generateRun{}
		return r.BBRun(ctx)
	})
}

func (r *generateRun) DualRun(ctx context.Context) error {
	step, ctx := build.StartStep(ctx, "Hello world")
	var err error
	logging.Infof(ctx, "hello world of logging")
	defer func() { step.End(err) }()

	return nil
}

// Called by a BB invocation
func (r *generateRun) BBRun(ctx context.Context) error {
	return r.DualRun(ctx)
}

// Called by cmdline invocation
func (r *generateRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := r.logCfg.Use(cli.GetContext(a, r, env))

	err := r.DualRun(ctx)
	if err != nil {
		return 1
	}

	return 0
}
