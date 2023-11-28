// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package common has utilities that are not context specific and can be used by
// all packages.
package common

import (
	"time"
)

const (
	DefaultHoursAhead = -1 * time.Hour
	DefaultInt64      = int64(-1)
	DefaultString     = ""

	// This is the value mapping that legacy SuiteScheduler uses for days in
	// weekly and Fortnightly configs.

	Monday                     = 0
	Tuesday                    = 1
	Wednesday                  = 2
	Thursday                   = 3
	Friday                     = 4
	Saturday                   = 5
	Sunday                     = 6
	FortnightlySecondMonday    = 7
	FortnightlySecondTuesday   = 8
	FortnightlySecondWednesday = 9
	FortnightlySecondThursday  = 10
	FortnightlySecondFriday    = 11
	FortnightlySecondSaturday  = 12
	FortnightlySecondSunday    = 13

	SuiteSchedulerCfgURL = "https://chromium.googlesource.com/chromiumos/infra/suite_scheduler/+/refs/heads/main/generated_configs/suite_scheduler.cfg?format=text"
	LabCfgURL            = "https://chromium.googlesource.com/chromiumos/infra/suite_scheduler/+/refs/heads/main/generated_configs/lab_config.cfg?format=text"
	SuiteSchedulerIniURL = "https://chromium.googlesource.com/chromiumos/infra/suite_scheduler/+/refs/heads/main/generated_configs/suite_scheduler.ini?format=text"
	LabIniURL            = "https://chromium.googlesource.com/chromiumos/infra/suite_scheduler/+/refs/heads/main/generated_configs/lab_config.ini?format=text"

	SuSchCfgPath = "configparser/generated/suite_scheduler.cfg"
	LabCfgPath   = "configparser/generated/lab_config.cfg"
	SuSchIniPath = "configparser/generated/suite_scheduler.ini"
	LabIniPath   = "configparser/generated/lab_config.ini"
)
