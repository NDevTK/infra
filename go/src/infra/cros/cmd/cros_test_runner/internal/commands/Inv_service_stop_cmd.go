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
