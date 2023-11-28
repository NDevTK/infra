// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package configparser implements logic to handle SuiteScheduler configuration files.
package configparser

import (
	"fmt"

	infrapb "go.chromium.org/chromiumos/infra/proto/go/testplans"
	"google.golang.org/protobuf/encoding/protojson"
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
		dailyMap:       map[Hour]ConfigList{},
		weeklyMap:      map[Day]HourMap{},
		fortnightlyMap: map[Day]HourMap{},
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
