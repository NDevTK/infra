// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cache

import (
	"regexp"
	"strconv"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components"
)

// ExtractHttpResponseCode extracts the HTTP Response Code from an
// error object.
func ExtractHttpResponseCode(err error) int {
	var httpResponseCode int
	stdErr, ok := errors.TagValueIn(components.StdErrTag, err)
	if !ok {
		return 0
	}
	stdErrStr := stdErr.(string)
	re := regexp.MustCompile("(returned error: )([0-9]*)")
	matchParts := re.FindAllStringSubmatch(stdErrStr, -1)
	if len(matchParts) == 1 {
		httpResponseCode, _ = strconv.Atoi(matchParts[0][2])
	}
	return httpResponseCode
}
