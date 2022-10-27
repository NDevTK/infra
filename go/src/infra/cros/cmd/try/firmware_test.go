// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"context"
	"strings"
	"testing"

	"infra/cros/internal/assert"
	"infra/cros/internal/cmd"
)

// TestDoesFWBranchHaveBuilder tests doesFWBranchHaveBuilder.
func TestDoesFWBranchHaveBuilder(t *testing.T) {
	t.Parallel()
	const (
		eveBranch    = "firmware-eve-9584.B"
		gruntBranch  = "firmware-grunt-11031.B"
		namiBranch   = "firmware-nami-10775.B"
		eveBuilder   = "chromeos/firmware/firmware-eve-9584.B-branch"
		gruntBuilder = "chromeos/firmware/firmware-grunt-11031.B-branch"
		namiBuilder  = "chromeos/firmware/firmware-nami-10775.B-branch"
	)
	f := firmwareRun{
		tryRunBase: tryRunBase{
			cmdRunner: cmd.FakeCommandRunner{
				ExpectedCmd: []string{"bb", "builders", "chromeos/firmware"},
				Stdout:      strings.Join([]string{eveBuilder, gruntBuilder}, "\n"),
			},
		},
	}
	ctx := context.Background()
	for i, tc := range []struct {
		branch     string
		production bool
		expected   bool
	}{
		{eveBranch, true, true},
		{namiBranch, true, false},
	} {
		builderExists, err := f.doesFWBranchHaveBuilder(ctx, tc.branch, !tc.production)
		if err != nil {
			t.Errorf("#%d: Unexpected error calling doesFWBranchHaveBuilder: %+v", i, err)
		}
		if builderExists != tc.expected {
			t.Errorf("#%d: Unexpected response from doesFWBranchHaveBuilder: got %v; want %v", i, builderExists, tc.expected)
		}
	}
}

// TestGetFWBuilderFullName tests getFWBuilderFullName.
func TestGetFWBuilderFullName(t *testing.T) {
	t.Parallel()
	const (
		eveBranch         = "firmware-eve-9584.B"
		eveBuilder        = "chromeos/firmware/firmware-eve-9584.B-branch"
		eveStagingBuilder = "chromeos/staging/staging-firmware-eve-9584.B-branch"
	)
	assert.StringsEqual(t, getFWBuilderFullName(eveBranch, false), eveBuilder)
	assert.StringsEqual(t, getFWBuilderFullName(eveBranch, true), eveStagingBuilder)
}

func TestValidate_firmwareRun(t *testing.T) {
	t.Parallel()
	const (
		eveFWBuilder  = "chromeos/firmware/firmware-eve-9584.B-branch"
		eveFWBranch   = "firmware-eve-9584.B"
		gruntFWBranch = "firmware-grunt-11031.B"
		releaseBranch = "release-R106.15054.B"
	)
	ctx := context.Background()

	// Test the good workflow
	f := firmwareRun{
		tryRunBase: tryRunBase{
			branch:     eveFWBranch,
			production: true,
			cmdRunner: cmd.FakeCommandRunner{
				ExpectedCmd: []string{"bb", "builders", "chromeos/firmware"},
				Stdout:      eveFWBuilder,
			},
		},
	}
	assert.NilError(t, f.validate(ctx))

	// No branch provided
	f.tryRunBase.branch = ""
	assert.NonNilError(t, f.validate(ctx))

	// Non-firmware branch
	f.tryRunBase.branch = releaseBranch
	assert.NonNilError(t, f.validate(ctx))

	// Firmware branch that doesn't have a builder
	f.tryRunBase.branch = gruntFWBranch
	assert.NonNilError(t, f.validate(ctx))

	// Patch set provided for production builder
	f.tryRunBase.branch = eveFWBranch
	f.tryRunBase.patches = []string{"crrev.com/c/1234567"}
	assert.NonNilError(t, f.validate(ctx))

	// Patch set provided for staging builder
	f.tryRunBase.production = false
	f.cmdRunner = cmd.FakeCommandRunner{
		ExpectedCmd: []string{"bb", "builders", "chromeos/staging"},
		Stdout:      "chromeos/staging/staging-firmware-eve-9584.B-branch",
	}
	assert.NilError(t, f.validate(ctx))
}
