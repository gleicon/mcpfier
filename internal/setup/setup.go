package setup

import (
	"fmt"
	"os"

	"github.com/gleicon/mcpfier/internal/config"
)

// PrintInstructions prints setup instructions for Claude Desktop integration
func PrintInstructions() error {
	configPath := config.FindConfigFile()
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
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
	
	printClaudeDesktopConfig(execPath, cwd)
	printAvailableTools(cfg)
	printDockerRequirements(cfg)
	printTestingInstructions(cfg)
	printTroubleshooting(execPath, configPath, cwd)
	
	return nil
}

func printClaudeDesktopConfig(execPath, cwd string) {
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
}

func printAvailableTools(cfg *config.Config) {
	fmt.Printf("## Available Tools\n\n")
	fmt.Printf("Once configured, these %d tools will be available:\n\n", len(cfg.Commands))
	
	for _, cmd := range cfg.Commands {
		fmt.Printf("### %s\n", cmd.Name)
		if cmd.Description != "" {
			fmt.Printf("**Description**: %s\n\n", cmd.Description)
		}
		
		if cmd.IsContainerized() {
			fmt.Printf("**Execution**: Docker container (`%s`)\n\n", cmd.Container)
		} else {
			fmt.Printf("**Execution**: Local system\n\n")
		}
		
		fmt.Printf("**Test command**: \"Use the %s tool\"\n\n", cmd.Name)
	}
}

func printDockerRequirements(cfg *config.Config) {
	dockerImages := getDockerImages(cfg)
	
	if len(dockerImages) > 0 {
		fmt.Printf("## Docker Setup Required\n\n")
		fmt.Printf("Some tools require Docker images. Run these commands:\n\n")
		fmt.Printf("```bash\n")
		
		for image := range dockerImages {
			fmt.Printf("docker pull %s\n", image)
		}
		fmt.Printf("```\n\n")
	}
}

func printTestingInstructions(cfg *config.Config) {
	fmt.Printf("## Testing\n\n")
	fmt.Printf("1. Start Claude Desktop with the MCP configuration\n")
	fmt.Printf("2. Try these test commands:\n")
	for _, cmd := range cfg.Commands {
		if !cmd.IsContainerized() || cmd.Name == "echo-test" || cmd.Name == "list-files" {
			fmt.Printf("   - \"Use the %s tool\"\n", cmd.Name)
		}
	}
	fmt.Printf("\n")
}

func printTroubleshooting(execPath, configPath, cwd string) {
	fmt.Printf("## Troubleshooting\n\n")
	fmt.Printf("- **Binary path**: Ensure the command path points to: `%s`\n", execPath)
	fmt.Printf("- **Config file**: Using config at: `%s`\n", configPath)
	fmt.Printf("- **Working directory**: Current directory: `%s`\n", cwd)
	fmt.Printf("- **Docker**: Ensure Docker is running for containerized tools\n")
	fmt.Printf("- **Permissions**: Ensure MCPFier has execute permissions\n\n")
}

func getDockerImages(cfg *config.Config) map[string]bool {
	images := make(map[string]bool)
	for _, cmd := range cfg.Commands {
		if cmd.IsContainerized() {
			images[cmd.Container] = true
		}
	}
	return images
}