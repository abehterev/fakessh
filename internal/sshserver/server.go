/*
 * FakeSSH - SSH server honeypot for monitoring brute force attacks
 * Copyright (C) 2023 Andrey Bekhterev
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License along
 * with this program; if not, write to the Free Software Foundation, Inc.,
 * 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
 */

package sshserver

import (
	cryptoRand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"time"

	"github.com/abehterev/fakessh/internal/config"
	"github.com/abehterev/fakessh/internal/logger"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
)

// Server represents a fake SSH server
type Server struct {
	config     *config.Config
	sshConfig  *ssh.ServerConfig
	logger     *logger.CredentialsLogger
	privateKey ssh.Signer
}

func init() {
	// Initialize the random number generator
	rand.Seed(time.Now().UnixNano())
}

// NewServer creates a new SSH server instance
func NewServer(config *config.Config, logger *logger.CredentialsLogger) (*Server, error) {
	// Get private key
	var privateKey ssh.Signer
	var err error

	if config.GenerateKey {
		// Generate a new private key
		privateKey, err = generatePrivateKey()
		if err != nil {
			return nil, fmt.Errorf("key generation error: %w", err)
		}
	} else if config.PrivateKeyPath != "" {
		// Load key from file
		privateKey, err = loadPrivateKey(config.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("key loading error: %w", err)
		}
	} else {
		// Use built-in key
		privateKey, err = ssh.ParsePrivateKey([]byte(defaultHostKey))
		if err != nil {
			return nil, fmt.Errorf("built-in key parsing error: %w", err)
		}
	}

	server := &Server{
		config:     config,
		logger:     logger,
		privateKey: privateKey,
	}

	// Configure SSH server
	sshConfig := &ssh.ServerConfig{
		PasswordCallback: server.passwordCallback,
		BannerCallback:   server.bannerCallback,
		ServerVersion:    config.GetFullServerVersion(),
	}

	// Add private key to configuration
	sshConfig.AddHostKey(privateKey)

	server.sshConfig = sshConfig

	return server, nil
}

// Start launches the SSH server
func (s *Server) Start() error {
	// Listen for connections on the specified port
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.config.Port))
	if err != nil {
		return fmt.Errorf("server start error: %w", err)
	}
	defer listener.Close()

	fmt.Printf("Fake SSH server started on port %d\n", s.config.Port)
	fmt.Printf("Server version: %s\n", s.config.GetFullServerVersion())

	// Print SSH key fingerprint for debugging
	if pubKey, ok := s.privateKey.PublicKey().(ssh.PublicKey); ok {
		fmt.Printf("Server fingerprint: %s\n", ssh.FingerprintSHA256(pubKey))
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Connection acceptance error: %v\n", err)
			continue
		}

		// Handle connection in a separate goroutine
		go s.handleConnection(conn)
	}
}

// handleConnection processes an incoming connection
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Perform SSH handshake
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, s.sshConfig)
	if err != nil {
		// Error is expected here as we always reject authentication
		return
	}
	defer sshConn.Close()

	// Process global requests (we reject them)
	go ssh.DiscardRequests(reqs)

	// Process incoming channels (shouldn't reach here due to authentication rejection)
	for newChannel := range chans {
		newChannel.Reject(ssh.Prohibited, "connection rejected")
	}
}

// passwordCallback handles password authentication attempts
func (s *Server) passwordCallback(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	// Log login attempt
	attempt := logger.CredentialAttempt{
		Timestamp:  time.Now(),
		RemoteAddr: conn.RemoteAddr().String(),
		Username:   conn.User(),
		Password:   string(password),
	}

	if err := s.logger.Log(attempt); err != nil {
		log.Error().Err(err).Msg("logging error")
	}

	// Always reject authentication with a delay to simulate a real server
	time.Sleep(time.Duration(200+rand.Intn(300)) * time.Millisecond)
	return nil, fmt.Errorf("permission denied (password), please try again")
}

// bannerCallback returns a greeting banner
func (s *Server) bannerCallback(conn ssh.ConnMetadata) string {
	return fmt.Sprintf("Welcome to Ubuntu %s (GNU/Linux 5.4.0-109-generic x86_64)\n\n", s.config.Banner)
}

// generatePrivateKey generates a new RSA private key for SSH server
func generatePrivateKey() (ssh.Signer, error) {
	// Generate a new RSA key
	key, err := rsa.GenerateKey(cryptoRand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Convert to PEM format
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	// Convert to SSH key format
	parsedKey, err := ssh.ParsePrivateKey(pem.EncodeToMemory(privateKeyPEM))
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH key: %w", err)
	}

	return parsedKey, nil
}

