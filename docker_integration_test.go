//go:build docker
// +build docker

package main

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

// TestDockerEnvironmentVariables tests the Docker container with environment variables
// This test is only run when the docker tag is specified
// Run with: go test -v -tags=docker
func TestDockerEnvironmentVariables(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	// Build Docker image
	buildCmd := exec.Command("docker", "build", "-f", "Dockerfile.alpine", "-t", "fakessh", ".")
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build Docker image: %v\nOutput: %s", err, buildOutput)
	}

	// Clean up existing containers with the same name if they exist
	containerName := "fakessh-env-test"
	exec.Command("docker", "stop", containerName).Run()
	exec.Command("docker", "rm", containerName).Run()

	// Run Docker container with environment variables
	runCmd := exec.Command("docker", "run", "--name", containerName, "-d", "-p", "2233:2222",
		"-e", "FAKESSH_LOG_FILE=stdout",
		"-e", "FAKESSH_LOG_FORMAT=pretty",
		"-e", "FAKESSH_BANNER=TestEnvBanner",
		"-e", "FAKESSH_SERVER_VERSION=TestEnvVersion_1.0",
		"fakessh")

	runOutput, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run Docker container: %v\nOutput: %s", err, runOutput)
	}

	// Make sure to clean up the container after the test
	defer func() {
		stopCmd := exec.Command("docker", "stop", containerName)
		stopCmd.Run()
		rmCmd := exec.Command("docker", "rm", containerName)
		rmCmd.Run()
	}()

	// Wait for container to start
	time.Sleep(2 * time.Second)

	// Verify the environment variables were applied by checking logs
	logsCmd := exec.Command("docker", "logs", containerName)
	logsOutput, err := logsCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to get Docker logs: %v", err)
	}

	logs := string(logsOutput)

	// Check if custom banner and version are present in logs
	if !strings.Contains(logs, "TestEnvBanner") {
		t.Errorf("Custom banner 'TestEnvBanner' not found in logs")
	}

	if !strings.Contains(logs, "TestEnvVersion_1.0") {
		t.Errorf("Custom server version 'TestEnvVersion_1.0' not found in logs")
	}

	// Try connecting to the SSH server to verify it's running
	clientConfig := &ssh.ClientConfig{
		User: "testuser",
		Auth: []ssh.AuthMethod{
			ssh.Password("testpassword"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	// Connect to the server
	client, err := ssh.Dial("tcp", "127.0.0.1:2233", clientConfig)
	if err == nil {
		client.Close()
		t.Fatalf("Expected authentication to fail, but it succeeded")
	} else {
		// Verify that the error is related to authentication
		if !strings.Contains(err.Error(), "unable to authenticate") {
			t.Errorf("Expected authentication failure error, got: %v", err)
		}
	}

	// Check logs again after connection attempt to verify credentials were logged
	logsCmd = exec.Command("docker", "logs", containerName)
	logsOutput, err = logsCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to get Docker logs after connection: %v", err)
	}

	logsAfterConnection := string(logsOutput)
	if !strings.Contains(logsAfterConnection, "testuser") || !strings.Contains(logsAfterConnection, "testpassword") {
		t.Errorf("Credentials not found in logs after connection attempt")
	}
}
