// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package configparser implements logic to handle SuiteScheduler configuration files.
package configparser

import (
	"fmt"

	infrapb "go.chromium.org/chromiumos/infra/proto/go/testplans"
)

// getBuildTargets forms a list of all BuildTarget tracked within the SuSch config.
func getBuildTargets(config *infrapb.SchedulerConfig, lab LabConfig) ([]BuildTarget, error) {
	targets := []BuildTarget{}

	excludeVariantsMap := make(map[Board]map[Variant]bool)

	for _, excludeConfig := range config.TargetOptions.ExcludeVariants {
		if _, ok := excludeVariantsMap[Board(excludeConfig.Board)]; !ok {
			excludeVariantsMap[Board(excludeConfig.Board)] = make(map[Variant]bool)
		}
		excludeVariantsMap[Board(excludeConfig.Board)][Variant(excludeConfig.Variant)] = true
	}

	for _, board := range config.TargetOptions.BoardsList {
		tempTargets := []BuildTarget{}

		if _, ok := lab[Board(board)]; !ok {
			return nil, fmt.Errorf("Config %s target board which isn't present in the lab config", config.Name)
		}

		labEntry := lab[Board(board)]

		tempTargets = append(tempTargets, BuildTarget(board))

		// If variants are being skipped then continue onto the next board option.
		if config.TargetOptions.SkipVariants {
			continue
		}

		// If no exclude variant check is in place then add the variant to
		// the target options.
		for _, variant := range labEntry.board.Variants {
			// No variants on this board are excluded.
			if _, ok := excludeVariantsMap[Board(board)]; !ok {
				tempTargets = append(tempTargets, BuildTarget(board+variant))
				continue
			}

			// No exclude variant config found, add variant to targeted options.
			if _, ok := excludeVariantsMap[Board(board)][Variant(variant)]; !ok {
				tempTargets = append(tempTargets, BuildTarget(board+variant))
			}
		}

		targets = append(targets, tempTargets...)
	}

	return targets, nil
}

// IngestSuSchConfigs takes in all of the raw Suite Scheduler and Lab configs and ingests
// them into a more usage structure.
func IngestSuSchConfigs(configs ConfigList, lab LabConfig) (*SuiteSchedulerConfigs, error) {
	configDS := &SuiteSchedulerConfigs{
		configStore:    ConfigList{},
		newBuildMap:    make(map[BuildTarget]ConfigList),
		configMap:      make(map[TestPlanName]*infrapb.SchedulerConfig),
		dailyMap:       make(map[Hour]ConfigList),
		weeklyMap:      make(map[Day]HourMap),
		fortnightlyMap: make(map[Day]HourMap),
	}

	for _, config := range configs {
		// Add the configuration to the map which holds stores information on
		// its LaunchProfile type
		switch config.LaunchCriteria.LaunchProfile {
		case infrapb.SchedulerConfig_LaunchCriteria_NEW_BUILD:
			err := configDS.addConfigToNewBuildMap(config, lab)
			if err != nil {
				return nil, err
			}
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
func IngestLabConfigs(labConfig *infrapb.LabConfig) (LabConfig, error) {
	tempConfig := make(LabConfig)

	for _, board := range labConfig.Boards {
		tempConfig[Board(board.Name)] = BoardEntry{
			isAndroid: false,
			board:     board,
		}
	}

	for _, board := range labConfig.AndroidBoards {
		tempConfig[Board(board.Name)] = BoardEntry{
			isAndroid: true,
			board:     board,
		}
	}

	return tempConfig, nil
}
