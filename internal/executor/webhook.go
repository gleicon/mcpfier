package executor

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gleicon/mcpfier/internal/config"
)

// WebhookExecutor handles webhook/API calls
type WebhookExecutor struct {
	client *http.Client
}

// NewWebhookExecutor creates a new webhook executor
func NewWebhookExecutor() *WebhookExecutor {
	return &WebhookExecutor{
		client: &http.Client{
			Timeout: 30 * time.Second, // Default timeout
		},
	}
}

// Execute performs a webhook/API call
func (e *WebhookExecutor) Execute(ctx context.Context, cmd *config.Command) (string, error) {
	if cmd.Webhook == nil {
		return "", fmt.Errorf("webhook configuration is nil")
	}

	webhook := cmd.Webhook
	
	// Set default method if not specified
	method := webhook.Method
	if method == "" {
		method = "GET"
	}

	// Create request body
	var body io.Reader
	if webhook.Body != "" {
		bodyContent, err := e.prepareRequestBody(webhook.Body, webhook.BodyFormat)
		if err != nil {
			return "", fmt.Errorf("failed to prepare request body: %w", err)
		}
		body = bytes.NewReader(bodyContent)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, method, webhook.URL, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	e.setHeaders(req, webhook)
	
	// Set authentication
	if err := e.setAuthentication(req, webhook.Auth); err != nil {
		return "", fmt.Errorf("failed to set authentication: %w", err)
	}

	// Apply timeout from command config
	if cmd.Timeout != "" {
		if duration, err := time.ParseDuration(cmd.Timeout); err == nil {
			e.client.Timeout = duration
		}
	}

	// Execute request with retries
	response, err := e.executeWithRetry(req, webhook.Retry)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer response.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check if response indicates an error
	if response.StatusCode >= 400 {
		return string(responseBody), fmt.Errorf("HTTP %d: %s", response.StatusCode, response.Status)
	}

	return string(responseBody), nil
}

// prepareRequestBody formats the request body according to the specified format
func (e *WebhookExecutor) prepareRequestBody(body, format string) ([]byte, error) {
	switch strings.ToLower(format) {
	case "json":
		// Validate JSON
		var jsonData interface{}
		if err := json.Unmarshal([]byte(body), &jsonData); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
		return []byte(body), nil
	case "xml", "text", "":
		return []byte(body), nil
	case "form":
		// URL-encoded form data - body should be key=value&key2=value2 format
		return []byte(body), nil
	default:
		return nil, fmt.Errorf("unsupported body format: %s", format)
	}
}

// setHeaders sets HTTP headers
func (e *WebhookExecutor) setHeaders(req *http.Request, webhook *config.WebhookConfig) {
	// Set content type based on body format
	if webhook.Body != "" && req.Header.Get("Content-Type") == "" {
		switch strings.ToLower(webhook.BodyFormat) {
		case "json":
			req.Header.Set("Content-Type", "application/json")
		case "xml":
			req.Header.Set("Content-Type", "application/xml")
		case "form":
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		default:
			req.Header.Set("Content-Type", "text/plain")
		}
	}

	// Set custom headers
	for key, value := range webhook.Headers {
		req.Header.Set(key, value)
	}

	// Set User-Agent if not provided
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "MCPFier/1.0.0")
	}
}

// setAuthentication sets authentication headers
func (e *WebhookExecutor) setAuthentication(req *http.Request, auth *config.WebhookAuth) error {
	if auth == nil {
		return nil
	}

	switch strings.ToLower(auth.Type) {
	case "bearer":
		if auth.Token == "" {
			return fmt.Errorf("bearer token is required")
		}
		req.Header.Set("Authorization", "Bearer "+auth.Token)
	
	case "api_key":
		if auth.Key == "" {
			return fmt.Errorf("API key is required")
		}
		header := auth.Header
		if header == "" {
			header = "X-API-Key"
		}
		req.Header.Set(header, auth.Key)
	
	case "basic":
		if auth.User == "" || auth.Pass == "" {
			return fmt.Errorf("username and password are required for basic auth")
		}
		credentials := base64.StdEncoding.EncodeToString([]byte(auth.User + ":" + auth.Pass))
		req.Header.Set("Authorization", "Basic "+credentials)
	
	case "oauth":
		return fmt.Errorf("OAuth authentication not yet implemented")
	
	default:
		return fmt.Errorf("unsupported authentication type: %s", auth.Type)
	}

	return nil
}

// executeWithRetry executes the request with retry logic
func (e *WebhookExecutor) executeWithRetry(req *http.Request, retryConfig *config.WebhookRetry) (*http.Response, error) {
	maxRetries := 3
	delay := time.Second
	backoff := "exponential"
	
	if retryConfig != nil {
		if retryConfig.MaxRetries > 0 {
			maxRetries = retryConfig.MaxRetries
		}
		if retryConfig.Delay != "" {
			if d, err := time.ParseDuration(retryConfig.Delay); err == nil {
				delay = d
			}
		}
		if retryConfig.Backoff != "" {
			backoff = retryConfig.Backoff
		}
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Clone request for retry (body needs to be reset)
		reqClone := req.Clone(req.Context())
		if req.Body != nil && req.GetBody != nil {
			body, err := req.GetBody()
			if err != nil {
				return nil, fmt.Errorf("failed to clone request body: %w", err)
			}
			reqClone.Body = body
		}

		resp, err := e.client.Do(reqClone)
		if err != nil {
			lastErr = err
		} else if e.shouldRetry(resp.StatusCode, retryConfig) {
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		} else {
			return resp, nil
		}

		// Don't sleep after the last attempt
		if attempt < maxRetries {
			sleepDuration := e.calculateDelay(delay, attempt, backoff)
			time.Sleep(sleepDuration)
		}
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries+1, lastErr)
}

// shouldRetry determines if a status code should be retried
func (e *WebhookExecutor) shouldRetry(statusCode int, retryConfig *config.WebhookRetry) bool {
	// Default retry on server errors and specific client errors
	defaultRetryCodes := []int{429, 502, 503, 504}
	
	if retryConfig != nil && len(retryConfig.StatusCodes) > 0 {
		for _, code := range retryConfig.StatusCodes {
			if code == statusCode {
				return true
			}
		}
		return false
	}

	// Check default codes
	for _, code := range defaultRetryCodes {
		if code == statusCode {
			return true
		}
	}

	return false
}

// calculateDelay calculates the delay for the next retry attempt
func (e *WebhookExecutor) calculateDelay(baseDelay time.Duration, attempt int, backoff string) time.Duration {
	switch strings.ToLower(backoff) {
	case "exponential":
		return baseDelay * time.Duration(1<<uint(attempt))
	case "linear":
		return baseDelay * time.Duration(attempt+1)
	case "fixed":
		return baseDelay
	default:
		return baseDelay * time.Duration(1<<uint(attempt)) // Default to exponential
	}
}