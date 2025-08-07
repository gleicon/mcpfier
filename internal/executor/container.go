package executor

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/gleicon/mcpfier/internal/config"
)

// ContainerExecutor executes commands in Docker containers
type ContainerExecutor struct {
	localExecutor *LocalExecutor
}

// NewContainerExecutor creates a new container executor
func NewContainerExecutor() *ContainerExecutor {
	return &ContainerExecutor{
		localExecutor: NewLocalExecutor(),
	}
}

// Execute runs a command in a Docker container, with fallback to local execution
func (e *ContainerExecutor) Execute(ctx context.Context, cmd *config.Command) (string, error) {
	if cmd.Container == "" {
		return e.localExecutor.Execute(ctx, cmd)
	}
	
	// Build docker run command
	dockerArgs := []string{"run", "--rm"}
	
	// Add environment variables
	for k, v := range cmd.Env {
		dockerArgs = append(dockerArgs, "-e", fmt.Sprintf("%s=%s", k, v))
	}
	
	// Add container image
	dockerArgs = append(dockerArgs, cmd.Container)
	
	// Add script and args
	dockerArgs = append(dockerArgs, cmd.Script)
	dockerArgs = append(dockerArgs, cmd.Args...)
	
	execCmd := exec.CommandContext(ctx, "docker", dockerArgs...)
	output, err := execCmd.CombinedOutput()
	return string(output), err
}