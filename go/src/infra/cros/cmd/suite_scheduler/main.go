// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"

	// NOTE: These imports are specific to how infra/infra imports packages. If
	// this project moves to a stand alone service these import will need to be
	// updated to reflect ./suite_scheduler/ as the root project directory.
	"infra/cros/cmd/suite_scheduler/configparser"
	"infra/cros/cmd/suite_scheduler/ctp_request"

	"google.golang.org/protobuf/encoding/protojson"
)

const (
	suSchCfgPath  = "configparser/generated/suite_scheduler.cfg"
	labCfgPath    = "configparser/generated/lab_config.cfg"
	suSchInigPath = "configparser/generated/suite_scheduler.ini"
	labIniPath    = "configparser/generated/lab_config.ini"
)

var (
	flags  = log.Default().Flags() | log.Lshortfile
	stdout = log.New(os.Stdout, "", flags)
	stderr = log.New(os.Stderr, "", flags)
)

// SetUp fetches the .ini files from the SuSch repo.
func SetUp() error {
	err := configparser.FetchAndWriteFile(configparser.SuiteSchedulerCfgURL, suSchCfgPath)
	if err != nil {
		return err
	}

	err = configparser.FetchAndWriteFile(configparser.LabCfgURL, labCfgPath)
	if err != nil {
		return err
	}

	err = configparser.FetchAndWriteFile(configparser.SuiteSchedulerIniURL, suSchInigPath)
	if err != nil {
		return err
	}

	err = configparser.FetchAndWriteFile(configparser.LabIniURL, labIniPath)
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

	suschCfg, err := configparser.ReadLocalFile(suSchCfgPath)
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

	labCfg, err := configparser.ReadLocalFile(labCfgPath)
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

	labMap := configparser.IngestLabConfigs(labConfigs)

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

	testConfig, err := suSuchConfigs.FetchConfigByName("CFTNightly")
	if err != nil {
		stderr.Println(err)
		return 1
	}

	boards, err := configparser.GetTargetOptions(testConfig, labMap)
	if err != nil {
		stderr.Println(err)
		return 1
	}

	ctpRequests := ctp_request.BuildAllCTPRequests(testConfig, boards)

	stdout.Printf("# of CTP requests generated for %s: %d", testConfig.Name, len(ctpRequests))

	if len(ctpRequests) >= 1 {
		tempConfig := ctpRequests[0]
		str, err := protojson.Marshal(tempConfig)
		if err != nil {
			stderr.Println(err)
			return 1
		}

		var printBuffer bytes.Buffer
		err = json.Indent(&printBuffer, str, "", "\t")

		stdout.Printf("Sample Request:\n%s", string(printBuffer.Bytes()))

	}
	return 0
}

func main() {
	os.Exit(innerRun())
}
