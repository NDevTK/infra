// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package util

var CronJobNames = map[string]string{
	"mainBQCronDaily":            "ufs.dumper.daily",
	"mainBQCronHourly":           "ufs.dumper.hourly",
	"changeEventToBQCron":        "ufs.change_event.BqDump",
	"snapshotToBQCron":           "ufs.snapshot_msg.BqDump",
	"networkConfigToBQCron":      "ufs.cros_network.dump",
	"hartSyncCron":               "ufs.sync_devices.sync",
	"droneQueenSyncCron":         "ufs.push_to_drone_queen",
	"InventoryMetricsReportCron": "ufs.report_inventory",
	"goldeneyeDevicesSyncCron":   "ufs.sync_goldeneye_devices.sync",
	"SwarmingLabelsDiffCron":     "ufs.swarming_labels_diff",
	"botConfigSyncCron":          "ufs.sync_bot_config.sync",
	"deviceConfigSyncCron":       "ufs.device_config.sync",
	"indexAssets":                "ufs.indexer.asset",
	"indexMachines":              "ufs.indexer.machine",
	"indexRacks":                 "ufs.indexer.rack",
	"indexMachineLSEs":           "ufs.indexer.machinelse",
	"indexDutStates":             "ufs.indexer.dutstate",
}
