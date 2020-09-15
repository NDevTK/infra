// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dumper

import (
	"context"
	"time"

	"cloud.google.com/go/bigquery"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server"

	bqlib "infra/libs/cros/lab_inventory/bq"
	"infra/unifiedfleet/app/cron"
	"infra/unifiedfleet/app/model/configuration"
	"infra/unifiedfleet/app/util"
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
	srv.RunInBackground("ufs.cros_inventory.dump", func(ctx context.Context) {
		cron.Run(ctx, 60*time.Minute, dumpCrosInventory)
	})
	srv.RunInBackground("ufs.change_event.BqDump", func(ctx context.Context) {
		cron.Run(ctx, 10*time.Minute, dumpChangeEvent)
	})
	srv.RunInBackground("ufs.change_event.BqDump", func(ctx context.Context) {
		cron.Run(ctx, 10*time.Minute, dumpChangeSnapshots)
	})
	srv.RunInBackground("ufs.cros_network.dump", func(ctx context.Context) {
		cron.Run(ctx, 60*time.Minute, dumpCrosNetwork)
	})
	// TODO(anushruth): Disabled machine sync as it's broken currently. Will enable it later
	// srv.RunInBackground("ufs.sync_machines.sync", func(ctx context.Context) {
	// 	cron.Run(ctx, 60*time.Minute, SyncMachinesFromAssets)
	// })
	srv.RunInBackground("ufs.sync_devices.sync", func(ctx context.Context) {
		cron.Run(ctx, 60*time.Minute, SyncAssetInfoFromHaRT)
	})
	srv.RunInBackground("ufs.sync_assets.sync", func(ctx context.Context) {
		cron.Run(ctx, 60*time.Minute, SyncAssetsFromIV2)
	})
}

func run(ctx context.Context, minInterval time.Duration) {
	cron.Run(ctx, minInterval, dump)
}

func dump(ctx context.Context) error {
	ctx = logging.SetLevel(ctx, logging.Info)
	// Execute importing before dumping
	err1 := importCrimson(ctx)
	err2 := dumpToBQ(ctx)
	if err1 == nil && err2 == nil {
		return nil
	}
	return errors.NewMultiError(err1, err2)
}

func dumpToBQ(ctx context.Context) (err error) {
	defer func() {
		dumpToBQTick.Add(ctx, 1, err == nil)
	}()
	logging.Infof(ctx, "Dumping to BQ")
	curTime := time.Now()
	curTimeStr := bqlib.GetPSTTimeStamp(curTime)
	if err := configuration.SaveProjectConfig(ctx, &configuration.ProjectConfigEntity{
		Name:             getProject(ctx),
		DailyDumpTimeStr: curTimeStr,
	}); err != nil {
		return err
	}
	bqClient := get(ctx)
	if err := dumpConfigurations(ctx, bqClient, curTimeStr); err != nil {
		return errors.Annotate(err, "dump configurations").Err()
	}
	if err := dumpRegistration(ctx, bqClient, curTimeStr); err != nil {
		return errors.Annotate(err, "dump registrations").Err()
	}
	if err := dumpInventory(ctx, bqClient, curTimeStr); err != nil {
		return errors.Annotate(err, "dump inventories").Err()
	}
	if err := dumpState(ctx, bqClient, curTimeStr); err != nil {
		return errors.Annotate(err, "dump states").Err()
	}
	logging.Debugf(ctx, "Dump is successfully finished")
	return nil
}

func dumpChangeEvent(ctx context.Context) (err error) {
	defer func() {
		dumpChangeEventTick.Add(ctx, 1, err == nil)
	}()
	ctx = logging.SetLevel(ctx, logging.Info)
	logging.Debugf(ctx, "Dumping change event to BQ")
	return dumpChangeEventHelper(ctx, get(ctx))
}

func dumpChangeSnapshots(ctx context.Context) (err error) {
	defer func() {
		dumpChangeSnapshotTick.Add(ctx, 1, err == nil)
	}()
	ctx = logging.SetLevel(ctx, logging.Info)
	logging.Debugf(ctx, "Dumping change snapshots to BQ")
	return dumpChangeSnapshotHelper(ctx, get(ctx))
}

func dumpCrosInventory(ctx context.Context) (err error) {
	defer func() {
		dumpCrosInventoryTick.Add(ctx, 1, err == nil)
	}()
	ctx = logging.SetLevel(ctx, logging.Info)
	return importCrosInventory(ctx)
}

func dumpCrosNetwork(ctx context.Context) (err error) {
	defer func() {
		dumpCrosNetworkTick.Add(ctx, 1, err == nil)
	}()
	return importCrosNetwork(ctx)
}

// unique key used to store and retrieve context.
var contextKey = util.Key("ufs bigquery-client key")
var projectKey = util.Key("ufs project key")

// Use installs bigquery client to context.
func Use(ctx context.Context, bqClient *bigquery.Client) context.Context {
	return context.WithValue(ctx, contextKey, bqClient)
}

func get(ctx context.Context) *bigquery.Client {
	return ctx.Value(contextKey).(*bigquery.Client)
}

// UseProject installs project name to context.
func UseProject(ctx context.Context, project string) context.Context {
	return context.WithValue(ctx, projectKey, project)
}

func getProject(ctx context.Context) string {
	return ctx.Value(projectKey).(string)
}
