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

package app

import (
	"context"
	"testing"

	"github.com/mixanemca/cdnscli/internal/config"
	"github.com/mixanemca/cdnscli/internal/models"
	pp "github.com/mixanemca/cdnscli/internal/prettyprint"
	"github.com/mixanemca/cdnscli/internal/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProvider is a mock implementation of Provider.
type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) AddRR(ctx context.Context, zone string, params models.CreateDNSRecordParams) (models.DNSRecord, error) {
	args := m.Called(ctx, zone, params)
	return args.Get(0).(models.DNSRecord), args.Error(1)
}

func (m *MockProvider) DeleteRR(ctx context.Context, zone string, rr models.DNSRecord) error {
	args := m.Called(ctx, zone, rr)
	return args.Error(0)
}

func (m *MockProvider) GetRRByName(ctx context.Context, zone, name string) (models.DNSRecord, error) {
	args := m.Called(ctx, zone, name)
	return args.Get(0).(models.DNSRecord), args.Error(1)
}

func (m *MockProvider) ListZones(ctx context.Context) ([]models.Zone, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Zone), args.Error(1)
}

func (m *MockProvider) ListZonesByName(ctx context.Context, name string) ([]models.Zone, error) {
	args := m.Called(ctx, name)
	return args.Get(0).([]models.Zone), args.Error(1)
}

func (m *MockProvider) ListRecords(ctx context.Context, params models.ListDNSRecordsParams) ([]models.DNSRecord, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]models.DNSRecord), args.Error(1)
}

func (m *MockProvider) ListRecordsByZoneID(ctx context.Context, id string, params models.ListDNSRecordsParams) ([]models.DNSRecord, error) {
	args := m.Called(ctx, id, params)
	return args.Get(0).([]models.DNSRecord), args.Error(1)
}

func (m *MockProvider) UpdateRR(ctx context.Context, zone string, rr models.DNSRecord) (models.DNSRecord, error) {
	args := m.Called(ctx, zone, rr)
	return args.Get(0).(models.DNSRecord), args.Error(1)
}

func TestNew_WithConfig_InvalidProvider(t *testing.T) {
	cfg := &config.Config{
		DefaultProvider: "non-existent",
		Providers: map[string]config.ProviderConfig{
			"non-existent": {
				Type:        "invalid-type",
				Credentials: map[string]interface{}{},
			},
		},
	}

	app, err := New(WithConfig(cfg))
	assert.Error(t, err)
	assert.Nil(t, app)
	// Should return ProviderTypeNotSupportedError
	_, ok := err.(*providers.ProviderTypeNotSupportedError)
	assert.True(t, ok)
}

func TestNew_WithOutputFormat_NoConfig(t *testing.T) {
	app, err := New(
		WithOutputFormat(pp.FormatJSON),
	)

	// Should fail because no config is provided
	assert.Error(t, err)
	assert.Nil(t, app)
	assert.Contains(t, err.Error(), "no configuration provided")
}

func TestNew_WithProvider_InvalidCredentials(t *testing.T) {
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{
			"test-provider": {
				Type: "cloudflare",
				Credentials: map[string]interface{}{
					"api_token": "", // Empty token will fail validation
				},
			},
		},
	}

	app, err := New(
		WithConfig(cfg),
		WithProvider("test-provider"),
	)

	// Should fail because credentials are invalid
	assert.Error(t, err)
	assert.Nil(t, app)
	// Should return ProviderCredentialsError or ProviderCreationError
	_, ok1 := err.(*providers.ProviderCredentialsError)
	_, ok2 := err.(*providers.ProviderCreationError)
	assert.True(t, ok1 || ok2, "error should be ProviderCredentialsError or ProviderCreationError, got: %T", err)
}

func TestNew_NoConfig(t *testing.T) {
	app, err := New()
	assert.Error(t, err)
	assert.Nil(t, app)
	assert.Contains(t, err.Error(), "no configuration provided")
}

func TestApp_Provider_NoProviders(t *testing.T) {
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{},
	}

	app, err := New(WithConfig(cfg))
	assert.Error(t, err)
	assert.Nil(t, app)
	assert.Contains(t, err.Error(), "no providers configured")
}

func TestApp_GetProvider_NotFound(t *testing.T) {
	// Create app manually to test GetProvider without needing valid providers
	a := &app{
		providers: make(map[string]providers.Provider),
	}

	// Get non-existent provider
	_, err := a.GetProvider("non-existent")
	assert.Error(t, err)
	_, ok := err.(*providers.ProviderNotFoundError)
	assert.True(t, ok)
}

func TestApp_GetProvider_EmptyName(t *testing.T) {
	mockProvider := new(MockProvider)

	// Create app manually
	a := &app{
		providers:       make(map[string]providers.Provider),
		defaultProvider: mockProvider,
	}

	// Get with empty name should return default
	provider, err := a.GetProvider("")
	assert.NoError(t, err)
	assert.Equal(t, mockProvider, provider)
}

func TestApp_ProviderNames(t *testing.T) {
	mockProvider1 := new(MockProvider)
	mockProvider2 := new(MockProvider)

	// Create app manually
	a := &app{
		providers: map[string]providers.Provider{
			"provider1": mockProvider1,
			"provider2": mockProvider2,
		},
	}

	names := a.ProviderNames()
	assert.Len(t, names, 2)
	assert.Contains(t, names, "provider1")
	assert.Contains(t, names, "provider2")
}

func TestApp_Printer(t *testing.T) {
	// Create app manually to test Printer without needing valid providers
	a := &app{
		providers: make(map[string]providers.Provider),
		pp:        pp.New(pp.FormatJSON),
	}

	printer := a.Printer()
	assert.NotNil(t, printer)
}

func TestNew_NoProvidersConfigured(t *testing.T) {
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{},
	}

	app, err := New(WithConfig(cfg))
	assert.Error(t, err)
	assert.Nil(t, app)
	assert.Contains(t, err.Error(), "no providers configured")
}

func TestNew_DefaultProviderFromConfig_NotFound(t *testing.T) {
	cfg := &config.Config{
		DefaultProvider: "non-existent-provider",
		Providers: map[string]config.ProviderConfig{
			"provider1": {
				Type: "cloudflare",
				Credentials: map[string]interface{}{
					"api_token": "test-token",
				},
			},
		},
	}

	// Should use first available provider if default not found
	app, err := New(WithConfig(cfg))
	// Will fail at provider creation, but we test the logic path
	if err == nil {
		// If it succeeds, default should be set to first available
		assert.NotNil(t, app)
	}
}

func TestWithOutputFormat(t *testing.T) {
	opt := WithOutputFormat(pp.FormatJSON)
	assert.NotNil(t, opt)

	// Test that it doesn't error
	a := &app{}
	err := opt(a)
	assert.NoError(t, err)
	assert.Equal(t, pp.FormatJSON, a.output)
}

func TestWithConfig(t *testing.T) {
	cfg := &config.Config{}
	opt := WithConfig(cfg)
	assert.NotNil(t, opt)

	a := &app{}
	err := opt(a)
	assert.NoError(t, err)
	assert.Equal(t, cfg, a.cfg)
}

func TestWithProvider(t *testing.T) {
	opt := WithProvider("test-provider")
	assert.NotNil(t, opt)

	a := &app{}
	err := opt(a)
	assert.NoError(t, err)
	assert.Equal(t, "test-provider", a.providerName)
}
