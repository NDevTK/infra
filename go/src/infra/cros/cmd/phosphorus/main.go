// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Command autotest_status_parser extracts individual test case results from status.log.
package main

import (
	"context"
	"os"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/gcloud/gs"
	"go.chromium.org/luci/common/logging/gologger"
	"go.chromium.org/luci/hardcoded/chromeinfra"
	serverauth "go.chromium.org/luci/server/auth"

	parser "infra/cros/cmd/phosphorus/internal/autotest_status_parser/cmd"
	"infra/cros/cmd/phosphorus/internal/cmd"
	"infra/cros/cmd/phosphorus/internal/parallels"
	localstate "infra/cros/cmd/phosphorus/internal/skylab_local_state/cmd"
)

func getApplication(authOpts auth.Options) *cli.Application {
	return &cli.Application{
		Name:  "phosphorus",
		Title: "A tool for running Autotest tests and uploading their results.",
		Context: func(ctx context.Context) context.Context {
			return gologger.StdConfig.Use(ctx)
		},
		Commands: []*subcommands.Command{
			subcommands.CmdHelp,

			subcommands.Section("Authentication"),
			authcli.SubcommandInfo(authOpts, "whoami", false),
			authcli.SubcommandLogin(authOpts, "login", false),
			authcli.SubcommandLogout(authOpts, "logout", false),

			subcommands.Section("Main commands"),
			cmd.Prejob,
			cmd.RunTest,
			cmd.UploadToTKO,
			cmd.UploadToGS(authOpts),
			cmd.FetchCrashes,
			localstate.Load(authOpts),
			localstate.Remove(authOpts),
			localstate.Save(authOpts),
			parser.Parse,

			subcommands.Section("Build Parallels Image commands"),
			parallels.Provision,
			parallels.Save(authOpts),
		},
	}
}

func main() {
	auth := chromeinfra.DefaultAuthOptions()
	auth.Scopes = append(auth.Scopes, serverauth.CloudOAuthScopes...)
	auth.Scopes = append(auth.Scopes, gs.ReadWriteScopes...)
	os.Exit(subcommands.Run(getApplication(auth), nil))
}
