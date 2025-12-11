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
	"context"
	"fmt"
	"strings"

	"github.com/mixanemca/cdnscli/internal/config"
	"github.com/mixanemca/regru-go"
)

// regruFactory creates RegRu providers.
type regruFactory struct{}

// NewRegRuFactory creates a new RegRu provider factory.
func NewRegRuFactory() ProviderFactory {
	return &regruFactory{}
}

// Type returns the provider type name.
func (f *regruFactory) Type() string {
	return TypeRegRu
}

// CreateProvider creates a RegRu provider from configuration.
func (f *regruFactory) CreateProvider(cfg *config.ProviderConfig) (Provider, error) {
	if cfg.Type != TypeRegRu {
		return nil, NewProviderConfigError("", TypeRegRu, "type",
			fmt.Sprintf("invalid provider type for RegRu factory: %q", cfg.Type), nil)
	}

	creds, err := cfg.GetRegRuCredentials()
	if err != nil {
		return nil, NewProviderCredentialsError("regru", "failed to get credentials", err)
	}

	// Validate credentials - check for empty strings after trimming whitespace
	username := strings.TrimSpace(creds.Username)
	password := strings.TrimSpace(creds.Password)

	if username == "" {
		return nil, NewProviderCredentialsError("regru",
			"username is required but not provided in credentials (check config file)", nil)
	}
	if password == "" {
		return nil, NewProviderCredentialsError("regru",
			"password is required but not provided in credentials (check config file)", nil)
	}

	// Create RegRu client with trimmed credentials
	client := regru.NewClient(username, password)

	// Verify credentials by trying to list zones
	ctx := context.Background()
	_, err = client.ListZones(ctx)
	if err != nil {
		return nil, NewProviderCredentialsError("regru",
			"failed to verify credentials (username/password may be invalid)", err)
	}

	repo := NewRepoRegRu(client)
	return NewProvider(repo), nil
}
