// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package configparser implements logic to handle SuiteScheduler configuration files.
package configparser

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"

	infrapb "go.chromium.org/chromiumos/infra/proto/go/testplans"

	"infra/cros/cmd/suite_scheduler/common"
)

// IngestSuSchConfigs takes in all of the raw Suite Scheduler and Lab configs and ingests
// them into a more usage structure.
func IngestSuSchConfigs(configs ConfigList, lab *LabConfigs) (*SuiteSchedulerConfigs, error) {
	configDS := &SuiteSchedulerConfigs{
		configList:     ConfigList{},
		newBuildList:   []*infrapb.SchedulerConfig{},
		newBuildMap:    map[BuildTarget]ConfigList{},
		configTargets:  map[string]TargetOptions{},
		configMap:      map[TestPlanName]*infrapb.SchedulerConfig{},
		dailyMap:       map[common.Hour]ConfigList{},
		weeklyMap:      map[common.Day]HourMap{},
		fortnightlyMap: map[common.Day]HourMap{},
	}

	for _, config := range configs {

		targetOptions, err := GetTargetOptions(config, lab)
		if err != nil {
			return nil, err
		}

		// Cache the calculated target options.
		configDS.configTargets[config.Name] = targetOptions

		// Add the configuration to the map which holds stores information on
		// its LaunchProfile type
		switch config.LaunchCriteria.LaunchProfile {
		case infrapb.SchedulerConfig_LaunchCriteria_NEW_BUILD:
			configDS.addConfigToNewBuildMap(config, lab, targetOptions)
		case infrapb.SchedulerConfig_LaunchCriteria_DAILY:
			err := configDS.addConfigToDailyMap(config)
			if err != nil {
				return nil, err
			}
		case infrapb.SchedulerConfig_LaunchCriteria_WEEKLY:
			err := configDS.addConfigToWeeklyMap(config)
			if err != nil {
				return nil, err
			}
		case infrapb.SchedulerConfig_LaunchCriteria_FORTNIGHTLY:
			err := configDS.addConfigToFortnightlyMap(config)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unsupported or unknown launch profile encountered in config %s", config.Name)
		}
	}

	return configDS, nil
}

// IngestLabConfigs takes in all of the raw Lab configs and ingests
// them into a more usage structure.
func IngestLabConfigs(labConfig *infrapb.LabConfig) *LabConfigs {
	tempConfig := &LabConfigs{
		Models: map[Model]*BoardEntry{},
		Boards: map[Board]*BoardEntry{},
	}

	for _, board := range labConfig.Boards {
		entry := &BoardEntry{
			isAndroid: false,
			board:     board,
		}
		tempConfig.Boards[Board(board.Name)] = entry

		for _, model := range board.Models {
			tempConfig.Models[Model(model)] = entry
		}
	}

	for _, board := range labConfig.AndroidBoards {
		entry := &BoardEntry{
			isAndroid: true,
			board:     board,
		}
		tempConfig.Boards[Board(board.Name)] = entry

		for _, model := range board.Models {
			tempConfig.Models[Model(model)] = entry
		}
	}

	return tempConfig
}

// BytesToLabProto takes a JSON formatted string and transforms it into an
// infrapb.LabConfig object.
func BytesToLabProto(configsBuffer []byte) (*infrapb.LabConfig, error) {
	configs := &infrapb.LabConfig{}

	err := protojson.Unmarshal(configsBuffer, configs)
	if err != nil {
		return nil, err
	}

	return configs, nil
}

// BytesToSchedulerProto takes a JSON formatted string and transforms it into an
// infrapb.SchedulerCfg object.
func BytesToSchedulerProto(configsBuffer []byte) (*infrapb.SchedulerCfg, error) {
	configs := &infrapb.SchedulerCfg{}

	err := protojson.Unmarshal(configsBuffer, configs)
	if err != nil {
		return nil, err
	}

	return configs, nil
}

// FetchLabConfigs fetches and ingests the lab configs. It will
// determine where to read the configs from based on the user provided flags.
func FetchLabConfigs(path string) (*LabConfigs, error) {
	var err error
	var labBytes []byte

	// If a file path was passed in for the Lab then parse that file. If not
	// then fetch the LabConfig from the ToT .cfg and ingest it in memory.
	if path != common.DefaultString {
		labBytes, err = common.ReadLocalFile(path)
		if err != nil {
			return nil, err
		}
	} else {
		labBytes, err = common.FetchFileFromURL(common.LabCfgURL)
		if err != nil {
			return nil, err
		}

	}

	labProto, err := BytesToLabProto(labBytes)
	if err != nil {
		return nil, err
	}

	labConfigs := IngestLabConfigs(labProto)

	return labConfigs, nil
}

// FetchSchedulerConfigs fetches and ingests the SuiteScheduler configs. It will
// determine where to read the configs from based on the user provided flags.
func FetchSchedulerConfigs(path string, labConfigs *LabConfigs) (*SuiteSchedulerConfigs, error) {
	var err error
	var schedulerBytes []byte

	// If a file path was passed in for the ScheduleConfigs then parse that file. If not
	// then fetch the SuiteSchedulerConfigs from the ToT .cfg and ingest it in memory.
	if path != common.DefaultString {
		schedulerBytes, err = common.ReadLocalFile(path)
		if err != nil {
			return nil, err
		}

	} else {
		schedulerBytes, err = common.FetchFileFromURL(common.SuiteSchedulerCfgURL)
		if err != nil {
			return nil, err
		}
		err := common.WriteToFile("configparser/generated/suite_scheduler.ini", schedulerBytes)
		if err != nil {
			return nil, err
		}

	}

	// Convert from []byte to a usable object type.
	scheduleProto, err := BytesToSchedulerProto(schedulerBytes)
	if err != nil {
		return nil, err
	}

	// Ingest the configs into a data structure which easier and more efficient
	// to search.
	schedulerConfigs, err := IngestSuSchConfigs(scheduleProto.Configs, labConfigs)
	if err != nil {
		return nil, err
	}

	return schedulerConfigs, nil
}
