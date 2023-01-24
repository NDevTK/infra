// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
)

// InvServiceStartCmd represents inventory service start cmd.
type InvServiceStartCmd struct {
	*interfaces.SingleCmdByExecutor
}

func NewInvServiceStartCmd(executor interfaces.ExecutorInterface) *InvServiceStartCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(InvServiceStartCmdType, executor)
	cmd := &InvServiceStartCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
