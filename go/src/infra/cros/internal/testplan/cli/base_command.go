// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package cli defines different commands for the test_plan tool.
package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/maruel/subcommands"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/proto"

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/bq"
	"go.chromium.org/luci/common/data/text"
	"go.chromium.org/luci/common/logging"
)

func errToCode(a subcommands.Application, err error) int {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", a.GetName(), err)
		return 1
	}

	return 0
}

// baseTestPlanRun embeds subcommands.CommandRunBase and implements flags shared
// across commands, such as logging and auth flags. It should be embedded in
// another struct that implements Run() for a specific command. baseTestPlanRun
// implements cli.ContextModificator, to set the log level based on flags.
type baseTestPlanRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
	logLevel  logging.Level
}

// addSharedFlags adds shared auth and logging flags.
func (r *baseTestPlanRun) addSharedFlags(authOpts auth.Options) {
	r.authFlags = authcli.Flags{}
	r.authFlags.Register(r.GetFlags(), authOpts)

	r.logLevel = logging.Info
	r.Flags.Var(&r.logLevel, "loglevel", text.Doc(`
	Log level, valid options are "debug", "info", "warning", "error". Default is "info".
	`))
}

// ModifyContext returns a new Context with the log level set in the flags.
func (r *baseTestPlanRun) ModifyContext(ctx context.Context) context.Context {
	return logging.SetLevel(ctx, r.logLevel)
}

// bqUpdateRun embeds baseTestPlanRun and implements flags shared across
// commands that update BigQuery, such as table and expiration flags.
// It should be embedded in another struct that implements Run() for a specific
// command.
type bqUpdateRun struct {
	baseTestPlanRun
	expiration time.Duration
	tableRef   string
}

// addBigQueryFlags defines flags for updating BigQuery.
func (r *bqUpdateRun) addBigQueryFlags() {
	r.Flags.StringVar(
		&r.tableRef,
		"table",
		"",
		text.Doc(`
			BigQuery table to upload to, in the form
			<project>.<dataset>.<table>. Required. The table will be
			created if it doesn't already exist, and the schema will be
			updated if needed.
		`),
	)
	r.Flags.DurationVar(
		&r.expiration,
		"expiration",
		time.Hour*24*90,
		text.Doc(`
			The expiration on the rows uploaded during execution,
			see https://cloud.google.com/bigquery/docs/managing-partitioned-tables#partition-expiration.
			Defaults to 90 days.
		`),
	)
}

// getClientAndTable parses tableRef to get a bigquery client and table.
func (r *bqUpdateRun) getClientAndTable(ctx context.Context) (*bigquery.Client, *bigquery.Table, error) {
	components := strings.Split(r.tableRef, ".")
	if len(components) != 3 {
		return nil, nil, fmt.Errorf("-table must be in the form <project>.<dataset>.<table>, got %q", r.tableRef)
	}

	authOpts, err := r.authFlags.Options()
	if err != nil {
		return nil, nil, err
	}

	tokenSource, err := auth.NewAuthenticator(ctx, auth.SilentLogin, authOpts).TokenSource()
	if err != nil {
		return nil, nil, err
	}

	client, err := bigquery.NewClient(ctx, components[0], option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, nil, err
	}

	table := client.Dataset(components[1]).Table(components[2])
	return client, table, nil
}

// ensureTableAndUploadRows ensures that the table exists and updates its schema
// if needed, then uploads rows.
func ensureTableAndUploadRows[M proto.Message](
	ctx context.Context,
	client *bigquery.Client,
	table *bigquery.Table,
	tableMetadata *bigquery.TableMetadata,
	rows []M,
) error {
	logging.Infof(ctx, "ensuring table exists and updating schema")
	if err := bq.EnsureTable(ctx, table, tableMetadata); err != nil {
		return err
	}

	logging.Infof(ctx, "uploading rows to BigQuery.")
	uploader := bq.NewUploader(ctx, client, table.DatasetID, table.TableID)
	// Uploader takes v2 proto rows, so we need to convert DirBQRow to
	// MessageV2.
	v2Rows := make([]proto.Message, len(rows))
	for i, row := range rows {
		v2Rows[i] = row
	}
	return uploader.Put(ctx, v2Rows...)
}
