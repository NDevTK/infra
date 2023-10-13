// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package configparser implements logic to handle SuiteScheduler configuration files.
package configparser

import (
	"fmt"

	infrapb "go.chromium.org/chromiumos/infra/proto/go/testplans"
)

type (
	/*
	 * SuiteSchedulerConfigs based types
	 */

	// Hour is bounded to [0,23]
	Hour int32

	// Day is bounded to [0,13]:
	// 		Weekly will only use [0,6].
	// 		Fortnightly can use the full [0,13].
	Day int32

	TestPlanName string

	HourMap map[Hour]ConfigList

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

	LabConfig map[Board]BoardEntry
)

// BoardEntry is a slight wrapper on the infrapb Board type
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

// SuiteSchedulerConfigs represents the ADS which will be used for accessing
// SuiteScheduler configurations.
type SuiteSchedulerConfigs struct {
	// Array of all configs. Allows quick access to all configurations.
	configStore ConfigList

	// Array of all configs. Allows quick access to all new build configurations.
	newBuildStore ConfigList

	// newBuildMap stores a mapping of build target to relevant NEW_BUILD
	// configs. Allows for retrieval of configs when searching by build target.
	newBuildMap map[BuildTarget]ConfigList

	// This map provides a quick direct access option for fetching configs by name.
	configMap map[TestPlanName]*infrapb.SchedulerConfig

	// The following maps correspond to the specific set of TimedEvents the
	// configuration is of.
	dailyMap       HourMap
	weeklyMap      map[Day]HourMap
	fortnightlyMap map[Day]HourMap
}

// isDayCompliant checks the day int type to ensure that it is within the
// accepted bounds. A flag for fortnightly is required for calculation of day
// range values.
func isDayCompliant(day Day, isFortnightly bool) error {
	highBound := Day(6)

	if isFortnightly {
		highBound = Day(13)
	}

	if day < 0 || day > highBound {
		return fmt.Errorf("Day %d is not within the supported range [0,%d]", day, highBound)
	}

	return nil
}

// isHourCompliant checks the hour int type to ensure that it is within the
// accepted bounds.
func isHourCompliant(hour Hour) error {
	if hour < 0 || hour > 23 {
		return fmt.Errorf("Hour %d is not within the supported range [0,23]", hour)
	}

	return nil
}

// addConfigToNewBuildMap takes a newBuild configuration and inserts it into the
// appropriate tracking lists.
func (s *SuiteSchedulerConfigs) addConfigToNewBuildMap(config *infrapb.SchedulerConfig, lab LabConfig) error {
	// Fetch all build targets which can trigger this configuration.
	targets, err := getBuildTargets(config, lab)
	if err != nil {
		return err
	}

	for _, target := range targets {
		// Add entry if no config with this build target has been
		// ingested yet.
		if _, ok := s.newBuildMap[target]; !ok {
			s.newBuildMap[target] = ConfigList{}
		}

		// Add the pointer to the config into the tracking set.
		s.newBuildMap[target] = append(s.newBuildMap[target], config)
	}

	// Add to the array tracking all SuSch configs.
	s.configStore = append(s.configStore, config)
	s.newBuildStore = append(s.newBuildStore, config)

	// Add to the direct access map.
	s.configMap[TestPlanName(config.Name)] = config

	return nil
}

// addConfigToDailyMap takes a daily configuration and inserts it into the
// appropriate tracking lists.
func (s *SuiteSchedulerConfigs) addConfigToDailyMap(config *infrapb.SchedulerConfig) error {
	configHour := Hour(config.LaunchCriteria.Hour)
	err := isHourCompliant(configHour)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Ingesting %s encountered %s", config.Name, err))
	}

	if _, ok := s.dailyMap[configHour]; !ok {
		s.dailyMap[configHour] = ConfigList{}
	}

	s.dailyMap[configHour] = append(s.dailyMap[configHour], config)

	// Add to the array tracking all SuSch configs.
	s.configStore = append(s.configStore, config)

	// Add to the direct access map.
	s.configMap[TestPlanName(config.Name)] = config

	return nil
}

// addConfigToWeeklyMap takes a weekly configuration and inserts it into the
// appropriate tracking lists.
func (s *SuiteSchedulerConfigs) addConfigToWeeklyMap(config *infrapb.SchedulerConfig) error {
	configDay := Day(config.LaunchCriteria.Day)
	err := isDayCompliant(configDay, false)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Ingesting %s encountered %s", config.Name, err))
	}
	configHour := Hour(config.LaunchCriteria.Hour)
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
	s.configStore = append(s.configStore, config)

	// Add to the direct access map.
	s.configMap[TestPlanName(config.Name)] = config

	return nil
}

// addConfigToFortnightlyMap takes a fortnightly configuration and inserts it into the
// appropriate tracking lists.
func (s *SuiteSchedulerConfigs) addConfigToFortnightlyMap(config *infrapb.SchedulerConfig) error {
	configDay := Day(config.LaunchCriteria.Day)
	err := isDayCompliant(configDay, false)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Ingesting %s encountered %s", config.Name, err))
	}
	configHour := Hour(config.LaunchCriteria.Hour)
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
	s.configStore = append(s.configStore, config)

	// Add to the direct access map.
	s.configMap[TestPlanName(config.Name)] = config

	return nil
}

