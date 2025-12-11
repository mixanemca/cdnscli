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
	"testing"

	"github.com/mixanemca/cdnscli/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestRegRuFactory_Type(t *testing.T) {
	factory := NewRegRuFactory()
	assert.Equal(t, "regru", factory.Type())
}

func TestRegRuFactory_CreateProvider_InvalidType(t *testing.T) {
	factory := NewRegRuFactory()
	cfg := &config.ProviderConfig{
		Type: "invalid-type",
	}

	provider, err := factory.CreateProvider(cfg)
	assert.Nil(t, provider)
	assert.Error(t, err)

	configErr, ok := err.(*ProviderConfigError)
	assert.True(t, ok)
	assert.Contains(t, configErr.Error(), "invalid provider type")
}

func TestRegRuFactory_CreateProvider_MissingCredentials(t *testing.T) {
	factory := NewRegRuFactory()
	cfg := &config.ProviderConfig{
		Type:        "regru",
		Credentials: map[string]interface{}{},
	}

	provider, err := factory.CreateProvider(cfg)
	assert.Nil(t, provider)
	assert.Error(t, err)

	credsErr, ok := err.(*ProviderCredentialsError)
	assert.True(t, ok)
	assert.Contains(t, credsErr.Error(), "username is required")
}

func TestRegRuFactory_CreateProvider_OnlyUsername(t *testing.T) {
	factory := NewRegRuFactory()
	cfg := &config.ProviderConfig{
		Type: "regru",
		Credentials: map[string]interface{}{
			"username": "test-user",
			// Missing password
		},
	}

	provider, err := factory.CreateProvider(cfg)
	assert.Nil(t, provider)
	assert.Error(t, err)

	credsErr, ok := err.(*ProviderCredentialsError)
	assert.True(t, ok)
	assert.Contains(t, credsErr.Error(), "password is required")
}

func TestRegRuFactory_CreateProvider_OnlyPassword(t *testing.T) {
	factory := NewRegRuFactory()
	cfg := &config.ProviderConfig{
		Type: "regru",
		Credentials: map[string]interface{}{
			"password": "test-password",
			// Missing username
		},
	}

	provider, err := factory.CreateProvider(cfg)
	assert.Nil(t, provider)
	assert.Error(t, err)

	credsErr, ok := err.(*ProviderCredentialsError)
	assert.True(t, ok)
	assert.Contains(t, credsErr.Error(), "username is required")
}

func TestRegRuFactory_CreateProvider_EmptyUsername(t *testing.T) {
	factory := NewRegRuFactory()
	cfg := &config.ProviderConfig{
		Type: "regru",
		Credentials: map[string]interface{}{
			"username": "",
			"password": "test-password",
		},
	}

	provider, err := factory.CreateProvider(cfg)
	assert.Nil(t, provider)
	assert.Error(t, err)

	credsErr, ok := err.(*ProviderCredentialsError)
	assert.True(t, ok)
	assert.Contains(t, credsErr.Error(), "username is required")
}

func TestRegRuFactory_CreateProvider_EmptyPassword(t *testing.T) {
	factory := NewRegRuFactory()
	cfg := &config.ProviderConfig{
		Type: "regru",
		Credentials: map[string]interface{}{
			"username": "test-user",
			"password": "",
		},
	}

	provider, err := factory.CreateProvider(cfg)
	assert.Nil(t, provider)
	assert.Error(t, err)

	credsErr, ok := err.(*ProviderCredentialsError)
	assert.True(t, ok)
	assert.Contains(t, credsErr.Error(), "password is required")
}

// Note: Testing actual provider creation with real RegRu API would require
// either integration tests with valid credentials or more sophisticated mocking.
// These tests focus on validation and error handling which can be tested
// without making actual API calls.
