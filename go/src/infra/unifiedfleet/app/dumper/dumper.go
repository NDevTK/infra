// Copyright 2023 The Chromium Authors
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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	bqlib "infra/cros/lab_inventory/bq"
	"infra/unifiedfleet/app/cron"
	"infra/unifiedfleet/app/model/configuration"
	"infra/unifiedfleet/app/util"
)

// Jobs is a list of all the cron jobs that are currently available for running
var Jobs = []*cron.CronTab{
	{
		// Dump configs, registrations, inventory and states to BQ
		Name:     util.CronJobNames["mainBQCronDaily"],
		Time:     20 * time.Minute,
		TrigType: cron.DAILY,
		Job:      dump,
	},
	{
		// Dump configs, registrations, inventory and states to BQ
		Name:     util.CronJobNames["mainBQCronHourly"],
		Time:     30 * time.Minute,
		TrigType: cron.HOURLY,
		Job:      dumpHourly,
	},
	{
		// Dump change events to BQ
		Name:     util.CronJobNames["changeEventToBQCron"],
		Time:     10 * time.Minute,
		TrigType: cron.EVERY,
		Job:      dumpChangeEvent,
	},
	{
		// Dump snapshots to BQ
		Name:     util.CronJobNames["snapshotToBQCron"],
		Time:     10 * time.Minute,
		TrigType: cron.EVERY,
		Job:      dumpChangeSnapshots,
	},
	{
		// Dump network configs to BQ
		Name:     util.CronJobNames["networkConfigToBQCron"],
		Time:     60 * time.Minute,
		TrigType: cron.EVERY,
		Job:      dumpCrosNetwork,
	},
	{
		// Sync asset info from HaRT
		Name:     util.CronJobNames["hartSyncCron"],
		TrigType: cron.HOURLY,
		Job:      SyncAssetInfoFromHaRT,
	},
	{
		// Push changes to dron queen
		Name:     util.CronJobNames["droneQueenSyncCron"],
		Time:     10 * time.Minute,
		TrigType: cron.EVERY,
		Job:      pushToDroneQueen,
	},
	{
		// Report UFS metrics
		Name:     util.CronJobNames["InventoryMetricsReportCron"],
		Time:     5 * time.Minute,
		TrigType: cron.EVERY,
		Job:      reportUFSInventoryCronHandler,
	},
	{
		// Sync Goldeneye Data
		Name:     util.CronJobNames["goldeneyeDevicesSyncCron"],
		Time:     12 * time.Hour,
		TrigType: cron.EVERY,
		Job:      getGoldenEyeData,
	},
	{
		// Compare differences between Swarming label generators
		Name:     util.CronJobNames["SwarmingLabelsDiffCron"],
		Time:     5 * time.Minute,
		TrigType: cron.DAILY,
		Job:      swarmingLabelsDiffHandler,
	},
	{
		// Sync ENC bot and security configs
		Name:     util.CronJobNames["botConfigSyncCron"],
		Time:     10 * time.Minute,
		TrigType: cron.EVERY,
		Job:      getBotConfigs,
	},
	{
		Name:     util.CronJobNames["deviceConfigSyncCron"],
		Time:     1 * time.Hour,
		TrigType: cron.EVERY,
		Job:      syncDeviceConfigs,
	},
	{
		// This job is not meant to be run by the cron. This will
		// be triggered by shivas at whatever time oncall deems
		// appropriate. There is a potential for this to block
		// other jobs as it updates ~500 rows every time. Only run
		// during low-traffic non-critical times. The long wait time
		// ensures that this will not be run under normal circumstances
		// (unless we manage to not update/reboot UFS for 200 years)
		Name:     util.CronJobNames["indexAssets"],
		Time:     200 * 365 * 24 * time.Hour, // when I'm gone carry on, don't mourn
		TrigType: cron.EVERY,
		Job:      IndexAssets,
	},
	{
		// This job is not meant to be run by the cron. This will
		// be triggered by shivas at whatever time oncall deems
		// appropriate. There is a potential for this to block
		// other jobs as it updates ~500 rows every time. Only run
		// during low-traffic non-critical times. The long wait time
		// ensures that this will not be run under normal circumstances
		// (unless we manage to not update/reboot UFS for 200 years)
		Name:     util.CronJobNames["indexMachines"],
		Time:     200 * 365 * 24 * time.Hour,
		TrigType: cron.EVERY,
		Job:      IndexMachines,
	},
	{
		// This job is not meant to be run by the cron. This will
		// be triggered by shivas at whatever time oncall deems
		// appropriate. There is a potential for this to block
		// other jobs as it updates ~500 rows every time. Only run
		// during low-traffic non-critical times. The long wait time
		// ensures that this will not be run under normal circumstances
		// (unless we manage to not update/reboot UFS for 200 years)
		Name:     util.CronJobNames["indexRacks"],
		Time:     200 * 365 * 24 * time.Hour,
		TrigType: cron.EVERY,
		Job:      indexRacks,
	},
	{
		// This job is not meant to be run by the cron. This will
		// be triggered by shivas at whatever time oncall deems
		// appropriate. There is a potential for this to block
		// other jobs as it updates ~500 rows every time. Only run
		// during low-traffic non-critical times. The long wait time
		// ensures that this will not be run under normal circumstances
		// (unless we manage to not update/reboot UFS for 200 years)
		Name:     util.CronJobNames["indexMachineLSEs"],
		Time:     200 * 365 * 24 * time.Hour, // Wake me up when September ends
		TrigType: cron.EVERY,
		Job:      indexMachineLSEs,
	},
	{
		// This job is not meant to be run by the cron. This will
		// be triggered by shivas at whatever time oncall deems
		// appropriate. There is a potential for this to block
		// other jobs as it updates ~500 rows every time. Only run
		// during low-traffic non-critical times. The long wait time
		// ensures that this will not be run under normal circumstances
		// (unless we manage to not update/reboot UFS for 200 years)
		Name:     util.CronJobNames["indexDutStates"],
		Time:     200 * 365 * 24 * time.Hour,
		TrigType: cron.EVERY,
		Job:      indexDutStates,
	},
}

