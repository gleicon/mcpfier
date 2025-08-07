package executor

import (
	"context"
	"fmt"

	"github.com/gleicon/mcpfier/internal/config"
)

// Executor defines the interface for command execution
type Executor interface {
	Execute(ctx context.Context, cmd *config.Command) (string, error)
}

// Service handles command execution with fallback strategies
type Service struct {
	local     *LocalExecutor
	container *ContainerExecutor
}

// New creates a new executor service
func New() *Service {
	return &Service{
		local:     NewLocalExecutor(),
		container: NewContainerExecutor(),
	}
}

// Execute runs a command using the appropriate executor
func (s *Service) Execute(ctx context.Context, cmd *config.Command) (string, error) {
	if cmd.IsContainerized() {
		return s.container.Execute(ctx, cmd)
	}
	return s.local.Execute(ctx, cmd)
}

// ExecuteByName finds and executes a command by name from config
func (s *Service) ExecuteByName(ctx context.Context, cfg *config.Config, commandName string) (string, error) {
	var foundCmd *config.Command
	for _, cmd := range cfg.Commands {
		if cmd.Name == commandName {
			foundCmd = &cmd
			break
		}
	}

	if foundCmd == nil {
		return "", fmt.Errorf("command '%s' not found", commandName)
	}

	return s.Execute(ctx, foundCmd)
}