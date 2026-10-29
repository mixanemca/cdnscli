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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mixanemca/cdnscli/internal/config"
	namecheap "github.com/namecheap/go-namecheap-sdk/v2/namecheap"
	"github.com/namecheap/go-namecheap-sdk/v2/namecheaptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNamecheapFactory_Type(t *testing.T) {
	factory := NewNamecheapFactory()
	assert.Equal(t, "namecheap", factory.Type())
}

func TestNamecheapFactory_CreateProvider_InvalidType(t *testing.T) {
	factory := NewNamecheapFactory()
	cfg := &config.ProviderConfig{Type: "invalid-type"}

	provider, err := factory.CreateProvider(cfg)
	assert.Nil(t, provider)
	require.Error(t, err)

	configErr, ok := err.(*ProviderConfigError)
	assert.True(t, ok)
	assert.Contains(t, configErr.Error(), "invalid provider type")
}

func TestNamecheapFactory_CreateProvider_MissingAPIUser(t *testing.T) {
	factory := NewNamecheapFactory()
	cfg := &config.ProviderConfig{
		Type:        "namecheap",
		Credentials: map[string]interface{}{"api_key": "some-key"},
	}

	provider, err := factory.CreateProvider(cfg)
	assert.Nil(t, provider)
	require.Error(t, err)

	credsErr, ok := err.(*ProviderCredentialsError)
	assert.True(t, ok)
	assert.Contains(t, credsErr.Error(), "api_user is required")
}

func TestNamecheapFactory_CreateProvider_EmptyAPIUser(t *testing.T) {
	factory := NewNamecheapFactory()
	cfg := &config.ProviderConfig{
		Type:        "namecheap",
		Credentials: map[string]interface{}{"api_user": "  ", "api_key": "some-key"},
	}

	provider, err := factory.CreateProvider(cfg)
	assert.Nil(t, provider)
	require.Error(t, err)

	credsErr, ok := err.(*ProviderCredentialsError)
	assert.True(t, ok)
	assert.Contains(t, credsErr.Error(), "api_user is required")
}

func TestNamecheapFactory_CreateProvider_MissingAPIKey(t *testing.T) {
	factory := NewNamecheapFactory()
	cfg := &config.ProviderConfig{
		Type:        "namecheap",
		Credentials: map[string]interface{}{"api_user": "some-user"},
	}

	provider, err := factory.CreateProvider(cfg)
	assert.Nil(t, provider)
	require.Error(t, err)

	credsErr, ok := err.(*ProviderCredentialsError)
	assert.True(t, ok)
	assert.Contains(t, credsErr.Error(), "api_key is required")
}

func TestNamecheapFactory_CreateProvider_EmptyAPIKey(t *testing.T) {
	factory := NewNamecheapFactory()
	cfg := &config.ProviderConfig{
		Type:        "namecheap",
		Credentials: map[string]interface{}{"api_user": "some-user", "api_key": ""},
	}

	provider, err := factory.CreateProvider(cfg)
	assert.Nil(t, provider)
	require.Error(t, err)

	credsErr, ok := err.(*ProviderCredentialsError)
	assert.True(t, ok)
	assert.Contains(t, credsErr.Error(), "api_key is required")
}

func TestNamecheapFactory_CreateProvider_IPDetectionFails(t *testing.T) {
	// Use a URL that immediately returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	factory := NewNamecheapFactory()
	cfg := &config.ProviderConfig{
		Type: "namecheap",
		Credentials: map[string]interface{}{
			"api_user": "some-user",
			"api_key":  "some-key",
		},
		Options: map[string]interface{}{
			"ip_provider_url": server.URL,
		},
	}

	provider, err := factory.CreateProvider(cfg)
	assert.Nil(t, provider)
	require.Error(t, err)

	credsErr, ok := err.(*ProviderCredentialsError)
	assert.True(t, ok)
	assert.Contains(t, credsErr.Error(), "failed to detect public IP")
}

func TestDetectPublicIP_PlainText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("1.2.3.4\n"))
	}))
	defer server.Close()

	ip, err := detectPublicIP(server.URL)
	require.NoError(t, err)
	assert.Equal(t, "1.2.3.4", ip)
}

func TestDetectPublicIP_JSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ip":"5.6.7.8"}`))
	}))
	defer server.Close()

	ip, err := detectPublicIP(server.URL)
	require.NoError(t, err)
	assert.Equal(t, "5.6.7.8", ip)
}

