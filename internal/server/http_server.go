package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gleicon/mcpfier/internal/analytics"
	"github.com/gleicon/mcpfier/internal/auth"
	"github.com/gleicon/mcpfier/internal/config"
	"github.com/gleicon/mcpfier/internal/executor"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// HTTPServer wraps the HTTP MCP server
type HTTPServer struct {
	config       *config.Config
	mcpServer    *server.MCPServer
	httpServer   *server.StreamableHTTPServer
	executor     *executor.Service
	analytics    analytics.Analytics
}

// NewHTTP creates a new HTTP MCP server instance
func NewHTTP(cfg *config.Config) *HTTPServer {
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
	
	// Create MCP server
	mcpServer := server.NewMCPServer(
		"mcpfier",
		"1.0.0",
		server.WithToolCapabilities(true),
	)
	
	// Create HTTP server instance
	httpSrv := &HTTPServer{
		config:    cfg,
		mcpServer: mcpServer,
		executor:  executorService,
		analytics: analyticsService,
	}
	
	// Register tools
	httpSrv.registerTools()
	
	// Create StreamableHTTP server with authentication context and stateless mode
	httpSrv.httpServer = server.NewStreamableHTTPServer(
		mcpServer,
		server.WithHTTPContextFunc(auth.ContextFunc(&cfg.Server.HTTP.Auth)),
		server.WithStateLess(true), // Enable stateless mode for simpler usage
	)
	
	return httpSrv
}

// registerTools registers all configured commands as MCP tools
func (s *HTTPServer) registerTools() {
	for _, cmd := range s.config.Commands {
		cmdCopy := cmd // Capture loop variable
		s.mcpServer.AddTool(
			mcp.NewTool(cmdCopy.Name,
				mcp.WithDescription(cmdCopy.GetDescription()),
			),
			func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return s.executeCommand(ctx, cmdCopy.Name)
			},
		)
	}
}

// executeCommand executes a command with authentication checks
func (s *HTTPServer) executeCommand(ctx context.Context, commandName string) (*mcp.CallToolResult, error) {
	// Check authentication and permissions
	authCtx, hasAuth := auth.AuthContextFromRequest(ctx)
	if s.config.Server.HTTP.Auth.Enabled {
		if !hasAuth {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: "Authentication required",
					},
				},
				IsError: true,
			}, nil
		}
		
		// Check permissions
		if !authCtx.HasPermission(commandName) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("Permission denied for tool '%s'", commandName),
					},
				},
				IsError: true,
			}, nil
		}
	}
	
	// Execute the command
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

// Start starts the HTTP MCP server
func (s *HTTPServer) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Server.HTTP.Host, s.config.Server.HTTP.Port)
	
	// Create custom HTTP server with middleware stack
	mux := http.NewServeMux()
	
	// Add health check endpoint (no auth required)
	mux.HandleFunc("/health", s.healthCheck)
	
	// Add analytics web interface
	mux.HandleFunc("/mcpfier/analytics", s.analyticsWeb)
	
	if s.config.Server.HTTP.Auth.Enabled {
		log.Printf("MCPFier HTTP server starting on %s", addr)
		log.Printf("Authentication: enabled (%s mode)", s.config.Server.HTTP.Auth.Mode)
		log.Printf("MCP endpoint: http://%s", addr)
		log.Printf("Analytics web UI: http://%s/mcpfier/analytics", addr)
		
		// Create authentication middleware
		authMiddleware := auth.Middleware(&s.config.Server.HTTP.Auth)
		
		// Wrap the StreamableHTTP server with auth middleware (analytics applied globally)
		mux.Handle("/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Apply CORS if enabled
			if s.config.Server.HTTP.CORS.Enabled {
				s.applyCORSHeaders(w, r)
				if r.Method == "OPTIONS" {
					w.WriteHeader(http.StatusOK)
					return
				}
			}
			
			// Let mcp-go handle the request directly
			s.httpServer.ServeHTTP(w, r)
		})))
	} else {
		log.Printf("MCPFier HTTP server starting on %s (no authentication)", addr)
		log.Printf("MCP endpoint: http://%s", addr)
		log.Printf("Analytics web UI: http://%s/mcpfier/analytics", addr)
		
		// No authentication (analytics applied globally)
		mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Apply CORS if enabled
			if s.config.Server.HTTP.CORS.Enabled {
				s.applyCORSHeaders(w, r)
				if r.Method == "OPTIONS" {
					w.WriteHeader(http.StatusOK)
					return
				}
			}
			
			// Let mcp-go handle the request directly
			s.httpServer.ServeHTTP(w, r)
		}))
	}
	
	// Apply analytics middleware to ALL requests, then logging middleware
	analyticsHandler := s.analyticsMiddleware(mux)
	handler := LoggingMiddleware()(analyticsHandler)
	
	return http.ListenAndServe(addr, handler)
}