// loadPrivateKey loads a private key from a file
func loadPrivateKey(path string) (ssh.Signer, error) {
	// Read the key file
	keyData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	// Parse the key
	privateKey, err := ssh.ParsePrivateKey(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SSH key: %w", err)
	}

	return privateKey, nil
}

// Built-in SSH key
const defaultHostKey = `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAABFwAAAAdzc2gtcn
NhAAAAAwEAAQAAAQEAqKTkizhpgCf/7SpCvWhAyzVdysTAel9F+fKmzDUPkTz1lAeLpU2e
RiJMOl58c9ZlEMl2Gy+p4N0G1z9+9ekoDZHAlM/NhzkO9T3p5GgAntfg/VnBv5+Kw/fxee
G/xEy/yzrDr7los3apBDS8o0ejQMwST5QpYc5K8+ImO8be05PZP1wYIdJ+R9bbMaYMVF7R
3Og2LvikvrjSiruR0DZD7AVRHQmdqXNHXJ9PY5m3nY6whXMFEVgom+xgLpcq8XUS08l8/8
B03dHAAbwoNXMSgeXL63UI7LhniYBDrB8j94l58hDU7auOo/BQsBfhcF+2/o6PeU6j0HTa
Gljr90Et+QAAA9Dt0r417dK+NQAAAAdzc2gtcnNhAAABAQCopOSLOGmAJ//tKkK9aEDLNV
3KxMB6X0X58qbMNQ+RPPWUB4ulTZ5GIkw6Xnxz1mUQyXYbL6ng3QbXP3716SgNkcCUz82H
OQ71PenkaACe1+D9WcG/n4rD9/F54b/ETL/LOsOvuWizdqkENLyjR6NAzBJPlClhzkrz4i
Y7xt7Tk9k/XBgh0n5H1tsxpgxUXtHc6DYu+KS+uNKKu5HQNkPsBVEdCZ2pc0dcn09jmbed
jrCFcwURWCib7GAulyrxdRLTyXz/wHTd0cABvCg1cxKB5cvrdQjsuGeJgEOsHyP3iXnyEN
Ttq46j8FCwF+FwX7b+jo95TqPQdNoaWOv3QS35AAAAAwEAAQAAAQAJvbiLyB7j7auNPe8r
9J0lf7gisbmyd9VZaigzTG9RQtWmjscErdaSE4IWrwV+RWiCDzj4ugiUef/eqAbD2otbOU
uH7PbgtC2GgeSEMnOyuSKAT9JuqJ8B0c0Lbrw+cPZ1HThXapy/HQAHQ6qPveASqpb2LMc1
JI7Uxn/R3Rta2ivNlJrPbiMi65Wlzlc3cne+0mEX0ti5v0uViNkQSjlz/DiXKXdpkRghJl
27NbrrxwVvJARTVruBZpgvVMAF68Z1QH9w0QB/MaV9oyY/5ZFZEGX8xcu6tvknkB6eApVW
pEZfkxQ5/avwJjTV/TrB6qmJOzT+TINNU9xuMGVBe5sXAAAAgQC3VQJwevtmy/F+wlBoVU
TkRySMivS1PsHy5lXxX7aRuTsia7mSczEG+zcVzd1ucY8Bnf3GRJQSiORYAa4QfHTk7wd2
8RD9A8oAhxXhBy+2pZ7//bOUs5LmMg5ZKtxZk4we3ZiOv66MC+TGPDMtMbfaYnq2AVTEf4
vV9JhRyjDhJAAAAIEA4xR8LkRil4rDV/GvRJdiE4cWLrlH3rdj3PgJfBRjSvVlTHL4bbzJ
//ZQhoIX98a0B9w20lJ/s4lTDjuHRNJ5G6/X2T8+OjaB4/5GuO73t47vWEIiPvF+u3v4kz
qElm2E6x86BGc+wyqip5u1EY+FHJ41KPiv7MXWNqhUZCGjlWMAAACBAL4fNyAQs5dGCee3
ZKNNSimsNG5w7UvA7mpGJabidAIuInbtWgQRApD+mlTczzTmcilgij3P6tQULm5BlLF+/l
/TCCVwNMaPOiwzBkOv2zMBuPEqIfZpRg57EP4an11XDbhj5q+4tk9BkcLmU/OKmg6IMtaP
IByGmg5krFlzbmvzAAAAE3ZzY29kZUA4MDJjZTk4YWJhYjkBAgMEBQYH
-----END OPENSSH PRIVATE KEY-----`
