package config

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// envVarPattern matches environment variables in the format ${VAR_NAME}
var envVarPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

// ExpandEnv expands environment variables in the format ${VAR_NAME}
func ExpandEnv(s string) string {
	return envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		// Extract variable name (remove ${ and })
		varName := match[2 : len(match)-1]
		if value := os.Getenv(varName); value != "" {
			return value
		}
		return match // Return original if env var not found
	})
}

// HashAPIKey creates a SHA-256 hash of an API key
// For storing hashed keys in configuration
func HashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// CompareAPIKey compares a plain text API key against a hash
// Returns true if they match
func CompareAPIKey(plainKey, hashedKey string) bool {
	computedHash := HashAPIKey(plainKey)
	return computedHash == hashedKey
}

// ProcessAPIKeys processes API keys in the configuration
// - Expands environment variables (${VAR_NAME})
// - Handles hashed keys (prefix: hash:)
// - Validates and normalizes keys
func (c *Config) ProcessAPIKeys() error {
	for i := range c.Server.HTTP.Auth.Simple.APIKeys {
		key := &c.Server.HTTP.Auth.Simple.APIKeys[i]

		// Expand environment variables in the key
		key.Key = ExpandEnv(key.Key)

		// Check if expansion failed (key still contains ${...})
		if strings.Contains(key.Key, "${") {
			return fmt.Errorf("API key for '%s' contains unexpanded environment variable", key.Name)
		}

		// Validate key is not empty after expansion
		if key.Key == "" {
			return fmt.Errorf("API key for '%s' is empty", key.Name)
		}

		// Check if it's a hashed key (prefix: hash:)
		if strings.HasPrefix(key.Key, "hash:") {
			// Store the hash without prefix
			key.Key = strings.TrimPrefix(key.Key, "hash:")
			// Validate it's a valid hex string
			if _, err := hex.DecodeString(key.Key); err != nil {
				return fmt.Errorf("invalid hashed API key for '%s': must be valid hex", key.Name)
			}
		}
	}

	return nil
}

// IsHashedKey checks if a key string is a hashed key
func IsHashedKey(key string) bool {
	// Check if it's a 64-character hex string (SHA-256)
	if len(key) != 64 {
		return false
	}
	_, err := hex.DecodeString(key)
	return err == nil
}
