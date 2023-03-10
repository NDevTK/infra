// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
)

// DutServiceStartCmd represents dut service start cmd.
type SshStopTunnelsCmd struct {
	*interfaces.SingleCmdByExecutor
}

func NewSshStopTunnelsCmd(executor interfaces.ExecutorInterface) *SshStopTunnelsCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(SshStopTunnelsCmdType, executor)
	cmd := &SshStopTunnelsCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
