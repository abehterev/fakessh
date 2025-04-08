package sshserver

import (
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/abehterev/fakessh/internal/config"
	"github.com/abehterev/fakessh/internal/logger"
)

// mockLogger is a mock implementation of logger.CredentialsLogger for testing
type mockLogger struct {
	attempts []logger.CredentialAttempt
}

// Log saves the login attempt
func (m *mockLogger) Log(attempt logger.CredentialAttempt) error {
	m.attempts = append(m.attempts, attempt)
	return nil
}

// Close implements the Close method required by the interface
func (m *mockLogger) Close() {
	// No-op for mock
}

// NewMockLogger creates a new mock logger
func NewMockLogger() *mockLogger {
	return &mockLogger{
		attempts: make([]logger.CredentialAttempt, 0),
	}
}

func TestNewServer(t *testing.T) {
	// Create a temporary directory for the tests
	tmpDir, err := ioutil.TempDir("", "ssh-server-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name          string
		config        *config.Config
		shouldSucceed bool
	}{
		{
			name: "Default configuration",
			config: &config.Config{
				Port:   2222,
				Banner: "Test",
				Log: config.LogConfig{
					File:   filepath.Join(tmpDir, "test.log"),
					Format: "json",
				},
				ServerVersion: "8.2p1",
				GenerateKey:   true,
			},
			shouldSucceed: true,
		},
		{
			name: "With private key file that doesn't exist",
			config: &config.Config{
				Port:   2222,
				Banner: "Test",
				Log: config.LogConfig{
					File:   filepath.Join(tmpDir, "test.log"),
					Format: "json",
				},
				ServerVersion:  "8.2p1",
				PrivateKeyPath: "/non/existing/path.key",
				GenerateKey:    false,
			},
			shouldSucceed: false,
		},
		{
			name: "With built-in key",
			config: &config.Config{
				Port:   2222,
				Banner: "Test",
				Log: config.LogConfig{
					File:   filepath.Join(tmpDir, "test.log"),
					Format: "json",
				},
				ServerVersion: "8.2p1",
				GenerateKey:   false,
			},
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary log file for each test
			tmpFile, err := ioutil.TempFile(tmpDir, "test-log-*.log")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			// Create a logger config
			logConfig := logger.Config{
				LogFile:   tmpFile.Name(),
				LogFormat: "json",
			}

			// Create a credentials logger
			credLogger, err := logger.NewCredentialsLogger(logConfig)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			defer credLogger.Close()

			// Create a new server
			server, err := NewServer(tt.config, credLogger)

			if tt.shouldSucceed {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if server == nil {
					t.Errorf("Expected server not to be nil")
				} else {
					if server.sshConfig == nil {
						t.Errorf("Expected sshConfig not to be nil")
					}
					if server.privateKey == nil {
						t.Errorf("Expected privateKey not to be nil")
					}
					if server.config != tt.config {
						t.Errorf("Expected config to match")
					}
					if server.sshConfig.ServerVersion != tt.config.GetFullServerVersion() {
						t.Errorf("Expected server version to be %s, got %s", tt.config.GetFullServerVersion(), server.sshConfig.ServerVersion)
					}
				}
			} else {
				if err == nil {
					t.Errorf("Expected an error, got nil")
				}
				if server != nil {
					t.Errorf("Expected server to be nil")
				}
			}
		})
	}
}

func TestPasswordCallback(t *testing.T) {
	// Create a temporary log file
	tmpFile, err := ioutil.TempFile("", "ssh-test-log-*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Create a minimal config for testing
	cfg := &config.Config{
		Port:   2222,
		Banner: "Test",
		Log: config.LogConfig{
			File:   tmpFile.Name(),
			Format: "json",
		},
		ServerVersion: "8.2p1",
		GenerateKey:   true,
	}

	// Create a real logger for testing
	logConfig := logger.Config{
		LogFile:   tmpFile.Name(),
		LogFormat: "json",
	}
	credLogger, err := logger.NewCredentialsLogger(logConfig)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer credLogger.Close()

	// Create a new server
	server, err := NewServer(cfg, credLogger)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	if server == nil {
		t.Fatalf("Server should not be nil")
	}

	// Create a mock connection metadata
	connMeta := &mockConnMetadata{
		user:       "testuser",
		remoteAddr: "127.0.0.1:12345",
	}

	// Test password callback
	password := []byte("password123")
	perm, err := server.passwordCallback(connMeta, password)

	// Should always reject authentication
	if err == nil {
		t.Errorf("Authentication should be rejected")
	}
	if perm != nil {
		t.Errorf("Permissions should be nil")
	}

	// Check that the file has content (we can't check the exact content easily since we're using zerolog)
	content, err := ioutil.ReadFile(tmpFile.Name())
	if err != nil {
		t.Errorf("Failed to read log file: %v", err)
	}
	if len(content) == 0 {
		t.Errorf("Log file is empty, expected content")
	}
}

// mockConnMetadata is a mock implementation of ssh.ConnMetadata for testing
type mockConnMetadata struct {
	user       string
	remoteAddr string
}

func (m *mockConnMetadata) User() string          { return m.user }
func (m *mockConnMetadata) SessionID() []byte     { return []byte("session-id") }
func (m *mockConnMetadata) ClientVersion() []byte { return []byte("SSH-2.0-OpenSSH_8.2p1") }
func (m *mockConnMetadata) ServerVersion() []byte { return []byte("SSH-2.0-OpenSSH_8.2p1") }
func (m *mockConnMetadata) RemoteAddr() net.Addr  { return mockAddr(m.remoteAddr) }
func (m *mockConnMetadata) LocalAddr() net.Addr   { return mockAddr("127.0.0.1:22") }

type mockAddr string

func (a mockAddr) Network() string { return "tcp" }
func (a mockAddr) String() string  { return string(a) }
