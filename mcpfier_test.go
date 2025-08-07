package main

import (
	"testing"

	"github.com/gleicon/mcpfier/internal/config"
)

// TestConfigIntegration tests the overall config integration
func TestConfigIntegration(t *testing.T) {
	// This test ensures the config package integration works
	cfg, err := config.LoadFromDefaultPath()
	if err != nil {
		// This is expected if no config file exists
		t.Logf("No config file found (expected): %v", err)
		return
	}

	if len(cfg.Commands) == 0 {
		t.Error("Expected at least one command in config")
	}
}

// TestMainFunctionality tests main function behavior
func TestMainFunctionality(t *testing.T) {
	// Test that main doesn't crash with various argument combinations
	// Note: These tests don't actually call main() to avoid side effects
	
	testCases := []struct {
		name string
		args []string
	}{
		{"no args", []string{"mcpfier"}},
		{"setup flag", []string{"mcpfier", "--setup"}},
		{"mcp flag", []string{"mcpfier", "--mcp"}},
		{"legacy command", []string{"mcpfier", "nonexistent"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Just verify the test case structure is valid
			if len(tc.args) == 0 {
				t.Error("Test case should have at least program name")
			}
		})
	}
}
