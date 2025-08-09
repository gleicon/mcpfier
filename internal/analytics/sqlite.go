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

	CREATE TABLE IF NOT EXISTS http_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		session_id TEXT,
		method TEXT,
		path TEXT,
		status_code INTEGER,
		duration_ms INTEGER,
		client_ip TEXT,
		user_agent TEXT,
		auth_method TEXT,
		auth_success BOOLEAN,
		response_size INTEGER
	);

	CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
	CREATE INDEX IF NOT EXISTS idx_events_command ON events(command_name);
	CREATE INDEX IF NOT EXISTS idx_http_events_timestamp ON http_events(timestamp);
	CREATE INDEX IF NOT EXISTS idx_http_events_path ON http_events(path);
	CREATE INDEX IF NOT EXISTS idx_http_events_status ON http_events(status_code);
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
		log.Printf("Analytics command recording failed: %v", err)
	}
}

// RecordHTTPEvent records an HTTP server event
func (a *SQLiteAnalytics) RecordHTTPEvent(ctx context.Context, event HTTPEvent) {
	// Sync insert for now to ensure data is written
	_, err := a.db.Exec(`
		INSERT INTO http_events (session_id, method, path, status_code, duration_ms,
								client_ip, user_agent, auth_method, auth_success, response_size)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		event.SessionID, event.Method, event.Path, event.StatusCode, event.Duration.Milliseconds(),
		event.ClientIP, event.UserAgent, event.AuthMethod, event.AuthSuccess, event.ResponseSize)
	
	if err != nil {
		log.Printf("Analytics HTTP recording failed: %v", err)
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

// GetHTTPStats returns HTTP server statistics for the specified number of days
func (a *SQLiteAnalytics) GetHTTPStats(days int) (*HTTPStats, error) {
	// Main HTTP stats query
	row := a.db.QueryRow(`
		SELECT 
			COUNT(*) as total_requests,
			COALESCE(AVG(CASE WHEN status_code < 400 THEN 1.0 ELSE 0.0 END), 0) as success_rate,
			COALESCE(AVG(CASE WHEN auth_success THEN 1.0 ELSE 0.0 END), 0) as auth_success_rate,
			COALESCE(AVG(duration_ms), 0) as avg_duration,
			COALESCE(SUM(CASE WHEN timestamp > datetime('now', '-1 day') AND status_code >= 400 THEN 1 ELSE 0 END), 0) as errors_24h,
			COALESCE(SUM(CASE WHEN timestamp > datetime('now', '-1 day') AND auth_success = 0 THEN 1 ELSE 0 END), 0) as auth_errors_24h
		FROM http_events 
		WHERE timestamp > datetime('now', '-' || ? || ' days')`, days)

	var stats HTTPStats
	var avgDuration float64
	err := row.Scan(&stats.TotalRequests, &stats.SuccessRate, &stats.AuthSuccessRate, 
		&avgDuration, &stats.ErrorsLast24h, &stats.AuthErrorsLast24h)
	if err != nil {
		return nil, err
	}

	stats.AvgDurationMs = int64(avgDuration)
	stats.SuccessRate = stats.SuccessRate * 100 // Convert to percentage
	stats.AuthSuccessRate = stats.AuthSuccessRate * 100 // Convert to percentage

	// Top paths query
	rows, err := a.db.Query(`
		SELECT 
			path,
			COUNT(*) as count,
			COALESCE(AVG(CASE WHEN status_code < 400 THEN 1.0 ELSE 0.0 END), 0) * 100 as success_rate,
			COALESCE(AVG(duration_ms), 0) as avg_duration
		FROM http_events 
		WHERE timestamp > datetime('now', '-' || ? || ' days')
		GROUP BY path 
		ORDER BY count DESC 
		LIMIT 10`, days)
	
	if err == nil {
		defer rows.Close()
		var topPaths []PathSummary
		for rows.Next() {
			var path PathSummary
			var avgDur float64
			err := rows.Scan(&path.Path, &path.Count, &path.SuccessRate, &avgDur)
			if err == nil {
				path.AvgDuration = int64(avgDur)
				topPaths = append(topPaths, path)
			}
		}
		stats.TopPaths = topPaths
	}

	// Status codes breakdown
	statusRows, err := a.db.Query(`
		SELECT status_code, COUNT(*) as count
		FROM http_events 
		WHERE timestamp > datetime('now', '-' || ? || ' days')
		GROUP BY status_code`, days)
	
	if err == nil {
		defer statusRows.Close()
		stats.StatusCodes = make(map[string]int)
		for statusRows.Next() {
			var statusCode int
			var count int
			if statusRows.Scan(&statusCode, &count) == nil {
				stats.StatusCodes[fmt.Sprintf("%d", statusCode)] = count
			}
		}
	}

	// Auth methods breakdown
	authRows, err := a.db.Query(`
		SELECT auth_method, COUNT(*) as count
		FROM http_events 
		WHERE timestamp > datetime('now', '-' || ? || ' days')
		GROUP BY auth_method`, days)
	
	if err == nil {
		defer authRows.Close()
		stats.AuthMethods = make(map[string]int)
		for authRows.Next() {
			var authMethod string
			var count int
			if authRows.Scan(&authMethod, &count) == nil {
				stats.AuthMethods[authMethod] = count
			}
		}
	}

	return &stats, nil
}

// GetWebhookStats returns webhook/API call statistics for the specified number of days
func (a *SQLiteAnalytics) GetWebhookStats(days int) (*WebhookStats, error) {
	// Main webhook stats query (only webhook execution mode)
	row := a.db.QueryRow(`
		SELECT 
			COUNT(*) as total_calls,
			COALESCE(AVG(CASE WHEN success THEN 1.0 ELSE 0.0 END), 0) as success_rate,
			COALESCE(AVG(duration_ms), 0) as avg_latency,
			COALESCE(SUM(CASE WHEN timestamp > datetime('now', '-1 day') AND success = 0 THEN 1 ELSE 0 END), 0) as errors_24h
		FROM events 
		WHERE timestamp > datetime('now', '-' || ? || ' days')
		AND execution_mode = 'webhook'`, days)

	var stats WebhookStats
	var avgLatency float64
	err := row.Scan(&stats.TotalCalls, &stats.SuccessRate, &avgLatency, &stats.ErrorsLast24h)
	if err != nil {
		return nil, err
	}

	stats.AvgLatencyMs = int64(avgLatency)
	stats.SuccessRate = stats.SuccessRate * 100 // Convert to percentage

	// Top webhooks query (only webhook execution mode)
	rows, err := a.db.Query(`
		SELECT 
			command_name,
			COUNT(*) as count,
			COALESCE(AVG(CASE WHEN success THEN 1.0 ELSE 0.0 END), 0) * 100 as success_rate,
			COALESCE(AVG(duration_ms), 0) as avg_latency
		FROM events 
		WHERE timestamp > datetime('now', '-' || ? || ' days')
		AND execution_mode = 'webhook'
		GROUP BY command_name 
		ORDER BY count DESC 
		LIMIT 10`, days)
	
	if err != nil {
		return &stats, nil // Return partial stats if top webhooks query fails
	}
	defer rows.Close()

	var topWebhooks []WebhookSummary
	for rows.Next() {
		var webhook WebhookSummary
		var avgLatency float64
		err := rows.Scan(&webhook.Name, &webhook.Count, &webhook.SuccessRate, &avgLatency)
		if err != nil {
			continue
		}
		webhook.AvgLatencyMs = int64(avgLatency)
		topWebhooks = append(topWebhooks, webhook)
	}

	stats.TopWebhooks = topWebhooks

	// Error breakdown query (only webhook execution mode, last 24h)
	errorRows, err := a.db.Query(`
		SELECT 
			CASE 
				WHEN error_message LIKE '%HTTP 4%' THEN '4xx Client Errors'
				WHEN error_message LIKE '%HTTP 5%' THEN '5xx Server Errors'
				WHEN error_message LIKE '%timeout%' OR error_message LIKE '%deadline%' THEN 'Timeout/Deadline'
				WHEN error_message LIKE '%connection%' OR error_message LIKE '%network%' THEN 'Connection Issues'
				ELSE 'Other Errors'
			END as error_type,
			COUNT(*) as count
		FROM events 
		WHERE timestamp > datetime('now', '-1 day')
		AND execution_mode = 'webhook'
		AND success = 0
		AND error_message != ''
		GROUP BY error_type
		ORDER BY count DESC`)
	
	if err == nil {
		defer errorRows.Close()
		stats.ErrorBreakdown = make(map[string]int)
		for errorRows.Next() {
			var errorType string
			var count int
			if errorRows.Scan(&errorType, &count) == nil {
				stats.ErrorBreakdown[errorType] = count
			}
		}
	}

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