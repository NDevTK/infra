// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"os"
	"time"

	"infra/cr_builder_health/healthpb"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/flag/fixflagpos"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
	"go.chromium.org/luci/luciexe/build"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type luciexeGenerateRun struct {
	generateRun
}

type generateRun struct {
	subcommands.CommandRunBase
	logCfg gologger.LoggerConfig

	dateString string // cmdline input
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
				LongDesc: `Takes metrics from cr-buildbucket BigQuery tables and processes them into health indicators stored back in cr-builder-health-indicators tables

				Also sends rpcs to Buildbucket to update live Builder Metadata, which is shown in Milo builder pages and consoles.

				Required ACLs: BigQuery read and write permissions in cr-builder-health-indicators.
				`,
				CommandRun: func() subcommands.CommandRun {
					r := &generateRun{}
					r.logCfg = gologger.LoggerConfig{Out: os.Stderr}
					r.Flags.StringVar(&r.dateString, "date", "", "The date to generate for in ISO 8601 (YYYY-MM-DD). The default date is yesterday.")

					return r
				},
			},
		},
	}

	os.Exit(subcommands.Run(cliApp, fixflagpos.FixSubcommands(os.Args[1:])))
}

func Run(ctx context.Context, input *healthpb.InputParams) error {
	var err error
	step, ctx := build.StartStep(ctx, "Hello world")
	defer func() { step.End(err) }()
	logging.Infof(ctx, "date is %+v", input.Date)

	return nil
}

// Called by bb invocation
func (r *luciexeGenerateRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	input := healthpb.InputParams{}

	build.Main(&input, nil, nil, func(ctx context.Context, userArgs []string, state *build.State) error {
		return Run(ctx, &input)
	})

	return 0
}

func (r *generateRun) ParseFlags(ctx context.Context) (*healthpb.InputParams, error) {
	input := &healthpb.InputParams{}
	if r.dateString == "" {
		// The default date is yesterday
		yesterday := time.Now().Add(-24 * time.Hour)
		input.Date = timestamppb.New(yesterday)
		return input, nil
	}

	iso8601Format := "2006-01-02"
	t, err := time.Parse(iso8601Format, r.dateString)
	if err != nil {
		return input, errors.New("Error parsing -date flag. Please specify date like YYYY-MM-DD")
	}
	input.Date = timestamppb.New(t)

	return input, nil
}

// Called by cmdline invocation
func (r *generateRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	// Setup
	ctx := r.logCfg.Use(cli.GetContext(a, r, env))
	input, err := r.ParseFlags(ctx)
	if err != nil {
		logging.Errorf(ctx, "Error parsing flags: %v", err)
		return 1
	}

	// Run
	err = Run(ctx, input)
	if err != nil {
		logging.Errorf(ctx, "Error in Run: %v", err)
		return 1
	}

	return 0
}
