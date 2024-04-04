// Copyright 2024 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components/cros"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
)

// printUptimeExec read and print uptime of the host to logs.
func printUptimeExec(ctx context.Context, info *execs.ExecInfo) error {
	dur, err := cros.Uptime(ctx, info.DefaultRunner())
	if err != nil {
		return errors.Annotate(err, "print uptime").Err()
	}
	log.Debugf(ctx, "Device %q uptime: current uptime: %s.", info.GetActiveResource(), dur)
	return nil
}

func init() {
	execs.Register("cros_uptime_print", printUptimeExec)
}
