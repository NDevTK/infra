// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package configparser implements logic to handle SuiteScheduler configuration
// files.
package configparser

import (
	"io"
	"os"

	infrapb "go.chromium.org/chromiumos/infra/proto/go/testplans"
	"google.golang.org/protobuf/encoding/protojson"
)

// StringToLabProto takes a JSON formatted string and transforms it into an
// infrapb.LabConfig object.
func StringToLabProto(configsBuffer []byte) (*infrapb.LabConfig, error) {
	configs := &infrapb.LabConfig{}

	err := protojson.Unmarshal(configsBuffer, configs)
	if err != nil {
		return nil, err
	}

	return configs, nil
}

// StringToSchedulerProto takes a JSON formatted string and transforms it into an
// infrapb.SchedulerCfg object.
func StringToSchedulerProto(configsBuffer []byte) (*infrapb.SchedulerCfg, error) {
	configs := &infrapb.SchedulerCfg{}

	err := protojson.Unmarshal(configsBuffer, configs)
	if err != nil {
		return nil, err
	}

	return configs, nil
}

// ReadLocalFile reads a file at the given path into memory and returns it's contents.
func ReadLocalFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	err = file.Close()
	return data, err
}
