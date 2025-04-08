//go:build !docker
// +build !docker

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/abehterev/fakessh/internal/config"
	"github.com/abehterev/fakessh/internal/logger"
	"github.com/abehterev/fakessh/internal/sshserver"
	"golang.org/x/crypto/ssh"
)

// TestFakeSSHServerIntegration tests the fake SSH server functionality
// by launching the server and attempting to connect to it
func TestFakeSSHServerIntegration(t *testing.T) {
	// Create a temporary directory for logs
	tempDir, err := ioutil.TempDir("", "fakessh-integration")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create log file path
	logFile := filepath.Join(tempDir, "credentials.log")

	// Setup the server configuration
	cfg := &config.Config{
		Port:          2229,
		Banner:        "Ubuntu-4ubuntu0.5",
		ServerVersion: "OpenSSH_8.2p1",
		Log: config.LogConfig{
			File:   logFile,
			Format: "pretty",
		},
		GenerateKey: true, // use generated key for test
	}

	// Create logger
	logConfig := logger.Config{
		LogFile:   logFile,
		LogFormat: "pretty",
	}
	credLogger, err := logger.NewCredentialsLogger(logConfig)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer credLogger.Close()

	// Create and start the server
	server, err := sshserver.NewServer(cfg, credLogger)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start the server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()

	// Wait for the server to start
	time.Sleep(1 * time.Second)

	// Create an SSH client config
	clientConfig := &ssh.ClientConfig{
		User: "testuser",
		Auth: []ssh.AuthMethod{
			ssh.Password("testpassword"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	// Connect to the server
	client, err := ssh.Dial("tcp", "127.0.0.1:2229", clientConfig)
	if err == nil {
		client.Close()
		t.Fatalf("Expected authentication to fail, but it succeeded")
	} else {
		// Verify that the error is related to authentication
		if !strings.Contains(err.Error(), "unable to authenticate") {
			t.Errorf("Expected authentication failure error, got: %v", err)
		}
	}

	// Wait a moment for the log to be written
	time.Sleep(500 * time.Millisecond)

	// Check if the log file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("Log file was not created")
	}

	// Verify log file has content
	content, err := ioutil.ReadFile(logFile)
	if err != nil {
		t.Errorf("Failed to read log file: %v", err)
	}
	if len(content) == 0 {
		t.Errorf("Log file is empty")
	}
}
