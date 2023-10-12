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
		"CROS Deploy",
		tlw.DUTSetupTypeCros,
		buildbucket.Deploy,
		"cros_deploy.json",
	},
	{
		"CROS Audit RPM",
		tlw.DUTSetupTypeCros,
		buildbucket.AuditRPM,
		"cros_audit_rpm.json",
	},
	{
		"CROS Audit Storage",
		tlw.DUTSetupTypeCros,
		buildbucket.AuditStorage,
		"cros_audit_storage.json",
	},
	{
		"CROS Audit USB stick",
		tlw.DUTSetupTypeCros,
		buildbucket.AuditUSB,
		"cros_audit_usb_stick.json",
	},
	{
		"Android AutoRepair",
		tlw.DUTSetupTypeAndroid,
		buildbucket.Recovery,
		"android_repair.json",
	},
	{
		"Android Deploy",
		tlw.DUTSetupTypeAndroid,
		buildbucket.Deploy,
		"android_deploy.json",
	},
	{
		"Labstation AutoRepair",
		tlw.DUTSetupTypeLabstation,
		buildbucket.Recovery,
		"labstation_repair.json",
	},
	{
		"Labstation Deploy",
		tlw.DUTSetupTypeLabstation,
		buildbucket.Deploy,
		"labstation_deploy.json",
	},
	{
		"Chrome Browser Repair",
		tlw.DUTSetupTypeCrosBrowser,
		buildbucket.Recovery,
		"browser_repair.json",
	},
	{
		"Chrome Browser Deploy",
		tlw.DUTSetupTypeCrosBrowser,
		buildbucket.Deploy,
		"browser_deploy.json",
	},
}

// CreateTreeFileIfNotExist controls whether we create tree files.
// When set to true, tree files are created in their appropriate location if
// no such tree files already exist.
var createTreeFileIfNotExist = os.Getenv("RECOVERY_GENERATE_CONFIG_TREE") != ""

func TestConfigTreeGeneraterOff(t *testing.T) {
	if createTreeFileIfNotExist {
		t.Errorf("TestConfigTreeGeneraterOff: please keep const `createTreeFileIfNotExist` in state `false` to avoid cases when files are regenerated!")
	}
}

// TestConfigTreeChanges checks that configs trees matches to current configs.
// Please run `make trees` in recovery folder to regenerate tree files.
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
			treeFilepath := filepath.Join("config_trees", cs.treeFilename)
			if createTreeFileIfNotExist {
				if err := os.WriteFile(treeFilepath, treeBytes, 0644); err != nil {
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
