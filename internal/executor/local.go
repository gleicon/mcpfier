package executor

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/gleicon/mcpfier/internal/config"
)

// LocalExecutor executes commands directly on the host system
type LocalExecutor struct{}

// NewLocalExecutor creates a new local executor
func NewLocalExecutor() *LocalExecutor {
	return &LocalExecutor{}
}

// Execute runs a command locally
func (e *LocalExecutor) Execute(ctx context.Context, cmd *config.Command) (string, error) {
	execCmd := exec.CommandContext(ctx, cmd.Script, cmd.Args...)
	
	// Set environment variables
	for k, v := range cmd.Env {
		execCmd.Env = append(execCmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	
	output, err := execCmd.CombinedOutput()
	return string(output), err
}