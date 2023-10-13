// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"log"
	"os"

	// NOTE: These imports are specific to how infra/infra imports packages. If
	// this project moves to a stand alone service these import will need to be
	// updated to reflect ./suite_scheduler/ as the root project directory.
	"infra/cros/cmd/suite_scheduler/configparser"
)

var (
	flags  = log.Default().Flags() | log.Lshortfile
	stdout = log.New(os.Stdout, "", flags)
	stderr = log.New(os.Stderr, "", flags)

	suSchIniPath = "configparser/fetched/suite_scheduler.ini"
	labIniPath   = "configparser/fetched/lab_config.ini"

	suSchCFG = "configparser/cfgs/suite_scheduler.cfg"
	labCFG   = "configparser/cfgs/lab_config.cfg"
)

// SetUp fetches the .ini files from the SuSch repo.
func SetUp() error {
	err := configparser.FetchAndWriteFile(configparser.SuiteSchedulerCfgURL, suSchIniPath)
	if err != nil {
		return err
	}

	err = configparser.FetchAndWriteFile(configparser.LabCfgURL, labIniPath)
	if err != nil {
		return err
	}

	return nil
}

// innerRun wraps the business logic of the application.
func innerRun() int {
	err := SetUp()
	if err != nil {
		stderr.Println(err)
		return 1
	}

	suschCfg, err := configparser.ReadLocalFile(suSchCFG)
	if err != nil {
		stderr.Println(err)
		return 1
	}

	schedulerConfigs, err := configparser.StringToSchedulerProto(suschCfg)
	if err != nil {
		stderr.Println(err)
		return 1
	}

	stdout.Printf("# of configs: %d\n", len(schedulerConfigs.Configs))

	labCfg, err := configparser.ReadLocalFile(labCFG)
	if err != nil {
		stderr.Println(err)
		return 1
	}

	labConfigs, err := configparser.StringToLabProto(labCfg)
	if err != nil {
		stderr.Println(err)
		return 1
	}
	stdout.Printf("# of boards: %d\n", len(labConfigs.Boards))
	stdout.Printf("# of Android boards: %d\n", len(labConfigs.AndroidBoards))

	labMap, err := configparser.IngestLabConfigs(labConfigs)
	if err != nil {
		stderr.Println(err)
		return 1
	}

	suSuchConfigs, err := configparser.IngestSuSchConfigs(schedulerConfigs.Configs, labMap)
	if err != nil {
		stderr.Println(err)
		return 1
	}

	configList, err := suSuchConfigs.FetchNewBuildConfigsByBuildTarget(configparser.BuildTarget("brya-kernelnext"))
	if err != nil {
		stderr.Println(err)
		return 1
	}
	stdout.Printf("# of byra-kernelnext targeting builds: %d\n", len(configList))

	configList = suSuchConfigs.FetchAllBewBuildConfigs()
	stdout.Printf("# of new build configs: %d\n", len(configList))
	configList = suSuchConfigs.FetchAllDailyConfigs()
	stdout.Printf("# of daily configs: %d\n", len(configList))
	configList = suSuchConfigs.FetchAllWeeklyConfigs()
	stdout.Printf("# of Weekly configs: %d\n", len(configList))
	configList = suSuchConfigs.FetchAllFortnightlyConfigs()
	stdout.Printf("# of Fortnightly configs: %d\n", len(configList))

	return 0
}

func main() {
	os.Exit(innerRun())
}
