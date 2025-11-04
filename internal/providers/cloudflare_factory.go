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
	"fmt"

	"github.com/cloudflare/cloudflare-go"
	"github.com/mixanemca/cdnscli/internal/config"
)

// cloudflareFactory creates Cloudflare providers.
type cloudflareFactory struct{}

// NewCloudflareFactory creates a new Cloudflare provider factory.
func NewCloudflareFactory() ProviderFactory {
	return &cloudflareFactory{}
}

// Type returns the provider type name.
func (f *cloudflareFactory) Type() string {
	return "cloudflare"
}

// CreateProvider creates a Cloudflare provider from configuration.
func (f *cloudflareFactory) CreateProvider(cfg *config.ProviderConfig) (Provider, error) {
	if cfg.Type != "cloudflare" {
		return nil, fmt.Errorf("invalid provider type for Cloudflare factory: %q", cfg.Type)
	}

	creds, err := cfg.GetCloudflareCredentials()
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

	repo := NewRepoCloudFlare(api)
	return NewProvider(repo), nil
}

