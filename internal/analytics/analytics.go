package analytics

import (
	"context"
	"time"
)

// Analytics defines the interface for embedded analytics
type Analytics interface {
	RecordCommand(ctx context.Context, event CommandEvent)
	RecordHTTPEvent(ctx context.Context, event HTTPEvent)
	GetStats(days int) (*UsageStats, error)
	GetHTTPStats(days int) (*HTTPStats, error)
	GetWebhookStats(days int) (*WebhookStats, error)
	Close() error
}

// CommandEvent represents a command execution event
type CommandEvent struct {
	SessionID     string
	CommandName   string
	Duration      time.Duration
	Success       bool
	OutputSize    int64
	ExecutionMode string // "local" or "container"
	Error         string
}

// UsageStats represents aggregated usage statistics
type UsageStats struct {
	TotalCommands    int64            `json:"total_commands"`
	SuccessRate      float64          `json:"success_rate"`
	TopCommands      []CommandSummary `json:"top_commands"`
	ErrorsLast24h    int64            `json:"errors_last_24h"`
	AvgDurationMs    int64            `json:"avg_duration_ms"`
}

// CommandSummary represents a command usage summary
type CommandSummary struct {
	Name         string  `json:"name"`
	Count        int64   `json:"count"`
	SuccessRate  float64 `json:"success_rate"`
	AvgDuration  int64   `json:"avg_duration_ms"`
}

// HTTPEvent represents an HTTP server event
type HTTPEvent struct {
	SessionID    string
	Method       string
	Path         string
	StatusCode   int
	Duration     time.Duration
	ClientIP     string
	UserAgent    string
	AuthMethod   string
	AuthSuccess  bool
	ResponseSize int
}

// HTTPStats contains HTTP server statistics
type HTTPStats struct {
	TotalRequests     int64
	SuccessRate       float64
	AuthSuccessRate   float64
	AvgDurationMs     int64
	ErrorsLast24h     int64
	AuthErrorsLast24h int64
	TopPaths          []PathSummary
	StatusCodes       map[string]int
	AuthMethods       map[string]int
}

// PathSummary contains statistics for a specific path
type PathSummary struct {
	Path        string
	Count       int64
	SuccessRate float64
	AvgDuration int64
}

// WebhookStats contains upstream API/webhook statistics
type WebhookStats struct {
	TotalCalls        int64                `json:"total_calls"`
	SuccessRate       float64              `json:"success_rate"`
	AvgLatencyMs      int64                `json:"avg_latency_ms"`
	ErrorsLast24h     int64                `json:"errors_last_24h"`
	TopWebhooks       []WebhookSummary     `json:"top_webhooks"`
	ErrorBreakdown    map[string]int       `json:"error_breakdown"`
}

// WebhookSummary contains statistics for a specific webhook/API endpoint
type WebhookSummary struct {
	Name         string  `json:"name"`
	Count        int64   `json:"count"`
	SuccessRate  float64 `json:"success_rate"`
	AvgLatencyMs int64   `json:"avg_latency_ms"`
}

// NoOpAnalytics provides a no-operation implementation
type NoOpAnalytics struct{}

func (n *NoOpAnalytics) RecordCommand(ctx context.Context, event CommandEvent) {}
func (n *NoOpAnalytics) RecordHTTPEvent(ctx context.Context, event HTTPEvent) {}
func (n *NoOpAnalytics) GetStats(days int) (*UsageStats, error) {
	return &UsageStats{}, nil
}
func (n *NoOpAnalytics) GetHTTPStats(days int) (*HTTPStats, error) {
	return &HTTPStats{}, nil
}
func (n *NoOpAnalytics) GetWebhookStats(days int) (*WebhookStats, error) {
	return &WebhookStats{}, nil
}
func (n *NoOpAnalytics) Close() error { return nil }