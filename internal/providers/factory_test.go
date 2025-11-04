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
	"context"
	"errors"
	"testing"

	"github.com/mixanemca/cdnscli/internal/config"
	"github.com/mixanemca/cdnscli/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProviderFactory is a mock implementation of ProviderFactory.
type MockProviderFactory struct {
	mock.Mock
}

func (m *MockProviderFactory) CreateProvider(cfg *config.ProviderConfig) (Provider, error) {
	args := m.Called(cfg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(Provider), args.Error(1)
}

func (m *MockProviderFactory) Type() string {
	args := m.Called()
	return args.String(0)
}

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

func TestNewProviderRegistry(t *testing.T) {
	registry := NewProviderRegistry()
	assert.NotNil(t, registry)
	assert.Equal(t, []string{}, registry.GetSupportedTypes())
}

func TestProviderRegistry_Register(t *testing.T) {
	registry := NewProviderRegistry()
	factory := new(MockProviderFactory)
	factory.On("Type").Return("test-type")

	registry.Register(factory)

	types := registry.GetSupportedTypes()
	assert.Contains(t, types, "test-type")
	factory.AssertExpectations(t)
}

func TestProviderRegistry_GetSupportedTypes(t *testing.T) {
	registry := NewProviderRegistry()
	
	// Initially empty
	types := registry.GetSupportedTypes()
	assert.Empty(t, types)

	// Register multiple factories
	factory1 := new(MockProviderFactory)
	factory1.On("Type").Return("type1")
	factory2 := new(MockProviderFactory)
	factory2.On("Type").Return("type2")

	registry.Register(factory1)
	registry.Register(factory2)

	types = registry.GetSupportedTypes()
	assert.Len(t, types, 2)
	assert.Contains(t, types, "type1")
	assert.Contains(t, types, "type2")
}

func TestProviderRegistry_CreateProvider_Success(t *testing.T) {
	registry := NewProviderRegistry()
	
	// Setup mock factory
	factory := new(MockProviderFactory)
	factory.On("Type").Return("test-type")
	
	mockProvider := new(MockProvider)
	providerCfg := &config.ProviderConfig{
		Type: "test-type",
		Credentials: map[string]interface{}{
			"token": "test-token",
		},
	}
	factory.On("CreateProvider", providerCfg).Return(mockProvider, nil)

	registry.Register(factory)

	// Setup config
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{
			"test-provider": *providerCfg,
		},
	}

	provider, err := registry.CreateProvider("test-provider", cfg)
	assert.NoError(t, err)
	assert.Equal(t, mockProvider, provider)
	factory.AssertExpectations(t)
}

func TestProviderRegistry_CreateProvider_ProviderNotFound(t *testing.T) {
	registry := NewProviderRegistry()
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{},
	}

	provider, err := registry.CreateProvider("non-existent", cfg)
	assert.Nil(t, provider)
	assert.Error(t, err)
	
	configErr, ok := err.(*ProviderConfigError)
	assert.True(t, ok)
	assert.Contains(t, configErr.Error(), "failed to get provider configuration")
}

func TestProviderRegistry_CreateProvider_UnsupportedType(t *testing.T) {
	registry := NewProviderRegistry()
	
	providerCfg := &config.ProviderConfig{
		Type: "unsupported-type",
	}
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{
			"test-provider": *providerCfg,
		},
	}

	provider, err := registry.CreateProvider("test-provider", cfg)
	assert.Nil(t, provider)
	assert.Error(t, err)
	
	_, ok := err.(*ProviderTypeNotSupportedError)
	assert.True(t, ok)
}

func TestProviderRegistry_CreateProvider_FactoryError(t *testing.T) {
	registry := NewProviderRegistry()
	
	factory := new(MockProviderFactory)
	factory.On("Type").Return("test-type")
	
	providerCfg := &config.ProviderConfig{
		Type: "test-type",
	}
	factoryError := errors.New("factory creation failed")
	factory.On("CreateProvider", providerCfg).Return(nil, factoryError)

	registry.Register(factory)

	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{
			"test-provider": *providerCfg,
		},
	}

	provider, err := registry.CreateProvider("test-provider", cfg)
	assert.Nil(t, provider)
	assert.Error(t, err)
	
	creationErr, ok := err.(*ProviderCreationError)
	assert.True(t, ok)
	assert.Equal(t, factoryError, creationErr.Unwrap())
	factory.AssertExpectations(t)
}

func TestProviderRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewProviderRegistry()
	
	// Register factories concurrently
	factory1 := new(MockProviderFactory)
	factory1.On("Type").Return("type1")
	factory2 := new(MockProviderFactory)
	factory2.On("Type").Return("type2")

	done := make(chan bool, 2)
	
	go func() {
		registry.Register(factory1)
		done <- true
	}()
	
	go func() {
		registry.Register(factory2)
		done <- true
	}()
	
	<-done
	<-done

	types := registry.GetSupportedTypes()
	assert.Len(t, types, 2)
	assert.Contains(t, types, "type1")
	assert.Contains(t, types, "type2")
}
