// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/data/rand/mathrand"
	"infra/cmd/labtool/meta"
	"infra/cmd/labtool/site"
	"infra/cmd/labtool/ufs/cmds/resources"
)

func getApplication() *cli.Application {
	return &cli.Application{
		Name:  "labtool",
		Title: "Unified Fleet System Management",
		Context: func(ctx context.Context) context.Context {
			return ctx
		},
		Commands: []*subcommands.Command{
			subcommands.CmdHelp,
			subcommands.Section("Meta"),
			meta.Version,
			meta.Update,
			subcommands.Section("Authentication"),
			authcli.SubcommandInfo(site.DefaultAuthOptions, "whoami", false),
			authcli.SubcommandLogin(site.DefaultAuthOptions, "login", false),
			authcli.SubcommandLogout(site.DefaultAuthOptions, "logout", false),
			subcommands.Section("Registration"),
			resources.SwitchCmd,
		},
	}
}

func main() {
	mathrand.SeedRandomly()
	log.SetOutput(ioutil.Discard)
	os.Exit(subcommands.Run(getApplication(), nil))
}
