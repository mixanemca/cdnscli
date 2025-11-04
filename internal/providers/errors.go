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
	"fmt"
)

// ProviderError represents a provider-related error.
type ProviderError struct {
	ProviderName string
	ProviderType string
	Message      string
	Cause        error
}

// Error implements the error interface.
func (e *ProviderError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("provider %q (type: %s): %s: %v", e.ProviderName, e.ProviderType, e.Message, e.Cause)
	}
	return fmt.Sprintf("provider %q (type: %s): %s", e.ProviderName, e.ProviderType, e.Message)
}

// Unwrap returns the underlying error.
func (e *ProviderError) Unwrap() error {
	return e.Cause
}

// ProviderNotFoundError indicates that a provider was not found.
type ProviderNotFoundError struct {
	ProviderName string
	Available    []string
}

// Error implements the error interface.
func (e *ProviderNotFoundError) Error() string {
	if len(e.Available) > 0 {
		return fmt.Sprintf("provider %q not found (available providers: %v)", e.ProviderName, e.Available)
	}
	return fmt.Sprintf("provider %q not found", e.ProviderName)
}

// ProviderTypeNotSupportedError indicates that a provider type is not supported.
type ProviderTypeNotSupportedError struct {
	ProviderType string
	Supported    []string
}

// Error implements the error interface.
func (e *ProviderTypeNotSupportedError) Error() string {
	if len(e.Supported) > 0 {
		return fmt.Sprintf("unsupported provider type: %q (supported types: %v)", e.ProviderType, e.Supported)
	}
	return fmt.Sprintf("unsupported provider type: %q", e.ProviderType)
}

// ProviderCreationError indicates that a provider could not be created.
type ProviderCreationError struct {
	ProviderName string
	ProviderType string
	Message      string
	Cause        error
}

// Error implements the error interface.
func (e *ProviderCreationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("failed to create provider %q (type: %s): %s: %v", e.ProviderName, e.ProviderType, e.Message, e.Cause)
	}
	return fmt.Sprintf("failed to create provider %q (type: %s): %s", e.ProviderName, e.ProviderType, e.Message)
}

// Unwrap returns the underlying error.
func (e *ProviderCreationError) Unwrap() error {
	return e.Cause
}

// ProviderConfigError indicates a configuration error for a provider.
type ProviderConfigError struct {
	ProviderName string
	ProviderType string
	Field        string
	Message      string
	Cause        error
}

// Error implements the error interface.
func (e *ProviderConfigError) Error() string {
	msg := fmt.Sprintf("invalid configuration for provider %q (type: %s)", e.ProviderName, e.ProviderType)
	if e.Field != "" {
		msg += fmt.Sprintf(", field %q", e.Field)
	}
	if e.Message != "" {
		msg += ": " + e.Message
	}
	if e.Cause != nil {
		msg += ": " + e.Cause.Error()
	}
	return msg
}

// Unwrap returns the underlying error.
func (e *ProviderConfigError) Unwrap() error {
	return e.Cause
}

// ProviderCredentialsError indicates an error with provider credentials.
type ProviderCredentialsError struct {
	ProviderType string
	Message      string
	Cause        error
}

// Error implements the error interface.
func (e *ProviderCredentialsError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("credentials error for provider type %q: %s: %v", e.ProviderType, e.Message, e.Cause)
	}
	return fmt.Sprintf("credentials error for provider type %q: %s", e.ProviderType, e.Message)
}

// Unwrap returns the underlying error.
func (e *ProviderCredentialsError) Unwrap() error {
	return e.Cause
}

// Helper functions to create errors

// NewProviderNotFoundError creates a new ProviderNotFoundError.
func NewProviderNotFoundError(name string, available []string) *ProviderNotFoundError {
	return &ProviderNotFoundError{
		ProviderName: name,
		Available:    available,
	}
}

// NewProviderTypeNotSupportedError creates a new ProviderTypeNotSupportedError.
func NewProviderTypeNotSupportedError(providerType string, supported []string) *ProviderTypeNotSupportedError {
	return &ProviderTypeNotSupportedError{
		ProviderType: providerType,
		Supported:    supported,
	}
}

// NewProviderCreationError creates a new ProviderCreationError.
func NewProviderCreationError(name, providerType, message string, cause error) *ProviderCreationError {
	return &ProviderCreationError{
		ProviderName: name,
		ProviderType: providerType,
		Message:      message,
		Cause:        cause,
	}
}

// NewProviderConfigError creates a new ProviderConfigError.
func NewProviderConfigError(name, providerType, field, message string, cause error) *ProviderConfigError {
	return &ProviderConfigError{
		ProviderName: name,
		ProviderType: providerType,
		Field:        field,
		Message:      message,
		Cause:        cause,
	}
}

// NewProviderCredentialsError creates a new ProviderCredentialsError.
func NewProviderCredentialsError(providerType, message string, cause error) *ProviderCredentialsError {
	return &ProviderCredentialsError{
		ProviderType: providerType,
		Message:      message,
		Cause:        cause,
	}
}
