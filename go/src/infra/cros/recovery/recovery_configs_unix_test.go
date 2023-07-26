//go:build linux
// +build linux

// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package recovery

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"infra/cros/recovery/config/tree"
	"infra/cros/recovery/tlw"
	"infra/libs/skylab/buildbucket"
)

// Test cases for TestConfigTreeChanges
var configTreeChangesCases = []struct {
	name         string
	setupType    tlw.DUTSetupType
	taskName     buildbucket.TaskName
	treeFilename string
}{
	{
		"CROS AutoRepair",
		tlw.DUTSetupTypeCros,
		buildbucket.Recovery,
		"cros_repair.json",
	},
	{
		"Android AutoRepair",
		tlw.DUTSetupTypeAndroid,
		buildbucket.Recovery,
		"android_repair.json",
	},
	{
		"Labstation AutoRepair",
		tlw.DUTSetupTypeLabstation,
		buildbucket.Recovery,
		"labstation_repair.json",
	},
}

// Key to create Tree Files if it is not exist.
// Please keep const in state `false` to keep checks in place.
// The const can be switch to true to regenerate files.
const createTreeFileIfNotExist = false

func TestConfigTreeGeneraterOff(t *testing.T) {
	if createTreeFileIfNotExist {
		t.Errorf("TestConfigTreeGeneraterOff: please keep const `createTreeFileIfNotExist` in state `false` to avoid cases when files are regenerated!")
	}
}

func TestConfigTreeChanges(t *testing.T) {
	t.Parallel()
	for _, c := range configTreeChangesCases {
		cs := c
		t.Run(cs.name, func(t *testing.T) {
			ctx := context.Background()
			config, err := ParsedDefaultConfiguration(ctx, c.taskName, c.setupType)
			if err != nil {
				t.Errorf("TestConfigTreeChanges:%q -> fail to read configuration", cs.name)
				return
			}
			configTree := tree.ConvertConfiguration(config)
			treeBytes, err := json.MarshalIndent(configTree, "", "  ")
			if err != nil {
				t.Errorf("TestConfigTreeChanges:%q -> fail to convert configuration", cs.name)
				return
			}
			// Read config from file.
			treeFilepath := filepath.Join("config_tries", cs.treeFilename)
			if createTreeFileIfNotExist {
				if err := os.WriteFile(treeFilepath, treeBytes, 0750); err != nil {
					t.Errorf("TestConfigTreeChanges:%q -> fail to create file by request: %q", cs.name, treeFilepath)
				}
				return
			}
			if _, err := os.Stat(treeFilepath); os.IsNotExist(err) {
				t.Errorf("TestConfigTreeChanges:%q -> the tree file %q is not exist", cs.name, treeFilepath)
				return
			}
			body, err := os.ReadFile(treeFilepath)
			if err != nil {
				t.Errorf("TestConfigTreeChanges:%q -> fail to read tree file %q", cs.name, treeFilepath)
				return
			}
			if diff := cmp.Diff(body, treeBytes); diff != "" {
				t.Errorf("TestConfigTreeChanges:%q mismatch (-want +got):\n%s", cs.name, diff)
			}
		})
	}
}
