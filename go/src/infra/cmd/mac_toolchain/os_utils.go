// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"regexp"

	"go.chromium.org/luci/common/errors"
)

var MajorVersionRegex = regexp.MustCompile(`(\d+)\..*`)

func isMacOS13OrLater(ctx context.Context) (bool, error) {
	curVersion, err := getCurVersion(ctx)
	if err != nil {
		return false, errors.Annotate(err, "failed to get current os version").Err()
	}

	result := MajorVersionRegex.FindStringSubmatch(curVersion)
	if len(result) == 0 {
		error := errors.Reason("unable to parse MacOS Version from %s", result).Err()
		return false, error
	} else {
		curMajorVersion := result[1]
		if curMajorVersion >= "13" {
			return true, nil
		}
	}
	return false, nil
}

func getCurVersion(ctx context.Context) (string, error) {
	out, err := RunOutput(ctx, "sw_vers", "-productVersion")
	if err != nil {
		return "", errors.Annotate(err, "failed to run sw_vers -productVersion").Err()
	}
	return out, nil
}
