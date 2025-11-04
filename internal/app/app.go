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

// Package app represent cdnscli application for work with providers API's.
package app

import (
	"fmt"
	"os"
	"sync"

	"github.com/mixanemca/cdnscli/internal/config"
	pp "github.com/mixanemca/cdnscli/internal/prettyprint"
	"github.com/mixanemca/cdnscli/internal/providers"
)

var (
	// defaultRegistry is the default provider registry with all providers registered.
	defaultRegistry providers.ProviderRegistry
	registryOnce     sync.Once
)

// initDefaultRegistry initializes the default registry with all available providers.
func initDefaultRegistry() {
	registryOnce.Do(func() {
		defaultRegistry = providers.NewProviderRegistry()
		// Register all available providers
		defaultRegistry.Register(providers.NewCloudflareFactory())
		// Add more providers here as they are implemented
		// defaultRegistry.Register(providers.NewRoute53Factory())
		// defaultRegistry.Register(providers.NewDigitalOceanFactory())
	})
}

type app struct {
	providers        map[string]providers.Provider
	providerDisplayNames map[string]string // Maps provider name to display name
	defaultProvider  providers.Provider
	pp               pp.PrettyPrinter
	output           pp.OutputFormat
	cfg              *config.Config
	providerName     string
	registry         providers.ProviderRegistry
}

// Option options for app
type Option func(c *app) error

// New creates a new application instance. Various client options can be used to configure
// the application.
func New(opts ...Option) (App, error) {
	// Initialize default registry
	initDefaultRegistry()

	// App with default values
	a := &app{
		providerName:         providers.TypeCloudflare, // Default provider
		providers:            make(map[string]providers.Provider),
		providerDisplayNames: make(map[string]string),
		registry:             defaultRegistry,
	}

	for _, opt := range opts {
		if err := opt(a); err != nil {
			return nil, err
		}
	}

	// If config is provided, use it to initialize providers
	if a.cfg != nil {
		// Initialize all providers from config
		for name := range a.cfg.Providers {
			provider, err := a.registry.CreateProvider(name, a.cfg)
			if err != nil {
				// Errors from registry are already properly typed, just return them
				return nil, err
			}
			a.providers[name] = provider
			
			// Store display name for the provider
			providerCfg := a.cfg.Providers[name]
			displayName := providers.GetDisplayName(providerCfg.Type, providerCfg.DisplayName)
			a.providerDisplayNames[name] = displayName
		}

		// Set default provider
		defaultName := a.providerName
		if a.cfg.DefaultProvider != "" {
			defaultName = a.cfg.DefaultProvider
		}

		// Try to get default provider from config
		if defaultProvider, exists := a.providers[defaultName]; exists {
			a.defaultProvider = defaultProvider
		} else if len(a.providers) > 0 {
			// If requested provider not found, use first available
			for _, p := range a.providers {
				a.defaultProvider = p
				break
			}
		} else {
			return nil, fmt.Errorf("no providers configured")
		}
	} else {
		// Fallback to environment variable for backward compatibility
		provider, err := createProviderFromEnv(a.registry)
		if err != nil {
			return nil, err
		}
		a.defaultProvider = provider
		providerName := providers.TypeCloudflare
		a.providers[providerName] = provider
		// Use default display name for Cloudflare
		a.providerDisplayNames[providerName] = providers.GetDisplayName(providers.TypeCloudflare, "")
	}

	a.pp = pp.New(pp.OutputFormat(a.output))

	return a, nil
}

// createProviderFromEnv creates a provider from environment variables (backward compatibility).
func createProviderFromEnv(registry providers.ProviderRegistry) (providers.Provider, error) {
	// Check for Cloudflare API token in environment
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	if apiToken == "" {
		return nil, fmt.Errorf("no configuration provided and CLOUDFLARE_API_TOKEN not set")
	}

	// Create temporary config with Cloudflare provider from env
	cfg := &config.Config{
		Providers: make(map[string]config.ProviderConfig),
	}
	cfg.Providers[providers.TypeCloudflare] = config.ProviderConfig{
		Type: providers.TypeCloudflare,
		Credentials: map[string]interface{}{
			"api_token": apiToken,
		},
	}

	return registry.CreateProvider(providers.TypeCloudflare, cfg)
}

func (a *app) Provider() providers.Provider {
	return a.defaultProvider
}

func (a *app) GetProvider(name string) (providers.Provider, error) {
	if name == "" {
		return a.Provider(), nil
	}

	provider, exists := a.providers[name]
	if !exists {
		return nil, providers.NewProviderNotFoundError(name, a.ProviderNames())
	}

	return provider, nil
}

func (a *app) ProviderNames() []string {
	names := make([]string, 0, len(a.providers))
	for name := range a.providers {
		names = append(names, name)
	}
	return names
}

func (a *app) DefaultProviderName() string {
	// Find the provider name that matches the default provider
	for name, provider := range a.providers {
		if provider == a.defaultProvider {
			// Return display name if available, otherwise return provider name
			if displayName, ok := a.providerDisplayNames[name]; ok {
				return displayName
			}
			return name
		}
	}
	// Fallback: try to get display name for stored providerName
	if displayName, ok := a.providerDisplayNames[a.providerName]; ok {
		return displayName
	}
	// Final fallback: use default display name for provider type if available
	if a.cfg != nil {
		if providerCfg, exists := a.cfg.Providers[a.providerName]; exists {
			return providers.GetDisplayName(providerCfg.Type, providerCfg.DisplayName)
		}
	}
	// Last resort: use provider name as-is
	return a.providerName
}

func (a *app) Printer() pp.PrettyPrinter {
	return a.pp
}
