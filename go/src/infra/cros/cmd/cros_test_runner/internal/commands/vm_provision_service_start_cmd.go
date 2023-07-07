// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
)

// VMProvisionServiceStartCmd represents vm-provision service start cmd.
type VMProvisionServiceStartCmd struct {
	*interfaces.SingleCmdByExecutor
}

func NewVMProvisionServiceStartCmd(executor interfaces.ExecutorInterface) *VMProvisionServiceStartCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(VMProvisionServiceStartCmdType, executor)
	cmd := &VMProvisionServiceStartCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
