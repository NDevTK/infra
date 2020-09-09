// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package cron implements handlers for appengine cron targets in this app.
package cron

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"go.chromium.org/luci/appengine/gaemiddleware"
	authclient "go.chromium.org/luci/auth"
	gitilesapi "go.chromium.org/luci/common/api/gitiles"
	"go.chromium.org/luci/common/bq"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	ds "go.chromium.org/luci/gae/service/datastore"
	"go.chromium.org/luci/gae/service/info"
	"go.chromium.org/luci/grpc/prpc"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/router"
	"golang.org/x/oauth2"
	"google.golang.org/api/iterator"

	apibq "infra/appengine/cros/lab_inventory/api/bigquery"
	"infra/appengine/cros/lab_inventory/app/config"
	"infra/appengine/cros/lab_inventory/app/converter"
	dronequeenapi "infra/appengine/drone-queen/api"
	"infra/libs/cros/git"
	"infra/libs/cros/gs"
	bqlib "infra/libs/cros/lab_inventory/bq"
	"infra/libs/cros/lab_inventory/cfg2datastore"
	"infra/libs/cros/lab_inventory/changehistory"
	"infra/libs/cros/lab_inventory/datastore"
	"infra/libs/cros/lab_inventory/deviceconfig"
	"infra/libs/cros/lab_inventory/dronecfg"
	"infra/libs/cros/lab_inventory/hart"
	"infra/libs/cros/lab_inventory/manufacturingconfig"
)

// InstallHandlers installs handlers for cron jobs that are part of this app.
//
// All handlers serve paths under /internal/cron/*
// These handlers can only be called by appengine's cron service.
func InstallHandlers(r *router.Router, mwBase router.MiddlewareChain) {
	mwCron := mwBase.Extend(gaemiddleware.RequireCron)

	r.GET("/internal/cron/import-service-config", mwCron, logAndSetHTTPErr(importServiceConfig))

	r.GET("/internal/cron/dump-to-bq", mwCron, logAndSetHTTPErr(dumpToBQCronHandler))

	r.GET("/internal/cron/dump-registered-assets-snapshot", mwCron, logAndSetHTTPErr(dumpRegisteredAssetsCronHandler))

	r.GET("/internal/cron/dump-inventory-snapshot", mwCron, logAndSetHTTPErr(dumpInventorySnapshot))

	r.GET("/internal/cron/dump-other-configs-snapshot", mwCron, logAndSetHTTPErr(dumpOtherConfigsCronHandler))

	r.GET("/internal/cron/dump-asset-info-to-bq", mwCron, logAndSetHTTPErr(dumpAssetInfoToBQHandler))

	r.GET("/internal/cron/sync-dev-config", mwCron, logAndSetHTTPErr(syncDevConfigHandler))

	r.GET("/internal/cron/sync-manufacturing-config", mwCron, logAndSetHTTPErr(syncManufacturingConfigHandler))

	r.GET("/internal/cron/changehistory-to-bq", mwCron, logAndSetHTTPErr(dumpChangeHistoryToBQCronHandler))

	r.GET("/internal/cron/push-to-drone-queen", mwCron, logAndSetHTTPErr(pushToDroneQueenCronHandler))

	r.GET("/internal/cron/report-inventory", mwCron, logAndSetHTTPErr(reportInventoryCronHandler))

	r.GET("/internal/cron/sync-device-list-to-drone-config", mwCron, logAndSetHTTPErr(syncDeviceListToDroneConfigHandler))

	r.GET("/internal/cron/sync-asset-info-from-hart", mwCron, logAndSetHTTPErr(syncAssetInfoFromHaRT))

	r.GET("/internal/cron/backfill-asset-tags", mwCron, logAndSetHTTPErr(backfillAssetTagsToDevicesHandler))
}

const pageSize = 500

func importServiceConfig(c *router.Context) error {
	return config.Import(c.Context)
}

func dumpToBQCronHandler(c *router.Context) (err error) {
	logging.Infof(c.Context, "not implemented yet")
	return nil
}

func syncDevConfigHandler(c *router.Context) error {
	logging.Infof(c.Context, "Start syncing device_config repo")
	cfg := config.Get(c.Context)
	if cfg.GetProjectConfigSource().GetEnableProjectConfig() {
		t, err := auth.GetRPCTransport(c.Context, auth.AsSelf, auth.WithScopes(authclient.OAuthScopeEmail, gitilesapi.OAuthScope))
		if err != nil {
			return err
		}
		bsCfg := cfg.GetProjectConfigSource()
		gitClient, err := git.NewClient(c.Context, &http.Client{Transport: t}, "", bsCfg.GetGitilesHost(), bsCfg.GetProject(), bsCfg.GetBranch())
		if err != nil {
			return err
		}
		gsClient, err := gs.NewClient(c.Context, t)
		if err != nil {
			return err
		}
		return deviceconfig.UpdateDatastoreFromBoxster(c.Context, gitClient, gsClient, bsCfg.GetProgramConfigsGsPath())
	}
	dCcfg := cfg.GetDeviceConfigSource()
	cli, err := cfg2datastore.NewGitilesClient(c.Context, dCcfg.GetHost())
	if err != nil {
		return err
	}
	return deviceconfig.UpdateDatastore(c.Context, cli, dCcfg.GetProject(), dCcfg.GetCommittish(), dCcfg.GetPath())
}

