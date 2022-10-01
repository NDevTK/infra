// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package main implements the `cros try` CLI, which enables users to
// run ChromeOS builders with certain common configurations.
package main

import (
	"os"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"
)

func newApplication() *cli.Application {
	return &cli.Application{
		Name:  "try",
		Title: "cros try CLI",
		Commands: []*subcommands.Command{
			subcommands.CmdHelp,
			getCmdRelease(),
		},
	}
}

// Main is the main entrypoint to the application.
func main() {
	os.Exit(subcommands.Run(newApplication(), nil))
}
