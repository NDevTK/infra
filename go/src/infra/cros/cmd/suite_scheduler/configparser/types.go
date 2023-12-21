// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package configparser implements logic to handle SuiteScheduler configuration files.
package configparser

import (
	"fmt"

	infrapb "go.chromium.org/chromiumos/infra/proto/go/testplans"

	"infra/cros/cmd/suite_scheduler/common"
)

type (
	/*
	 * SuiteSchedulerConfigs based types
	 */

	TestPlanName string

	HourMap map[common.Hour]ConfigList

	ConfigList []*infrapb.SchedulerConfig

	/*
	 * LabConfig based types
	 */

	// These fields ensure that we aren't using magic strings elsewhere in the
	// code.

	Board   string
	Variant string
	Model   string
	// BuildTarget is in the form board(-<variant>) with the variant being optional.
	BuildTarget string

	// TargetOptions is a map of board->TargetOption to easily retrieve information.
	TargetOptions map[Board]*TargetOption
)

// TargetOption is a struct which contains all information for a targeted piece
// of hardware to be tested.
type TargetOption struct {
	Board    string
	Models   []string
	Variants []string
}

// LabConfigs is a wrapper to provide quick access to boards and models in the lab.
type LabConfigs struct {
	Models map[Model]*BoardEntry
	Boards map[Board]*BoardEntry
}

// BoardEntry is a wrapper on the infrapb Board type.
type BoardEntry struct {
	isAndroid bool
	board     *infrapb.Board
}

func (b *BoardEntry) isAndroidBoard() bool {
	return b.isAndroid
}

func (b *BoardEntry) GetBoard() *infrapb.Board {
	return b.board
}

func (b *BoardEntry) GetName() string {
	return b.board.Name
}

// SuiteSchedulerConfigs represents the ADS which will be used for accessing
// SuiteScheduler configurations.
type SuiteSchedulerConfigs struct {
	// Array of all configs. Allows quick access to all configurations.
	configList ConfigList

	// Array of all configs. Allows quick access to all new build configurations.
	newBuildList ConfigList

	// newBuildMap stores a mapping of build target to relevant NEW_BUILD
	// configs. Allows for retrieval of configs when searching by build target.
	newBuildMap map[BuildTarget]ConfigList

	// configTargets will provided a cached version of the, computationally
	// expensive to build, target options per config.
	configTargets map[string]TargetOptions

	// This map provides a quick direct access option for fetching configs by name.
	configMap map[TestPlanName]*infrapb.SchedulerConfig

	// The following maps correspond to the specific set of TimedEvents the
	// configuration is of.
	dailyMap       HourMap
	weeklyMap      map[common.Day]HourMap
	fortnightlyMap map[common.Day]HourMap
}

// addConfigToNewBuildMap takes a newBuild configuration and inserts it into the
// appropriate tracking lists.
func (s *SuiteSchedulerConfigs) addConfigToNewBuildMap(config *infrapb.SchedulerConfig, lab *LabConfigs, targetOptions TargetOptions) {

	// Fetch all build buildTargets which can trigger this configuration.
	buildTargets := GetBuildTargets(targetOptions)

	for _, target := range buildTargets {
		// Add entry if no config with this build target has been
		// ingested yet.
		if _, ok := s.newBuildMap[target]; !ok {
			s.newBuildMap[target] = ConfigList{}
		}

		// Add the pointer to the config into the tracking set.
		s.newBuildMap[target] = append(s.newBuildMap[target], config)
	}

	// Add to the array tracking all SuSch configs.
	s.configList = append(s.configList, config)
	s.newBuildList = append(s.newBuildList, config)

	// Add to the direct access map.
	s.configMap[TestPlanName(config.Name)] = config
}

// addConfigToDailyMap takes a daily configuration and inserts it into the
// appropriate tracking lists.
func (s *SuiteSchedulerConfigs) addConfigToDailyMap(config *infrapb.SchedulerConfig) error {
	configHour := common.Hour(config.LaunchCriteria.Hour)
	err := isHourCompliant(configHour)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Ingesting %s encountered %s", config.Name, err))
	}

	if _, ok := s.dailyMap[configHour]; !ok {
		s.dailyMap[configHour] = ConfigList{}
	}

	s.dailyMap[configHour] = append(s.dailyMap[configHour], config)

	// Add to the array tracking all SuSch configs.
	s.configList = append(s.configList, config)

	// Add to the direct access map.
	s.configMap[TestPlanName(config.Name)] = config

	return nil
}

// addConfigToWeeklyMap takes a weekly configuration and inserts it into the
// appropriate tracking lists.
func (s *SuiteSchedulerConfigs) addConfigToWeeklyMap(config *infrapb.SchedulerConfig) error {
	configDay := common.Day(config.LaunchCriteria.Day)
	err := isDayCompliant(configDay, false)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Ingesting %s encountered %s", config.Name, err))
	}
	configHour := common.Hour(config.LaunchCriteria.Hour)
	err = isHourCompliant(configHour)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Ingesting %s encountered %s", config.Name, err))
	}

	if _, ok := s.weeklyMap[configDay]; !ok {
		s.weeklyMap[configDay] = make(HourMap)
	}

	dayMap := s.weeklyMap[configDay]

	if _, ok := dayMap[configHour]; !ok {
		dayMap[configHour] = ConfigList{}
	}
	dayMap[configHour] = append(dayMap[configHour], config)

	// Add to the array tracking all SuSch configs.
	s.configList = append(s.configList, config)

	// Add to the direct access map.
	s.configMap[TestPlanName(config.Name)] = config

	return nil
}

// addConfigToFortnightlyMap takes a fortnightly configuration and inserts it into the
// appropriate tracking lists.
func (s *SuiteSchedulerConfigs) addConfigToFortnightlyMap(config *infrapb.SchedulerConfig) error {
	configDay := common.Day(config.LaunchCriteria.Day)
	err := isDayCompliant(configDay, true)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Ingesting %s encountered %s", config.Name, err))
	}
	configHour := common.Hour(config.LaunchCriteria.Hour)
	err = isHourCompliant(configHour)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Ingesting %s encountered %s", config.Name, err))
	}

	if _, ok := s.fortnightlyMap[configDay]; !ok {
		s.fortnightlyMap[configDay] = make(HourMap)
	}

	dayMap := s.fortnightlyMap[configDay]

	if _, ok := dayMap[configHour]; !ok {
		dayMap[configHour] = ConfigList{}
	}
	dayMap[configHour] = append(dayMap[configHour], config)

	// Add to the array tracking all SuSch configs.
	s.configList = append(s.configList, config)

	// Add to the direct access map.
	s.configMap[TestPlanName(config.Name)] = config

	return nil
}
