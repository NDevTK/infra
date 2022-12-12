// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package stableversion

import (
	"errors"
	"fmt"
	"regexp"
)

// capture groups:
// platform, release, tip, branch, branchbranch
var faftVersionPattern *regexp.Regexp = regexp.MustCompile(`\A(?P<platform>[A-Za-z0-9_]+)-(?P<kind>[A-Za-z]+)/R(?P<release>[0-9]+)-(?P<tip>[0-9]+)\.(?P<branch>[0-9]+)\.(?P<branchbranch>[0-9]+)\z`)

// FaftVersionResult stores the result of parsing a FAFT version.
type FaftVersionResult struct {
	// Platform is the name of board in question e.g. octopus.
	Platform string
	// Kind is "firmware" or "release".
	Kind string
	// Release is the release milestone.
	Release int
	// Tip is the 2nd number in a release version string.
	Tip int
	// Branch is 3rd number in a release version string.
	Branch int
	// BranchBranch is the 4th number in a release version string.
	BranchBranch int
}

// ParseFaftVersion takes a version string and extracts version info
func ParseFaftVersion(fv string) (*FaftVersionResult, error) {
	if fv == "" {
		return nil, fmt.Errorf("empty faft version string is invalid")
	}
	if faftVersionPattern.FindString(fv) == "" {
		return nil, fmt.Errorf("faft version string is not valid")
	}
	m, err := findMatchMap(faftVersionPattern, fv)
	if err != nil {
		return nil, err
	}
	platform, err := extractString(m, "platform")
	if err != nil {
		return nil, err
	}
	kind, err := extractString(m, "kind")
	if err != nil {
		return nil, err
	}
	if err := ValidateFaftKind(kind); err != nil {
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
	return &FaftVersionResult{
		Platform:     platform,
		Kind:         kind,
		Release:      release,
		Tip:          tip,
		Branch:       branch,
		BranchBranch: branchBranch,
	}, nil
}

// ValidateFaftVersion checks that a given faft version is well-formed
// such as "octopus-firmware/R72-11297.75.0"
// or      "octopus-release/R72-11297.75.0"
func ValidateFaftVersion(v string) error {
	_, err := ParseFaftVersion(v)
	return err
}

func ValidateFaftKind(v string) error {
	if v == "" {
		return errors.New("validate faft kind: kind cannot be empty")
	}
	switch v {
	case "firmware":
		return nil
	case "release":
		return nil
	}
	return fmt.Errorf("validate faft kind: bad kind %q", v)
}
