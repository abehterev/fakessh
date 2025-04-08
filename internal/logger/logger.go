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

package logger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// CredentialsLogger provides functionality for logging authentication attempts
type CredentialsLogger struct {
	logger zerolog.Logger
	output io.Writer
}

// CredentialAttempt represents information about an authentication attempt
type CredentialAttempt struct {
	Timestamp  time.Time
	RemoteAddr string
	Username   string
	Password   string
}

// Config contains settings for the logger
type Config struct {
	// Path to log file or "stdout" for console output
	LogFile string
	// Log format: "json" or "pretty"
	LogFormat string
}

// NewCredentialsLogger creates a new credentials logger
func NewCredentialsLogger(config Config) (*CredentialsLogger, error) {
	var output io.Writer

	// Determine where to output logs
	if config.LogFile == "stdout" {
		output = os.Stdout
	} else {
		// Check if the file can be opened for writing
		f, err := os.OpenFile(config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		output = f
	}

	// Configure zerolog
	zerolog.TimeFieldFormat = time.RFC3339
	var logger zerolog.Logger

	// Determine output format
	if config.LogFormat == "pretty" {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: output, TimeFormat: time.RFC3339}).
			With().Timestamp().Str("component", "auth").Logger()
	} else {
		// Default is JSON
		logger = zerolog.New(output).With().Timestamp().Str("component", "auth").Logger()
	}

	return &CredentialsLogger{
		logger: logger,
		output: output,
	}, nil
}

// Log records information about an authentication attempt
func (l *CredentialsLogger) Log(attempt CredentialAttempt) error {
	// Use global logger if logging to stdout
	// Otherwise use local logger for file or other outputs
	if _, ok := l.output.(*os.File); ok && l.output == os.Stdout {
		log.Info().
			Str("component", "auth").
			Str("event", "auth_attempt").
			Str("remote_addr", attempt.RemoteAddr).
			Str("username", attempt.Username).
			Str("password", attempt.Password).
			Msg("authentication attempt")
	} else {
		// Use local logger configured for current format
		l.logger.Info().
			Str("event", "auth_attempt").
			Str("remote_addr", attempt.RemoteAddr).
			Str("username", attempt.Username).
			Str("password", attempt.Password).
			Msg("authentication attempt")
	}

	return nil
}

// Close closes the logger and releases resources
func (l *CredentialsLogger) Close() {
	// If output implements io.Closer, close it
	if closer, ok := l.output.(io.Closer); ok && l.output != os.Stdout {
		closer.Close()
	}
}
