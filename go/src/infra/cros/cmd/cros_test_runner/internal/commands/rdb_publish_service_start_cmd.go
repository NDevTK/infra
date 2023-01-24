// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
)

// RdbPublishServiceStartCmd represents rdb publish service start cmd.
type RdbPublishServiceStartCmd struct {
	*interfaces.SingleCmdByExecutor
}

func NewRdbPublishServiceStartCmd(executor interfaces.ExecutorInterface) *RdbPublishServiceStartCmd {
	cmd := &RdbPublishServiceStartCmd{SingleCmdByExecutor: interfaces.NewSingleCmdByExecutor(RdbPublishStartCmdType, executor)}
	cmd.ConcreteCmd = cmd
	return cmd
}
