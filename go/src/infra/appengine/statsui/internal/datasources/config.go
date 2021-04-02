// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package datasources

import (
	"gopkg.in/yaml.v2"
)

// Config is used to configure the data source client
type Config struct {
	Sources map[string]SourceConfig `yaml:"sources"`
}

// SourceConfig specifies the configuration for a single data source
type SourceConfig struct {
	Queries map[Period]string `yaml:"queries,flow"`
}

// Period that defines what time period to look at for the dates given.
type Period string

const (
	// Day specifies a time period of a day.
	Day Period = "day"
	// Week specifies a time period of a week.
	Week Period = "week"
)

func UnmarshallConfig(yamlConfig []byte) (*Config, error) {
	sources := Config{}
	err := yaml.Unmarshal(yamlConfig, &sources)
	if err != nil {
		return nil, err
	}
	return &sources, nil
}
