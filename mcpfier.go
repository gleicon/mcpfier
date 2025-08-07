package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"gopkg.in/yaml.v2"
)

// Command represents the structure of a command in the config
type Command struct {
	Name        string            `yaml:"name"`
	Script      string            `yaml:"script"`
	Args        []string          `yaml:"args"`
	Description string            `yaml:"description"`
	Container   string            `yaml:"container"`
	Timeout     string            `yaml:"timeout"`
	Env         map[string]string `yaml:"env"`
}

// Config holds all the commands from the YAML file
type Config struct {
	Commands []Command `yaml:"commands"`
}

// LoadConfig loads the configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// RunCommand executes the command based on the configuration
func RunCommand(command Command) error {
	cmd := exec.Command(command.Script, command.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// MCPFierServer wraps the configuration and provides MCP tools
type MCPFierServer struct {
	config *Config
}

// NewMCPFierServer creates a new MCPFier server instance
func NewMCPFierServer(configPath string) (*MCPFierServer, error) {
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return &MCPFierServer{config: config}, nil
}

// ExecuteCommandLocal executes a command locally
func (s *MCPFierServer) ExecuteCommandLocal(ctx context.Context, cmd *Command) (string, error) {
	execCmd := exec.CommandContext(ctx, cmd.Script, cmd.Args...)
	for k, v := range cmd.Env {
		execCmd.Env = append(execCmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	
	output, err := execCmd.CombinedOutput()
	return string(output), err
}

// ExecuteCommandContainer executes a command in a Docker container
func (s *MCPFierServer) ExecuteCommandContainer(ctx context.Context, cmd *Command) (string, error) {
	if cmd.Container == "" {
		return s.ExecuteCommandLocal(ctx, cmd)
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

// ExecuteCommand executes a command and returns the result
func (s *MCPFierServer) ExecuteCommand(ctx context.Context, commandName string, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var foundCmd *Command
	for _, cmd := range s.config.Commands {
		if cmd.Name == commandName {
			foundCmd = &cmd
			break
		}
	}

	if foundCmd == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Command '%s' not found", commandName),
				},
			},
			IsError: true,
		}, nil
	}

	// Execute command (container-aware)
	output, err := s.ExecuteCommandContainer(ctx, foundCmd)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Command execution failed: %v\nOutput: %s", err, output),
				},
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: output,
			},
		},
	}, nil
}

func main() {
	// Check if running in legacy command mode
	if len(os.Args) >= 2 && os.Args[1] != "--mcp" {
		// Legacy command wrapper mode
		commandName := os.Args[1]

		config, err := LoadConfig("config.yaml")
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		var foundCmd *Command
		for _, cmd := range config.Commands {
			if cmd.Name == commandName {
				foundCmd = &cmd
				break
			}
		}

		if foundCmd == nil {
			log.Fatalf("Command '%s' not found in config", commandName)
		}

		if err := RunCommand(*foundCmd); err != nil {
			log.Fatalf("Failed to run command '%s': %v", commandName, err)
		}
		return
	}

	// MCP Server mode
	mcpfierServer, err := NewMCPFierServer("config.yaml")
	if err != nil {
		log.Fatalf("Failed to create MCP server: %v", err)
	}

	// Create MCP server
	s := server.NewMCPServer(
		"mcpfier",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Register tools for each command
	for _, cmd := range mcpfierServer.config.Commands {
		cmdCopy := cmd // Capture loop variable
		s.AddTool(
			mcp.NewTool(cmdCopy.Name,
				mcp.WithDescription(getCommandDescription(cmdCopy)),
			),
			func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return mcpfierServer.ExecuteCommand(ctx, cmdCopy.Name, request)
			},
		)
	}

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// getCommandDescription returns a description for the command
func getCommandDescription(cmd Command) string {
	if cmd.Description != "" {
		return cmd.Description
	}
	return fmt.Sprintf("Execute %s with configured arguments", cmd.Name)
}

