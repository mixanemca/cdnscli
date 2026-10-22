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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudflare/cloudflare-go/v7/option"
	"github.com/mixanemca/cdnscli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestCloudflareFactory_CreateProvider_GlobalKeyInTokenField(t *testing.T) {
	factory := NewCloudflareFactory()
	cfg := &config.ProviderConfig{
		Type: "cloudflare",
		Credentials: map[string]interface{}{
			"api_token": "cfk_0123456789abcdef0123456789abcdef01234567abcd",
		},
	}

	provider, err := factory.CreateProvider(cfg)
	assert.Nil(t, provider)
	require.Error(t, err)

	credsErr, ok := err.(*ProviderCredentialsError)
	require.True(t, ok)
	assert.Contains(t, credsErr.Error(), "looks like a Global API Key")
}

func TestCloudflareFactory_CreateProvider_TokenInKeyField(t *testing.T) {
	factory := NewCloudflareFactory()
	cfg := &config.ProviderConfig{
		Type: "cloudflare",
		Credentials: map[string]interface{}{
			"api_key": "cfat_0123456789abcdef0123456789abcdef01234567abcd",
			"email":   "test@example.com",
		},
	}

	provider, err := factory.CreateProvider(cfg)
	assert.Nil(t, provider)
	require.Error(t, err)

	credsErr, ok := err.(*ProviderCredentialsError)
	require.True(t, ok)
	assert.Contains(t, credsErr.Error(), "looks like an API token")
}

// fakeCloudflare starts a fake Cloudflare API that records the last request
// and answers GET /zones with the given handler.
func fakeCloudflare(t *testing.T, zones http.HandlerFunc) (*httptest.Server, *http.Request) {
	t.Helper()
	last := &http.Request{}
	mux := http.NewServeMux()
	mux.HandleFunc("/zones", func(w http.ResponseWriter, r *http.Request) {
		*last = *r.Clone(r.Context())
		zones(w, r)
	})
	mux.HandleFunc("/user/tokens/verify", func(w http.ResponseWriter, _ *http.Request) {
		t.Error("/user/tokens/verify must not be used: it rejects account tokens")
		w.WriteHeader(http.StatusBadRequest)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, last
}

func zonesOK(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"success":true,"errors":[],"messages":[],"result":[],` +
		`"result_info":{"page":1,"per_page":1,"count":0,"total_count":0,"total_pages":1}}`))
}

func zonesInvalidToken(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte(`{"success":false,` +
		`"errors":[{"code":1000,"message":"Invalid API Token"}],"messages":[],"result":null}`))
}

func newTestCloudflareFactory(srv *httptest.Server) *cloudflareFactory {
	return &cloudflareFactory{opts: []option.RequestOption{
		option.WithBaseURL(srv.URL),
		option.WithMaxRetries(0),
	}}
}

func TestCloudflareFactory_CreateProvider_AccountToken(t *testing.T) {
	srv, last := fakeCloudflare(t, zonesOK)
	token := "cfat_0123456789abcdef0123456789abcdef01234567abcd"
	cfg := &config.ProviderConfig{
		Type:        "cloudflare",
		Credentials: map[string]interface{}{"api_token": token},
	}

	provider, err := newTestCloudflareFactory(srv).CreateProvider(cfg)
	require.NoError(t, err)
	assert.NotNil(t, provider)

	assert.Equal(t, "Bearer "+token, last.Header.Get("Authorization"))
	assert.Equal(t, "1", last.URL.Query().Get("per_page"))
}

func TestCloudflareFactory_CreateProvider_LegacyToken(t *testing.T) {
	srv, last := fakeCloudflare(t, zonesOK)
	token := "0123456789abcdefABCDEF0123456789abcdefAB"
	cfg := &config.ProviderConfig{
		Type:        "cloudflare",
		Credentials: map[string]interface{}{"api_token": token},
	}

	provider, err := newTestCloudflareFactory(srv).CreateProvider(cfg)
	require.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, "Bearer "+token, last.Header.Get("Authorization"))
}

func TestCloudflareFactory_CreateProvider_GlobalKey(t *testing.T) {
	srv, last := fakeCloudflare(t, zonesOK)
	key := "cfk_0123456789abcdef0123456789abcdef01234567abcd"
	cfg := &config.ProviderConfig{
		Type: "cloudflare",
		Credentials: map[string]interface{}{
			"api_key": key,
			"email":   "test@example.com",
		},
	}

	provider, err := newTestCloudflareFactory(srv).CreateProvider(cfg)
	require.NoError(t, err)
	assert.NotNil(t, provider)

	assert.Empty(t, last.Header.Get("Authorization"))
	assert.Equal(t, key, last.Header.Get("X-Auth-Key"))
	assert.Equal(t, "test@example.com", last.Header.Get("X-Auth-Email"))
}

func TestCloudflareFactory_CreateProvider_InvalidToken(t *testing.T) {
	srv, _ := fakeCloudflare(t, zonesInvalidToken)
	cfg := &config.ProviderConfig{
		Type:        "cloudflare",
		Credentials: map[string]interface{}{"api_token": "cfut_definitely-not-a-valid-token"},
	}

	provider, err := newTestCloudflareFactory(srv).CreateProvider(cfg)
	assert.Nil(t, provider)
	require.Error(t, err)

	credsErr, ok := err.(*ProviderCredentialsError)
	require.True(t, ok)
	assert.Contains(t, credsErr.Error(), "failed to verify API token")
	assert.Contains(t, credsErr.Error(), "cfut_/cfat_ prefix")
	assert.Contains(t, credsErr.Error(), "Invalid API Token")
}

func TestCloudflareFactory_CreateProvider_InvalidKey(t *testing.T) {
	srv, _ := fakeCloudflare(t, zonesInvalidToken)
	cfg := &config.ProviderConfig{
		Type: "cloudflare",
		Credentials: map[string]interface{}{
			"api_key": "0123456789abcdef0123456789abcdef01234",
			"email":   "test@example.com",
		},
	}

	provider, err := newTestCloudflareFactory(srv).CreateProvider(cfg)
	assert.Nil(t, provider)
	require.Error(t, err)

	credsErr, ok := err.(*ProviderCredentialsError)
	require.True(t, ok)
	assert.Contains(t, credsErr.Error(), "failed to verify API key and email")
	assert.Contains(t, credsErr.Error(), "cfk_ prefix")
}
