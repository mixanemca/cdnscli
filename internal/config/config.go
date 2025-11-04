/*
Copyright © 2024 Michael Bruskov <mixanemca@yandex.ru>

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

// Package config holds configuration structures and loading logic.
package config

import (
	"time"
)

// Config represents the main configuration structure.
type Config struct {
	// DefaultProvider is the name of the default provider to use
	DefaultProvider string `mapstructure:"default_provider" yaml:"default-provider,omitempty"`

	// Providers contains configuration for all DNS providers
	Providers map[string]ProviderConfig `mapstructure:"providers" yaml:"providers"`

	// ClientTimeout is the default timeout for API requests
	ClientTimeout time.Duration `mapstructure:"client_timeout" yaml:"client-timeout,omitempty"`

	// OutputFormat is the default output format
	OutputFormat string `mapstructure:"output_format" yaml:"output-format,omitempty"`

	// Debug enables debug output
	Debug bool `mapstructure:"debug" yaml:"debug"`
}

// ProviderConfig holds configuration for a specific DNS provider.
type ProviderConfig struct {
	// Type is the provider type (e.g., "cloudflare", "route53", "digitalocean")
	Type string `mapstructure:"type" yaml:"type"`

	// Credentials holds provider-specific credentials
	Credentials map[string]interface{} `mapstructure:"credentials" yaml:"credentials"`

	// Options holds provider-specific options
	Options map[string]interface{} `mapstructure:"options" yaml:"options"`
}

// CloudflareCredentials holds Cloudflare-specific credentials.
type CloudflareCredentials struct {
	// APIToken is the Cloudflare API token
	APIToken string `mapstructure:"api_token" yaml:"api_token"`

	// APIKey is the Cloudflare API key (alternative to token)
	APIKey string `mapstructure:"api_key" yaml:"api_key"`

	// Email is the Cloudflare account email (required with APIKey)
	Email string `mapstructure:"email" yaml:"email"`
}

// GetCloudflareCredentials extracts Cloudflare credentials from provider config.
func (pc *ProviderConfig) GetCloudflareCredentials() (*CloudflareCredentials, error) {
	creds := &CloudflareCredentials{}

	// Support both api_token and api-token (with dash)
	if apiToken, ok := pc.Credentials["api_token"].(string); ok {
		creds.APIToken = apiToken
	} else if apiToken, ok := pc.Credentials["api-token"].(string); ok {
		creds.APIToken = apiToken
	}

	// Support both api_key and api-key (with dash)
	if apiKey, ok := pc.Credentials["api_key"].(string); ok {
		creds.APIKey = apiKey
	} else if apiKey, ok := pc.Credentials["api-key"].(string); ok {
		creds.APIKey = apiKey
	}

	if email, ok := pc.Credentials["email"].(string); ok {
		creds.Email = email
	}

	return creds, nil
}