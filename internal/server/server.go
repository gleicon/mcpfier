package server

import (
	"context"
	"fmt"

	"github.com/gleicon/mcpfier/internal/analytics"
	"github.com/gleicon/mcpfier/internal/config"
	"github.com/gleicon/mcpfier/internal/executor"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MCPFierServer wraps the configuration and provides MCP tools
type MCPFierServer struct {
	config    *config.Config
	executor  *executor.Service
	server    *server.MCPServer
	analytics analytics.Analytics
}

// New creates a new MCPFier STDIO server instance
func New(cfg *config.Config) *MCPFierServer {
	// Initialize analytics
	var analyticsService analytics.Analytics = &analytics.NoOpAnalytics{}
	
	if cfg.Analytics.Enabled {
		if cfg.Analytics.DatabasePath == "" {
			cfg.Analytics.DatabasePath = "./analytics.db"
		}
		
		if sqliteAnalytics, err := analytics.NewSQLiteAnalytics(cfg.Analytics.DatabasePath); err == nil {
			analyticsService = sqliteAnalytics
		}
	}
	
	executorService := executor.New().WithAnalytics(analyticsService)
	
	return &MCPFierServer{
		config:    cfg,
		executor:  executorService,
		analytics: analyticsService,
		server: server.NewMCPServer(
			"mcpfier",
			"1.0.0",
			server.WithToolCapabilities(true),
		),
	}
}

// RegisterTools registers all configured commands as MCP tools
func (s *MCPFierServer) RegisterTools() {
	for _, cmd := range s.config.Commands {
		cmdCopy := cmd // Capture loop variable
		s.server.AddTool(
			mcp.NewTool(cmdCopy.Name,
				mcp.WithDescription(cmdCopy.GetDescription()),
			),
			func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return s.executeCommand(ctx, cmdCopy.Name)
			},
		)
	}
}

// executeCommand executes a command and returns MCP-formatted result
func (s *MCPFierServer) executeCommand(ctx context.Context, commandName string) (*mcp.CallToolResult, error) {
	output, err := s.executor.ExecuteByName(ctx, s.config, commandName)
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

// Start starts the MCP stdio server
func (s *MCPFierServer) Start() error {
	s.RegisterTools()
	return server.ServeStdio(s.server)
}

// Close closes the server and analytics
func (s *MCPFierServer) Close() error {
	return s.analytics.Close()
}

// GetAnalytics returns the analytics service
func (s *MCPFierServer) GetAnalytics() analytics.Analytics {
	return s.analytics
}