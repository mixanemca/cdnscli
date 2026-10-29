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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mixanemca/cdnscli/internal/config"
	namecheap "github.com/namecheap/go-namecheap-sdk/v2/namecheap"
)

const (
	defaultIPProviderURL = "https://api.ipify.org"

	// namecheapUserAgent is appended to the SDK User-Agent so Namecheap support can
	// identify the calling client.
	namecheapUserAgent = "cdnscli"

	// namecheapVerifyTimeout bounds the credentials check performed when the
	// provider is created.
	namecheapVerifyTimeout = 30 * time.Second

	// namecheapIPDetectTimeout bounds the public IP lookup.
	namecheapIPDetectTimeout = 10 * time.Second

	// namecheapVerifyPageSize is the smallest page size domains.getList accepts.
	namecheapVerifyPageSize = 10
)

// namecheapAPIBaseURL overrides the Namecheap API endpoint when non-empty.
// It exists for tests that point the factory at a mock server.
var namecheapAPIBaseURL string

type namecheapFactory struct{}

// NewNamecheapFactory creates a new Namecheap provider factory.
func NewNamecheapFactory() ProviderFactory {
	return &namecheapFactory{}
}

// Type returns the provider type name.
func (f *namecheapFactory) Type() string {
	return TypeNamecheap
}

// CreateProvider creates a Namecheap provider from configuration.
func (f *namecheapFactory) CreateProvider(cfg *config.ProviderConfig) (Provider, error) {
	if cfg.Type != TypeNamecheap {
		return nil, NewProviderConfigError("", TypeNamecheap, "type",
			fmt.Sprintf("invalid provider type for Namecheap factory: %q", cfg.Type), nil)
	}

	creds, err := cfg.GetNamecheapCredentials()
	if err != nil {
		return nil, NewProviderCredentialsError("namecheap", "failed to get credentials", err)
	}

	apiUser := strings.TrimSpace(creds.APIUser)
	apiKey := strings.TrimSpace(creds.APIKey)

	if apiUser == "" {
		return nil, NewProviderCredentialsError("namecheap",
			"api_user is required but not provided in credentials (check config file)", nil)
	}
	if apiKey == "" {
		return nil, NewProviderCredentialsError("namecheap",
			"api_key is required but not provided in credentials (check config file)", nil)
	}

	userName := strings.TrimSpace(creds.UserName)
	if userName == "" {
		userName = apiUser
	}

	ipProviderURL := defaultIPProviderURL
	if cfg.Options != nil {
		if u, ok := cfg.Options["ip_provider_url"].(string); ok && strings.TrimSpace(u) != "" {
			ipProviderURL = strings.TrimSpace(u)
		}
	}

	clientIP, err := detectPublicIP(ipProviderURL)
	if err != nil {
		return nil, NewProviderCredentialsError("namecheap",
			fmt.Sprintf("failed to detect public IP from %s", ipProviderURL), err)
	}

	client := namecheap.NewClient(&namecheap.ClientOptions{
		UserName:   userName,
		ApiUser:    apiUser,
		ApiKey:     apiKey,
		ClientIp:   clientIP,
		UseSandbox: creds.UseSandbox,
		UserAgent:  namecheapUserAgent,
	})
	if namecheapAPIBaseURL != "" {
		client.BaseURL = namecheapAPIBaseURL
	}

	if err := verifyNamecheapCredentials(client, clientIP); err != nil {
		return nil, err
	}

	repo := NewRepoNamecheap(client)
	return NewProvider(repo), nil
}

// verifyNamecheapCredentials performs the cheapest authenticated call to make sure
// the API user, key and client IP are accepted before any command runs.
func verifyNamecheapCredentials(client *namecheap.Client, clientIP string) error {
	ctx, cancel := context.WithTimeout(context.Background(), namecheapVerifyTimeout)
	defer cancel()

	_, err := client.Domains.GetListWithContext(ctx, &namecheap.DomainsGetListArgs{
		PageSize: namecheap.Int(namecheapVerifyPageSize),
	})
	if err == nil {
		return nil
	}

	if errors.Is(err, namecheap.ErrAccessDenied) {
		return NewProviderCredentialsError("namecheap",
			fmt.Sprintf("access denied: api_user/api_key may be invalid "+
				"or client IP %s is not whitelisted in the Namecheap API settings", clientIP), err)
	}

	return NewProviderCredentialsError("namecheap", "failed to verify credentials", err)
}

// detectPublicIP fetches the public IP from the given URL.
// Supports plain-text response ("1.2.3.4") and JSON ({"ip":"1.2.3.4"}).
func detectPublicIP(ipProviderURL string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), namecheapIPDetectTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ipProviderURL, nil)
	if err != nil {
		return "", fmt.Errorf("invalid IP provider URL: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected HTTP status %s from %s", resp.Status, ipProviderURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Try JSON {"ip":"..."} first
	var jsonResp struct {
		IP string `json:"ip"`
	}
	if err := json.Unmarshal(body, &jsonResp); err == nil && jsonResp.IP != "" {
		return strings.TrimSpace(jsonResp.IP), nil
	}

	ip := strings.TrimSpace(string(body))
	if ip == "" {
		return "", fmt.Errorf("empty response from %s", ipProviderURL)
	}
	return ip, nil
}
