// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package stableversion

import (
	"fmt"
	"regexp"
)

// FirmwareImageResult is the result of parsing a firmware image.
type FirmwareImageResult struct {
	Platform     string
	Release      int
	Tip          int
	Branch       int
	BranchBranch int
}

// GetPlatform returns the platform or ""
func (r *FirmwareImageResult) GetPlatform() string {
	if r == nil {
		return ""
	}
	return r.Platform
}

// GetRelease returns the release or 0
func (r *FirmwareImageResult) GetRelease() int {
	if r == nil {
		return 0
	}
	return r.Release
}

// GetTip returns the platform or 0
func (r *FirmwareImageResult) GetTip() int {
	if r == nil {
		return 0
	}
	return r.Tip
}

// GetBranch returns the platform or 0
func (r *FirmwareImageResult) GetBranch() int {
	if r == nil {
		return 0
	}
	return r.Branch
}

// GetBranchBranch returns the platform or 0
func (r *FirmwareImageResult) GetBranchBranch() int {
	if r == nil {
		return 0
	}
	return r.BranchBranch
}

// capture groups:
// platform, release, tip, branch, branchbranch
var firmwareImagePattern *regexp.Regexp = regexp.MustCompile(`\A(?P<platform>[A-Za-z0-9_]+)-firmware/R(?P<release>[0-9]+)-(?P<tip>[0-9]+)\.(?P<branch>[0-9]+)\.(?P<branchbranch>[0-9]+)\z`)

// ParseFirmwareImage takes a version string and extracts version info
func ParseFirmwareImage(fv string) (*FirmwareImageResult, error) {
	if fv == "" {
		return nil, fmt.Errorf("empty firmware version string is invalid")
	}
	if firmwareImagePattern.FindString(fv) == "" {
		return nil, fmt.Errorf("firmware version string is not valid")
	}
	m, err := findMatchMap(firmwareImagePattern, fv)
	if err != nil {
		return nil, err
	}
	platform, err := extractString(m, "platform")
	if err != nil {
		return nil, err
	}
	release, err := extractInt(m, "release")
	if err != nil {
		return nil, err
	}
	tip, err := extractInt(m, "tip")
	if err != nil {
		return nil, err
	}
	branch, err := extractInt(m, "branch")
	if err != nil {
		return nil, err
	}
	branchBranch, err := extractInt(m, "branchbranch")
	if err != nil {
		return nil, err
	}
	return &FirmwareImageResult{
		Platform:     platform,
		Release:      release,
		Tip:          tip,
		Branch:       branch,
		BranchBranch: branchBranch,
	}, nil
}

// ValidateFirmwareImage checks that a given firmware version is well-formed
// such as "octopus-firmware/R72-11297.75.0"
func ValidateFirmwareImage(v string) error {
	_, err := ParseFirmwareImage(v)
	return err
}

// SerializeFirmwareImage takes arguments describing a firmware version
// and produces a string in the canonical format.
func SerializeFirmwareImage(platform string, release, tip, branch, branchBranch int) string {
	return fmt.Sprintf("%s-firmware/R%d-%d.%d.%d", platform, release, tip, branch, branchBranch)
}
