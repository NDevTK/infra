// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dumper

import (
	"go.chromium.org/luci/common/tsmon/field"
	"go.chromium.org/luci/common/tsmon/metric"
)

var (
	dumpToBQTick = metric.NewCounter(
		"chromeos/ufs/dumper/dump_to_bq",
		"dumpToBQ attempt",
		nil,
		field.Bool("success"), // If the attempt succeed
	)
	dumpChangeEventTick = metric.NewCounter(
		"chromeos/ufs/dumper/dump_change_event",
		"dumpChangeEvent attempt",
		nil,
		field.Bool("success"),
	)
	dumpChangeSnapshotTick = metric.NewCounter(
		"chromeos/ufs/dumper/dump_change_snapshot",
		"dumpChangeSnapshot attempt",
		nil,
		field.Bool("success"),
	)
	dumpCrimsonTick = metric.NewCounter(
		"chromeos/ufs/dumper/import_crimson",
		"import crimson attempt",
		nil,
		field.Bool("success"),
	)
	dumpCrosInventoryTick = metric.NewCounter(
		"chromeos/ufs/dumper/import_cros_inventory",
		"import cros inventory attempt",
		nil,
		field.Bool("success"),
	)
	dumpCrosNetworkTick = metric.NewCounter(
		"chromeos/ufs/dumper/import_cros_network",
		"import cros network attempt",
		nil,
		field.Bool("success"),
	)
	dumpPushToDroneQueenTick = metric.NewCounter(
		"chromeos/ufs/dumper/push_to_drone_queen",
		"push to drone queen attempt",
		nil,
		field.Bool("success"),
	)
)
