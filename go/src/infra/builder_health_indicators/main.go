// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"os"

	"infra/builder_health_indicators/indicators_pb"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/flag/fixflagpos"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
	"go.chromium.org/luci/luciexe/build"
)

type luciexeGenerateRun struct {
	generateRun
}

type generateRun struct {
	subcommands.CommandRunBase
	logCfg gologger.LoggerConfig

	day string
}

func main() {
	cliApp := &cli.Application{
		Name:  "builder-health-indicators",
		Title: "Builder Health Indicators track Chromium builders' long term health",
		Commands: []*subcommands.Command{
			{
				UsageLine: `luciexe`,
				ShortDesc: "Run as a luciexe and do what generate_indicators does",
				LongDesc:  "Run as a luciexe and do what generate_indicators does",
				CommandRun: func() subcommands.CommandRun {
					r := &luciexeGenerateRun{}
					r.logCfg = gologger.LoggerConfig{Out: os.Stderr}
					return r
				},
			},
			{
				UsageLine: `generate_indicators`,
				ShortDesc: "Generate builder health indicators",
				LongDesc:  "Generate builder health indicators",
				CommandRun: func() subcommands.CommandRun {
					r := &generateRun{}
					r.logCfg = gologger.LoggerConfig{Out: os.Stderr}
					r.Flags.StringVar(&r.day, "day", "", "The day to generate for in ISO 8601 (YYYY-MM-DD)")

					return r
				},
			},
		},
	}

	os.Exit(subcommands.Run(cliApp, fixflagpos.FixSubcommands(os.Args[1:])))
}

func (r *generateRun) DualRun(ctx context.Context) error {
	step, ctx := build.StartStep(ctx, "Hello world")
	var err error
	logging.Infof(ctx, "hello world of logging")
	defer func() { step.End(err) }()

	return nil
}

// Called by bb invocation
func (r *luciexeGenerateRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	input := indicators_pb.InputParams{}

	build.Main(&input, nil, nil, func(ctx context.Context, userArgs []string, state *build.State) error {
		r.day = input.Day
		return r.DualRun(ctx)
	})

	return 0
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
