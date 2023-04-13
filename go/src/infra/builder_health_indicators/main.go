// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"os"

	"infra/builder_health_indicators/cmd/generate_indicators"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/flag/fixflagpos"
	"go.chromium.org/luci/common/logging/gologger"
)

var logCfg = gologger.LoggerConfig{
	Format: `%{message}`,
	Out:    os.Stderr,
}

func main() {
	app := &cli.Application{
		Name:  "builder-health-indicators",
		Title: "Builder Health Indicators track Chromium builders' long term health",
		Context: func(ctx context.Context) context.Context {
			return logCfg.Use(ctx)
		},
		Commands: []*subcommands.Command{
			generate_indicators.CmdGenerate(),
		},
	}

	os.Exit(subcommands.Run(app, fixflagpos.FixSubcommands(os.Args[1:])))
}
