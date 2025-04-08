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

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/abehterev/fakessh/internal/config"
	"github.com/abehterev/fakessh/internal/logger"
	"github.com/abehterev/fakessh/internal/sshserver"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	cfgFile        string
	port           int
	logFile        string
	logFormat      string
	banner         string
	serverVersion  string
	privateKeyPath string
	generateKey    bool
)

// rootCmd represents the base command when the application is called
var rootCmd = &cobra.Command{
	Use:   "fakessh",
	Short: "Fake SSH server for credential harvesting",
	Long: `Fake SSH server that emulates OpenSSH server behavior,
but always rejects authentication attempts and logs credentials.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Configure global zerolog logger
		zerolog.TimeFieldFormat = time.RFC3339
		if logFormat == "pretty" {
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
		} else {
			// Configure JSON output
			log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
		}

		// Load configuration
		cfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("configuration loading error: %w", err)
		}

		// Command line flags take precedence
		if cmd.Flags().Changed("port") {
			cfg.Port = port
		}
		if cmd.Flags().Changed("log") {
			cfg.Log.File = logFile
		}
		if cmd.Flags().Changed("log-format") {
			cfg.Log.Format = logFormat
		}
		if cmd.Flags().Changed("banner") {
			cfg.Banner = banner
		}
		if cmd.Flags().Changed("server-version") {
			cfg.ServerVersion = serverVersion
		}
		if cmd.Flags().Changed("key") {
			cfg.PrivateKeyPath = privateKeyPath
		}
		if cmd.Flags().Changed("generate-key") {
			cfg.GenerateKey = generateKey
		}

		// Validate configuration
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("invalid configuration: %w", err)
		}

		// Create credentials logger
		loggerConfig := logger.Config{
			LogFile:   cfg.Log.File,
			LogFormat: cfg.Log.Format,
		}

		credLogger, err := logger.NewCredentialsLogger(loggerConfig)
		if err != nil {
			return fmt.Errorf("logger creation error: %w", err)
		}
		defer credLogger.Close()

		// Create SSH server
		server, err := sshserver.NewServer(cfg, credLogger)
		if err != nil {
			return fmt.Errorf("SSH server creation error: %w", err)
		}

		// Launch server
		log.Info().
			Int("port", cfg.Port).
			Str("log_file", cfg.Log.File).
			Str("version", cfg.GetFullServerVersion()).
			Msg("Starting fake SSH server")

		// Start SSH server
		if err := server.Start(); err != nil {
			return fmt.Errorf("server runtime error: %w", err)
		}

		return nil
	},
}

func init() {
	// Command line flags
	rootCmd.Flags().StringVar(&cfgFile, "config", "", "path to configuration file")
	rootCmd.Flags().IntVar(&port, "port", 2222, "SSH server port")
	rootCmd.Flags().StringVar(&logFile, "log", "credentials.log", "path to credentials log file (stdout for console output)")
	rootCmd.Flags().StringVar(&logFormat, "log-format", "json", "log format (json, pretty or text)")
	rootCmd.Flags().StringVar(&banner, "banner", "Ubuntu-4ubuntu0.5", "SSH banner (version part)")
	rootCmd.Flags().StringVar(&serverVersion, "server-version", "OpenSSH_8.2p1", "SSH server version")
	rootCmd.Flags().StringVar(&privateKeyPath, "key", "", "path to SSH private key (if not specified, built-in or newly generated will be used)")
	rootCmd.Flags().BoolVar(&generateKey, "generate-key", true, "generate a new SSH key on each start")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
