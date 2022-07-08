package server

import (
	"context"

	"infra/cros/cmd/cros-tool-runner/internal/v2/commands"
)

// CommandExecutor proxies command execution to provide an abstraction layer for interception.
type CommandExecutor interface {
	// Execute returns the same output as the original command.
	Execute(context.Context, commands.Command) (string, string, error)
}

// DefaultCommandExecutor is the default implementation that executes command as is.
type DefaultCommandExecutor struct{ CommandExecutor }

func (*DefaultCommandExecutor) Execute(ctx context.Context, cmd commands.Command) (string, string, error) {
	return cmd.Execute(ctx)
}
