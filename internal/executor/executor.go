package executor

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/gleicon/mcpfier/internal/analytics"
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
	analytics analytics.Analytics
}

// New creates a new executor service
func New() *Service {
	return &Service{
		local:     NewLocalExecutor(),
		container: NewContainerExecutor(),
		analytics: &analytics.NoOpAnalytics{},
	}
}

// WithAnalytics sets the analytics instance
func (s *Service) WithAnalytics(a analytics.Analytics) *Service {
	s.analytics = a
	return s
}

// Execute runs a command using the appropriate executor
func (s *Service) Execute(ctx context.Context, cmd *config.Command) (string, error) {
	sessionID := getSessionID(ctx)
	start := time.Now()
	
	var output string
	var err error
	
	if cmd.IsContainerized() {
		output, err = s.container.Execute(ctx, cmd)
	} else {
		output, err = s.local.Execute(ctx, cmd)
	}
	
	// Record analytics
	s.analytics.RecordCommand(ctx, analytics.CommandEvent{
		SessionID:     sessionID,
		CommandName:   cmd.Name,
		Duration:      time.Since(start),
		Success:       err == nil,
		OutputSize:    int64(len(output)),
		ExecutionMode: getExecutionMode(cmd),
		Error:         getErrorString(err),
	})
	
	return output, err
}

// getSessionID gets or creates a session ID from context
func getSessionID(ctx context.Context) string {
	if sessionID, ok := ctx.Value("session_id").(string); ok {
		return sessionID
	}
	
	// Generate a simple session ID
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// getExecutionMode returns the execution mode string
func getExecutionMode(cmd *config.Command) string {
	if cmd.IsContainerized() {
		return "container"
	}
	return "local"
}

// getErrorString safely converts error to string
func getErrorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
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