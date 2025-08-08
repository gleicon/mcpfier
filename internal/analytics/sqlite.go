package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// SQLiteAnalytics implements Analytics using SQLite
type SQLiteAnalytics struct {
	db   *sql.DB
	path string
}

// NewSQLiteAnalytics creates a new SQLite analytics instance
func NewSQLiteAnalytics(dbPath string) (*SQLiteAnalytics, error) {
	// Resolve path (handle ~ and relative paths)
	resolvedPath, err := resolvePath(dbPath)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", resolvedPath)
	if err != nil {
		return nil, err
	}

	a := &SQLiteAnalytics{db: db, path: resolvedPath}
	if err := a.createTables(); err != nil {
		return nil, err
	}

	return a, nil
}

// createTables creates the necessary database tables
func (a *SQLiteAnalytics) createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		session_id TEXT,
		command_name TEXT,
		duration_ms INTEGER,
		success BOOLEAN,
		error_message TEXT,
		output_size INTEGER,
		execution_mode TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
	CREATE INDEX IF NOT EXISTS idx_events_command ON events(command_name);
	`

	_, err := a.db.Exec(schema)
	return err
}

// RecordCommand records a command execution event
func (a *SQLiteAnalytics) RecordCommand(ctx context.Context, event CommandEvent) {
	// Sync insert for now to ensure data is written
	_, err := a.db.Exec(`
		INSERT INTO events (session_id, command_name, duration_ms, success, 
						   error_message, output_size, execution_mode)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		event.SessionID, event.CommandName, event.Duration.Milliseconds(),
		event.Success, event.Error, event.OutputSize, event.ExecutionMode)
	
	if err != nil {
		log.Printf("Analytics recording failed: %v", err)
	}
}

// GetStats returns usage statistics for the specified number of days
func (a *SQLiteAnalytics) GetStats(days int) (*UsageStats, error) {
	// Main stats query
	row := a.db.QueryRow(`
		SELECT 
			COUNT(*) as total_commands,
			COALESCE(AVG(CASE WHEN success THEN 1.0 ELSE 0.0 END), 0) as success_rate,
			COALESCE(AVG(duration_ms), 0) as avg_duration,
			COALESCE(SUM(CASE WHEN timestamp > datetime('now', '-1 day') AND success = 0 THEN 1 ELSE 0 END), 0) as errors_24h
		FROM events 
		WHERE timestamp > datetime('now', '-' || ? || ' days')`, days)

	var stats UsageStats
	var avgDuration float64
	err := row.Scan(&stats.TotalCommands, &stats.SuccessRate, &avgDuration, &stats.ErrorsLast24h)
	if err != nil {
		return nil, err
	}

	stats.AvgDurationMs = int64(avgDuration)
	stats.SuccessRate = stats.SuccessRate * 100 // Convert to percentage

	// Top commands query
	rows, err := a.db.Query(`
		SELECT 
			command_name,
			COUNT(*) as count,
			COALESCE(AVG(CASE WHEN success THEN 1.0 ELSE 0.0 END), 0) * 100 as success_rate,
			COALESCE(AVG(duration_ms), 0) as avg_duration
		FROM events 
		WHERE timestamp > datetime('now', '-' || ? || ' days')
		GROUP BY command_name 
		ORDER BY count DESC 
		LIMIT 10`, days)
	
	if err != nil {
		return &stats, nil // Return partial stats if top commands query fails
	}
	defer rows.Close()

	var topCommands []CommandSummary
	for rows.Next() {
		var cmd CommandSummary
		var avgDur float64
		err := rows.Scan(&cmd.Name, &cmd.Count, &cmd.SuccessRate, &avgDur)
		if err != nil {
			continue
		}
		cmd.AvgDuration = int64(avgDur)
		topCommands = append(topCommands, cmd)
	}

	stats.TopCommands = topCommands
	return &stats, nil
}

// Close closes the database connection
func (a *SQLiteAnalytics) Close() error {
	return a.db.Close()
}

// resolvePath resolves ~ and relative paths to absolute paths
func resolvePath(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	// Handle ~ for home directory
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(homeDir, path[2:])
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	// Check if directory exists, return error if it doesn't
	dir := filepath.Dir(absPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", fmt.Errorf("directory does not exist: %s", dir)
	}

	return absPath, nil
}