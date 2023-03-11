// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cron

import (
	"go.chromium.org/luci/common/tsmon/field"
	"go.chromium.org/luci/common/tsmon/metric"
)

var (
	freeInvalidDUTsTick = metric.NewCounter(
		"chromeos/drone-queen/cron/free-invalid-duts/success",
		"success of free-invalid-duts cron jobs",
		nil,
		field.String("instance"),
		field.Bool("success"),
	)
	pruneExpiredDronesTick = metric.NewCounter(
		"chromeos/drone-queen/cron/prune-expired-drones/success",
		"success of prune-expired-drones cron jobs",
		nil,
		field.String("instance"),
		field.Bool("success"),
	)
	pruneDrainedDUTsTick = metric.NewCounter(
		"chromeos/drone-queen/cron/prune-drained-duts/success",
		"success of prune-drained-duts cron jobs",
		nil,
		field.String("instance"),
		field.Bool("success"),
	)
)