func TestDetectPublicIP_JSONWithExtraFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ip":"9.10.11.12","output":"json"}`))
	}))
	defer server.Close()

	ip, err := detectPublicIP(server.URL)
	require.NoError(t, err)
	assert.Equal(t, "9.10.11.12", ip)
}

func TestDetectPublicIP_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("   "))
	}))
	defer server.Close()

	ip, err := detectPublicIP(server.URL)
	assert.Empty(t, ip)
	assert.ErrorContains(t, err, "empty response")
}

func TestDetectPublicIP_ServerError(t *testing.T) {
	_, err := detectPublicIP("http://127.0.0.1:1") // nothing listening
	assert.Error(t, err)
}

// newIPProviderServer serves a fixed public IP for the factory to detect.
func newIPProviderServer(t *testing.T, ip string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, ip)
	}))
	t.Cleanup(server.Close)
	return server
}

// useNamecheapMockAPI points the factory at the mock Namecheap API for the test.
func useNamecheapMockAPI(t *testing.T, srv *namecheaptest.Server) {
	t.Helper()
	namecheapAPIBaseURL = srv.URL()
	t.Cleanup(func() { namecheapAPIBaseURL = "" })
}

func TestNamecheapFactory_CreateProvider_Success(t *testing.T) {
	srv := namecheaptest.NewServer(t)
	srv.StubFixture("namecheap.domains.getList", "domains_getList")
	useNamecheapMockAPI(t, srv)
	ipServer := newIPProviderServer(t, "203.0.113.7")

	factory := NewNamecheapFactory()
	cfg := &config.ProviderConfig{
		Type: "namecheap",
		Credentials: map[string]interface{}{
			"api_user": "some-user",
			"api_key":  "some-key",
			"username": "other-user",
		},
		Options: map[string]interface{}{
			"ip_provider_url": ipServer.URL,
		},
	}

	provider, err := factory.CreateProvider(cfg)

	require.NoError(t, err)
	assert.NotNil(t, provider)
	srv.AssertCalled(t, "namecheap.domains.getList", map[string]string{
		"ApiUser":  "some-user",
		"ApiKey":   "some-key",
		"Username": "other-user",
		"ClientIp": "203.0.113.7",
		"PageSize": "10",
	})
}

func TestNamecheapFactory_CreateProvider_AccessDenied(t *testing.T) {
	srv := namecheaptest.NewServer(t)
	srv.StubError("namecheap.domains.getList", 4011103, "Access denied")
	useNamecheapMockAPI(t, srv)
	ipServer := newIPProviderServer(t, "203.0.113.7")

	factory := NewNamecheapFactory()
	cfg := &config.ProviderConfig{
		Type: "namecheap",
		Credentials: map[string]interface{}{
			"api_user": "some-user",
			"api_key":  "some-key",
		},
		Options: map[string]interface{}{
			"ip_provider_url": ipServer.URL,
		},
	}

	provider, err := factory.CreateProvider(cfg)

	assert.Nil(t, provider)
	require.Error(t, err)
	var credsErr *ProviderCredentialsError
	require.ErrorAs(t, err, &credsErr)
	assert.Contains(t, credsErr.Error(), "203.0.113.7 is not whitelisted")
	assert.ErrorIs(t, err, namecheap.ErrAccessDenied)
}

func TestNamecheapFactory_CreateProvider_VerifyFails(t *testing.T) {
	srv := namecheaptest.NewServer(t)
	srv.StubError("namecheap.domains.getList", 2030166, "Domain is invalid")
	useNamecheapMockAPI(t, srv)
	ipServer := newIPProviderServer(t, "203.0.113.7")

	factory := NewNamecheapFactory()
	cfg := &config.ProviderConfig{
		Type: "namecheap",
		Credentials: map[string]interface{}{
			"api_user": "some-user",
			"api_key":  "some-key",
		},
		Options: map[string]interface{}{
			"ip_provider_url": ipServer.URL,
		},
	}

	provider, err := factory.CreateProvider(cfg)

	assert.Nil(t, provider)
	require.Error(t, err)
	assert.ErrorContains(t, err, "failed to verify credentials")
	assert.ErrorIs(t, err, namecheap.ErrDomainInvalid)
}

func TestDetectPublicIP_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	_, err := detectPublicIP(server.URL)

	assert.ErrorContains(t, err, "unexpected HTTP status")
}
