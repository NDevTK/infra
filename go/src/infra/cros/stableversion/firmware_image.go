// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package stableversion

import (
	"fmt"
	"regexp"
)

// FirmwareImageResult is the result of parsing a firmware image.
// for example: "octopus-firmware/R72-11297.75.0"
type FirmwareImageResult struct {
	Platform     string // would be "octopus"
	ReleaseKind  string // would be "firmware"
	Release      int    // would be 72
	Tip          int    // would be 11297
	Branch       int    // would be 75
	BranchBranch int    // would be 0
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
// platform, releaseKind, release, tip, branch, branchbranch
var firmwareImagePattern *regexp.Regexp = regexp.MustCompile(`\A(?P<platform>[A-Za-z0-9_]+)-(?P<releaseKind>[A-Za-z0-9_]+)/R(?P<release>[0-9]+)-(?P<tip>[0-9]+)\.(?P<branch>[0-9]+)\.(?P<branchbranch>[0-9]+)\z`)

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
	releaseKind, err := extractString(m, "releaseKind")
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
		ReleaseKind:  releaseKind,
		Release:      release,
		Tip:          tip,
		Branch:       branch,
		BranchBranch: branchBranch,
	}, nil
}

// SerializeFirmwareImage takes arguments describing a firmware version
// and produces a string in the canonical format.
func SerializeFirmwareImage(r FirmwareImageResult) string {
	return fmt.Sprintf("%s-%s/R%d-%d.%d.%d", r.Platform, r.ReleaseKind, r.Release, r.Tip, r.Branch, r.BranchBranch)
}
