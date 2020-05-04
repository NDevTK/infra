// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dumper

import (
	"context"
	"infra/unifiedfleet/app/cron"
	"time"

	"cloud.google.com/go/bigquery"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server"

	bqlib "infra/libs/cros/lab_inventory/bq"
)

// Options is the dumper server configuration.
type Options struct {
	// CronInterval setups the user-specific cron interval for data dumping
	CronInterval time.Duration
}

// InitServer initializes a purger server.
func InitServer(srv *server.Server, opts Options) {
	srv.RunInBackground("ufs.dumper", func(ctx context.Context) {
		minInterval := 24 * 60 * time.Minute
		if opts.CronInterval > 0 {
			minInterval = opts.CronInterval
		}
		run(ctx, minInterval)
	})
}

func run(ctx context.Context, minInterval time.Duration) {
	cron.Run(ctx, minInterval, dumpConfigurations)
}

func dumpConfigurations(ctx context.Context) error {
	logging.Debugf(ctx, "Dumping configuration subsystems")
	curTime := time.Now()
	curTimeStr := bqlib.GetPSTTimeStamp(curTime)
	bqClient := get(ctx)
	if err := dumpChromePlatform(ctx, bqClient, curTimeStr); err != nil {
		return errors.Annotate(err, "dump configurations").Err()
	}

	logging.Debugf(ctx, "Dump is successfully finished")
	return nil
}

// unique key used to store and retrieve context.
var contextKey = "ufs bigquery-client key"

// Use installs bigquery client to context.
func Use(ctx context.Context, bqClient *bigquery.Client) context.Context {
	return context.WithValue(ctx, &contextKey, bqClient)
}

func get(ctx context.Context) *bigquery.Client {
	return ctx.Value(&contextKey).(*bigquery.Client)
}
