// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package log adds an internal logging API.
// This makes us future-proof in case we depend on another package for
// logging and need to switch.
package log

import (
	"context"

	"go.chromium.org/luci/common/logging"
)

// TODO(ayatane): log to cloud or something

// Infof logs information a developer might find useful for debugging.
func Infof(ctx context.Context, format string, args ...interface{}) {
	logging.Get(ctx).LogCall(logging.Info, 1, format, args)
}

// Errorf logs fatal errors when handling a request.
func Errorf(ctx context.Context, format string, args ...interface{}) {
	logging.Get(ctx).LogCall(logging.Error, 1, format, args)
}
