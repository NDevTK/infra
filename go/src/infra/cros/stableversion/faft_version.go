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

// ParseFaftVersion takes a version string and extracts version info
func ParseFaftVersion(fv string) (string, string, int, int, int, int, error) {
	if fv == "" {
		return "", "", 0, 0, 0, 0, fmt.Errorf("empty faft version string is invalid")
	}
	if faftVersionPattern.FindString(fv) == "" {
		return "", "", 0, 0, 0, 0, fmt.Errorf("faft version string is not valid")
	}
	m, err := findMatchMap(faftVersionPattern, fv)
	if err != nil {
		return "", "", 0, 0, 0, 0, err
	}
	platform, err := extractString(m, "platform")
	if err != nil {
		return "", "", 0, 0, 0, 0, err
	}
	kind, err := extractString(m, "kind")
	if err != nil {
		return "", "", 0, 0, 0, 0, err
	}
	if err := ValidateFaftKind(kind); err != nil {
		return "", "", 0, 0, 0, 0, err
	}
	release, err := extractInt(m, "release")
	if err != nil {
		return "", "", 0, 0, 0, 0, err
	}
	tip, err := extractInt(m, "tip")
	if err != nil {
		return "", "", 0, 0, 0, 0, err
	}
	branch, err := extractInt(m, "branch")
	if err != nil {
		return "", "", 0, 0, 0, 0, err
	}
	branchBranch, err := extractInt(m, "branchbranch")
	if err != nil {
		return "", "", 0, 0, 0, 0, err
	}
	return platform, kind, release, tip, branch, branchBranch, nil
}

// ValidateFaftVersion checks that a given faft version is well-formed
// such as "octopus-firmware/R72-11297.75.0"
// or      "octopus-release/R72-11297.75.0"
func ValidateFaftVersion(v string) error {
	_, _, _, _, _, _, err := ParseFaftVersion(v)
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

// SerializeFaftVersion takes arguments describing a faft version
// and produces a string in the canonical format.
func SerializeFaftVersion(platform string, kind string, release int, tip int, branch int, branchBranch int) string {
	return fmt.Sprintf("%s-%s/R%d-%d.%d.%d", platform, kind, release, tip, branch, branchBranch)
}
