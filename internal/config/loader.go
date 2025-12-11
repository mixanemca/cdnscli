/*
Copyright © 2024-2025 Michael Bruskov <mixanemca@yandex.ru>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

const (
	// DefaultConfigName is the default name of the config file
	DefaultConfigName = ".cdnscli"

	// DefaultConfigType is the default config file type
	DefaultConfigType = "yaml"

	// DefaultClientTimeout is the default client timeout
	DefaultClientTimeout = 10 * time.Second

	// DefaultOutputFormat is the default output format
	DefaultOutputFormat = "text"
)

// Load loads configuration from file, environment variables, and command line flags.
// Priority order: flags > env > config file > defaults
func Load(cfgFile string) (*Config, error) {
	cfg := &Config{
		DefaultProvider: "",
		Providers:       make(map[string]ProviderConfig),
		ClientTimeout:   DefaultClientTimeout,
		OutputFormat:    DefaultOutputFormat,
		Debug:           false,
	}

	// Setup Viper
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := homedir.Dir()
		if err != nil {
			return nil, fmt.Errorf("failed to find home directory: %w", err)
		}

		// Search config in home directory with name ".cdnscli" (without extension)
		viper.AddConfigPath(home)
		viper.AddConfigPath(".") // Also check current directory
		viper.SetConfigName(DefaultConfigName)
		viper.SetConfigType(DefaultConfigType)
	}

	// Environment variables are not used - only config file

	// Set defaults
	viper.SetDefault("client_timeout", DefaultClientTimeout)
	viper.SetDefault("output_format", DefaultOutputFormat)
	viper.SetDefault("debug", false)

	// Read config file (optional - don't fail if it doesn't exist)
	if err := viper.ReadInConfig(); err != nil {
		// Config file not found; this is OK if we have env vars or flags
		// Check if it's a "file not found" error by checking the error message
		if err.Error() != "config file not found" && !contains(err.Error(), "not found") {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal config
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

// GetConfigPath returns the path to the config file that would be used.
func GetConfigPath() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", fmt.Errorf("failed to find home directory: %w", err)
	}

	return filepath.Join(home, DefaultConfigName+"."+DefaultConfigType), nil
}

// Save saves the configuration to the default config file.
func Save(cfg *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Set config file path
	viper.SetConfigFile(configPath)
	viper.SetConfigType(DefaultConfigType)

	// Set values (use dashes for consistency with flags)
	viper.Set("default-provider", cfg.DefaultProvider)
	viper.Set("client-timeout", cfg.ClientTimeout)
	viper.Set("output-format", cfg.OutputFormat)
	viper.Set("debug", cfg.Debug)

	// Set providers
	for name, provider := range cfg.Providers {
		viper.Set(fmt.Sprintf("providers.%s.type", name), provider.Type)
		if provider.Credentials != nil {
			for key, value := range provider.Credentials {
				// Normalize keys to use dashes (api_token -> api-token)
				normalizedKey := strings.ReplaceAll(key, "_", "-")
				viper.Set(fmt.Sprintf("providers.%s.credentials.%s", name, normalizedKey), value)
			}
		}
		if provider.Options != nil {
			for key, value := range provider.Options {
				viper.Set(fmt.Sprintf("providers.%s.options.%s", name, key), value)
			}
		}
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config file
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
