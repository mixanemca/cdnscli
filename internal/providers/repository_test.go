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
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	"github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/option"
	"github.com/mixanemca/cdnscli/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeCloudflareAPI is a minimal in-memory stand-in for the Cloudflare v4 API,
// covering only the endpoints the repository uses.
type fakeCloudflareAPI struct {
	mu sync.Mutex
	// zones served by GET /zones, filtered by the name query parameter.
	zones []map[string]any
	// records served by GET /zones/{zone}/dns_records, one slice per page.
	recordPages [][]map[string]any
	// record served by GET /zones/{zone}/dns_records/{record}.
	record map[string]any
	// requests received, in order.
	requests []recordedRequest
	// unauthorized makes every endpoint answer with an authentication error.
	unauthorized bool
}

type recordedRequest struct {
	Method string
	Path   string
	Query  map[string]string
	Header http.Header
	Body   map[string]any
}

func (f *fakeCloudflareAPI) server(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /zones", f.handleListZones)
	mux.HandleFunc("GET /zones/{zone}/dns_records", f.handleListRecords)
	mux.HandleFunc("GET /zones/{zone}/dns_records/{record}", f.handleGetRecord)
	mux.HandleFunc("POST /zones/{zone}/dns_records", f.handleEchoRecord)
	mux.HandleFunc("PUT /zones/{zone}/dns_records/{record}", f.handleEchoRecord)
	mux.HandleFunc("DELETE /zones/{zone}/dns_records/{record}", f.handleDeleteRecord)

	srv := httptest.NewServer(f.capture(mux))
	t.Cleanup(srv.Close)

	return srv
}

// capture wraps the mux to record every request before it is handled.
func (f *fakeCloudflareAPI) capture(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := recordedRequest{
			Method: r.Method,
			Path:   r.URL.Path,
			Query:  map[string]string{},
			Header: r.Header.Clone(),
		}
		for k := range r.URL.Query() {
			req.Query[k] = r.URL.Query().Get(k)
		}
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&req.Body)
		}

		f.mu.Lock()
		f.requests = append(f.requests, req)
		unauthorized := f.unauthorized
		f.mu.Unlock()

		if unauthorized {
			writeEnvelope(w, http.StatusForbidden, false, nil, nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (f *fakeCloudflareAPI) lastRequest() recordedRequest {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.requests[len(f.requests)-1]
}

func (f *fakeCloudflareAPI) setUnauthorized(v bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.unauthorized = v
}

func (f *fakeCloudflareAPI) handleListZones(w http.ResponseWriter, r *http.Request) {
	if pageOf(r) > 1 {
		writeEnvelope(w, http.StatusOK, true, []map[string]any{}, nil)
		return
	}

	name := r.URL.Query().Get("name")
	result := make([]map[string]any, 0, len(f.zones))
	for _, z := range f.zones {
		if name == "" || z["name"] == name {
			result = append(result, z)
		}
	}

	writeEnvelope(w, http.StatusOK, true, result, nil)
}

func (f *fakeCloudflareAPI) handleListRecords(w http.ResponseWriter, r *http.Request) {
	page := pageOf(r)
	if page > len(f.recordPages) {
		writeEnvelope(w, http.StatusOK, true, []map[string]any{}, nil)
		return
	}

	writeEnvelope(w, http.StatusOK, true, f.recordPages[page-1], nil)
}

func (f *fakeCloudflareAPI) handleGetRecord(w http.ResponseWriter, _ *http.Request) {
	writeEnvelope(w, http.StatusOK, true, f.record, nil)
}

// handleEchoRecord answers create/update calls with the submitted record
// (already decoded by capture), assigning it an identifier.
func (f *fakeCloudflareAPI) handleEchoRecord(w http.ResponseWriter, r *http.Request) {
	result := map[string]any{"id": "echo-id"}
	if id := r.PathValue("record"); id != "" {
		result["id"] = id
	}
	for k, v := range f.lastRequest().Body {
		result[k] = v
	}

	writeEnvelope(w, http.StatusOK, true, result, nil)
}

func (f *fakeCloudflareAPI) handleDeleteRecord(w http.ResponseWriter, r *http.Request) {
	writeEnvelope(w, http.StatusOK, true, map[string]any{"id": r.PathValue("record")}, nil)
}

func pageOf(r *http.Request) int {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		return 1
	}

	return page
}

func writeEnvelope(
	w http.ResponseWriter, status int, success bool, result any, errs []map[string]any,
) {
	if errs == nil {
		errs = []map[string]any{}
	}
	if !success {
		errs = []map[string]any{{"code": 10000, "message": "Authentication error"}}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success":  success,
		"errors":   errs,
		"messages": []any{},
		"result":   result,
	})
}

