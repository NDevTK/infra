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
