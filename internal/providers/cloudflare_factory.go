/*
Copyright © 2024-2026 Michael Bruskov <mixanemca@yandex.ru>

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
	"net/http"

	"github.com/cloudflare/cloudflare-go"
	"github.com/mixanemca/cdnscli/internal/config"
)

// cloudflareFactory creates Cloudflare providers.
type cloudflareFactory struct {
	// opts are extra options passed to the Cloudflare API client
	// (used by tests to point the client at a fake API server).
	opts []cloudflare.Option
}

// NewCloudflareFactory creates a new Cloudflare provider factory.
func NewCloudflareFactory() ProviderFactory {
	return &cloudflareFactory{}
}

// Type returns the provider type name.
func (f *cloudflareFactory) Type() string {
	return TypeCloudflare
}

// CreateProvider creates a Cloudflare provider from configuration.
func (f *cloudflareFactory) CreateProvider(cfg *config.ProviderConfig) (Provider, error) {
	if cfg.Type != TypeCloudflare {
		return nil, NewProviderConfigError("", TypeCloudflare, "type",
			fmt.Sprintf("invalid provider type for Cloudflare factory: %q", cfg.Type), nil)
	}

	creds, err := cfg.GetCloudflareCredentials()
	if err != nil {
		return nil, NewProviderCredentialsError(TypeCloudflare, "failed to get credentials", err)
	}

	// Validate credentials
	if creds.APIToken == "" && (creds.APIKey == "" || creds.Email == "") {
		return nil, NewProviderCredentialsError(TypeCloudflare,
			"incomplete credentials: need either api_token or (api_key + email)", nil)
	}
	if err := creds.CheckPrefixes(); err != nil {
		return nil, NewProviderCredentialsError(TypeCloudflare, err.Error(), nil)
	}

	var api *cloudflare.API

	if creds.APIToken != "" {
		api, err = cloudflare.NewWithAPIToken(creds.APIToken, f.opts...)
		if err != nil {
			return nil, NewProviderCredentialsError(TypeCloudflare,
				"failed to create API client with token", err)
		}
	} else {
		api, err = cloudflare.New(creds.APIKey, creds.Email, f.opts...)
		if err != nil {
			return nil, NewProviderCredentialsError(TypeCloudflare,
				"failed to create API client with API key and email", err)
		}
	}

	if err := verifyCloudflareCredentials(context.Background(), api); err != nil {
		return nil, NewProviderCredentialsError(TypeCloudflare, cloudflareVerifyHint(creds), err)
	}

	repo := NewRepoCloudFlare(api)
	return NewProvider(repo), nil
}

// verifyCloudflareCredentials checks that Cloudflare accepts the configured
// credentials by requesting a single zone.
//
// The dedicated /user/tokens/verify endpoint is deliberately not used: it only
// accepts User API Tokens and rejects Account API Tokens (cfat_ prefix) and
// Global API Keys with "Invalid API Token (1000)". Listing zones works with all
// credential types and additionally checks the Zone:Read permission that every
// cdnscli command needs anyway.
func verifyCloudflareCredentials(ctx context.Context, api *cloudflare.API) error {
	_, err := api.Raw(ctx, http.MethodGet, "/zones?per_page=1", nil, nil)
	return err
}

// cloudflareVerifyHint returns a human-readable explanation of a failed
// credentials check, tailored to the authentication method in use.
func cloudflareVerifyHint(creds *config.CloudflareCredentials) string {
	if creds.APIToken != "" {
		return fmt.Sprintf("failed to verify API token (make sure it is copied in full "+
			"including the %s/%s prefix, has Zone:Read permission and is not expired)",
			config.CloudflareUserTokenPrefix, config.CloudflareAccountTokenPrefix)
	}
	return fmt.Sprintf("failed to verify API key and email (make sure the Global API Key "+
		"is copied in full including the %s prefix and email matches the account)",
		config.CloudflareGlobalKeyPrefix)
}