func newTestRepo(t *testing.T, fake *fakeCloudflareAPI) Repo {
	t.Helper()

	srv := fake.server(t)
	client := cloudflare.NewClient(
		option.WithAPIToken("test-token"),
		option.WithBaseURL(srv.URL),
		option.WithMaxRetries(0),
	)

	return NewRepoCloudFlare(client)
}

func TestRepoCloudFlare_ListZones(t *testing.T) {
	fake := &fakeCloudflareAPI{
		zones: []map[string]any{
			{
				"id":           "zone-1",
				"name":         "example.com",
				"name_servers": []string{"ns1.example.com", "ns2.example.com"},
				"status":       "active",
			},
			{
				"id":           "zone-2",
				"name":         "example.org",
				"name_servers": []string{"ns1.example.org"},
				"status":       "pending",
			},
		},
	}
	repo := newTestRepo(t, fake)

	t.Run("all zones", func(t *testing.T) {
		result, err := repo.ListZones(context.Background())
		require.NoError(t, err)
		assert.Equal(t, []models.Zone{
			{
				ID:          "zone-1",
				Name:        "example.com",
				NameServers: []string{"ns1.example.com", "ns2.example.com"},
				Status:      "active",
			},
			{
				ID:          "zone-2",
				Name:        "example.org",
				NameServers: []string{"ns1.example.org"},
				Status:      "pending",
			},
		}, result)
		assert.Equal(t, "", fake.lastRequest().Query["name"])
	})

	t.Run("filtered by name", func(t *testing.T) {
		result, err := repo.ListZones(context.Background(), "example.org")
		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "zone-2", result[0].ID)
		assert.Equal(t, "example.org", fake.lastRequest().Query["name"])
	})

	t.Run("punycode name is normalized", func(t *testing.T) {
		_, err := repo.ListZones(context.Background(), "xn--e1afmkfd.xn--p1ai")
		require.NoError(t, err)
		assert.Equal(t, "пример.рф", fake.lastRequest().Query["name"])
	})
}

func TestRepoCloudFlare_ZoneIDByName(t *testing.T) {
	fake := &fakeCloudflareAPI{
		zones: []map[string]any{
			{"id": "zone-1", "name": "example.com", "status": "active"},
			{"id": "zone-dup-a", "name": "dup.example", "status": "active"},
			{"id": "zone-dup-b", "name": "dup.example", "status": "active"},
		},
	}
	repo := newTestRepo(t, fake)

	t.Run("found", func(t *testing.T) {
		id, err := repo.ZoneIDByName("example.com")
		require.NoError(t, err)
		assert.Equal(t, "zone-1", id)
	})

	t.Run("not found", func(t *testing.T) {
		id, err := repo.ZoneIDByName("missing.example")
		assert.EqualError(t, err, "zone could not be found")
		assert.Equal(t, "", id)
	})

	t.Run("ambiguous", func(t *testing.T) {
		id, err := repo.ZoneIDByName("dup.example")
		assert.EqualError(t, err, "ambiguous zone name; an account ID might help")
		assert.Equal(t, "", id)
	})

	t.Run("api error", func(t *testing.T) {
		fake.setUnauthorized(true)
		defer fake.setUnauthorized(false)

		id, err := repo.ZoneIDByName("example.com")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "zones list command failed")
		assert.Equal(t, "", id)
	})
}

func TestRepoCloudFlare_ListDNSRecords(t *testing.T) {
	fake := &fakeCloudflareAPI{
		recordPages: [][]map[string]any{
			{
				{
					"id": "rec-1", "name": "a.example.com", "type": "A",
					"content": "192.0.2.1", "ttl": 3600, "proxied": true,
				},
				{
					"id": "rec-2", "name": "b.example.com", "type": "CNAME",
					"content": "a.example.com", "ttl": 1, "proxied": false,
				},
			},
			{
				{
					"id": "rec-3", "name": "example.com", "type": "MX",
					"content": "mail.example.com", "ttl": 300, "priority": 10,
				},
			},
		},
	}
	repo := newTestRepo(t, fake)

	result, err := repo.ListDNSRecords(context.Background(), "zone-1")
	require.NoError(t, err)
	assert.Equal(t, []models.DNSRecord{
		{
			ID: "rec-1", Name: "a.example.com", Type: "A",
			Content: "192.0.2.1", TTL: 3600, Proxied: true,
		},
		{ID: "rec-2", Name: "b.example.com", Type: "CNAME", Content: "a.example.com", TTL: 1},
		{ID: "rec-3", Name: "example.com", Type: "MX", Content: "mail.example.com", TTL: 300},
	}, result)
	assert.Equal(t, "/zones/zone-1/dns_records", fake.lastRequest().Path)
}

