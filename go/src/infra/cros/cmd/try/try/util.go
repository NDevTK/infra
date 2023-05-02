// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package try

import (
	"fmt"
	"regexp"
	"strings"
)

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

// formatPatchURL formats a changeNumber and gerritInstance into a patch string like "crrev.com/c/1234567".
func formatPatchURL(gerritInstance string, changeNumber int) string {
	return fmt.Sprintf("crrev.com/%s/%d", gerritInstance, changeNumber)
}
