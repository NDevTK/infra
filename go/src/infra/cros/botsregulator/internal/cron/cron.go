// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package cron defines the service's cron job.
package cron

import (
	"context"

	"go.chromium.org/luci/common/logging"
)

func Regulate(ctx context.Context) error {
	logging.Infof(ctx, "hello from cron!")
	return nil
}
