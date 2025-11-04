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

// Package app represent cdnscli application for work with providers API's.
package app

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudflare/cloudflare-go"
	"github.com/mixanemca/cdnscli/internal/config"
	pp "github.com/mixanemca/cdnscli/internal/prettyprint"
	"github.com/mixanemca/cdnscli/internal/providers"
)

type app struct {
	provider providers.Provider
	pp       pp.PrettyPrinter
	output   pp.OutputFormat
	cfg      *config.Config
	providerName string
}

// Option options for app
type Option func(c *app) error

// New creates a new application instance. Various client options can be used to configure
// the application.
func New(opts ...Option) (App, error) {
	// App with default values
	a := &app{
		providerName: "cloudflare", // Default provider
	}

	for _, opt := range opts {
		if err := opt(a); err != nil {
			return nil, err
		}
	}

	// If config is provided, use it to initialize provider
	if a.cfg != nil {
		provider, err := a.cfg.GetProvider(a.providerName)
		if err != nil {
			return nil, fmt.Errorf("failed to get provider %q: %w", a.providerName, err)
		}

		// Initialize provider based on type
		switch provider.Type {
		case "cloudflare":
			repo, err := a.initCloudflareProvider(provider)
			if err != nil {
				return nil, fmt.Errorf("failed to initialize Cloudflare provider: %w", err)
			}
			a.provider = providers.NewProvider(repo)
		default:
			return nil, fmt.Errorf("unsupported provider type: %q", provider.Type)
		}
	} else {
		// Fallback to environment variable for backward compatibility
		apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
		if apiToken == "" {
			return nil, fmt.Errorf("no configuration provided and CLOUDFLARE_API_TOKEN not set")
		}

		api, err := cloudflare.NewWithAPIToken(apiToken)
		if err != nil {
			return nil, fmt.Errorf("failed to create Cloudflare API client: %w", err)
		}

		// Verify token
		_, err = api.VerifyAPIToken(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to verify Cloudflare API token: %w", err)
		}

		repo := providers.NewRepoCloudFlare(api)
		a.provider = providers.NewProvider(repo)
	}

	a.pp = pp.New(pp.OutputFormat(a.output))

	return a, nil
}

// initCloudflareProvider initializes Cloudflare provider from config.
func (a *app) initCloudflareProvider(provider *config.ProviderConfig) (providers.Repo, error) {
	creds, err := provider.GetCloudflareCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to get Cloudflare credentials: %w", err)
	}

	var api *cloudflare.API

	if creds.APIToken != "" {
		api, err = cloudflare.NewWithAPIToken(creds.APIToken)
		if err != nil {
			return nil, fmt.Errorf("failed to create Cloudflare API client with token: %w", err)
		}
	} else if creds.APIKey != "" && creds.Email != "" {
		api, err = cloudflare.New(creds.APIKey, creds.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to create Cloudflare API client with key: %w", err)
		}
	} else {
		return nil, fmt.Errorf("cloudflare credentials incomplete: need either api_token or (api_key + email)")
	}

	// Verify token/key
	_, err = api.VerifyAPIToken(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to verify Cloudflare API credentials: %w", err)
	}

	return providers.NewRepoCloudFlare(api), nil
}

func (a *app) Provider() providers.Provider {
	return a.provider
}

func (a *app) Printer() pp.PrettyPrinter {
	return a.pp
}
