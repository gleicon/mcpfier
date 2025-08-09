package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gleicon/mcpfier/internal/config"
)

// AuthContext represents authentication context
type AuthContext struct {
	UserID      string   `json:"user_id"`
	ClientName  string   `json:"client_name"`
	Permissions []string `json:"permissions"`
	Method      string   `json:"method"` // "api_key", "oauth", etc.
}

// contextKey is a custom type for context keys to avoid collisions
type contextKey struct{}

var authContextKey = contextKey{}

// WithAuthContext adds authentication context to the request context
func WithAuthContext(ctx context.Context, auth *AuthContext) context.Context {
	return context.WithValue(ctx, authContextKey, auth)
}

// AuthContextFromRequest extracts authentication context from request context
func AuthContextFromRequest(ctx context.Context) (*AuthContext, bool) {
	auth, ok := ctx.Value(authContextKey).(*AuthContext)
	return auth, ok
}

// Middleware creates HTTP authentication middleware
func Middleware(cfg *config.AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication if disabled
			if !cfg.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Extract authentication from request
			authCtx, err := extractAuth(r, cfg)
			if err != nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Add authentication context to request
			ctx := WithAuthContext(r.Context(), authCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractAuth extracts authentication information from the request
func extractAuth(r *http.Request, cfg *config.AuthConfig) (*AuthContext, error) {
	switch cfg.Mode {
	case "simple":
		return extractSimpleAuth(r, &cfg.Simple)
	default:
		return nil, errors.New("unsupported authentication mode")
	}
}

// extractSimpleAuth handles simple authentication (API keys)
func extractSimpleAuth(r *http.Request, cfg *config.SimpleAuthConfig) (*AuthContext, error) {
	// Try X-API-Key header first
	apiKey := r.Header.Get("X-API-Key")
	
	// Try Authorization header with ApiKey prefix
	if apiKey == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "ApiKey ") {
			apiKey = strings.TrimPrefix(authHeader, "ApiKey ")
		}
	}

	if apiKey == "" {
		return nil, errors.New("missing API key")
	}

	// Find matching API key in configuration
	for _, key := range cfg.APIKeys {
		if key.Key == apiKey {
			return &AuthContext{
				UserID:      key.Name,
				ClientName:  key.Name,
				Permissions: key.Permissions,
				Method:      "api_key",
			}, nil
		}
	}

	return nil, errors.New("invalid API key")
}

// HasPermission checks if the authenticated user has permission for a specific tool
func (a *AuthContext) HasPermission(toolName string) bool {
	if a == nil {
		return false
	}

	// Check for wildcard permission
	for _, perm := range a.Permissions {
		if perm == "*" {
			return true
		}
		if perm == toolName {
			return true
		}
	}

	return false
}

// ContextFunc creates a context function for mcp-go server
func ContextFunc(cfg *config.AuthConfig) func(context.Context, *http.Request) context.Context {
	return func(ctx context.Context, r *http.Request) context.Context {
		// Skip authentication if disabled
		if !cfg.Enabled {
			return ctx
		}

		// Extract authentication from request
		authCtx, err := extractAuth(r, cfg)
		if err != nil {
			// Return context without auth - will be handled by middleware
			return ctx
		}

		// Add authentication context
		return WithAuthContext(ctx, authCtx)
	}
}