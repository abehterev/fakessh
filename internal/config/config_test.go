package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Check default port
	if cfg.Port != 2222 {
		t.Errorf("Expected default port 2222, got %d", cfg.Port)
	}

	// Check default log file
	if cfg.Log.File != "credentials.log" {
		t.Errorf("Expected default log file 'credentials.log', got '%s'", cfg.Log.File)
	}

	// Check default log format
	if cfg.Log.Format != "json" {
		t.Errorf("Expected default log format 'json', got '%s'", cfg.Log.Format)
	}

	// Check default banner
	if cfg.Banner != "Ubuntu-4ubuntu0.5" {
		t.Errorf("Expected default banner 'Ubuntu-4ubuntu0.5', got '%s'", cfg.Banner)
	}

	// Check default SSH server version
	if cfg.ServerVersion != "OpenSSH_8.2p1" {
		t.Errorf("Expected default SSH server version 'OpenSSH_8.2p1', got '%s'", cfg.ServerVersion)
	}

	// Check default private key path
	if cfg.PrivateKeyPath != "" {
		t.Errorf("Expected default private key path to be empty, got '%s'", cfg.PrivateKeyPath)
	}

	// Check default generate key flag
	if !cfg.GenerateKey {
		t.Error("Expected default generate key flag to be true")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "Valid configuration",
			config: &Config{
				Port: 2222,
				Log: LogConfig{
					File:   "credentials.log",
					Format: "json",
				},
				Banner:         "Ubuntu-4ubuntu0.5",
				ServerVersion:  "OpenSSH_8.2p1",
				PrivateKeyPath: "",
				GenerateKey:    true,
			},
			expectError: false,
		},
		{
			name: "Negative port",
			config: &Config{
				Port: -1,
				Log: LogConfig{
					File:   "credentials.log",
					Format: "json",
				},
			},
			expectError: true,
		},
		{
			name: "Too large port",
			config: &Config{
				Port: 70000,
				Log: LogConfig{
					File:   "credentials.log",
					Format: "json",
				},
			},
			expectError: true,
		},
		{
			name: "Invalid log format",
			config: &Config{
				Port: 2222,
				Log: LogConfig{
					File:   "credentials.log",
					Format: "invalid",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError && err == nil {
				t.Errorf("Expected validation error, but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no validation error, but got: %v", err)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary YAML file
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write test configuration
	yamlContent := `
port: 2223
log:
  file: "test.log"
  format: "pretty"
banner: "Test-Banner"
server_version: "TestSSH_1.0"
private_key_path: ""
generate_key: false
`
	if _, err := tmpFile.Write([]byte(yamlContent)); err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	tmpFile.Close()

	// Load the configuration from file
	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Verify loaded values
	if cfg.Port != 2223 {
		t.Errorf("Expected port 2223, got %d", cfg.Port)
	}
	if cfg.Log.File != "test.log" {
		t.Errorf("Expected log file 'test.log', got '%s'", cfg.Log.File)
	}
	if cfg.Log.Format != "pretty" {
		t.Errorf("Expected log format 'pretty', got '%s'", cfg.Log.Format)
	}
	if cfg.Banner != "Test-Banner" {
		t.Errorf("Expected banner 'Test-Banner', got '%s'", cfg.Banner)
	}
	if cfg.ServerVersion != "TestSSH_1.0" {
		t.Errorf("Expected SSH server version 'TestSSH_1.0', got '%s'", cfg.ServerVersion)
	}
	if cfg.GenerateKey {
		t.Error("Expected generate key flag to be false")
	}

	// Test loading with environment variables
	os.Setenv("FAKESSH_PORT", "5555")
	os.Setenv("FAKESSH_LOG_FILE", "env.log")
	os.Setenv("FAKESSH_LOG_FORMAT", "json")
	os.Setenv("FAKESSH_BANNER", "Env-Banner")
	os.Setenv("FAKESSH_SERVER_VERSION", "EnvSSH_1.0")
	os.Setenv("FAKESSH_PRIVATE_KEY_PATH", "/path/to/key")
	os.Setenv("FAKESSH_GENERATE_KEY", "true")
	defer func() {
		os.Unsetenv("FAKESSH_PORT")
		os.Unsetenv("FAKESSH_LOG_FILE")
		os.Unsetenv("FAKESSH_LOG_FORMAT")
		os.Unsetenv("FAKESSH_BANNER")
		os.Unsetenv("FAKESSH_SERVER_VERSION")
		os.Unsetenv("FAKESSH_PRIVATE_KEY_PATH")
		os.Unsetenv("FAKESSH_GENERATE_KEY")
	}()

	// Load config with empty path to test environment variables
	cfg, err = LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Verify environment variable override
	if cfg.Port != 5555 {
		t.Errorf("Expected port 5555 from env var, got %d", cfg.Port)
	}
	if cfg.Log.File != "env.log" {
		t.Errorf("Expected log file 'env.log' from env var, got '%s'", cfg.Log.File)
	}
	if cfg.Log.Format != "json" {
		t.Errorf("Expected log format 'json' from env var, got '%s'", cfg.Log.Format)
	}
	if cfg.Banner != "Env-Banner" {
		t.Errorf("Expected banner 'Env-Banner' from env var, got '%s'", cfg.Banner)
	}
	if cfg.ServerVersion != "EnvSSH_1.0" {
		t.Errorf("Expected SSH server version 'EnvSSH_1.0' from env var, got '%s'", cfg.ServerVersion)
	}
	if cfg.PrivateKeyPath != "/path/to/key" {
		t.Errorf("Expected private key path '/path/to/key' from env var, got '%s'", cfg.PrivateKeyPath)
	}
	if !cfg.GenerateKey {
		t.Error("Expected generate key flag to be true from env var")
	}
}

func TestGetFullServerVersion(t *testing.T) {
	cfg := &Config{
		ServerVersion: "TestSSH_1.0",
		Banner:        "Test-Banner",
	}

	expected := "SSH-2.0-TestSSH_1.0 Test-Banner"
	version := cfg.GetFullServerVersion()

	if version != expected {
		t.Errorf("Expected version '%s', got '%s'", expected, version)
	}
}
