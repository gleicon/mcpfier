package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gleicon/mcpfier/internal/config"
)

func TestExtractSimpleAuth(t *testing.T) {
	// Create hashed version of test key
	hashedKey := config.HashAPIKey("hashed_key_789")

	cfg := &config.SimpleAuthConfig{
		APIKeys: []config.APIKey{
			{
				Key:         "test_key_123",
				Name:        "test-client",
				Permissions: []string{"tool1", "tool2"},
			},
			{
				Key:         "admin_key_456",
				Name:        "admin",
				Permissions: []string{"*"},
			},
			{
				Key:         hashedKey, // Hashed key (64-char hex)
				Name:        "hashed-client",
				Permissions: []string{"tool3"},
			},
		},
	}

	tests := []struct {
		name        string
		header      string
		headerValue string
		wantErr     bool
		wantName    string
		wantPerms   []string
	}{
		{
			name:        "valid X-API-Key header",
			header:      "X-API-Key",
			headerValue: "test_key_123",
			wantErr:     false,
			wantName:    "test-client",
			wantPerms:   []string{"tool1", "tool2"},
		},
		{
			name:        "valid Authorization ApiKey header",
			header:      "Authorization",
			headerValue: "ApiKey test_key_123",
			wantErr:     false,
			wantName:    "test-client",
			wantPerms:   []string{"tool1", "tool2"},
		},
		{
			name:        "admin with wildcard permission",
			header:      "X-API-Key",
			headerValue: "admin_key_456",
			wantErr:     false,
			wantName:    "admin",
			wantPerms:   []string{"*"},
		},
		{
			name:        "hashed key authentication",
			header:      "X-API-Key",
			headerValue: "hashed_key_789", // Plain key that matches the hash
			wantErr:     false,
			wantName:    "hashed-client",
			wantPerms:   []string{"tool3"},
		},
		{
			name:        "invalid API key",
			header:      "X-API-Key",
			headerValue: "invalid_key",
			wantErr:     true,
		},
		{
			name:    "missing API key",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.header != "" {
				req.Header.Set(tt.header, tt.headerValue)
			}

			authCtx, err := extractSimpleAuth(req, cfg)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if authCtx.ClientName != tt.wantName {
				t.Errorf("ClientName = %v, want %v", authCtx.ClientName, tt.wantName)
			}

			if len(authCtx.Permissions) != len(tt.wantPerms) {
				t.Errorf("Permissions length = %v, want %v", len(authCtx.Permissions), len(tt.wantPerms))
			}

			for i, perm := range tt.wantPerms {
				if authCtx.Permissions[i] != perm {
					t.Errorf("Permission[%d] = %v, want %v", i, authCtx.Permissions[i], perm)
				}
			}
		})
	}
}

func TestAuthContextHasPermission(t *testing.T) {
	tests := []struct {
		name        string
		authCtx     *AuthContext
		toolName    string
		wantAllowed bool
	}{
		{
			name: "wildcard permission allows all",
			authCtx: &AuthContext{
				Permissions: []string{"*"},
			},
			toolName:    "any_tool",
			wantAllowed: true,
		},
		{
			name: "specific permission allows tool",
			authCtx: &AuthContext{
				Permissions: []string{"tool1", "tool2"},
			},
			toolName:    "tool1",
			wantAllowed: true,
		},
		{
			name: "missing permission denies tool",
			authCtx: &AuthContext{
				Permissions: []string{"tool1", "tool2"},
			},
			toolName:    "tool3",
			wantAllowed: false,
		},
		{
			name:        "nil context denies",
			authCtx:     nil,
			toolName:    "any_tool",
			wantAllowed: false,
		},
		{
			name: "empty permissions denies",
			authCtx: &AuthContext{
				Permissions: []string{},
			},
			toolName:    "any_tool",
			wantAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed := tt.authCtx.HasPermission(tt.toolName)
			if allowed != tt.wantAllowed {
				t.Errorf("HasPermission() = %v, want %v", allowed, tt.wantAllowed)
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	cfg := &config.AuthConfig{
		Enabled: true,
		Mode:    "simple",
		Simple: config.SimpleAuthConfig{
			APIKeys: []config.APIKey{
				{
					Key:         "valid_key",
					Name:        "test",
					Permissions: []string{"*"},
				},
			},
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authCtx, ok := AuthContextFromRequest(r.Context())
		if !ok {
			t.Error("expected auth context in request")
			return
		}
		if authCtx.ClientName != "test" {
			t.Errorf("ClientName = %v, want test", authCtx.ClientName)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := Middleware(cfg)
	wrappedHandler := middleware(handler)

	tests := []struct {
		name           string
		apiKey         string
		wantStatusCode int
	}{
		{
			name:           "valid API key",
			apiKey:         "valid_key",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "invalid API key",
			apiKey:         "invalid_key",
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "missing API key",
			apiKey:         "",
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.apiKey != "" {
				req.Header.Set("X-API-Key", tt.apiKey)
			}
			rr := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Errorf("status code = %v, want %v", rr.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestMiddlewareDisabled(t *testing.T) {
	cfg := &config.AuthConfig{
		Enabled: false,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := Middleware(cfg)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	// No API key provided
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code = %v, want %v when auth disabled", rr.Code, http.StatusOK)
	}
}
