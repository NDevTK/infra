// Copyright 2023 The Chromium Authors
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

	"infra/cros/cmd/suite_scheduler/common"
	suschSubCommands "infra/cros/cmd/suite_scheduler/subcommands"
)

func getApplication(authOpts auth.Options) *subcommands.DefaultApplication {
	return &subcommands.DefaultApplication{
		Name:  "suite-scheduler",
		Title: "SuSch v1.5 golang implementation",
		Commands: []*subcommands.Command{
			subcommands.CmdHelp,
			suschSubCommands.GetConfigParserCommand(authOpts),
			authcli.SubcommandInfo(authOpts, "auth-info", false),
			authcli.SubcommandLogin(authOpts, "auth-login", false),
			authcli.SubcommandLogout(authOpts, "auth-logout", false),
		},
	}
}

type suiteSchedulerApplication struct {
	*subcommands.DefaultApplication
	stdoutLog *log.Logger
	stderrLog *log.Logger
}

type suiteSchedulerCommand interface {
	validate() error
}

func main() {
	opts := chromeinfra.DefaultAuthOptions()
	s := &suiteSchedulerApplication{
		getApplication(opts),
		common.Stdout,
		common.Stderr}
	os.Exit(subcommands.Run(s, nil))
}