// InitServer initializes a cron server.
func InitServer(srv *server.Server) {
	for _, job := range Jobs {
		// make a copy of the job to avoid race condition.
		t := job
		// Start all the cron jobs in background.
		srv.RunInBackground(job.Name, func(ctx context.Context) {
			cron.Run(ctx, t)
		})
	}
}

// TriggerJob triggers a job by name. Returns error if the job is not found.
func TriggerJob(name string) error {
	for _, job := range Jobs {
		if job.Name == name {
			return cron.Trigger(job)
		}
	}
	return status.Errorf(codes.NotFound, "Invalid cron job %s. Not found", name)
}

// dump a snapshot to BQ daily
func dump(ctx context.Context) error {
	ctx = logging.SetLevel(ctx, logging.Info)
	if err := exportToBQ(ctx, dumpToBQ); err != nil {
		return err
	}
	return nil
}

// dumpHourly Similar to dump, but hourly
func dumpHourly(ctx context.Context) error {
	ctx = logging.SetLevel(ctx, logging.Info)
	if err := exportToBQ(ctx, dumpToBQHourly); err != nil {
		return err
	}
	return nil
}

func dumpToBQ(ctx context.Context, bqClient *bigquery.Client) (err error) {
	defer func() {
		ns := util.GetNamespaceFromCtx(ctx)
		dumpToBQDailyTick.Add(ctx, 1, err == nil, ns)
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
	if err := dumpConfigurations(ctx, bqClient, curTimeStr, false); err != nil {
		return errors.Annotate(err, "dump configurations").Err()
	}
	if err := dumpRegistration(ctx, bqClient, curTimeStr, false); err != nil {
		return errors.Annotate(err, "dump registrations").Err()
	}
	if err := dumpInventory(ctx, bqClient, curTimeStr, false); err != nil {
		return errors.Annotate(err, "dump inventories").Err()
	}
	if err := dumpState(ctx, bqClient, curTimeStr, false); err != nil {
		return errors.Annotate(err, "dump states").Err()
	}
	logging.Debugf(ctx, "Dump is successfully finished")
	return nil
}

func dumpToBQHourly(ctx context.Context, bqClient *bigquery.Client) (err error) {
	defer func() {
		ns := util.GetNamespaceFromCtx(ctx)
		dumpToBQHourlyTick.Add(ctx, 1, err == nil, ns)
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
	if err := dumpConfigurations(ctx, bqClient, curTimeStr, true); err != nil {
		return errors.Annotate(err, "dump configurations").Err()
	}
	if err := dumpRegistration(ctx, bqClient, curTimeStr, true); err != nil {
		return errors.Annotate(err, "dump registrations").Err()
	}
	if err := dumpInventory(ctx, bqClient, curTimeStr, true); err != nil {
		return errors.Annotate(err, "dump inventories").Err()
	}
	if err := dumpState(ctx, bqClient, curTimeStr, true); err != nil {
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
	return exportToBQ(ctx, dumpChangeEventHelper)
}

func dumpChangeSnapshots(ctx context.Context) (err error) {
	defer func() {
		dumpChangeSnapshotTick.Add(ctx, 1, err == nil)
	}()
	ctx = logging.SetLevel(ctx, logging.Info)
	logging.Debugf(ctx, "Dumping change snapshots to BQ")
	return exportToBQ(ctx, dumpChangeSnapshotHelper)
}

func dumpCrosNetwork(ctx context.Context) (err error) {
	defer func() {
		dumpCrosNetworkTick.Add(ctx, 1, err == nil)
	}()
	// In UFS write to 'os' namespace
	ctx, err = util.SetupDatastoreNamespace(ctx, util.OSNamespace)
	if err != nil {
		return err
	}
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

func exportToBQ(ctx context.Context, f func(ctx context.Context, bqClient *bigquery.Client) error) error {
	var mErr error
	for _, ns := range util.ClientToDatastoreNamespace {
		newCtx, err := util.SetupDatastoreNamespace(ctx, ns)
		if ns == "" {
			// This is only for printing error message for default namespace.
			ns = "default (chrome)"
		}
		logging.Infof(newCtx, "Exporting to BQ for namespace %q", ns)
		if err != nil {
			logging.Errorf(ctx, "Setting namespace %q failed, BQ export skipped: %s", ns, err.Error())
			mErr = errors.NewMultiError(mErr, err)
			continue
		}
		err = f(newCtx, get(newCtx))
		if err != nil {
			logging.Errorf(ctx, "BQ export failed for the namespace %q: %s", ns, err.Error())
			mErr = errors.NewMultiError(mErr, err)
		}
	}
	return mErr
}
