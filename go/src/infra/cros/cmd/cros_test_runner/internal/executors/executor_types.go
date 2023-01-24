// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package executors

import (
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
)

// All supported executor types.
const (
	NoExecutorType         interfaces.ExecutorType = "NoExecutor"
	InvServiceExecutorType interfaces.ExecutorType = "InvServiceExecutor"
	CtrExecutorType        interfaces.ExecutorType = "CtrExecutor"
)
