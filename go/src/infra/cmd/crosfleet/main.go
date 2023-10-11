// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Command crosfleet is a Chrome OS hardware test lab tool.
package main

import (
	"context"
	"os"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/logging/gologger"

	"infra/cmd/crosfleet/internal/dut"
	"infra/cmd/crosfleet/internal/meta"
	"infra/cmd/crosfleet/internal/run"
	"infra/cmd/crosfleet/internal/site"
	"infra/cmd/crosfleet/internal/vm"
)

var application = &cli.Application{
	Name: "crosfleet",
	Title: `Universal tool for accessing ChromeOS hardware labs

Tool uses a default RPC retry strategy with five attempts and exponential backoff.
Full documentation http://go/crosfleet-cli`,
	Context: func(ctx context.Context) context.Context {
		return gologger.StdConfig.Use(ctx)
	},
	Commands: []*subcommands.Command{
		subcommands.CmdHelp,
		meta.Update,
		meta.Version,
		subcommands.Section("Authentication"),
		authcli.SubcommandLogin(site.DefaultAuthOptions, "login", false),
		authcli.SubcommandLogout(site.DefaultAuthOptions, "logout", false),
		authcli.SubcommandInfo(site.DefaultAuthOptions, "whoami", false),
		subcommands.Section("Tasks"),
		dut.CmdDut,
		run.CmdRun,
		vm.CmdVm,
		subcommands.CmdHelp,
	},
}

func updateThenRun() int {
	// Try to update crosfleet, but don't block users from running their
	// commands if the update fails for some transient reason.
	_ = subcommands.Run(application, []string{"update", "-silent"})
	return subcommands.Run(application, nil)
}

func main() {
	os.Exit(updateThenRun())
}
