// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"log"
	"os"

	"github.com/maruel/subcommands"

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/hardcoded/chromeinfra"

	"infra/cros/cmd/kron/common"
	kronSubCommands "infra/cros/cmd/kron/subcommands"
)

func getApplication(authOpts auth.Options) *subcommands.DefaultApplication {
	return &subcommands.DefaultApplication{
		Name:  "kron",
		Title: "Kron golang implementation",
		Commands: []*subcommands.Command{
			subcommands.CmdHelp,
			kronSubCommands.GetConfigParserCommand(authOpts),
			kronSubCommands.GetRunCommand(authOpts),
			kronSubCommands.GetFirestoreCommand(authOpts),
			authcli.SubcommandInfo(authOpts, "auth-info", false),
			authcli.SubcommandLogin(authOpts, "auth-login", false),
			authcli.SubcommandLogout(authOpts, "auth-logout", false),
		},
	}
}

type kronApplication struct {
	*subcommands.DefaultApplication
	stdoutLog *log.Logger
	stderrLog *log.Logger
}

func main() {
	opts := chromeinfra.DefaultAuthOptions()
	s := &kronApplication{
		getApplication(opts),
		common.Stdout,
		common.Stderr}
	os.Exit(subcommands.Run(s, nil))
}
