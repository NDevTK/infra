package commands

import (
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
)

// TestServiceStartCmd represents test service start cmd.
type TestServiceStartCmd struct {
	*interfaces.SingleCmdByExecutor
}

func NewTestServiceStartCmd(executor interfaces.ExecutorInterface) *TestServiceStartCmd {
	singleCmdByExec := interfaces.NewSingleCmdByExecutor(TestServiceStartCmdType, executor)
	cmd := &TestServiceStartCmd{SingleCmdByExecutor: singleCmdByExec}
	cmd.ConcreteCmd = cmd
	return cmd
}
