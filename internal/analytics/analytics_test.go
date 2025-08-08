package analytics

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestSQLiteAnalytics(t *testing.T) {
	// Create temporary database
	tmpfile, err := os.CreateTemp("", "test_analytics.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	// Initialize analytics
	analytics, err := NewSQLiteAnalytics(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to create analytics: %v", err)
	}
	defer analytics.Close()

	// Record some test events
	ctx := context.Background()
	
	analytics.RecordCommand(ctx, CommandEvent{
		SessionID:     "session1",
		CommandName:   "test-cmd",
		Duration:      time.Millisecond * 100,
		Success:       true,
		OutputSize:    42,
		ExecutionMode: "local",
		Error:         "",
	})

	analytics.RecordCommand(ctx, CommandEvent{
		SessionID:     "session2",
		CommandName:   "test-cmd",
		Duration:      time.Millisecond * 200,
		Success:       false,
		OutputSize:    0,
		ExecutionMode: "container",
		Error:         "command failed",
	})

	// Get stats
	stats, err := analytics.GetStats(7)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	// Verify stats
	if stats.TotalCommands != 2 {
		t.Errorf("Expected 2 commands, got %d", stats.TotalCommands)
	}

	if stats.SuccessRate != 50.0 {
		t.Errorf("Expected 50%% success rate, got %.1f", stats.SuccessRate)
	}

	if len(stats.TopCommands) != 1 {
		t.Errorf("Expected 1 top command, got %d", len(stats.TopCommands))
	}

	if len(stats.TopCommands) > 0 && stats.TopCommands[0].Name != "test-cmd" {
		t.Errorf("Expected top command 'test-cmd', got '%s'", stats.TopCommands[0].Name)
	}
}

func TestNoOpAnalytics(t *testing.T) {
	analytics := &NoOpAnalytics{}
	
	// These should not panic or error
	analytics.RecordCommand(context.Background(), CommandEvent{})
	
	stats, err := analytics.GetStats(7)
	if err != nil {
		t.Errorf("NoOp analytics should not error: %v", err)
	}
	
	if stats == nil {
		t.Error("NoOp analytics should return empty stats, not nil")
	}
	
	err = analytics.Close()
	if err != nil {
		t.Errorf("NoOp analytics close should not error: %v", err)
	}
}