// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"infra/cros/cmd/common_lib/interfaces"
)

// RdbPublishServiceStartCmd represents rdb publish service start cmd.
type RdbPublishServiceStartCmd struct {
	*interfaces.SingleCmdByExecutor
}

func NewRdbPublishServiceStartCmd(executor interfaces.ExecutorInterface) *RdbPublishServiceStartCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(RdbPublishStartCmdType, executor)
	cmd := &RdbPublishServiceStartCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
