// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Satlab is a wrapper around shivas.

package main

import (
	"context"
	"os"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/data/rand/mathrand"

	"infra/cros/cmd/satlab/internal/subcmds"
)

func getApplication() *cli.Application {
	return &cli.Application{
		Name:  "satlab",
		Title: `Satlab DUT Management Tool`,
		Context: func(ctx context.Context) context.Context {
			return ctx
		},
		Commands: []*subcommands.Command{
			subcommands.CmdHelp,
			subcommands.Section("Meta"),
			// TODO(gregorynisbet): Add commands for version and update.
			subcommands.Section("Authentication"),
			// TODO(gregorynisbet): Add commands for auth.
			subcmds.AddCmd,
			subcmds.DeleteCmd,
			subcmds.GetCmd,
		},
	}
}

func main() {
	mathrand.SeedRandomly()
	os.Exit(subcommands.Run(getApplication(), nil))
}
