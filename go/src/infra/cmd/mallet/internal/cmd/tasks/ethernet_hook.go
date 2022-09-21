// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tasks

import (
	"context"
	"fmt"

	"cloud.google.com/go/storage"
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"google.golang.org/api/option"

	"infra/cmd/mallet/internal/cmd/tasks/ethernethook"
	"infra/cmd/mallet/internal/site"
	"infra/cmdsupport/cmdlib"
)

// EthernetHook does nothing, but eventually it will print all the ethernet events for a single
// day as a convenient list of protos.
var EthernetHook = &subcommands.Command{
	UsageLine: "ethernet-hook",
	ShortDesc: "ethernet-hook",
	LongDesc:  `ethernet-hook prints ethernet hook events for a single day.`,
	CommandRun: func() subcommands.CommandRun {
		c := &ethernetHookRun{}
		c.authFlags.Register(&c.Flags, site.EthernetHookCallbackOptions)
		c.commonFlags.Register(&c.Flags)
		c.Flags.StringVar(&c.date, "date", "", "the date to process")
		c.Flags.StringVar(&c.bucket, "bucket", "chromeos-test-logs", "the base GS bucket to check for logs")
		c.Flags.StringVar(&c.prefix, "prefix", "", "prefix of the objects in question")
		c.Flags.StringVar(&c.delimiter, "delimiter", "", "delimiter of the objects in question")
		return c
	},
}

type ethernetHookRun struct {
	subcommands.CommandRunBase
	authFlags   authcli.Flags
	envFlags    site.EnvFlags
	commonFlags site.CommonFlags

	// Dates are given in YYYY-MM-DD format, the only correct format.
	date string

	// Add a default bucket.
	bucket string

	// The prefix of the Google Storage object within the Google Storage bucket.
	prefix string

	// The limit controls what kinds of results are returned: prefixes or full results.
	delimiter string
}

// Run runs the ethernet hook command.
func (c *ethernetHookRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := cli.GetContext(a, c, env)
	if c.commonFlags.Verbose {
		ctx = logging.SetLevel(ctx, logging.Debug)
	}
	if err := c.innerRun(ctx, a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

// Print out ethernet-hook-related records.
func (c *ethernetHookRun) innerRun(ctx context.Context, a subcommands.Application, args []string, env subcommands.Env) error {
	options, err := c.authFlags.Options()
	if err != nil {
		return errors.Annotate(err, "failed to get auth options").Err()
	}
	authenticator := auth.NewAuthenticator(ctx, auth.SilentLogin, options)
	httpClient, err := authenticator.Client()
	if err != nil {
		return errors.Annotate(err, "failed to set up http client").Err()
	}
	email, err := authenticator.GetEmail()
	if err != nil {
		return errors.Annotate(err, "failed to get email").Err()
	}
	tokenSource, err := authenticator.TokenSource()
	if err != nil {
		return errors.Annotate(err, "failed to get token source").Err()
	}
	rawStorageClient, err := storage.NewClient(
		ctx,
		option.WithHTTPClient(httpClient),
		option.WithTokenSource(tokenSource),
		option.WithScopes(options.Scopes...),
	)
	if err != nil {
		return errors.Annotate(err, "failed to set up storage client as %q", email).Err()
	}
	storageClient, err := ethernethook.NewExtendedGSClient(rawStorageClient)
	if err != nil {
		return errors.Annotate(err, "failed to wrap storage client").Err()
	}

	out, err := storageClient.LsSmall(ctx, c.bucket, &storage.Query{
		Delimiter:                c.delimiter,
		Prefix:                   c.prefix,
		IncludeTrailingDelimiter: true,
	})
	if err != nil {
		return err
	}

	for _, item := range out {
		fmt.Fprintf(a.GetOut(), "%s\n", storageClient.ExpandName(c.bucket, item))
	}

	return nil
}
