package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

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
	// Check for special flags first
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "--setup":
			printSetupInstructions()
			return
		case "--mcp":
			// Continue to MCP server mode below
		default:
			// Legacy command mode - execute command directly
			commandName := os.Args[1]

			configPath := findConfigFile()
			config, err := LoadConfig(configPath)
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
	}

	// MCP Server mode
	configPath := findConfigFile()
	mcpfierServer, err := NewMCPFierServer(configPath)
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

func findConfigFile() string {
	// Check environment variable first
	if configPath := os.Getenv("MCPFIER_CONFIG"); configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}
	
	// Try various config file locations
	candidates := []string{
		"config.yaml",                    // Current directory
		"./config.yaml",                 // Explicit current directory
	}
	
	// Try config next to executable
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		candidates = append(candidates, filepath.Join(execDir, "config.yaml"))
	}
	
	// Try user's home directory
	if homeDir, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(homeDir, ".mcpfier", "config.yaml"))
		candidates = append(candidates, filepath.Join(homeDir, "mcpfier", "config.yaml"))
	}
	
	// Try system-wide locations
	candidates = append(candidates, "/etc/mcpfier/config.yaml")
	
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	
	// Default to config.yaml if nothing found
	return "config.yaml"
}

func printSetupInstructions() {
	configPath := findConfigFile()
	config, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Println("# MCPFier Setup Instructions")
	fmt.Println()
	
	// Get current working directory and binary path
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	
	execPath, err := os.Executable()
	if err != nil {
		execPath = "./mcpfier"
	}
	
	fmt.Printf("## Claude Desktop Configuration\n\n")
	fmt.Printf("Add this to your Claude Desktop MCP settings:\n\n")
	fmt.Printf("```json\n")
	fmt.Printf("{\n")
	fmt.Printf("  \"mcpServers\": {\n")
	fmt.Printf("    \"mcpfier\": {\n")
	fmt.Printf("      \"command\": \"%s\",\n", execPath)
	fmt.Printf("      \"args\": [\"--mcp\"],\n")
	fmt.Printf("      \"cwd\": \"%s\"\n", cwd)
	fmt.Printf("    }\n")
	fmt.Printf("  }\n")
	fmt.Printf("}\n")
	fmt.Printf("```\n\n")
	
	fmt.Printf("## Available Tools\n\n")
	fmt.Printf("Once configured, these %d tools will be available:\n\n", len(config.Commands))
	
	for _, cmd := range config.Commands {
		fmt.Printf("### %s\n", cmd.Name)
		if cmd.Description != "" {
			fmt.Printf("**Description**: %s\n\n", cmd.Description)
		}
		
		if cmd.Container != "" {
			fmt.Printf("**Execution**: Docker container (`%s`)\n\n", cmd.Container)
		} else {
			fmt.Printf("**Execution**: Local system\n\n")
		}
		
		fmt.Printf("**Test command**: \"Use the %s tool\"\n\n", cmd.Name)
	}
	
	// Check for Docker requirements
	dockerCommands := []string{}
	for _, cmd := range config.Commands {
		if cmd.Container != "" {
			dockerCommands = append(dockerCommands, cmd.Container)
		}
	}
	
	if len(dockerCommands) > 0 {
		fmt.Printf("## Docker Setup Required\n\n")
		fmt.Printf("Some tools require Docker images. Run these commands:\n\n")
		fmt.Printf("```bash\n")
		
		// Deduplicate images
		imageSet := make(map[string]bool)
		for _, image := range dockerCommands {
			imageSet[image] = true
		}
		
		for image := range imageSet {
			fmt.Printf("docker pull %s\n", image)
		}
		fmt.Printf("```\n\n")
	}
	
	fmt.Printf("## Testing\n\n")
	fmt.Printf("1. Start Claude Desktop with the MCP configuration\n")
	fmt.Printf("2. Try these test commands:\n")
	for _, cmd := range config.Commands {
		if cmd.Container == "" || cmd.Name == "echo-test" || cmd.Name == "list-files" {
			fmt.Printf("   - \"Use the %s tool\"\n", cmd.Name)
		}
	}
	fmt.Printf("\n")
	
	fmt.Printf("## Troubleshooting\n\n")
	fmt.Printf("- **Binary path**: Ensure the command path points to: `%s`\n", execPath)
	fmt.Printf("- **Config file**: Using config at: `%s`\n", configPath)
	fmt.Printf("- **Working directory**: Current directory: `%s`\n", cwd)
	fmt.Printf("- **Docker**: Ensure Docker is running for containerized tools\n")
	fmt.Printf("- **Permissions**: Ensure MCPFier has execute permissions\n\n")
}