func syncManufacturingConfigHandler(c *router.Context) error {
	logging.Infof(c.Context, "Start syncing manufacturing_config repo")
	cfg := config.Get(c.Context).GetManufacturingConfigSource()
	cli, err := cfg2datastore.NewGitilesClient(c.Context, cfg.GetHost())
	if err != nil {
		return err
	}
	project := cfg.GetProject()
	committish := cfg.GetCommittish()
	path := cfg.GetPath()
	return manufacturingconfig.UpdateDatastore(c.Context, cli, project, committish, path)
}

func dumpRegisteredAssetsCronHandler(c *router.Context) error {
	ctx := c.Context
	logging.Infof(ctx, "Start to dump registered assets to bigquery")

	uploader, err := bqlib.InitBQUploader(ctx, info.AppID(ctx), "inventory", fmt.Sprintf("registered_assets$%s", bqlib.GetPSTTimeStamp(time.Now())))
	if err != nil {
		return err
	}
	msgs := bqlib.GetRegisteredAssetsProtos(ctx)
	logging.Debugf(ctx, "Dumping %d records to bigquery", len(msgs))
	if err := uploader.Put(ctx, msgs...); err != nil {
		return err
	}
	logging.Debugf(ctx, "Dump is successfully finished")
	return nil
}

func dumpAssetInfoToBQHandler(c *router.Context) error {
	ctx := c.Context
	logging.Infof(ctx, "Starting to dump asset info to BQ")

	uploader, err := bqlib.InitBQUploader(ctx, info.AppID(ctx), "inventory", fmt.Sprintf("asset_info$%s", bqlib.GetPSTTimeStamp(time.Now())))
	if err != nil {
		return err
	}
	msgs, err := datastore.GetAllAssetInfo(ctx, false)

	// uploader only accepts proto.Message interface. Casting AssetInfo
	// to proto.Message interface
	data := make([]proto.Message, len(msgs))
	for idx, msg := range msgs {
		data[idx] = msg
	}
	if err != nil {
		return err
	}

	logging.Debugf(ctx, "Dumping %d asset info records to BQ", len(data))
	if err := uploader.Put(ctx, data...); err != nil {
		return err
	}
	logging.Debugf(ctx, "Dumped all asset info records to BQ")
	return nil
}

func dumpOtherConfigsCronHandler(c *router.Context) error {
	ctx := c.Context
	logging.Infof(ctx, "Start to dump related configs in inventory to bigquery")

	curTime := time.Now()
	curTimeStr := bqlib.GetPSTTimeStamp(curTime)
	client, err := bigquery.NewClient(ctx, info.AppID(ctx))
	if err != nil {
		return err
	}

	uploader := bqlib.InitBQUploaderWithClient(ctx, client, "inventory", fmt.Sprintf("deviceconfig$%s", curTimeStr))
	msgs := bqlib.GetDeviceConfigProtos(ctx)
	logging.Debugf(ctx, "Dumping %d records of device configs to bigquery", len(msgs))
	if err := uploader.Put(ctx, msgs...); err != nil {
		return err
	}

	uploader = bqlib.InitBQUploaderWithClient(ctx, client, "inventory", fmt.Sprintf("manufacturing$%s", curTimeStr))
	msgs = bqlib.GetManufacturingConfigProtos(ctx)
	logging.Debugf(ctx, "Dumping %d records of manufacturing configs to bigquery", len(msgs))
	if err := uploader.Put(ctx, msgs...); err != nil {
		return err
	}

	logging.Debugf(ctx, "Dump is successfully finished")
	return nil
}

