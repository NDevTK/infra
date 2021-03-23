// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package datasources

import (
	"gopkg.in/yaml.v2"
	"infra/appengine/statsui/internal/model"
)

type Config struct {
	Sources map[string]SourceConfig `yaml:"sources"`
}

type SourceConfig struct {
	Queries map[model.Period]string `yaml:"queries,flow"`
}

func LoadConfig(yamlConfig []byte) (*Config, error) {
	sources := Config{}
	err := yaml.Unmarshal(yamlConfig, &sources)
	if err != nil {
		return nil, err
	}
	return &sources, nil
}
