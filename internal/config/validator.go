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
	"strings"
	"time"
)

// ValidationError represents a configuration validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error in field %q: %s", e.Field, e.Message)
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	var errors []error

	// Validate client timeout
	if c.ClientTimeout <= 0 {
		errors = append(errors, &ValidationError{
			Field:   "client_timeout",
			Message: "must be greater than 0",
		})
	}

	// Validate output format
	validFormats := map[string]bool{
		"text": true,
		"json": true,
		"none": true,
	}
	if c.OutputFormat != "" && !validFormats[strings.ToLower(c.OutputFormat)] {
		errors = append(errors, &ValidationError{
			Field:   "output_format",
			Message: fmt.Sprintf("must be one of: text, json, none (got: %s)", c.OutputFormat),
		})
	}

	// Validate providers
	// It's OK if no providers are configured (credentials may come from env vars)
	// We'll validate provider-specific credentials when they're used
	if len(c.Providers) > 0 {
		for name, provider := range c.Providers {
			if err := provider.Validate(name); err != nil {
				errors = append(errors, err)
			}
		}
	}

	// Validate default provider
	if c.DefaultProvider != "" {
		if _, exists := c.Providers[c.DefaultProvider]; !exists {
			errors = append(errors, &ValidationError{
				Field:   "default_provider",
				Message: fmt.Sprintf("provider %q not found in providers list", c.DefaultProvider),
			})
		}
	}

	if len(errors) > 0 {
		var errMsgs []string
		for _, err := range errors {
			errMsgs = append(errMsgs, err.Error())
		}
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errMsgs, "; "))
	}

	return nil
}

// Validate validates a provider configuration.
func (pc *ProviderConfig) Validate(name string) error {
	var errors []error

	// Validate provider type
	if pc.Type == "" {
		errors = append(errors, &ValidationError{
			Field:   fmt.Sprintf("providers.%s.type", name),
			Message: "provider type is required",
		})
	}

	// Validate provider-specific credentials
	switch strings.ToLower(pc.Type) {
	case "cloudflare":
		if err := pc.validateCloudflare(name); err != nil {
			errors = append(errors, err)
		}
	default:
		// Unknown provider type - just warn but don't fail
		// This allows for future provider types
	}

	if len(errors) > 0 {
		var errMsgs []string
		for _, err := range errors {
			errMsgs = append(errMsgs, err.Error())
		}
		return fmt.Errorf("provider %q validation failed: %s", name, strings.Join(errMsgs, "; "))
	}

	return nil
}

// validateCloudflare validates Cloudflare-specific configuration.
func (pc *ProviderConfig) validateCloudflare(name string) error {
	var errors []error

	creds, err := pc.GetCloudflareCredentials()
	if err != nil {
		return fmt.Errorf("failed to get Cloudflare credentials: %w", err)
	}

	// Must have either API token OR (API key + email)
	hasToken := creds.APIToken != ""
	hasKeyAndEmail := creds.APIKey != "" && creds.Email != ""

	if !hasToken && !hasKeyAndEmail {
		errors = append(errors, &ValidationError{
			Field:   fmt.Sprintf("providers.%s.credentials", name),
			Message: "must have either api_token or (api_key + email)",
		})
	}

	if hasToken && hasKeyAndEmail {
		errors = append(errors, &ValidationError{
			Field:   fmt.Sprintf("providers.%s.credentials", name),
			Message: "cannot have both api_token and (api_key + email)",
		})
	}

	if len(errors) > 0 {
		var errMsgs []string
		for _, err := range errors {
			errMsgs = append(errMsgs, err.Error())
		}
		return fmt.Errorf("cloudflare provider validation failed: %s", strings.Join(errMsgs, "; "))
	}

	return nil
}

// GetProvider returns the provider configuration by name.
// Returns an error if the provider is not found.
func (c *Config) GetProvider(name string) (*ProviderConfig, error) {
	if name == "" {
		// Use default provider if name is empty
		if c.DefaultProvider == "" {
			return nil, fmt.Errorf("no provider specified and no default provider configured")
		}
		name = c.DefaultProvider
	}

	provider, exists := c.Providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %q not found", name)
	}

	return &provider, nil
}

// GetClientTimeout returns the client timeout, ensuring it's at least 1 second.
func (c *Config) GetClientTimeout() time.Duration {
	if c.ClientTimeout <= 0 {
		return DefaultClientTimeout
	}
	return c.ClientTimeout
}

