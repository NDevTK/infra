// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
)

// InvServiceStopCmd represents inventory service stop cmd.
type InvServiceStopCmd struct {
	*interfaces.SingleCmdByExecutor
}

func NewInvServiceStopCmd(executor interfaces.ExecutorInterface) *InvServiceStopCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(InvServiceStopCmdType, executor)
	cmd := &InvServiceStopCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
