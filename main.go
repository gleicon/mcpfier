package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gleicon/mcpfier/internal/analytics"
	"github.com/gleicon/mcpfier/internal/config"
	"github.com/gleicon/mcpfier/internal/executor"
	"github.com/gleicon/mcpfier/internal/server"
	"github.com/gleicon/mcpfier/internal/setup"
)

func main() {
	// Parse command line arguments
	args := parseArgs()
	
	// Set config path if provided
	if args.configPath != "" {
		os.Setenv("MCPFIER_CONFIG", args.configPath)
	}

	// Execute based on mode
	switch args.mode {
	case "setup":
		if err := setup.PrintInstructions(); err != nil {
			log.Fatalf("Setup failed: %v", err)
		}
	case "analytics":
		showAnalytics()
	case "mcp":
		startMCPServer()
	case "legacy":
		executeLegacyCommand(args.commandName)
	default:
		startMCPServer() // Default to MCP server mode
	}
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

func executeLegacyCommand(commandName string) {
	if commandName == "" {
		log.Fatal("Command name required")
	}

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

	// Initialize analytics for legacy mode too
	var analyticsService analytics.Analytics = &analytics.NoOpAnalytics{}
	if cfg.Analytics.Enabled {
		if cfg.Analytics.DatabasePath == "" {
			cfg.Analytics.DatabasePath = "./analytics.db"
		}
		if sqliteAnalytics, err := analytics.NewSQLiteAnalytics(cfg.Analytics.DatabasePath); err == nil {
			analyticsService = sqliteAnalytics
		}
	}

	// Execute using the executor service to get analytics
	executorService := executor.New().WithAnalytics(analyticsService)
	ctx := context.Background()
	
	output, err := executorService.Execute(ctx, foundCmd)
	if err != nil {
		log.Fatalf("Failed to run command '%s': %v", commandName, err)
	}
	
	// Print output to maintain compatibility
	fmt.Print(output)
}

func showAnalytics() {
	cfg, err := config.LoadFromDefaultPath()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if !cfg.Analytics.Enabled {
		fmt.Println("Analytics is disabled in configuration")
		return
	}

	dbPath := cfg.Analytics.DatabasePath
	if dbPath == "" {
		dbPath = "./analytics.db"
	}

	analyticsService, err := analytics.NewSQLiteAnalytics(dbPath)
	if err != nil {
		log.Fatalf("Failed to open analytics database: %v", err)
	}
	defer analyticsService.Close()

	days := 7 // Default to 7 days
	stats, err := analyticsService.GetStats(days)
	if err != nil {
		log.Fatalf("Failed to get analytics: %v", err)
	}

	// Pretty print JSON
	output, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		log.Fatalf("Failed to format output: %v", err)
	}

	fmt.Printf("MCPFier Analytics (Last %d days):\n", days)
	fmt.Println(string(output))
}

// cmdArgs represents parsed command line arguments
type cmdArgs struct {
	mode        string // "setup", "analytics", "mcp", "legacy"
	configPath  string
	commandName string // for legacy mode
}

// parseArgs parses command line arguments and returns structured args
func parseArgs() cmdArgs {
	args := cmdArgs{mode: "default"}
	
	if len(os.Args) < 2 {
		return args
	}

	i := 1
	for i < len(os.Args) {
		arg := os.Args[i]
		
		switch {
		case arg == "--config" || arg == "-c":
			// Next argument should be config path
			if i+1 >= len(os.Args) {
				log.Fatal("--config requires a path argument")
			}
			args.configPath = os.Args[i+1]
			i += 2
			
		case arg == "--setup":
			args.mode = "setup"
			i++
			
		case arg == "--analytics":
			args.mode = "analytics"
			i++
			
		case arg == "--mcp":
			args.mode = "mcp"
			i++
			
		case arg == "--help" || arg == "-h":
			printHelp()
			os.Exit(0)
			
		case strings.HasPrefix(arg, "--config="):
			// Handle --config=path format
			args.configPath = strings.TrimPrefix(arg, "--config=")
			i++
			
		case strings.HasPrefix(arg, "-"):
			log.Fatalf("Unknown flag: %s", arg)
			
		default:
			// Non-flag argument - assume it's a command name for legacy mode
			args.mode = "legacy"
			args.commandName = arg
			i++
		}
	}
	
	return args
}

// printHelp prints usage information
func printHelp() {
	fmt.Printf(`MCPFier - Model Context Protocol server for command execution

Usage:
  mcpfier [options] [command]
  mcpfier [options] --mcp
  mcpfier [options] --setup
  mcpfier [options] --analytics

Options:
  --config, -c PATH    Use specific configuration file
  --help, -h          Show this help message

Modes:
  --mcp               Start MCP server (default mode)
  --setup             Generate Claude Desktop configuration
  --analytics         Show usage statistics
  command-name        Execute command directly (legacy mode)

Examples:
  mcpfier --mcp                           # Start MCP server
  mcpfier --config /path/config.yaml --mcp  # Use custom config
  mcpfier --analytics                     # Show statistics  
  mcpfier --setup                         # Generate setup info
  mcpfier echo-test                       # Run command directly
  mcpfier -c ~/.mcpfier/config.yaml echo-test  # Custom config + command

Environment Variables:
  MCPFIER_CONFIG      Path to configuration file (overridden by --config)

For more information, see README.md
`)
}