package logger

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestCredentialsLogger(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "credentials_test*.log")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Create logger
	config := Config{
		LogFile:   tempFile.Name(),
		LogFormat: "json",
	}
	logger, err := NewCredentialsLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test data
	timestamp := time.Now()
	attempt := CredentialAttempt{
		Timestamp:  timestamp,
		RemoteAddr: "127.0.0.1:12345",
		Username:   "test_user",
		Password:   "test_password",
	}

	// Log an attempt
	if err := logger.Log(attempt); err != nil {
		t.Fatalf("Logging error: %v", err)
	}

	// Check log file content
	content, err := os.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	// Check that all data is present in the log
	if !strings.Contains(logContent, attempt.RemoteAddr) {
		t.Errorf("Log does not contain IP address: %s", attempt.RemoteAddr)
	}

	if !strings.Contains(logContent, attempt.Username) {
		t.Errorf("Log does not contain username: %s", attempt.Username)
	}

	if !strings.Contains(logContent, attempt.Password) {
		t.Errorf("Log does not contain password: %s", attempt.Password)
	}

	// Check timestamp format - convert to string again
	timestampStr := timestamp.Format(time.RFC3339)
	if !strings.Contains(logContent, timestampStr) {
		t.Errorf("Log does not contain correct timestamp: %s", timestampStr)
	}

	// Check concurrent access - multiple entries
	for i := 0; i < 5; i++ {
		attempt := CredentialAttempt{
			Timestamp:  time.Now(),
			RemoteAddr: "192.168.1.1:54321",
			Username:   "concurrent_user",
			Password:   "concurrent_password",
		}

		if err := logger.Log(attempt); err != nil {
			t.Fatalf("Error during concurrent logging: %v", err)
		}
	}

	// Check that all entries were saved
	content, err = os.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to read log file after concurrent writes: %v", err)
	}

	logContent = string(content)
	if !strings.Contains(logContent, "concurrent_user") {
		t.Errorf("Log does not contain entries with concurrent_user")
	}
}

func TestCredentialsLoggerWithPrettyFormat(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "credentials_pretty_test*.log")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Create logger with pretty format
	config := Config{
		LogFile:   tempFile.Name(),
		LogFormat: "pretty",
	}
	logger, err := NewCredentialsLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test data
	attempt := CredentialAttempt{
		Timestamp:  time.Now(),
		RemoteAddr: "127.0.0.1:12345",
		Username:   "pretty_user",
		Password:   "pretty_password",
	}

	// Log an attempt
	if err := logger.Log(attempt); err != nil {
		t.Fatalf("Logging error: %v", err)
	}

	// Check log file content
	content, err := os.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	// Check that all data is present in the log
	if !strings.Contains(logContent, attempt.RemoteAddr) {
		t.Errorf("Log does not contain IP address: %s", attempt.RemoteAddr)
	}

	if !strings.Contains(logContent, attempt.Username) {
		t.Errorf("Log does not contain username: %s", attempt.Username)
	}

	if !strings.Contains(logContent, attempt.Password) {
		t.Errorf("Log does not contain password: %s", attempt.Password)
	}
}
