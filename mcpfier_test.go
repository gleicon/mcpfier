package main

import (
	"io/ioutil"
	"os"
	"testing"
)

// TestLoadConfig tests the loading of the configuration from a YAML file
func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpfile, err := ioutil.TempFile("", "config.yaml")
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
`
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Load the configuration from the temporary file
	config, err := LoadConfig(tmpfile.Name())
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
}

// TestRunCommand tests running a command
// This is a basic sanity check since executing actual scripts could have side effects
func TestRunCommand(t *testing.T) {
	command := Command{
		Script: "echo",
		Args:   []string{"Hello, World!"},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := RunCommand(command)

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedOutput := "Hello, World!\n"
	if string(out) != expectedOutput {
		t.Errorf("Expected output '%s', got '%s'", expectedOutput, string(out))
	}
}
