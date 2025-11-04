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

package providers

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ProviderError
		expected string
	}{
		{
			name: "error with cause",
			err: &ProviderError{
				ProviderName: "test-provider",
				ProviderType: "cloudflare",
				Message:      "test message",
				Cause:        errors.New("underlying error"),
			},
			expected: `provider "test-provider" (type: cloudflare): test message: underlying error`,
		},
		{
			name: "error without cause",
			err: &ProviderError{
				ProviderName: "test-provider",
				ProviderType: "cloudflare",
				Message:      "test message",
			},
			expected: `provider "test-provider" (type: cloudflare): test message`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
			if tt.err.Cause != nil {
				assert.Equal(t, tt.err.Cause, tt.err.Unwrap())
			} else {
				assert.Nil(t, tt.err.Unwrap())
			}
		})
	}
}

func TestProviderNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ProviderNotFoundError
		expected string
	}{
		{
			name: "error with available providers",
			err: &ProviderNotFoundError{
				ProviderName: "missing-provider",
				Available:    []string{"cloudflare", "route53"},
			},
			expected: `provider "missing-provider" not found (available providers: [cloudflare route53])`,
		},
		{
			name: "error without available providers",
			err: &ProviderNotFoundError{
				ProviderName: "missing-provider",
				Available:    []string{},
			},
			expected: `provider "missing-provider" not found`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestProviderTypeNotSupportedError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ProviderTypeNotSupportedError
		expected string
	}{
		{
			name: "error with supported types",
			err: &ProviderTypeNotSupportedError{
				ProviderType: "unsupported",
				Supported:    []string{"cloudflare", "route53"},
			},
			expected: `unsupported provider type: "unsupported" (supported types: [cloudflare route53])`,
		},
		{
			name: "error without supported types",
			err: &ProviderTypeNotSupportedError{
				ProviderType: "unsupported",
				Supported:    []string{},
			},
			expected: `unsupported provider type: "unsupported"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestProviderCreationError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ProviderCreationError
		expected string
	}{
		{
			name: "error with cause",
			err: &ProviderCreationError{
				ProviderName: "test-provider",
				ProviderType: "cloudflare",
				Message:      "creation failed",
				Cause:        errors.New("network error"),
			},
			expected: `failed to create provider "test-provider" (type: cloudflare): creation failed: network error`,
		},
		{
			name: "error without cause",
			err: &ProviderCreationError{
				ProviderName: "test-provider",
				ProviderType: "cloudflare",
				Message:      "creation failed",
			},
			expected: `failed to create provider "test-provider" (type: cloudflare): creation failed`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
			if tt.err.Cause != nil {
				assert.Equal(t, tt.err.Cause, tt.err.Unwrap())
			} else {
				assert.Nil(t, tt.err.Unwrap())
			}
		})
	}
}

func TestProviderConfigError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ProviderConfigError
		expected string
	}{
		{
			name: "error with field and cause",
			err: &ProviderConfigError{
				ProviderName: "test-provider",
				ProviderType: "cloudflare",
				Field:        "credentials",
				Message:      "invalid format",
				Cause:        errors.New("parse error"),
			},
			expected: `invalid configuration for provider "test-provider" (type: cloudflare), field "credentials": invalid format: parse error`,
		},
		{
			name: "error without field",
			err: &ProviderConfigError{
				ProviderName: "test-provider",
				ProviderType: "cloudflare",
				Message:      "invalid format",
			},
			expected: `invalid configuration for provider "test-provider" (type: cloudflare): invalid format`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
			if tt.err.Cause != nil {
				assert.Equal(t, tt.err.Cause, tt.err.Unwrap())
			} else {
				assert.Nil(t, tt.err.Unwrap())
			}
		})
	}
}

func TestProviderCredentialsError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ProviderCredentialsError
		expected string
	}{
		{
			name: "error with cause",
			err: &ProviderCredentialsError{
				ProviderType: "cloudflare",
				Message:      "invalid token",
				Cause:        errors.New("unauthorized"),
			},
			expected: `credentials error for provider type "cloudflare": invalid token: unauthorized`,
		},
		{
			name: "error without cause",
			err: &ProviderCredentialsError{
				ProviderType: "cloudflare",
				Message:      "invalid token",
			},
			expected: `credentials error for provider type "cloudflare": invalid token`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
			if tt.err.Cause != nil {
				assert.Equal(t, tt.err.Cause, tt.err.Unwrap())
			} else {
				assert.Nil(t, tt.err.Unwrap())
			}
		})
	}
}

func TestNewProviderNotFoundError(t *testing.T) {
	err := NewProviderNotFoundError("missing", []string{"cloudflare", "route53"})
	assert.NotNil(t, err)
	assert.Equal(t, "missing", err.ProviderName)
	assert.Equal(t, []string{"cloudflare", "route53"}, err.Available)
}

func TestNewProviderTypeNotSupportedError(t *testing.T) {
	err := NewProviderTypeNotSupportedError("unsupported", []string{"cloudflare"})
	assert.NotNil(t, err)
	assert.Equal(t, "unsupported", err.ProviderType)
	assert.Equal(t, []string{"cloudflare"}, err.Supported)
}

func TestNewProviderCreationError(t *testing.T) {
	cause := errors.New("test cause")
	err := NewProviderCreationError("test", "cloudflare", "test message", cause)
	assert.NotNil(t, err)
	assert.Equal(t, "test", err.ProviderName)
	assert.Equal(t, "cloudflare", err.ProviderType)
	assert.Equal(t, "test message", err.Message)
	assert.Equal(t, cause, err.Cause)
}

func TestNewProviderConfigError(t *testing.T) {
	cause := errors.New("test cause")
	err := NewProviderConfigError("test", "cloudflare", "field", "test message", cause)
	assert.NotNil(t, err)
	assert.Equal(t, "test", err.ProviderName)
	assert.Equal(t, "cloudflare", err.ProviderType)
	assert.Equal(t, "field", err.Field)
	assert.Equal(t, "test message", err.Message)
	assert.Equal(t, cause, err.Cause)
}

func TestNewProviderCredentialsError(t *testing.T) {
	cause := errors.New("test cause")
	err := NewProviderCredentialsError("cloudflare", "test message", cause)
	assert.NotNil(t, err)
	assert.Equal(t, "cloudflare", err.ProviderType)
	assert.Equal(t, "test message", err.Message)
	assert.Equal(t, cause, err.Cause)
}
