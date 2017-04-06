// Copyright 2017 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"regexp"
	"strings"
)

// defaultGitRetry is the set of RE2-formatted Regular Expressions to add
// to the DefaultGitRetryRegexp.
//
// defaultGitRetryPOSIX was originally translated from "depot_tools":
// https://chromium.googlesource.com/chromium/tools/depot_tools/+/14db1dfd88b987c9119f59f75f1f27b8e7c49602/git_common.py
var defaultGitRetryRegexpSource = []string{
	// crbug.com/285832
	`!.*\[remote rejected\].*\(error in hook\)`,

	// crbug.com/289932
	`!.*\[remote rejected\].*\(failed to lock\)`,

	// crbug.com/307156
	`!.*\[remote rejected\].*\(error in Gerrit backend\)`,

	// crbug.com/285832
	`remote error: Internal Server Error`,

	// crbug.com/294449
	`fatal: Couldn't find remote ref `,

	// crbug.com/220543
	`git fetch_pack: expected ACK/NAK, got`,

	// crbug.com/189455
	`protocol error: bad pack header`,

	// crbug.com/202807
	`The remote end hung up unexpectedly`,

	// crbug.com/298189
	`TLS packet with unexpected length was received`,

	// crbug.com/187444
	`RPC failed; result=\d+, HTTP code = \d+`,

	// crbug.com/388876
	`Connection timed out`,

	// crbug.com/430343
	// TODO(dnj): Resync with Chromite.

	`The requested URL returned error: 5\d+`,
	`Connection reset by peer`,

	`Unable to look up`,
	`Couldn't resolve host`,
	`Unknown SSL protocol error`,
}

// DefaultGitRetryRegexp is the set of default transient regular expressions to
// retry on.
var DefaultGitRetryRegexp *regexp.Regexp

func init() {
	if len(defaultGitRetryRegexpSource) > 0 {
		DefaultGitRetryRegexp = regexp.MustCompile(mergeRegex(defaultGitRetryRegexpSource))
	}
}

// mergeRegex merges multiple regular expression strings together into a single
// "|"-delimited regular expression group. No capture groups are introduced in
// this merge.
func mergeRegex(regexps []string) string {
	// Merge all of the regex into a single regex.
	allRE := make([]string, len(regexps))
	for i, re := range regexps {
		allRE[i] = "(?:" + re + ")"
	}
	return "(?:" + strings.Join(allRE, "|") + ")"
}
