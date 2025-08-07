package main

import (
	"log"
	"os"
	"os/exec"

	"github.com/gleicon/mcpfier/internal/config"
	"github.com/gleicon/mcpfier/internal/server"
	"github.com/gleicon/mcpfier/internal/setup"
)

func main() {
	// Check for special flags first
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "--setup":
			if err := setup.PrintInstructions(); err != nil {
				log.Fatalf("Setup failed: %v", err)
			}
			return
		case "--mcp":
			startMCPServer()
			return
		default:
			// Legacy command mode - execute command directly
			executeLegacyCommand()
			return
		}
	}

	// Default to MCP server mode if no args
	startMCPServer()
}

func startMCPServer() {
	cfg, err := config.LoadFromDefaultPath()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	mcpServer := server.New(cfg)
	if err := mcpServer.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func executeLegacyCommand() {
	if len(os.Args) < 2 {
		log.Fatal("Command name required")
	}
	commandName := os.Args[1]

	cfg, err := config.LoadFromDefaultPath()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Find the command
	var foundCmd *config.Command
	for _, cmd := range cfg.Commands {
		if cmd.Name == commandName {
			foundCmd = &cmd
			break
		}
	}

	if foundCmd == nil {
		log.Fatalf("Command '%s' not found in config", commandName)
	}

	// Execute using legacy direct method for stdout/stderr compatibility
	cmd := exec.Command(foundCmd.Script, foundCmd.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to run command '%s': %v", commandName, err)
	}
}