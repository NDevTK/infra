// Copyright 2022 The Chromium Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package weekly

import (
	"context"

	"github.com/maruel/subcommands"

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/logging/gologger"
	"go.chromium.org/luci/hardcoded/chromeinfra"
)

func application(p Param) *cli.Application {
	p.Auth = chromeinfra.DefaultAuthOptions()
	return &cli.Application{
		Name:  "weekly",
		Title: "A CLI for chrome browser perf engprod weekly updates.",
		Context: func(ctx context.Context) context.Context {
			return gologger.StdConfig.Use(ctx)
		},
		Commands: []*subcommands.Command{
			cmdReportIncoming(p),
			authcli.SubcommandLogin(p.Auth, "auth-login", false),
			authcli.SubcommandLogout(p.Auth, "auth-logout", false),
			authcli.SubcommandInfo(p.Auth, "auth-info", false),

			subcommands.CmdHelp,
		},
	}
}

// Param includes the parameters to use for the CLI application.
type Param struct {
	DefaultServiceDomain, OIDCProviderURL string
	Auth                                  auth.Options
}

// Main invokes the subcommands for the application.
func Main(p Param, args []string) int {
	return subcommands.Run(application(p), nil)
}
