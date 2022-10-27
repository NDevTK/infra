// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"fmt"
	"regexp"
	"strings"
)

// prependString returns an array with an element at the beginning.
func prependString(newElem string, arr []string) []string {
	return append([]string{newElem}, arr...)
}

// separateBucketFromBuilder takes a full builder name (like chromeos/release/release-main-orchestrator),
// and separates it into a bucket (chromeos/release) and a builder (release-main-orchestrator).
func separateBucketFromBuilder(fullBuilderName string) (bucket string, builder string, err error) {
	parts := strings.Split(fullBuilderName, "/")
	if len(parts) != 3 {
		return "", "", fmt.Errorf("builder %s has %d slash-delimited parts; expect 3", fullBuilderName, len(parts))
	}
	bucket = strings.Join(parts[:2], "/")
	builder = parts[2]
	return bucket, builder, nil
}

// interfaceSlicetoStr converts a slice of interface{}s to a slice of strings.
func interfaceSliceToStr(s []interface{}) []string {
	ret := make([]string, len(s))
	for i := range s {
		ret[i] = s[i].(string)
	}
	return ret
}

// parseEmailFromAuthInfo parses an email from a `led auth-info` invocation.
func parseEmailFromAuthInfo(stdout string) (string, error) {
	reAuthUser := regexp.MustCompile(`^Logged in as ([A-Za-z0-9\-_.+]+@[A-Za-z0-9\-_.+]+\.\w+)\.(\s|$)`)
	submatch := reAuthUser.FindStringSubmatch(stdout)
	if len(submatch) == 0 {
		return "", fmt.Errorf("Could not find username in `luci auth-info` output:\n%s", stdout)
	}
	return strings.TrimSpace(submatch[1]), nil
}

// sliceContainsStr checks whether a []string contains a string.
func sliceContainsStr(slice []string, s string) bool {
	for _, x := range slice {
		if x == s {
			return true
		}
	}
	return false
}

// patchListToBBAddArgs converts a []string of patches to []string formatting expected by bb add (like ["crrev.com/c/1234567"] -> ["-cl", "crrev.com/c/123456"])
func patchListToBBAddArgs(patches []string) []string {
	bbAddArgs := make([]string, 0)
	for _, patch := range patches {
		bbAddArgs = append(bbAddArgs, []string{"-cl", patch}...)
	}
	return bbAddArgs
}
