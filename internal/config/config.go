/*
Copyright © 2024-2026 Michael Bruskov <mixanemca@yandex.ru>

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
	"fmt"
	"strings"
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

	// DisplayName is the custom display name for the provider (e.g., "Cloudflare", "Рога и Копыта")
	// If not set, a default display name will be used based on the provider type.
	DisplayName string `mapstructure:"display_name" yaml:"display-name,omitempty"`

	// Credentials holds provider-specific credentials
	Credentials map[string]interface{} `mapstructure:"credentials" yaml:"credentials"`

	// Options holds provider-specific options
	Options map[string]interface{} `mapstructure:"options" yaml:"options"`
}

// Prefixes of Cloudflare credentials in the scannable format (introduced in
// 2026). Legacy credentials have no prefix: API tokens are 40 alphanumeric
// characters, the Global API Key is a 37-45 character hex string. Both formats
// are accepted by the API.
// See https://developers.cloudflare.com/fundamentals/api/get-started/token-formats/
const (
	// CloudflareGlobalKeyPrefix marks a Global API Key (goes to api_key + email).
	CloudflareGlobalKeyPrefix = "cfk_"
	// CloudflareUserTokenPrefix marks a User API Token (goes to api_token).
	CloudflareUserTokenPrefix = "cfut_"
	// CloudflareAccountTokenPrefix marks an Account API Token (goes to api_token).
	CloudflareAccountTokenPrefix = "cfat_"
)

// CloudflareCredentials holds Cloudflare-specific credentials.
type CloudflareCredentials struct {
	// APIToken is the Cloudflare API token (User or Account API Token)
	APIToken string `mapstructure:"api_token" yaml:"api_token"`

	// APIKey is the Cloudflare Global API Key (alternative to token)
	APIKey string `mapstructure:"api_key" yaml:"api_key"`

	// Email is the Cloudflare account email (required with APIKey)
	Email string `mapstructure:"email" yaml:"email"`
}

// CheckPrefixes verifies that prefixed credentials are placed in the right
// fields: a Global API Key (cfk_) belongs to api_key, API tokens (cfut_, cfat_)
// belong to api_token. Unprefixed (legacy) values are accepted as is.
func (c *CloudflareCredentials) CheckPrefixes() error {
	if strings.HasPrefix(c.APIToken, CloudflareGlobalKeyPrefix) {
		return fmt.Errorf("api_token looks like a Global API Key (%s prefix): "+
			"put it into api_key and set email instead", CloudflareGlobalKeyPrefix)
	}
	if strings.HasPrefix(c.APIKey, CloudflareUserTokenPrefix) ||
		strings.HasPrefix(c.APIKey, CloudflareAccountTokenPrefix) {
		return fmt.Errorf("api_key looks like an API token (%s/%s prefix): "+
			"put it into api_token instead",
			CloudflareUserTokenPrefix, CloudflareAccountTokenPrefix)
	}
	return nil
}

// GetCloudflareCredentials extracts Cloudflare credentials from provider config.
func (pc *ProviderConfig) GetCloudflareCredentials() (*CloudflareCredentials, error) {
	creds := &CloudflareCredentials{}

	if apiToken, ok := pc.Credentials["api_token"].(string); ok {
		creds.APIToken = apiToken
	}

	if apiKey, ok := pc.Credentials["api_key"].(string); ok {
		creds.APIKey = apiKey
	}

	if email, ok := pc.Credentials["email"].(string); ok {
		creds.Email = email
	}

	return creds, nil
}

// RegRuCredentials holds RegRu-specific credentials.
type RegRuCredentials struct {
	// Username is the RegRu account username
	Username string `mapstructure:"username" yaml:"username"`

	// Password is the RegRu account password
	Password string `mapstructure:"password" yaml:"password"`
}

// GetRegRuCredentials extracts RegRu credentials from provider config.
func (pc *ProviderConfig) GetRegRuCredentials() (*RegRuCredentials, error) {
	creds := &RegRuCredentials{}

	if username, ok := pc.Credentials["username"].(string); ok {
		creds.Username = username
	}

	if password, ok := pc.Credentials["password"].(string); ok {
		creds.Password = password
	}

	return creds, nil
}

// Namecheap credential key names used in ProviderConfig.Credentials map.
const (
	NamecheapCredKeyAPIUser  = "api_user"
	NamecheapCredKeyAPIKey   = "api_key"
	NamecheapCredKeyUserName = "username"
	NamecheapCredKeySandbox  = "sandbox"
)

// NamecheapCredentials holds Namecheap-specific credentials.
type NamecheapCredentials struct {
	// APIUser is the Namecheap account username used for API calls
	APIUser string `mapstructure:"api_user" yaml:"api_user"`

	// APIKey is the Namecheap API key
	APIKey string `mapstructure:"api_key" yaml:"api_key"`

	// UserName is the Namecheap account username (optional, defaults to APIUser)
	UserName string `mapstructure:"username" yaml:"username"`

	// UseSandbox enables the Namecheap sandbox API
	UseSandbox bool `mapstructure:"sandbox" yaml:"sandbox"`
}

// GetNamecheapCredentials extracts Namecheap credentials from provider config.
func (pc *ProviderConfig) GetNamecheapCredentials() (*NamecheapCredentials, error) {
	creds := &NamecheapCredentials{}

	if apiUser, ok := pc.Credentials[NamecheapCredKeyAPIUser].(string); ok {
		creds.APIUser = apiUser
	}

	if apiKey, ok := pc.Credentials[NamecheapCredKeyAPIKey].(string); ok {
		creds.APIKey = apiKey
	}

	if userName, ok := pc.Credentials[NamecheapCredKeyUserName].(string); ok {
		creds.UserName = userName
	}

	if sandbox, ok := pc.Credentials[NamecheapCredKeySandbox].(bool); ok {
		creds.UseSandbox = sandbox
	}

	return creds, nil
}
