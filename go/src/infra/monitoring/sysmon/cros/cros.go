// Copyright (c) 2016 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"os/user"
	"path/filepath"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/tsmon"
	"go.chromium.org/luci/common/tsmon/field"
	"go.chromium.org/luci/common/tsmon/metric"
	"go.chromium.org/luci/common/tsmon/types"
	"golang.org/x/net/context"
)

var (
	cpuTemp = metric.NewFloat("dev/cros/cpu/temperature",
		"device CPU temperature in °C",
		&types.MetricMetadata{Units: types.DegreeCelsiusUnit},
		field.String("device_id"))
	battLevel = metric.NewFloat("dev/cros/battery/level",
		"percentage of energy left in battery",
		nil,
		field.String("device_id"))
	devOSVersion = metric.NewString("dev/cros/os/version",
		"CrOS version of the DUT",
		nil,
		field.String("device_id"))
	devOS = metric.NewString("dev/cros/os/name",
		"OS of the device",
		nil,
		field.String("device_id"))
	memFree = metric.NewInt("dev/cros/memory/free",
		"Available memory in kb",
		&types.MetricMetadata{Units: types.Kibibytes},
		field.String("device_id"))
	memTotal = metric.NewInt("dev/cros/memory/total",
		"Total memory in kb",
		&types.MetricMetadata{Units: types.Kibibytes},
		field.String("device_id"))

	allMetrics = []types.Metric{
		cpuTemp,
		battLevel,
		devOSVersion,
		devOS,
		memFree,
		memTotal,
	}
)

// Register adds tsmon callbacks to set metrics
func Register() {
	tsmon.RegisterGlobalCallback(func(c context.Context) {
		usr, err := user.Current()
		if err != nil {
			logging.Errorf(c, "Failed to get current user: %s",
				err)
		} else if err = update(c, usr.HomeDir); err != nil {
			logging.Errorf(c, "Failed to update DUT metrics: %s",
				err)
		}
	}, allMetrics...)
}

func update(c context.Context, usrHome string) (err error) {
	allFiles, err := filepath.Glob(
		filepath.Join(usrHome, fileGlob))
	if err != nil {
		return
	}
	if len(allFiles) == 0 {
		// This is usual case in most machines. So don't log an error
		// message.
		return
	}
	var lastErr error
	for _, filePath := range allFiles {
		statusFile, err := loadfile(c, filePath)
		if err != nil {
			logging.Errorf(c, "Failed to load file %s. %s",
				filePath, err)
			lastErr = err
			continue
		} else {
			updateFromFile(c, statusFile)
		}
	}
	err = lastErr
	return
}

func updateFromFile(c context.Context, f deviceStatusFile) {
	for name, d := range f.Devices {
		cpuTempValue := d.GetCPUTemp()
		if cpuTempValue != nil {
			cpuTemp.Set(c, *cpuTempValue, name)
		}
		battLevel.Set(c, d.Battery.Level, name)
		devOSVersion.Set(c, d.OSversion, name)
		devOS.Set(c, d.OS, name)
		memFree.Set(c, d.Mem.Avail, name)
		memTotal.Set(c, d.Mem.Total, name)
	}
}
