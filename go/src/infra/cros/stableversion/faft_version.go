// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package stableversion

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// capture groups:
// platform, release, tip, branch, branchbranch
var legacyFaftVersionPattern *regexp.Regexp = regexp.MustCompile(`\A(?P<platform>[A-Za-z0-9_]+)-(?P<kind>[A-Za-z]+)/R(?P<release>[0-9]+)-(?P<tip>[0-9]+)\.(?P<branch>[0-9]+)\.(?P<branchbranch>[0-9]+)\z`)

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
	if strings.HasPrefix(fv, "firmware") {
		return parseNewFaftVersion(fv)
	}
	return parseLegacyFaftVersion(fv)
}

// func parseLegacyFaftVersion takes a string and parses it as a legacy path.
func parseLegacyFaftVersion(fv string) (*FaftVersionResult, error) {
	if fv == "" {
		return nil, fmt.Errorf("empty faft version string is invalid")
	}
	if legacyFaftVersionPattern.FindString(fv) == "" {
		return nil, fmt.Errorf("faft version string is not valid")
	}
	m, err := findMatchMap(legacyFaftVersionPattern, fv)
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

// capture groups:
// builder, tip, tipSuffix
//
// Examples:
// 1) "firmware-a-99.B-branch-firmware"
// 2) "firmware-a-99-branch-firmware"
var newFaftPrefixPattern = regexp.MustCompile(`\Afirmware-(?P<builder>[0-9_a-zA-Z]+)-(?P<tip>[0-9]+)(\.(?P<tipSuffix>[0-9_a-zA-Z]+))?-branch-firmware\z`)

func parseNewFaftPrefix(prefix string) (map[string]string, error) {
	res, err := findMatchMap(newFaftPrefixPattern, prefix)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// capture groups: release, tip, branch, branchbranch, board
//
// Examples:
// 1) "R99-44.33.22",
// 2) "R99-44.33.22/octopus"
var newFaftSuffixPattern = regexp.MustCompile(`\A(?P<release>R\w*)-(?P<tip>\w*)\.(?P<branch>\w*)\.(?P<branchbranch>[0-9]+)(/(?P<board>\w*))?\z`)

func parseNewFaftSuffix(suffix string) (map[string]string, error) {
	res, err := findMatchMap(newFaftSuffixPattern, suffix)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// func parseNewFaftVersion takes a string and parses it as a legacy path.
func parseNewFaftVersion(fv string) (*FaftVersionResult, error) {
	segments := strings.SplitN(fv, "/", 2)
	switch len(segments) {
	case 0, 1:
		return nil, fmt.Errorf(`new faft version must contain "/"`)
	case 2:
		// do nothing
	default:
		panic("impossible")
	}
	prefixMap, err := parseNewFaftPrefix(segments[0])
	if err != nil {
		return nil, err
	}
	suffixMap, err := parseNewFaftSuffix(segments[1])
	if err != nil {
		return nil, err
	}

	// If we have a three-part path like "something/something/octopus" then the last segment is
	// definitely the board, even if the prefix contains a different platform.
	platform := suffixMap["board"]
	if platform == "" {
		platform = prefixMap["builder"]
	}

	out := &FaftVersionResult{
		Platform:     platform,
		Kind:         "firmware",
		Release:      asInt(strings.TrimPrefix(suffixMap["release"], "R")),
		Tip:          asInt(suffixMap["tip"]),
		Branch:       asInt(suffixMap["branch"]),
		BranchBranch: asInt(suffixMap["branchbranch"]),
	}
	return out, nil
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