// FetchAllBewBuildConfigs returns all NEW_BUILD type configs.
func (s *SuiteSchedulerConfigs) FetchAllBewBuildConfigs() ConfigList {
	return s.newBuildStore
}

// FetchNewBuildConfigsByBuildTarget returns all NEW_BUILD configs that are
// to be triggered by a new image of the given build target.
func (s *SuiteSchedulerConfigs) FetchNewBuildConfigsByBuildTarget(target BuildTarget) (ConfigList, error) {
	if obj, ok := s.newBuildMap[target]; ok {
		return obj, nil
	} else {
		return nil, fmt.Errorf("no NEW_BUILD configs found for build target %s", target)
	}
}

// FetchAllDailyConfigs returns all DAILY type configs.
func (s *SuiteSchedulerConfigs) FetchAllDailyConfigs() ConfigList {
	tempList := ConfigList{}

	for _, list := range s.dailyMap {
		tempList = append(tempList, list...)
	}

	return tempList
}

// FetchDailyByHour returns all DAILY configs that are to be scheduled at the
// specified hour.
func (s *SuiteSchedulerConfigs) FetchDailyByHour(hour Hour) (ConfigList, error) {
	err := isHourCompliant(hour)
	if err != nil {
		return nil, err
	}

	if obj, ok := s.dailyMap[hour]; ok {
		return obj, nil
	} else {
		return nil, fmt.Errorf("no DAILY configs found at hour %d", hour)
	}
}

// FetchAllWeeklyConfigs returns all WEEKLY type configs.
func (s *SuiteSchedulerConfigs) FetchAllWeeklyConfigs() ConfigList {
	tempList := ConfigList{}

	for _, mapobj := range s.weeklyMap {
		for _, list := range mapobj {
			tempList = append(tempList, list...)
		}
	}

	return tempList
}

// FetchWeeklyByDay returns all WEEKLY configs that are to be scheduled on the
// specified DAY.
func (s *SuiteSchedulerConfigs) FetchWeeklyByDay(day Day) (ConfigList, error) {
	err := isDayCompliant(day, false)
	if err != nil {
		return nil, err
	}

	if obj, ok := s.weeklyMap[day]; ok {
		tempList := ConfigList{}

		for _, hour := range obj {
			tempList = append(tempList, hour...)
		}

		return tempList, nil
	} else {
		return nil, fmt.Errorf("no WEEKLY configs found at Day %d", day)
	}
}

// FetchWeeklyByDayHour returns all WEEKLY configs that are to be scheduled on the
// specified DAY at the given HOUR.
func (s *SuiteSchedulerConfigs) FetchWeeklyByDayHour(day Day, hour Hour) (ConfigList, error) {
	err := isDayCompliant(day, false)
	if err != nil {
		return nil, err
	}

	if _, ok := s.weeklyMap[day]; !ok {
		return nil, fmt.Errorf("no WEEKLY configs found at Day %d", day)
	}

	if list, ok := s.weeklyMap[day][hour]; ok {
		return list, nil
	} else {
		return nil, fmt.Errorf("no WEEKLY configs found at Day:Hour %d:%d", day, hour)
	}
}

// FetchAllFortnightlyConfigs returns all FORTNIGHTLY type configs.
func (s *SuiteSchedulerConfigs) FetchAllFortnightlyConfigs() ConfigList {
	tempList := ConfigList{}

	for _, mapobj := range s.fortnightlyMap {
		for _, list := range mapobj {
			tempList = append(tempList, list...)
		}
	}

	return tempList
}

// FetchFortnightlyByDay returns all FORTNIGHTLY configs that are to be scheduled on the
// specified DAY.
func (s *SuiteSchedulerConfigs) FetchFortnightlyByDay(day Day) (ConfigList, error) {
	err := isDayCompliant(day, false)
	if err != nil {
		return nil, err
	}

	if obj, ok := s.fortnightlyMap[day]; ok {
		tempList := ConfigList{}

		for _, hour := range obj {
			tempList = append(tempList, hour...)
		}

		return tempList, nil
	} else {
		return nil, fmt.Errorf("no WEEKLY configs found at Day %d", day)
	}
}

// FetchFortnightlyByDayHour returns all FORTNIGHTLY configs that are to be scheduled on the
// specified DAY at the given HOUR.
func (s *SuiteSchedulerConfigs) FetchFortnightlyByDayHour(day Day, hour Hour) (ConfigList, error) {
	err := isDayCompliant(day, false)
	if err != nil {
		return nil, err
	}

	if _, ok := s.fortnightlyMap[day]; !ok {
		return nil, fmt.Errorf("no WEEKLY configs found at Day %d", day)
	}

	if list, ok := s.fortnightlyMap[day][hour]; ok {
		return list, nil
	} else {
		return nil, fmt.Errorf("no WEEKLY configs found at Day:Hour %d:%d", day, hour)
	}
}