// analyticsMiddleware creates middleware for recording HTTP analytics
func (s *HTTPServer) analyticsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Wrap response writer to capture status and size
		lrw := NewLoggingResponseWriter(w)
		
		// Process request
		next.ServeHTTP(lrw, r)
		
		// Record analytics event
		duration := time.Since(start)
		
		// Get client IP
		clientIP := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			clientIP = forwarded
		} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
			clientIP = realIP
		}
		
		// Determine auth method and success
		authMethod := "none"
		authSuccess := false
		
		// Check if auth headers are present
		if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
			authMethod = "api_key"
		} else if auth := r.Header.Get("Authorization"); auth != "" {
			if strings.HasPrefix(auth, "Bearer ") {
				authMethod = "bearer"
			} else if strings.HasPrefix(auth, "ApiKey ") {
				authMethod = "api_key"
			}
		}
		
		// Auth success is determined by: auth method present AND status code < 400
		if authMethod != "none" {
			authSuccess = lrw.statusCode < 400
		}
		
		// Record HTTP event in analytics
		event := analytics.HTTPEvent{
			SessionID:    r.Header.Get("X-Session-ID"),
			Method:       r.Method,
			Path:         r.URL.Path,
			StatusCode:   lrw.statusCode,
			Duration:     duration,
			ClientIP:     clientIP,
			UserAgent:    r.Header.Get("User-Agent"),
			AuthMethod:   authMethod,
			AuthSuccess:  authSuccess,
			ResponseSize: lrw.size,
		}
		
		s.analytics.RecordHTTPEvent(r.Context(), event)
	})
}

