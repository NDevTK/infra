// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package servo

import (
	"context"
	"strings"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
)

// validateServoVersion takes a servo version string and checks whether it refers to a real servo version.
func validateServoVersion(version string) error {
	switch version {
	case "":
		return errors.Reason("validate servo version: version cannot be empty").Err()
	case "v3", "v4", "v4p1":
		return nil
	default:
		return errors.Reason("validate servo version: bad version %q", version).Err()
	}
}

// hasServoVersion checks whether the servo's hardware version is the given version in UFS.
//
// TODO(gregorynisbet): Consider spinning up servod and get the true version of the servo in question.
func hasServoVersion(ctx context.Context, info *execs.ExecInfo) error {
	args := info.GetActionArgs(ctx)
	version := strings.ToUpper(args.AsString(ctx, "version", ""))
	if info.GetChromeos().GetServo() == nil {
		return errors.Reason("has servo version: nonexistent servo does not have version %q", version).Err()
	}
	if err := validateServoVersion(version); err != nil {
		return errors.Annotate(err, "has servo version").Err()
	}
	realServodType := info.GetChromeos().GetServo().GetServodType()
	log.Infof(ctx, "Read servod type %q from UFS", realServodType)
	realServodVersion := info.GetChromeos().GetServo().GetServodVersion()
	if realServodVersion == version {
		return nil
	}
	return errors.Reason("has servo version: want %q got %q", version, realServodVersion).Err()
}

func init() {
	execs.Register("servo_has_servo_version", hasServoVersion)
}
