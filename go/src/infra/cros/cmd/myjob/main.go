// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package main implements the `cros myjob` CLI, which enables users to
// run ChromeOS builders with certain common configurations.
package main

import (
	"os"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"
)

func newApplication() *cli.Application {
	return &cli.Application{
		Name:  "myjob",
		Title: "cros myjob CLI",
		Commands: []*subcommands.Command{
			subcommands.CmdHelp,
		},
	}
}

// Main is the main entrypoint to the application.
func main() {
	os.Exit(subcommands.Run(newApplication(), nil))
}