func dumpChangeHistoryToBQCronHandler(c *router.Context) error {
	ctx := c.Context
	logging.Infof(c.Context, "Start to dump change history to bigquery")
	project := info.AppID(ctx)
	dataset := "inventory"
	table := "changehistory"

	client, err := bigquery.NewClient(ctx, project)
	if err != nil {
		return err
	}
	up := bq.NewUploader(ctx, client, dataset, table)
	up.SkipInvalidRows = true
	up.IgnoreUnknownValues = true

	changes, err := changehistory.LoadFromDatastore(ctx)
	if err != nil {
		return err
	}
	msgs := make([]proto.Message, len(changes))
	for i, c := range changes {
		updatedTime, _ := ptypes.TimestampProto(c.Updated)
		msgs[i] = &apibq.ChangeHistory{
			Id:          c.DeviceID,
			Hostname:    c.Hostname,
			Label:       c.Label,
			OldValue:    c.OldValue,
			NewValue:    c.NewValue,
			UpdatedTime: updatedTime,
			ByWhom: &apibq.ChangeHistory_User{
				Name:  c.ByWhomName,
				Email: c.ByWhomEmail,
			},
			Comment: c.Comment,
		}
	}

	logging.Debugf(ctx, "Uploading %d records of change history", len(msgs))
	if err := up.Put(ctx, msgs...); err != nil {
		return err
	}
	logging.Debugf(ctx, "Cleaning %d records of change history from datastore", len(msgs))
	return changehistory.FlushDatastore(ctx, changes)
}

// dumpInventorySnapshot takes a snapshot of the inventory at the current time and
// uploads it to bigquery.
func dumpInventorySnapshot(c *router.Context) (err error) {
	ctx := c.Context
	defer func() {
		dumpInventorySnapshotTick.Add(ctx, 1, err == nil)
	}()

	logging.Infof(c.Context, "Start dumping inventory snapshot")
	project := info.AppID(ctx)
	dataset := "inventory"
	curTimeStr := bqlib.GetPSTTimeStamp(time.Now())
	client, err := bigquery.NewClient(ctx, project)
	labconfigUploader := bqlib.InitBQUploaderWithClient(ctx, client, dataset, fmt.Sprintf("lab$%s", curTimeStr))
	stateUploader := bqlib.InitBQUploaderWithClient(ctx, client, dataset, fmt.Sprintf("stateconfig$%s", curTimeStr))

	if err != nil {
		return fmt.Errorf("bq client: %s", err)
	}

	logging.Debugf(ctx, "getting all devices")
	for curPageToken := ""; ; {
		devices, nextPageToken, err := datastore.ListDevices(ctx, pageSize, curPageToken)
		if err != nil {
			return fmt.Errorf("gathering devices: %s", err)
		}
		labconfigs, stateconfigs, err := converter.DeviceToBQMsgsSeq(devices)
		if es := err.(errors.MultiError); len(es) > 0 {
			for _, e := range es {
				logging.Errorf(ctx, "failed to get devices: %s", e)
			}
		}
		if len(labconfigs) > 0 {
			logging.Debugf(ctx, "uploading %d lab configs to bigquery dataset (%s) table (lab)", len(labconfigs), dataset)
			if err := labconfigUploader.Put(ctx, labconfigs...); err != nil {
				return fmt.Errorf("labconfig put: %s", err)
			}
		}
		if len(stateconfigs) > 0 {
			logging.Debugf(ctx, "uploading %d state configs to bigquery dataset (%s) table (stateconfig)", len(stateconfigs), dataset)
			if err := stateUploader.Put(ctx, stateconfigs...); err != nil {
				return fmt.Errorf("stateconfig put: %s", err)
			}
		}
		if nextPageToken == "" {
			break
		}
		curPageToken = nextPageToken
	}

	logging.Debugf(ctx, "successfully uploaded lab inventory to bigquery")
	return nil
}

func pushToDroneQueenCronHandler(c *router.Context) error {
	ctx := c.Context
	logging.Infof(c.Context, "Start to push inventory to drone queen")
	queenHostname := config.Get(ctx).QueenService
	if queenHostname == "" {
		logging.Infof(ctx, "No drone queen service configured.")
		return nil
	}

	droneQueenRecord, err := dronecfg.Get(ctx, dronecfg.QueenDroneName(config.Get(ctx).Environment))
	if err != nil {
		return err
	}

	duts := make([]string, len(droneQueenRecord.DUTs))
	for i := range duts {
		duts[i] = droneQueenRecord.DUTs[i].Hostname
	}
	ts, err := auth.GetTokenSource(ctx, auth.AsSelf)
	if err != nil {
		return err
	}
	h := oauth2.NewClient(ctx, ts)
	client := dronequeenapi.NewInventoryProviderPRPCClient(&prpc.Client{
		C:    h,
		Host: queenHostname,
	})
	logging.Debugf(ctx, "DUTs to declare: %#v", duts)
	_, err = client.DeclareDuts(ctx, &dronequeenapi.DeclareDutsRequest{Duts: duts})
	if err != nil {
		return err
	}
	return nil
}

func reportInventoryCronHandler(c *router.Context) error {
	logging.Infof(c.Context, "start reporting inventory")
	if config.Get(c.Context).EnableInventoryReporting {
		return datastore.ReportInventory(c.Context, config.Get(c.Context).Environment)
	}
	logging.Infof(c.Context, "not enabled yet")
	return nil
}

