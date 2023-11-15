// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configparser

import (
	"fmt"

	infrapb "go.chromium.org/chromiumos/infra/proto/go/testplans"
)

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

// FetchConfigByName returns the config with the name provided. If it does not
// exist then an error is returned.
func (s *SuiteSchedulerConfigs) FetchConfigByName(name string) (*infrapb.SchedulerConfig, error) {
	if val, ok := s.configMap[TestPlanName(name)]; ok {
		return val, nil
	}

	return nil, fmt.Errorf("no config found with name %s", name)
}
