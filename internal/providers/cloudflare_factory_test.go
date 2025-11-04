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

package providers

import (
	"testing"

	"github.com/mixanemca/cdnscli/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestCloudflareFactory_Type(t *testing.T) {
	factory := NewCloudflareFactory()
	assert.Equal(t, "cloudflare", factory.Type())
}

func TestCloudflareFactory_CreateProvider_InvalidType(t *testing.T) {
	factory := NewCloudflareFactory()
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

func TestCloudflareFactory_CreateProvider_MissingCredentials(t *testing.T) {
	factory := NewCloudflareFactory()
	cfg := &config.ProviderConfig{
		Type:        "cloudflare",
		Credentials: map[string]interface{}{},
	}

	provider, err := factory.CreateProvider(cfg)
	assert.Nil(t, provider)
	assert.Error(t, err)
	
	credsErr, ok := err.(*ProviderCredentialsError)
	assert.True(t, ok)
	assert.Contains(t, credsErr.Error(), "incomplete credentials")
}

func TestCloudflareFactory_CreateProvider_OnlyAPIKey(t *testing.T) {
	factory := NewCloudflareFactory()
	cfg := &config.ProviderConfig{
		Type: "cloudflare",
		Credentials: map[string]interface{}{
			"api_key": "test-key",
			// Missing email
		},
	}

	provider, err := factory.CreateProvider(cfg)
	assert.Nil(t, provider)
	assert.Error(t, err)
	
	credsErr, ok := err.(*ProviderCredentialsError)
	assert.True(t, ok)
	assert.Contains(t, credsErr.Error(), "incomplete credentials")
}

func TestCloudflareFactory_CreateProvider_OnlyEmail(t *testing.T) {
	factory := NewCloudflareFactory()
	cfg := &config.ProviderConfig{
		Type: "cloudflare",
		Credentials: map[string]interface{}{
			"email": "test@example.com",
			// Missing api_key
		},
	}

	provider, err := factory.CreateProvider(cfg)
	assert.Nil(t, provider)
	assert.Error(t, err)
	
	credsErr, ok := err.(*ProviderCredentialsError)
	assert.True(t, ok)
	assert.Contains(t, credsErr.Error(), "incomplete credentials")
}

func TestCloudflareFactory_CreateProvider_EmptyAPIToken(t *testing.T) {
	factory := NewCloudflareFactory()
	cfg := &config.ProviderConfig{
		Type: "cloudflare",
		Credentials: map[string]interface{}{
			"api_token": "",
		},
	}

	provider, err := factory.CreateProvider(cfg)
	assert.Nil(t, provider)
	assert.Error(t, err)
	
	credsErr, ok := err.(*ProviderCredentialsError)
	assert.True(t, ok)
	assert.Contains(t, credsErr.Error(), "incomplete credentials")
}

// Note: Testing actual provider creation with real Cloudflare API would require
// either integration tests with a test token or more sophisticated mocking.
// These tests focus on validation and error handling which can be tested
// without making actual API calls.

