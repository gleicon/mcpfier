package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Command represents the structure of a command in the config
type Command struct {
	Name        string            `yaml:"name"`
	Script      string            `yaml:"script"`
	Args        []string          `yaml:"args"`
	Description string            `yaml:"description"`
	Container   string            `yaml:"container"`
	Timeout     string            `yaml:"timeout"`
	Env         map[string]string `yaml:"env"`
	// Webhook/API configuration
	Webhook     *WebhookConfig    `yaml:"webhook,omitempty"`
}

// WebhookConfig represents webhook/API call configuration
type WebhookConfig struct {
	URL         string            `yaml:"url"`
	Method      string            `yaml:"method"`      // GET, POST, PUT, DELETE, etc.
	Headers     map[string]string `yaml:"headers"`     // Custom headers
	Body        string            `yaml:"body"`        // Request body template
	BodyFormat  string            `yaml:"body_format"` // json, xml, form, text
	Auth        *WebhookAuth      `yaml:"auth,omitempty"`
	Retry       *WebhookRetry     `yaml:"retry,omitempty"`
}

// WebhookAuth represents authentication for webhook calls
type WebhookAuth struct {
	Type   string `yaml:"type"`   // bearer, api_key, basic, oauth
	Token  string `yaml:"token"`  // For bearer token
	Key    string `yaml:"key"`    // For API key auth
	Header string `yaml:"header"` // Header name for API key (default: X-API-Key)
	User   string `yaml:"user"`   // For basic auth
	Pass   string `yaml:"pass"`   // For basic auth
}

// WebhookRetry represents retry configuration
type WebhookRetry struct {
	MaxRetries int      `yaml:"max_retries"` // Default: 3
	Backoff    string   `yaml:"backoff"`     // exponential, linear, fixed
	Delay      string   `yaml:"delay"`       // Initial delay, e.g. "1s"
	StatusCodes []int   `yaml:"status_codes,omitempty"` // Which status codes to retry
}

// Config holds all the commands from the YAML file
type Config struct {
	Commands  []Command       `yaml:"commands"`
	Server    ServerConfig    `yaml:"server"`
	Analytics AnalyticsConfig `yaml:"analytics"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	HTTP HTTPConfig `yaml:"http"`
}

// HTTPConfig holds HTTP server configuration
type HTTPConfig struct {
	Port int        `yaml:"port"`
	Host string     `yaml:"host"`
	Auth AuthConfig `yaml:"auth"`
	CORS CORSConfig `yaml:"cors"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled bool             `yaml:"enabled"`
	Mode    string           `yaml:"mode"` // "simple" or "enterprise"
	Simple  SimpleAuthConfig `yaml:"simple"`
}

// SimpleAuthConfig holds simple authentication configuration
type SimpleAuthConfig struct {
	APIKeys []APIKey `yaml:"api_keys"`
}

// APIKey represents an API key configuration
type APIKey struct {
	Key         string   `yaml:"key"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Permissions []string `yaml:"permissions"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	Enabled        bool     `yaml:"enabled"`
	AllowedOrigins []string `yaml:"allowed_origins"`
	AllowedMethods []string `yaml:"allowed_methods"`
	AllowedHeaders []string `yaml:"allowed_headers"`
}

// AnalyticsConfig holds analytics configuration
type AnalyticsConfig struct {
	Enabled      bool   `yaml:"enabled"`
	DatabasePath string `yaml:"database_path"`
	RetentionDays int   `yaml:"retention_days"`
}

// Load loads the configuration from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Apply defaults
	config.applyDefaults()

	return &config, nil
}

// applyDefaults applies default values to the configuration
func (c *Config) applyDefaults() {
	// Server defaults
	if c.Server.HTTP.Port == 0 {
		c.Server.HTTP.Port = 8080
	}
	if c.Server.HTTP.Host == "" {
		c.Server.HTTP.Host = "localhost"
	}

	// Auth defaults
	if c.Server.HTTP.Auth.Mode == "" {
		c.Server.HTTP.Auth.Mode = "simple"
	}

	// CORS defaults
	if c.Server.HTTP.CORS.Enabled && len(c.Server.HTTP.CORS.AllowedMethods) == 0 {
		c.Server.HTTP.CORS.AllowedMethods = []string{"GET", "POST", "OPTIONS"}
	}
	if c.Server.HTTP.CORS.Enabled && len(c.Server.HTTP.CORS.AllowedHeaders) == 0 {
		c.Server.HTTP.CORS.AllowedHeaders = []string{"Authorization", "Content-Type", "X-API-Key"}
	}
}

// LoadFromDefaultPath loads configuration using the default search paths
func LoadFromDefaultPath() (*Config, error) {
	path := FindConfigFile()
	return Load(path)
}

// FindConfigFile searches for config.yaml in standard locations
func FindConfigFile() string {
	// Check environment variable first
	if configPath := os.Getenv("MCPFIER_CONFIG"); configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}
	
	// Try various config file locations
	candidates := []string{
		"config.yaml",                    // Current directory
		"./config.yaml",                 // Explicit current directory
	}
	
	// Try config next to executable
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		candidates = append(candidates, filepath.Join(execDir, "config.yaml"))
	}
	
	// Try user's home directory
	if homeDir, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(homeDir, ".mcpfier", "config.yaml"))
		candidates = append(candidates, filepath.Join(homeDir, "mcpfier", "config.yaml"))
	}
	
	// Try system-wide locations
	candidates = append(candidates, "/etc/mcpfier/config.yaml")
	
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	
	// Default to config.yaml if nothing found
	return "config.yaml"
}

// GetDescription returns a description for the command
func (c Command) GetDescription() string {
	if c.Description != "" {
		return c.Description
	}
	return "Execute " + c.Name + " with configured arguments"
}

// IsContainerized returns true if the command should run in a container
func (c Command) IsContainerized() bool {
	return c.Container != ""
}

// IsWebhook returns true if the command is a webhook/API call
func (c Command) IsWebhook() bool {
	return c.Webhook != nil && c.Webhook.URL != ""
}