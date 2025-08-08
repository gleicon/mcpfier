package analytics

import (
	"context"
	"time"
)

// Analytics defines the interface for embedded analytics
type Analytics interface {
	RecordCommand(ctx context.Context, event CommandEvent)
	GetStats(days int) (*UsageStats, error)
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

// NoOpAnalytics provides a no-operation implementation
type NoOpAnalytics struct{}

func (n *NoOpAnalytics) RecordCommand(ctx context.Context, event CommandEvent) {}
func (n *NoOpAnalytics) GetStats(days int) (*UsageStats, error) {
	return &UsageStats{}, nil
}
func (n *NoOpAnalytics) Close() error { return nil }