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
}

// Config holds all the commands from the YAML file
type Config struct {
	Commands  []Command       `yaml:"commands"`
	Analytics AnalyticsConfig `yaml:"analytics"`
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

	return &config, nil
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