// analyticsWeb serves the analytics web dashboard
func (s *HTTPServer) analyticsWeb(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Get HTTP stats for the last 7 days
	httpStats, err := s.analytics.GetHTTPStats(7)
	if err != nil {
		log.Printf("Failed to get HTTP stats: %v", err)
		httpStats = &analytics.HTTPStats{} // Empty stats on error
	}
	
	// Get command stats for the last 7 days
	commandStats, err := s.analytics.GetStats(7)
	if err != nil {
		log.Printf("Failed to get command stats: %v", err)
		commandStats = &analytics.UsageStats{} // Empty stats on error
	}

	// Get webhook stats for the last 7 days
	webhookStats, err := s.analytics.GetWebhookStats(7)
	if err != nil {
		log.Printf("Failed to get webhook stats: %v", err)
		webhookStats = &analytics.WebhookStats{} // Empty stats on error
	}
	
	// Generate HTML response with Tailwind CSS
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>MCPFier Analytics</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script>
        // Auto-refresh every 30 seconds
        setTimeout(() => location.reload(), 30000);
    </script>
</head>
<body class="bg-gray-100 min-h-screen">
    <div class="container mx-auto px-4 py-8">
        <h1 class="text-3xl font-bold text-gray-800 mb-8">MCPFier Analytics Dashboard</h1>
        
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-6 mb-8">
            <div class="bg-white rounded-lg shadow-md p-6">
                <div class="flex items-center">
                    <div class="text-2xl"></div>
                    <div class="ml-4">
                        <p class="text-sm font-medium text-gray-600">Total HTTP Requests</p>
                        <p class="text-2xl font-bold text-gray-900">%d</p>
                    </div>
                </div>
            </div>
            
            <div class="bg-white rounded-lg shadow-md p-6">
                <div class="flex items-center">
                    <div class="text-2xl"></div>
                    <div class="ml-4">
                        <p class="text-sm font-medium text-gray-600">Success Rate</p>
                        <p class="text-2xl font-bold text-green-600">%.1f%%</p>
                    </div>
                </div>
            </div>
            
            <div class="bg-white rounded-lg shadow-md p-6">
                <div class="flex items-center">
                    <div class="text-2xl"></div>
                    <div class="ml-4">
                        <p class="text-sm font-medium text-gray-600">Auth Success Rate</p>
                        <p class="text-2xl font-bold text-blue-600">%.1f%%</p>
                    </div>
                </div>
            </div>
            
            <div class="bg-white rounded-lg shadow-md p-6">
                <div class="flex items-center">
                    <div class="text-2xl"></div>
                    <div class="ml-4">
                        <p class="text-sm font-medium text-gray-600">Avg HTTP Request Time</p>
                        <p class="text-2xl font-bold text-purple-600">%dms</p>
                    </div>
                </div>
            </div>
            
            <div class="bg-white rounded-lg shadow-md p-6">
                <div class="flex items-center">
                    <div class="text-2xl"></div>
                    <div class="ml-4">
                        <p class="text-sm font-medium text-gray-600">Avg MCP Tool Time</p>
                        <p class="text-2xl font-bold text-indigo-600">%dms</p>
                    </div>
                </div>
            </div>
        </div>
        
        <!-- Error Summary -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
            <div class="bg-white rounded-lg shadow-md p-6">
                <h3 class="text-lg font-semibold text-gray-800 mb-4">Last 24h Errors</h3>
                <div class="space-y-2">
                    <div class="flex justify-between">
                        <span class="text-gray-600">HTTP Errors:</span>
                        <span class="font-semibold text-red-600">%d</span>
                    </div>
                    <div class="flex justify-between">
                        <span class="text-gray-600">Auth Errors:</span>
                        <span class="font-semibold text-orange-600">%d</span>
                    </div>
                    <div class="flex justify-between">
                        <span class="text-gray-600">Command Errors:</span>
                        <span class="font-semibold text-yellow-600">%d</span>
                    </div>
                </div>
            </div>
        </div>
        
        <!-- Upstream API/Webhook Metrics -->`+s.renderWebhookSection(webhookStats)+`
        
        <!-- MCP Tools Section -->
        <div class="bg-white rounded-lg shadow-md p-6 mb-8">
            <h3 class="text-lg font-semibold text-gray-800 mb-4">Popular MCP Tools (Last 7 days)</h3>
            <div class="overflow-x-auto">
                <table class="min-w-full table-auto">
                    <thead>
                        <tr class="bg-gray-50">
                            <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Tool</th>
                            <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Uses</th>
                            <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Success Rate</th>
                            <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Avg Duration</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-gray-200">`,
		httpStats.TotalRequests,
		httpStats.SuccessRate,
		httpStats.AuthSuccessRate,
		httpStats.AvgDurationMs,
		commandStats.AvgDurationMs,
		httpStats.ErrorsLast24h,
		httpStats.AuthErrorsLast24h,
		commandStats.ErrorsLast24h,
	)
	
	// Add top commands
	for _, cmd := range commandStats.TopCommands {
		html += fmt.Sprintf(`
                        <tr>
                            <td class="px-4 py-2 text-sm font-medium text-gray-900">%s</td>
                            <td class="px-4 py-2 text-sm text-gray-500">%d</td>
                            <td class="px-4 py-2 text-sm text-green-600">%.1f%%</td>
                            <td class="px-4 py-2 text-sm text-gray-500">%dms</td>
                        </tr>`,
			cmd.Name, cmd.Count, cmd.SuccessRate, cmd.AvgDuration,
		)
	}
	
	html += `
                    </tbody>
                </table>
            </div>
        </div>
        
        <!-- HTTP Endpoints Section -->
        <div class="bg-white rounded-lg shadow-md p-6">
            <h3 class="text-lg font-semibold text-gray-800 mb-4">Popular HTTP Endpoints (Last 7 days)</h3>
            <div class="overflow-x-auto">
                <table class="min-w-full table-auto">
                    <thead>
                        <tr class="bg-gray-50">
                            <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Path</th>
                            <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Requests</th>
                            <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Success Rate</th>
                            <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Avg Duration</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-gray-200">`
	
	// Add top paths
	for _, path := range httpStats.TopPaths {
		html += fmt.Sprintf(`
                        <tr>
                            <td class="px-4 py-2 text-sm font-medium text-gray-900">%s</td>
                            <td class="px-4 py-2 text-sm text-gray-500">%d</td>
                            <td class="px-4 py-2 text-sm text-green-600">%.1f%%</td>
                            <td class="px-4 py-2 text-sm text-gray-500">%dms</td>
                        </tr>`,
			path.Path, path.Count, path.SuccessRate, path.AvgDuration,
		)
	}
	
	html += `
                    </tbody>
                </table>
            </div>
        </div>
        
        <div class="text-center text-gray-500 text-sm mt-8">
            <p>Auto-refreshes every 30 seconds • Last updated: ` + time.Now().Format("2006-01-02 15:04:05") + `</p>
        </div>
    </div>
</body>
</html>`
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, html)
}

// healthCheck provides a health check endpoint
func (s *HTTPServer) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","version":"1.0.0","server":"mcpfier"}`)
}

// applyCORSHeaders applies CORS headers to the response
func (s *HTTPServer) applyCORSHeaders(w http.ResponseWriter, r *http.Request) {
	cors := s.config.Server.HTTP.CORS
	
	// Set CORS headers
	if len(cors.AllowedOrigins) > 0 {
		origin := r.Header.Get("Origin")
		for _, allowedOrigin := range cors.AllowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
				break
			}
		}
	}
	
	if len(cors.AllowedMethods) > 0 {
		methods := ""
		for i, method := range cors.AllowedMethods {
			if i > 0 {
				methods += ", "
			}
			methods += method
		}
		w.Header().Set("Access-Control-Allow-Methods", methods)
	}
	
	if len(cors.AllowedHeaders) > 0 {
		headers := ""
		for i, header := range cors.AllowedHeaders {
			if i > 0 {
				headers += ", "
			}
			headers += header
		}
		w.Header().Set("Access-Control-Allow-Headers", headers)
	}
}

