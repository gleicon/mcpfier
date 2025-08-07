package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpfile, err := os.CreateTemp("", "config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Write test data to the temporary config file
	content := `
commands:
  - name: test-command
    script: /path/to/test_script.py
    args: ["--test", "value"]
    description: "Test command description"
    container: "test:latest"
    timeout: "30s"
    env:
      TEST_VAR: "test_value"
`
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Load the configuration from the temporary file
	config, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify the config contents
	if len(config.Commands) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(config.Commands))
	}

	command := config.Commands[0]
	if command.Name != "test-command" {
		t.Errorf("Expected command name 'test-command', got '%s'", command.Name)
	}
	if command.Script != "/path/to/test_script.py" {
		t.Errorf("Expected script path '/path/to/test_script.py', got '%s'", command.Script)
	}
	if len(command.Args) != 2 || command.Args[0] != "--test" || command.Args[1] != "value" {
		t.Errorf("Arguments do not match expected values, got %v", command.Args)
	}
	if command.Description != "Test command description" {
		t.Errorf("Expected description 'Test command description', got '%s'", command.Description)
	}
	if command.Container != "test:latest" {
		t.Errorf("Expected container 'test:latest', got '%s'", command.Container)
	}
	if command.Timeout != "30s" {
		t.Errorf("Expected timeout '30s', got '%s'", command.Timeout)
	}
	if command.Env["TEST_VAR"] != "test_value" {
		t.Errorf("Expected env var TEST_VAR='test_value', got '%s'", command.Env["TEST_VAR"])
	}
}

func TestCommandMethods(t *testing.T) {
	cmd := Command{
		Name:        "test-cmd",
		Description: "Test description",
		Container:   "test:latest",
	}

	// Test GetDescription
	if desc := cmd.GetDescription(); desc != "Test description" {
		t.Errorf("Expected 'Test description', got '%s'", desc)
	}

	// Test GetDescription with empty description
	cmd.Description = ""
	if desc := cmd.GetDescription(); desc != "Execute test-cmd with configured arguments" {
		t.Errorf("Expected default description, got '%s'", desc)
	}

	// Test IsContainerized
	if !cmd.IsContainerized() {
		t.Error("Expected command to be containerized")
	}

	cmd.Container = ""
	if cmd.IsContainerized() {
		t.Error("Expected command to not be containerized")
	}
}