func TestRepoCloudFlare_ListDNSRecords_Error(t *testing.T) {
	fake := &fakeCloudflareAPI{unauthorized: true}
	repo := newTestRepo(t, fake)

	result, err := repo.ListDNSRecords(context.Background(), "zone-1")
	require.Error(t, err)
	assert.Equal(t, []models.DNSRecord{}, result)
}

func TestRepoCloudFlare_GetDNSRecord(t *testing.T) {
	fake := &fakeCloudflareAPI{
		record: map[string]any{
			"id": "rec-1", "name": "a.example.com", "type": "A",
			"content": "192.0.2.1", "ttl": 3600, "proxied": true,
		},
	}
	repo := newTestRepo(t, fake)

	result, err := repo.GetDNSRecord(context.Background(), "zone-1", "rec-1")
	require.NoError(t, err)
	assert.Equal(t, models.DNSRecord{
		ID: "rec-1", Name: "a.example.com", Type: "A",
		Content: "192.0.2.1", TTL: 3600, Proxied: true,
	}, result)
	assert.Equal(t, "/zones/zone-1/dns_records/rec-1", fake.lastRequest().Path)

	fake.setUnauthorized(true)
	result, err = repo.GetDNSRecord(context.Background(), "zone-1", "rec-1")
	require.Error(t, err)
	assert.Equal(t, models.DNSRecord{}, result)
}

func TestRepoCloudFlare_CreateDNSRecord(t *testing.T) {
	fake := &fakeCloudflareAPI{}
	repo := newTestRepo(t, fake)

	result, err := repo.CreateDNSRecord(context.Background(), models.CreateDNSRecordParams{
		Content: "192.0.2.1",
		Name:    "a.example.com",
		Proxied: true,
		TTL:     3600,
		Type:    "A",
		ZoneID:  "zone-1",
	})
	require.NoError(t, err)
	assert.Equal(t, models.DNSRecord{
		ID: "echo-id", Name: "a.example.com", Type: "A",
		Content: "192.0.2.1", TTL: 3600, Proxied: true,
	}, result)

	req := fake.lastRequest()
	assert.Equal(t, http.MethodPost, req.Method)
	assert.Equal(t, "/zones/zone-1/dns_records", req.Path)
	assert.Equal(t, map[string]any{
		"content": "192.0.2.1",
		"name":    "a.example.com",
		"proxied": true,
		"ttl":     float64(3600),
		"type":    "A",
	}, req.Body)
}

func TestRepoCloudFlare_CreateDNSRecord_ZeroTTLOmitted(t *testing.T) {
	fake := &fakeCloudflareAPI{}
	repo := newTestRepo(t, fake)

	_, err := repo.CreateDNSRecord(context.Background(), models.CreateDNSRecordParams{
		Content: "192.0.2.1",
		Name:    "a.example.com",
		Type:    "A",
		ZoneID:  "zone-1",
	})
	require.NoError(t, err)

	_, hasTTL := fake.lastRequest().Body["ttl"]
	assert.False(t, hasTTL, "zero TTL must not be sent")
}

func TestRepoCloudFlare_UpdateDNSRecord(t *testing.T) {
	fake := &fakeCloudflareAPI{}
	repo := newTestRepo(t, fake)

	result, err := repo.UpdateDNSRecord(context.Background(), models.UpdateDNSRecordParams{
		Content: "192.0.2.2",
		ID:      "rec-1",
		Name:    "a.example.com",
		Proxied: false,
		TTL:     300,
		Type:    "A",
		ZoneID:  "zone-1",
	})
	require.NoError(t, err)
	assert.Equal(t, models.DNSRecord{
		ID: "rec-1", Name: "a.example.com", Type: "A", Content: "192.0.2.2", TTL: 300,
	}, result)

	req := fake.lastRequest()
	assert.Equal(t, http.MethodPut, req.Method)
	assert.Equal(t, "/zones/zone-1/dns_records/rec-1", req.Path)
	assert.Equal(t, map[string]any{
		"content": "192.0.2.2",
		"name":    "a.example.com",
		"proxied": false,
		"ttl":     float64(300),
		"type":    "A",
	}, req.Body)
}

func TestRepoCloudFlare_DeleteDNSRecord(t *testing.T) {
	fake := &fakeCloudflareAPI{}
	repo := newTestRepo(t, fake)

	err := repo.DeleteDNSRecord(context.Background(), "zone-1", "rec-1")
	require.NoError(t, err)

	req := fake.lastRequest()
	assert.Equal(t, http.MethodDelete, req.Method)
	assert.Equal(t, "/zones/zone-1/dns_records/rec-1", req.Path)

	fake.setUnauthorized(true)
	err = repo.DeleteDNSRecord(context.Background(), "zone-1", "rec-1")
	require.Error(t, err)
}
