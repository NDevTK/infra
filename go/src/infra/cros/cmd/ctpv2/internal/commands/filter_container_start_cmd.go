// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"infra/cros/cmd/common_lib/interfaces"
)

// FilterStartCmd represents test service start cmd.
type FilterStartCmd struct {
	*interfaces.SingleCmdByExecutor
}

func NewFilterStartCmd(executor interfaces.ExecutorInterface) *FilterStartCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(FilterStartCmdType, executor)
	cmd := &FilterStartCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