func syncDeviceListToDroneConfigHandler(c *router.Context) error {
	ctx := c.Context
	logging.Infof(c.Context, "start to sync device list to drone config")
	return dronecfg.SyncDeviceList(ctx, dronecfg.QueenDroneName(config.Get(ctx).Environment))
}

func syncAssetInfoFromHaRT(c *router.Context) error {
	ctx := c.Context

	logging.Infof(ctx, "Sync AssetInfo from HaRT")

	cfg := config.Get(ctx).GetHart()
	proj := cfg.GetProject()
	topic := cfg.GetTopic()

	if proj == "" || topic == "" {
		return errors.Reason("Invalid config. Project[%v] Topic[%v]",
			proj, topic).Err()
	}

	assets, err := datastore.GetAllAssets(ctx, true)
	if err != nil {
		logging.Errorf(ctx, err.Error())
		return err
	}

	ids := make([]string, 0, len(assets))
	for _, a := range assets {
		ids = append(ids, a.Id)
	}

	// Filter assets not yet in datastore
	// TODO(anushruth): Change to exists after implementation in datastore.
	// Need not get complete AssetInfo to determine existence
	res := datastore.GetAssetInfo(ctx, ids)
	req := make([]string, 0, len(assets))
	for _, a := range res {
		if a.Err != nil && ds.IsErrNoSuchEntity(a.Err) {
			req = append(req, a.Entity.AssetTag)
		}
	}

	logging.Infof(ctx, "Syncing %v AssetInfo entit(y|ies) from HaRT", len(req))
	hart.SyncAssetInfoFromHaRT(ctx, proj, topic, req)

	return nil
}

// List of regexp to compare Servos. This is used to avoid backfilling
// asset tags to devices using servo asset tags
var notDUT = []*regexp.Regexp{
	regexp.MustCompile(`Servo\ v4\ Type\-[AC]`),
}

func isDUT(googleCodeName string) bool {
	for _, exp := range notDUT {
		if exp.MatchString(googleCodeName) {
			return false
		}
	}
	return true
}

// backfillAssetTagsToDevicesHandler uses BQ table in the project to
// update devices with their asset tag
//
// This job is run every hour and the 'rate per hour'/'batch size' can
// can be configured in the project configs.
func backfillAssetTagsToDevicesHandler(c *router.Context) error {
	ctx := c.Context

	cfg := config.Get(ctx).GetBackfillingConfig()

	batchSize := int(cfg.GetRate())

	if cfg.GetEnable() {
		logging.Infof(ctx, "Backfill asset tags to CrosDevices.")
	} else {
		logging.Infof(ctx, "Backfilling asset tags not enabled.")
		return nil
	}

	client, err := bigquery.NewClient(ctx, info.AppID(ctx))

	if err != nil {
		return err
	}

	// HostNameAT is used here to record the mapping between
	// HostName and Asset tag from bigquery
	type HostNameAT struct {
		Hostname string
		AssetTag string
	}

	// Get all the known hostname <-> Asset tag mapping
	q := client.Query("SELECT s_host_name as hostname, a_asset_tag as assetTag FROM `" +
		cfg.GetDataset() + "." + cfg.GetTable() +
		"` WHERE a_asset_tag NOT IN (SELECT id FROM `" + cfg.GetDataset() +
		".lab` WHERE hostname=s_host_name)")
	it, err := q.Read(ctx)
	if err != nil {
		return err
	}

	for batchSize > 0 {
		hostAT := make(map[string]string, batchSize)
		hosts := make([]string, 0, batchSize)

		for idx := 0; idx < batchSize; idx++ {
			var value HostNameAT
			err := it.Next(&value)
			if err != nil {
				if err == iterator.Done {
					batchSize = len(hostAT)
					break
				}
				return err
			}
			hostAT[value.Hostname] = value.AssetTag
			hosts = append(hosts, value.Hostname)
		}

		devOpRes := datastore.GetDevicesByHostnames(ctx, hosts)
		for _, dev := range devOpRes {
			if dev.Err == nil && string(dev.Entity.ID) != hostAT[dev.Entity.Hostname] {
				logging.Infof(ctx, "Updating %v from %v to %v",
					dev.Entity.Hostname, dev.Entity.ID,
					hostAT[dev.Entity.Hostname])
				datastore.UpdateDeviceID(ctx, string(dev.Entity.ID),
					hostAT[dev.Entity.Hostname])
				batchSize--
			}
		}
	}
	return nil
}

func logAndSetHTTPErr(f func(c *router.Context) error) func(*router.Context) {
	return func(c *router.Context) {
		if err := f(c); err != nil {
			logging.Errorf(c.Context, err.Error())
			http.Error(c.Writer, "Internal server error", http.StatusInternalServerError)
		}
	}
}