// Close closes the server and analytics
func (s *HTTPServer) Close() error {
	return s.analytics.Close()
}

// GetAnalytics returns the analytics service
func (s *HTTPServer) GetAnalytics() analytics.Analytics {
	return s.analytics
}

// renderWebhookSection renders the upstream API/webhook metrics section
func (s *HTTPServer) renderWebhookSection(stats *analytics.WebhookStats) string {
	if stats.TotalCalls == 0 {
		return `
        <div class="bg-white rounded-lg shadow-md p-6 mb-8">
            <h3 class="text-lg font-semibold text-gray-800 mb-4">Upstream API/Webhook Metrics (Last 7 days)</h3>
            <p class="text-gray-500">No webhook/API calls recorded yet.</p>
        </div>`
	}

	html := fmt.Sprintf(`
        <div class="bg-white rounded-lg shadow-md p-6 mb-8">
            <h3 class="text-lg font-semibold text-gray-800 mb-4">Upstream API/Webhook Metrics (Last 7 days)</h3>
            
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-6">
                <div class="bg-gradient-to-r from-blue-500 to-blue-600 rounded-lg shadow-md p-4 text-white">
                    <div class="flex items-center justify-between">
                        <div>
                            <p class="text-sm font-medium opacity-90">Total API Calls</p>
                            <p class="text-2xl font-bold">%d</p>
                        </div>
                        <div class="text-3xl opacity-80"></div>
                    </div>
                </div>
                
                <div class="bg-gradient-to-r from-green-500 to-green-600 rounded-lg shadow-md p-4 text-white">
                    <div class="flex items-center justify-between">
                        <div>
                            <p class="text-sm font-medium opacity-90">Success Rate</p>
                            <p class="text-2xl font-bold">%.1f%%%%</p>
                        </div>
                        <div class="text-3xl opacity-80"></div>
                    </div>
                </div>
                
                <div class="bg-gradient-to-r from-purple-500 to-purple-600 rounded-lg shadow-md p-4 text-white">
                    <div class="flex items-center justify-between">
                        <div>
                            <p class="text-sm font-medium opacity-90">Avg Latency</p>
                            <p class="text-2xl font-bold">%dms</p>
                        </div>
                        <div class="text-3xl opacity-80"></div>
                    </div>
                </div>
                
                <div class="bg-gradient-to-r from-red-500 to-red-600 rounded-lg shadow-md p-4 text-white">
                    <div class="flex items-center justify-between">
                        <div>
                            <p class="text-sm font-medium opacity-90">Errors (24h)</p>
                            <p class="text-2xl font-bold">%d</p>
                        </div>
                        <div class="text-3xl opacity-80"></div>
                    </div>
                </div>
            </div>`,
		stats.TotalCalls,
		stats.SuccessRate,
		stats.AvgLatencyMs,
		stats.ErrorsLast24h,
	)

	// Add top webhooks table
	if len(stats.TopWebhooks) > 0 {
		html += `
            <div class="overflow-x-auto">
                <table class="min-w-full table-auto">
                    <thead>
                        <tr class="bg-gray-50">
                            <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">API Endpoint</th>
                            <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Calls</th>
                            <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Success Rate</th>
                            <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Avg Latency</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-gray-200">`

		for _, webhook := range stats.TopWebhooks {
			html += fmt.Sprintf(`
                        <tr>
                            <td class="px-4 py-2 text-sm font-medium text-blue-600">%s</td>
                            <td class="px-4 py-2 text-sm text-gray-500">%d</td>
                            <td class="px-4 py-2 text-sm text-green-600">%.1f%%%%</td>
                            <td class="px-4 py-2 text-sm text-purple-600">%dms</td>
                        </tr>`,
				webhook.Name, webhook.Count, webhook.SuccessRate, webhook.AvgLatencyMs)
		}

		html += `
                    </tbody>
                </table>
            </div>`
	}

	// Add error breakdown if available
	if len(stats.ErrorBreakdown) > 0 {
		html += `
            <div class="mt-6">
                <h4 class="text-md font-medium text-gray-700 mb-3">Recent Error Types (Last 24h)</h4>
                <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">`

		for errorType, count := range stats.ErrorBreakdown {
			html += fmt.Sprintf(`
                    <div class="bg-red-50 border border-red-200 rounded-lg p-3">
                        <div class="flex items-center justify-between">
                            <span class="text-sm font-medium text-red-800">%s</span>
                            <span class="text-sm font-bold text-red-600">%d</span>
                        </div>
                    </div>`, errorType, count)
		}

		html += `
                </div>
            </div>`
	}

	html += `
        </div>`

	return html
}