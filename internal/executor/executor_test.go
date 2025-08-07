package executor

import (
	"context"
	"testing"

	"github.com/gleicon/mcpfier/internal/config"
)

func TestLocalExecutor(t *testing.T) {
	executor := NewLocalExecutor()
	
	cmd := &config.Command{
		Name:   "echo-test",
		Script: "echo",
		Args:   []string{"Hello, World!"},
	}

	ctx := context.Background()
	output, err := executor.Execute(ctx, cmd)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedOutput := "Hello, World!\n"
	if output != expectedOutput {
		t.Errorf("Expected output '%s', got '%s'", expectedOutput, output)
	}
}

func TestExecutorService(t *testing.T) {
	service := New()
	
	cfg := &config.Config{
		Commands: []config.Command{
			{
				Name:   "test-echo",
				Script: "echo",
				Args:   []string{"test output"},
			},
		},
	}

	ctx := context.Background()
	
	// Test ExecuteByName
	output, err := service.ExecuteByName(ctx, cfg, "test-echo")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	expectedOutput := "test output\n"
	if output != expectedOutput {
		t.Errorf("Expected output '%s', got '%s'", expectedOutput, output)
	}
	
	// Test command not found
	_, err = service.ExecuteByName(ctx, cfg, "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent command")
	}
}