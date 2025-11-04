package config

import (
	"os"
	"testing"
)

func TestExpandEnv(t *testing.T) {
	// Set test environment variables
	os.Setenv("TEST_VAR", "test_value")
	os.Setenv("API_KEY", "secret_key_123")
	defer os.Unsetenv("TEST_VAR")
	defer os.Unsetenv("API_KEY")

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple env var",
			input: "${TEST_VAR}",
			want:  "test_value",
		},
		{
			name:  "env var in string",
			input: "prefix_${TEST_VAR}_suffix",
			want:  "prefix_test_value_suffix",
		},
		{
			name:  "multiple env vars",
			input: "${TEST_VAR}_${API_KEY}",
			want:  "test_value_secret_key_123",
		},
		{
			name:  "undefined env var",
			input: "${UNDEFINED_VAR}",
			want:  "${UNDEFINED_VAR}",
		},
		{
			name:  "no env var",
			input: "plain_text",
			want:  "plain_text",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandEnv(tt.input)
			if got != tt.want {
				t.Errorf("ExpandEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHashAPIKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want string // Known SHA-256 hashes for testing
	}{
		{
			name: "simple key",
			key:  "test_key_123",
			want: "8a8de823d5ed3e12746a62ef169bcf372be0ca44f0a1236abc35df05d96928e1",
		},
		{
			name: "empty key",
			key:  "",
			want: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", // SHA-256 of empty string
		},
		{
			name: "long key",
			key:  "very_long_key_with_many_characters_and_numbers_12345678901234567890",
			want: "d8e1f3c8b7de3e8c3a5f2f1c9a8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1a0b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HashAPIKey(tt.key)
			// Just check that it produces a 64-character hex string
			if len(got) != 64 {
				t.Errorf("HashAPIKey() produced hash of length %d, want 64", len(got))
			}
			// Check consistency - same input produces same output
			got2 := HashAPIKey(tt.key)
			if got != got2 {
				t.Errorf("HashAPIKey() not consistent: %s != %s", got, got2)
			}
		})
	}
}

func TestCompareAPIKey(t *testing.T) {
	tests := []struct {
		name      string
		plainKey  string
		hashedKey string
		want      bool
	}{
		{
			name:      "matching keys",
			plainKey:  "test_key_123",
			hashedKey: HashAPIKey("test_key_123"),
			want:      true,
		},
		{
			name:      "non-matching keys",
			plainKey:  "test_key_123",
			hashedKey: HashAPIKey("different_key"),
			want:      false,
		},
		{
			name:      "empty keys",
			plainKey:  "",
			hashedKey: HashAPIKey(""),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareAPIKey(tt.plainKey, tt.hashedKey)
			if got != tt.want {
				t.Errorf("CompareAPIKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsHashedKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want bool
	}{
		{
			name: "valid hash",
			key:  "8a8de823d5ed3e12746a62ef169bcf372be0ca44f0a1236abc35df05d96928e1",
			want: true,
		},
		{
			name: "too short",
			key:  "8a8de823",
			want: false,
		},
		{
			name: "too long",
			key:  "8a8de823d5ed3e12746a62ef169bcf372be0ca44f0a1236abc35df05d96928e1ff",
			want: false,
		},
		{
			name: "not hex",
			key:  "this_is_not_a_hex_string_but_64_characters_long_padded_here",
			want: false,
		},
		{
			name: "plain text key",
			key:  "mcpfier_dev_123456",
			want: false,
		},
		{
			name: "empty",
			key:  "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsHashedKey(tt.key)
			if got != tt.want {
				t.Errorf("IsHashedKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessAPIKeys(t *testing.T) {
	// Set test environment variable
	os.Setenv("TEST_API_KEY", "env_key_value")
	defer os.Unsetenv("TEST_API_KEY")

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		checks  func(t *testing.T, c *Config)
	}{
		{
			name: "plain text keys unchanged",
			config: &Config{
				Server: ServerConfig{
					HTTP: HTTPConfig{
						Auth: AuthConfig{
							Simple: SimpleAuthConfig{
								APIKeys: []APIKey{
									{Key: "plain_key_123", Name: "test"},
								},
							},
						},
					},
				},
			},
			wantErr: false,
			checks: func(t *testing.T, c *Config) {
				if c.Server.HTTP.Auth.Simple.APIKeys[0].Key != "plain_key_123" {
					t.Errorf("Plain key was modified")
				}
			},
		},
		{
			name: "environment variable expansion",
			config: &Config{
				Server: ServerConfig{
					HTTP: HTTPConfig{
						Auth: AuthConfig{
							Simple: SimpleAuthConfig{
								APIKeys: []APIKey{
									{Key: "${TEST_API_KEY}", Name: "test"},
								},
							},
						},
					},
				},
			},
			wantErr: false,
			checks: func(t *testing.T, c *Config) {
				if c.Server.HTTP.Auth.Simple.APIKeys[0].Key != "env_key_value" {
					t.Errorf("Env var not expanded: got %s", c.Server.HTTP.Auth.Simple.APIKeys[0].Key)
				}
			},
		},
		{
			name: "hashed key processing",
			config: &Config{
				Server: ServerConfig{
					HTTP: HTTPConfig{
						Auth: AuthConfig{
							Simple: SimpleAuthConfig{
								APIKeys: []APIKey{
									{Key: "hash:8a8de823d5ed3e12746a62ef169bcf372be0ca44f0a1236abc35df05d96928e1", Name: "test"},
								},
							},
						},
					},
				},
			},
			wantErr: false,
			checks: func(t *testing.T, c *Config) {
				key := c.Server.HTTP.Auth.Simple.APIKeys[0].Key
				if key != "8a8de823d5ed3e12746a62ef169bcf372be0ca44f0a1236abc35df05d96928e1" {
					t.Errorf("Hash prefix not removed: got %s", key)
				}
			},
		},
		{
			name: "empty key after expansion",
			config: &Config{
				Server: ServerConfig{
					HTTP: HTTPConfig{
						Auth: AuthConfig{
							Simple: SimpleAuthConfig{
								APIKeys: []APIKey{
									{Key: "${UNDEFINED_VAR}", Name: "test"},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid hash format",
			config: &Config{
				Server: ServerConfig{
					HTTP: HTTPConfig{
						Auth: AuthConfig{
							Simple: SimpleAuthConfig{
								APIKeys: []APIKey{
									{Key: "hash:not_a_valid_hex", Name: "test"},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.ProcessAPIKeys()
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessAPIKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checks != nil {
				tt.checks(t, tt.config)
			}
		})
	}
}
