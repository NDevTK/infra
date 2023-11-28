// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configparser

import (
	"fmt"
	"infra/cros/cmd/suite_scheduler/common"

	infrapb "go.chromium.org/chromiumos/infra/proto/go/testplans"
)

// FetchAllConfigs returns all configs.
func (s *SuiteSchedulerConfigs) FetchAllConfigs() ConfigList {
	return s.configList
}

// FetchConfigTargetOptions returns all target options.
func (s *SuiteSchedulerConfigs) FetchConfigTargetOptions(configName string) (TargetOptions, error) {
	targetOptions, ok := s.configTargets[configName]
	if !ok {
		return nil, fmt.Errorf("target options for config %s not found", configName)
	}
	return targetOptions, nil
}

// FetchAllNewBuildConfigs returns all NEW_BUILD type configs.
func (s *SuiteSchedulerConfigs) FetchAllNewBuildConfigs() ConfigList {
	return s.newBuildList
}

// FetchNewBuildConfigsByBuildTarget returns all NEW_BUILD configs that are
// to be triggered by a new image of the given build target.
func (s *SuiteSchedulerConfigs) FetchNewBuildConfigsByBuildTarget(target BuildTarget) ConfigList {
	if obj, ok := s.newBuildMap[target]; ok {
		return obj
	}
	return nil
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
	}
	return nil, nil
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
		return nil, nil
	}

	if list, ok := s.weeklyMap[day][hour]; ok {
		return list, nil
	}

	return nil, nil
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
	}
	return nil, nil
}

// FetchFortnightlyByDayHour returns all FORTNIGHTLY configs that are to be scheduled on the
// specified DAY at the given HOUR.
func (s *SuiteSchedulerConfigs) FetchFortnightlyByDayHour(day Day, hour Hour) (ConfigList, error) {
	err := isDayCompliant(day, false)
	if err != nil {
		return nil, err
	}

	if _, ok := s.fortnightlyMap[day]; !ok {
		return nil, nil
	}

	if list, ok := s.fortnightlyMap[day][hour]; ok {
		return list, nil
	}
	return nil, nil
}

// FetchConfigByName returns the config with the name provided. If it does not
// exist then an error is returned.
func (s *SuiteSchedulerConfigs) FetchConfigByName(name string) *infrapb.SchedulerConfig {
	if val, ok := s.configMap[TestPlanName(name)]; ok {
		return val
	}

	return nil
}

// ValidateHoursAheadArgs will check that all of the arguments are within the
// specified bounds that can be worked with.
func ValidateHoursAheadArgs(startHour Hour, startDay Day, hoursAhead int64, isFortnightly bool) error {
	// Validate that all input values fit within the expected bounds.
	if hoursAhead < 0 {
		return fmt.Errorf("hours head must be a positive value, %d was given", hoursAhead)
	}

	if err := isHourCompliant(startHour); err != nil {
		return err
	}

	// This check allows the same function to be used by for the daily configs
	// function as long as it sends over the default int64 value stored as a
	// constant.
	if startDay != Day(common.DefaultInt64) {
		if err := isDayCompliant(startDay, isFortnightly); err != nil {
			return err
		}

	}

	return nil
}

func (s *SuiteSchedulerConfigs) FetchAllTargetOptions() map[string]TargetOptions {
	return s.configTargets
}
