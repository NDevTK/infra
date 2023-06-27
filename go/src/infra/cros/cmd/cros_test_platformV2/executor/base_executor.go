// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package executor

import (
	"context"
)

type Executor interface {

	// Execute runs the exector
	Execute(context.Context, string, *TestPlanResponse) error

	// Response
}

// AbstractExecutor satisfies the executor requirement that is common to all.
type AbstractExecutor struct {
	Executor
}
