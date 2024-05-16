//go:build linux
// +build linux

// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package recovery

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"

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
		"cros_repair",
	},
	{
		"CROS Deploy",
		tlw.DUTSetupTypeCros,
		buildbucket.Deploy,
		"cros_deploy",
	},
	{
		"CROS Audit RPM",
		tlw.DUTSetupTypeCros,
		buildbucket.AuditRPM,
		"cros_audit_rpm",
	},
	{
		"CROS Audit Storage",
		tlw.DUTSetupTypeCros,
		buildbucket.AuditStorage,
		"cros_audit_storage",
	},
	{
		"CROS Audit USB stick",
		tlw.DUTSetupTypeCros,
		buildbucket.AuditUSB,
		"cros_audit_usb_stick",
	},
	{
		"Android AutoRepair",
		tlw.DUTSetupTypeAndroid,
		buildbucket.Recovery,
		"android_repair",
	},
	{
		"Android Deploy",
		tlw.DUTSetupTypeAndroid,
		buildbucket.Deploy,
		"android_deploy",
	},
	{
		"Labstation AutoRepair",
		tlw.DUTSetupTypeLabstation,
		buildbucket.Recovery,
		"labstation_repair",
	},
	{
		"Labstation Deploy",
		tlw.DUTSetupTypeLabstation,
		buildbucket.Deploy,
		"labstation_deploy",
	},
	{
		"Chrome Browser Repair",
		tlw.DUTSetupTypeCrosBrowser,
		buildbucket.Recovery,
		"browser_repair",
	},
	{
		"Chrome Browser Deploy",
		tlw.DUTSetupTypeCrosBrowser,
		buildbucket.Deploy,
		"browser_deploy",
	},
	{
		"Browser Audit RPM",
		tlw.DUTSetupTypeCrosBrowser,
		buildbucket.AuditRPM,
		"browser_cros_audit_rpm",
	},
	{
		"Browser Audit Storage",
		tlw.DUTSetupTypeCrosBrowser,
		buildbucket.AuditStorage,
		"browser_cros_audit_storage",
	},
	{
		"Browser Audit USB stick",
		tlw.DUTSetupTypeCrosBrowser,
		buildbucket.AuditUSB,
		"browser_cros_audit_usb_stick",
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
			t.Parallel()
			testRun_ConfigTreeChanges(t, cs.name, cs.setupType, cs.taskName, cs.treeFilename, true)
		})
		t.Run(cs.name, func(t *testing.T) {
			t.Parallel()
			testRun_ConfigTreeChanges(t, cs.name, cs.setupType, cs.taskName, cs.treeFilename, false)
		})
	}
}

func testRun_ConfigTreeChanges(t *testing.T, name string, setupType tlw.DUTSetupType, taskName buildbucket.TaskName, treeFilename string, shortVersion bool) {
	ctx := context.Background()
	config, err := ParsedDefaultConfiguration(ctx, taskName, setupType)
	if err != nil {
		t.Errorf("TestConfigTreeChanges:%q -> fail to read configuration", name)
		return
	}
	configTree := tree.ConvertConfiguration(config, shortVersion)
	treeBytes, err := yaml.Marshal(configTree)
	if err != nil {
		t.Errorf("TestConfigTreeChanges:%q -> fail to convert configuration", name)
		return
	}
	// Read config from file.
	filename := treeFilename
	if shortVersion {
		filename += "_short"
	}
	treeFilepath := filepath.Join("config_trees", filename+".yaml")
	if createTreeFileIfNotExist {
		if err := os.WriteFile(treeFilepath, treeBytes, 0644); err != nil {
			t.Errorf("TestConfigTreeChanges:%q -> fail to create file by request: %q", name, treeFilepath)
		}
		return
	}
	if _, err := os.Stat(treeFilepath); os.IsNotExist(err) {
		t.Errorf("TestConfigTreeChanges:%q -> the tree file %q is not exist", name, treeFilepath)
		return
	}
	body, err := os.ReadFile(treeFilepath)
	if err != nil {
		t.Errorf("TestConfigTreeChanges:%q -> fail to read tree file %q", name, treeFilepath)
		return
	}
	if diff := cmp.Diff(body, treeBytes); diff != "" {
		t.Errorf("TestConfigTreeChanges:%q mismatch (-want +got):\n%s", name, diff)
	}
}
