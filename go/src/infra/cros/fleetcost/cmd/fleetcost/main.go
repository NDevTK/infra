// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"os"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/logging/gologger"
)

// getApplication returns the fleetcost command line application.
func getApplication() *cli.Application {
	return &cli.Application{
		Name:  "fleet cost",
		Title: "fleet cost command line tool",
		Context: func(ctx context.Context) context.Context {
			return gologger.StdConfig.Use(ctx)
		},
		Commands: []*subcommands.Command{
			subcommands.CmdHelp,
		},
	}
}

// main is the entrypoint to the fleet command line application.
func main() {
	os.Exit(subcommands.Run(getApplication(), nil))
}
