// Copyright 2098 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cmd

import (
	"context"
	"fmt"
	"infra/cmd/cros_test_platform/internal/site"
	"net/http"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/gcloud/gs"
)

// Enumerate subcommand: Enumerate test cases for a request.
var Enumerate = &subcommands.Command{
	UsageLine: "enumerate -input_json /path/to/input.json -output_json /path/to/output.json",
	ShortDesc: "Enumerate tasks to execute for a request.",
	LongDesc: `Enumerate tasks to execute for a request.

Step input and output uses JSON encoded protobufs defined at
https://chromium.googlesource.com/chromiumos/infra/proto/+/master/src/test_platform/steps/enumeration.proto`,
	CommandRun: func() subcommands.CommandRun {
		c := &enumerateRun{}
		c.authFlags.Register(&c.Flags, appendGSScopes(site.DefaultAuthOptions))
		c.Flags.StringVar(&c.inputPath, "input_json", "", "Path that contains JSON encoded test_platform.steps.EnumerationRequest")
		c.Flags.StringVar(&c.outputPath, "output_json", "", "Path where JSON encoded test_platform.steps.EnumerationResponse should be written.")
		return c
	},
}

type enumerateRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags

	inputPath  string
	outputPath string
}

func appendGSScopes(o auth.Options) auth.Options {
	o.Scopes = append(o.Scopes, gs.ReadOnlyScopes...)
	return o
}

func (c *enumerateRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		fmt.Fprintf(a.GetErr(), "%s\n", err)
		return 1
	}
	return 0
}

func (c *enumerateRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	return nil
}

// TODO: Move to common.go
// httpClient returns an HTTP client with authentication set up.
func commonHTTPClient(ctx context.Context, f *authcli.Flags) (*http.Client, error) {
	o, err := f.Options()
	if err != nil {
		return nil, errors.Annotate(err, "failed to get auth options").Err()
	}
	a := auth.NewAuthenticator(ctx, auth.OptionalLogin, o)
	c, err := a.Client()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create HTTP client").Err()
	}
	return c, nil
}
