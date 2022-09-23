// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package execs

import (
	"context"

	"infra/cros/recovery/logger"
	"infra/cros/recovery/tlw"
)

// NewLogger returns logger.
func (ei *ExecInfo) NewLogger() logger.Logger {
	return ei.runArgs.Logger
}

// GetLogRoot returns path to logs directory.
func (ei *ExecInfo) GetLogRoot() string {
	return ei.runArgs.LogRoot
}

// CopyFrom copies files from resource to localhost.
func (ei *ExecInfo) CopyFrom(ctx context.Context, resourceName, srcFile, destFile string) error {
	return ei.runArgs.Access.CopyFileFrom(ctx, &tlw.CopyRequest{
		Resource:        resourceName,
		PathSource:      srcFile,
		PathDestination: destFile,
	})
}
