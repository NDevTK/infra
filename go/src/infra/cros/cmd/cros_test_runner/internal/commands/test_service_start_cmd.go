package commands

import (
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
)

// TestServiceStartCmd represents test service start cmd.
type TestServiceStartCmd struct {
	*interfaces.SingleCmdByExecutor
}

func NewTestServiceStartCmd(executor interfaces.ExecutorInterface) *TestServiceStartCmd {
	cmd := &TestServiceStartCmd{SingleCmdByExecutor: interfaces.NewSingleCmdByExecutor(TestServiceStartCmdType, executor)}
	cmd.ConcreteCmd = cmd
	return cmd
}
