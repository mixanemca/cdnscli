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

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloudflareCredentials_CheckPrefixes(t *testing.T) {
	tests := []struct {
		name    string
		creds   CloudflareCredentials
		wantErr string
	}{
		{
			name:  "legacy token without prefix",
			creds: CloudflareCredentials{APIToken: "0123456789abcdefABCDEF0123456789abcdefAB"},
		},
		{
			name:  "user token",
			creds: CloudflareCredentials{APIToken: "cfut_0123456789abcdef0123456789abcdef01234567abcd"},
		},
		{
			name:  "account token",
			creds: CloudflareCredentials{APIToken: "cfat_0123456789abcdef0123456789abcdef01234567abcd"},
		},
		{
			name: "global key in api_key",
			creds: CloudflareCredentials{
				APIKey: "cfk_0123456789abcdef0123456789abcdef01234567abcd",
				Email:  "user@example.com",
			},
		},
		{
			name: "legacy global key in api_key",
			creds: CloudflareCredentials{
				APIKey: "0123456789abcdef0123456789abcdef01234",
				Email:  "user@example.com",
			},
		},
		{
			name:    "global key in api_token",
			creds:   CloudflareCredentials{APIToken: "cfk_0123456789abcdef0123456789abcdef01234567abcd"},
			wantErr: "api_token looks like a Global API Key (cfk_ prefix)",
		},
		{
			name: "user token in api_key",
			creds: CloudflareCredentials{
				APIKey: "cfut_0123456789abcdef0123456789abcdef01234567abcd",
				Email:  "user@example.com",
			},
			wantErr: "api_key looks like an API token (cfut_/cfat_ prefix)",
		},
		{
			name: "account token in api_key",
			creds: CloudflareCredentials{
				APIKey: "cfat_0123456789abcdef0123456789abcdef01234567abcd",
				Email:  "user@example.com",
			},
			wantErr: "api_key looks like an API token (cfut_/cfat_ prefix)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.creds.CheckPrefixes()
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestProviderConfig_Validate_Cloudflare(t *testing.T) {
	tests := []struct {
		name    string
		creds   map[string]interface{}
		wantErr string
	}{
		{
			name:  "api_token only",
			creds: map[string]interface{}{"api_token": "cfat_0123456789abcdef0123456789abcdef01234567abcd"},
		},
		{
			name: "api_key and email",
			creds: map[string]interface{}{
				"api_key": "cfk_0123456789abcdef0123456789abcdef01234567abcd",
				"email":   "user@example.com",
			},
		},
		{
			name:    "no credentials",
			creds:   map[string]interface{}{},
			wantErr: "must have either api_token or (api_key + email)",
		},
		{
			name:    "api_key without email",
			creds:   map[string]interface{}{"api_key": "cfk_0123456789abcdef0123456789abcdef01234567abcd"},
			wantErr: "must have either api_token or (api_key + email)",
		},
		{
			name: "both token and key",
			creds: map[string]interface{}{
				"api_token": "cfut_0123456789abcdef0123456789abcdef01234567abcd",
				"api_key":   "cfk_0123456789abcdef0123456789abcdef01234567abcd",
				"email":     "user@example.com",
			},
			wantErr: "cannot have both api_token and (api_key + email)",
		},
		{
			name:    "global key placed into api_token",
			creds:   map[string]interface{}{"api_token": "cfk_0123456789abcdef0123456789abcdef01234567abcd"},
			wantErr: "api_token looks like a Global API Key",
		},
		{
			name: "token placed into api_key",
			creds: map[string]interface{}{
				"api_key": "cfut_0123456789abcdef0123456789abcdef01234567abcd",
				"email":   "user@example.com",
			},
			wantErr: "api_key looks like an API token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := &ProviderConfig{Type: "cloudflare", Credentials: tt.creds}
			err := pc.Validate("cf")
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), `provider "cf" validation failed`)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}
