// Copyright 2017 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

/*
CLI tool to audit a googlesource gerrit host for appropriate permissions.
*/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/andygrunwald/go-gerrit"
	"github.com/luci/luci-go/common/auth"
	"github.com/luci/luci-go/common/logging"
	"github.com/luci/luci-go/common/logging/gologger"
	"golang.org/x/net/context"
)

const hostURLFormat = "https://%s-review.googlesource.com/"

var (
	host     = flag.String("host", "chromium", "Googlesource host name")
	template = flag.String("template", "Public-CQ", "Name of template project")
	login    = flag.Bool("login", false, "Use interactive login")
	creds    = flag.String("creds", "", "Path to service account JSON file")
	verbose  = flag.Bool("verbose", false, "Output more verbose logging")
)

func main() {
	flag.Parse()

	ctx := context.Background()
	ctx = gologger.StdConfig.Use(ctx)
	ctx = logging.SetLevel(ctx, logging.Warning)
	if *verbose {
		ctx = logging.SetLevel(ctx, logging.Debug)
	}

	authOptions := auth.Options{
		ServiceAccountJSONPath: *creds,
		Scopes: []string{
			auth.OAuthScopeEmail,
			"https://www.googleapis.com/auth/gerritcodereview",
		},
	}

	mode := auth.SilentLogin
	if *login {
		mode = auth.InteractiveLogin
	}

	httpclient, err := auth.NewAuthenticator(ctx, mode, authOptions).Client()
	if err != nil {
		logging.Errorf(ctx, "auth.NewAuthenticator: %v", err)
		if !*login {
			logging.Errorf(ctx, "Consider re-running with -login")
		}
		os.Exit(1)
	}

	client, err := gerrit.NewClient(fmt.Sprintf(hostURLFormat, *host), httpclient)
	if err != nil {
		logging.Errorf(ctx, "gerrit.NewClient: %v", err)
		os.Exit(1)
	}

	config, resp, err := client.Projects.GetConfig("All-Projects")
	if err != nil {
		logging.Errorf(ctx, "Error: %v", err)
		os.Exit(1)
	}
	if resp.StatusCode != 200 {
		// Seems like this shouldn't be necessary, but go-gerrit doesn't guarantee
		// that err will be non-nil if the response is not some flavor of success.
		logging.Errorf(ctx, "HTTP Error: %v", resp)
		os.Exit(1)
	}

	r, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		logging.Errorf(ctx, "Invalid JSON: %v", err)
		os.Exit(1)
	}
	fmt.Printf(string(r))
}
