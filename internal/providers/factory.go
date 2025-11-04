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
	"fmt"
	"sync"

	"github.com/mixanemca/cdnscli/internal/config"
)

// ProviderFactory creates a provider from configuration.
type ProviderFactory interface {
	// CreateProvider creates a new provider instance from the given configuration.
	CreateProvider(cfg *config.ProviderConfig) (Provider, error)
	// Type returns the provider type name (e.g., "cloudflare", "route53").
	Type() string
}

// ProviderRegistry manages provider factories and creates providers.
type ProviderRegistry interface {
	// Register registers a provider factory.
	Register(factory ProviderFactory)
	// CreateProvider creates a provider by name from the given configuration.
	CreateProvider(name string, cfg *config.Config) (Provider, error)
	// GetSupportedTypes returns a list of supported provider types.
	GetSupportedTypes() []string
}

type providerRegistry struct {
	mu       sync.RWMutex
	factories map[string]ProviderFactory
}

// NewProviderRegistry creates a new provider registry.
func NewProviderRegistry() ProviderRegistry {
	return &providerRegistry{
		factories: make(map[string]ProviderFactory),
	}
}

// Register registers a provider factory.
func (r *providerRegistry) Register(factory ProviderFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[factory.Type()] = factory
}

// CreateProvider creates a provider by name from the given configuration.
func (r *providerRegistry) CreateProvider(name string, cfg *config.Config) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Get provider configuration
	providerCfg, err := cfg.GetProvider(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider config: %w", err)
	}

	// Get factory for provider type
	factory, exists := r.factories[providerCfg.Type]
	if !exists {
		return nil, fmt.Errorf("unsupported provider type: %q (supported types: %v)", providerCfg.Type, r.GetSupportedTypes())
	}

	// Create provider using factory
	provider, err := factory.CreateProvider(providerCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider %q: %w", name, err)
	}

	return provider, nil
}

// GetSupportedTypes returns a list of supported provider types.
func (r *providerRegistry) GetSupportedTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.factories))
	for t := range r.factories {
		types = append(types, t)
	}
	return types
}

