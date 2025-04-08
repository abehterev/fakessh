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

package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config contains all settings for the fake SSH server
type Config struct {
	// Server port
	Port int `mapstructure:"port"`
	// Logging settings
	Log LogConfig `mapstructure:"log"`
	// SSH greeting banner
	Banner string `mapstructure:"banner"`
	// SSH server version
	ServerVersion string `mapstructure:"server_version"`
	// Path to SSH private key
	PrivateKeyPath string `mapstructure:"private_key_path"`
	// If true, will generate a new key on each start
	GenerateKey bool `mapstructure:"generate_key"`
}

// LogConfig contains logging settings
type LogConfig struct {
	// Path to log file, "stdout" for console
	File string `mapstructure:"file"`
	// Log format: "json" or "pretty"
	Format string `mapstructure:"format"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Port: 2222,
		Log: LogConfig{
			File:   "credentials.log",
			Format: "json",
		},
		Banner:         "Ubuntu-4ubuntu0.5",
		ServerVersion:  "OpenSSH_8.2p1",
		PrivateKeyPath: "",
		GenerateKey:    true,
	}
}

// LoadConfig loads configuration from file and/or environment variables
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	if configPath != "" {
		// Use viper to read configuration
		viper.SetConfigFile(configPath)

		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("error reading configuration file: %w", err)
		}

		if err := viper.Unmarshal(config); err != nil {
			return nil, fmt.Errorf("error parsing configuration: %w", err)
		}
	}

	// Override through environment variables
	viper.SetEnvPrefix("FAKESSH")
	viper.AutomaticEnv()

	if viper.IsSet("PORT") {
		config.Port = viper.GetInt("PORT")
	}

	if viper.IsSet("LOG_FILE") {
		config.Log.File = viper.GetString("LOG_FILE")
	}

	if viper.IsSet("LOG_FORMAT") {
		config.Log.Format = viper.GetString("LOG_FORMAT")
	}

	if viper.IsSet("BANNER") {
		config.Banner = viper.GetString("BANNER")
	}

	if viper.IsSet("SERVER_VERSION") {
		config.ServerVersion = viper.GetString("SERVER_VERSION")
	}

	if viper.IsSet("PRIVATE_KEY_PATH") {
		config.PrivateKeyPath = viper.GetString("PRIVATE_KEY_PATH")
	}

	if viper.IsSet("GENERATE_KEY") {
		config.GenerateKey = viper.GetBool("GENERATE_KEY")
	}

	return config, nil
}

// Validate checks the configuration validity
func (c *Config) Validate() error {
	// Check port range
	if c.Port < 0 {
		return fmt.Errorf("invalid port: must be positive")
	}
	if c.Port > 65535 {
		return fmt.Errorf("invalid port: must be less than 65536")
	}

	// Check log format
	if c.Log.Format != "json" && c.Log.Format != "pretty" && c.Log.Format != "text" {
		return fmt.Errorf("invalid log format: must be 'json', 'pretty', or 'text'")
	}

	// If a private key path is specified, check that it exists and is readable
	if c.PrivateKeyPath != "" && !c.GenerateKey {
		if _, err := os.Stat(c.PrivateKeyPath); os.IsNotExist(err) {
			return fmt.Errorf("private key not found: %s", c.PrivateKeyPath)
		}
	}

	return nil
}

// GetFullServerVersion returns the full SSH server version string
func (c *Config) GetFullServerVersion() string {
	return fmt.Sprintf("SSH-2.0-%s %s", c.ServerVersion, c.Banner)
}
