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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mixanemca/cdnscli/internal/models"
	namecheap "github.com/namecheap/go-namecheap-sdk/v2/namecheap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// namecheapXMLError returns an API error response envelope.
func namecheapXMLError(msg, number string) string {
	return `<?xml version="1.0" encoding="utf-8"?>
<ApiResponse Status="ERROR" xmlns="http://api.namecheap.com/xml.response">
  <Errors><Error Number="` + number + `">` + msg + `</Error></Errors>
  <CommandResponse />
</ApiResponse>`
}

// namecheapXMLSetHostsOK returns a successful SetHosts response.
func namecheapXMLSetHostsOK(domain string) string {
	return `<?xml version="1.0" encoding="utf-8"?>
<ApiResponse Status="OK" xmlns="http://api.namecheap.com/xml.response">
  <Errors />
  <CommandResponse>
    <DomainDNSSetHostsResult Domain="` + domain + `" IsSuccess="true" />
  </CommandResponse>
</ApiResponse>`
}

// multiHandler returns an HTTP handler that serves responses from the slice in order.
func multiHandler(responses []string) http.Handler {
	idx := 0
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if idx < len(responses) {
			_, _ = io.WriteString(w, responses[idx])
			idx++
		}
	})
}

// newTestNamecheapClient creates a Namecheap client pointed at the given mock server URL.
func newTestNamecheapClient(serverURL string) *namecheap.Client {
	c := namecheap.NewClient(&namecheap.ClientOptions{
		UserName: "testuser",
		ApiUser:  "testuser",
		ApiKey:   "testapikey",
		ClientIp: "127.0.0.1",
	})
	c.BaseURL = serverURL
	return c
}

// namecheapXMLResponse wraps a CommandResponse body in the standard ApiResponse envelope.
func namecheapXMLResponse(commandResponse string) string {
	return `<?xml version="1.0" encoding="utf-8"?>
<ApiResponse Status="OK" xmlns="http://api.namecheap.com/xml.response">
  <Errors />
  <Warnings />
  <CommandResponse>` + commandResponse + `</CommandResponse>
</ApiResponse>`
}

// --- relativeHostName ---

func TestRelativeHostName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		domain   string
		expected string
	}{
		{"root equals domain", "example.com", "example.com", "@"},
		{"root with trailing dot", "example.com.", "example.com", "@"},
		{"empty name", "", "example.com", "@"},
		{"subdomain stripped", "www.example.com", "example.com", "www"},
		{"deep subdomain stripped", "mail.sub.example.com", "example.com", "mail.sub"},
		{"already relative", "www", "example.com", "www"},
		{"at sign passthrough", "@", "example.com", "@"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, relativeHostName(tt.input, tt.domain))
		})
	}
}

// --- convFromNamecheapHost ---

func TestConvFromNamecheapHost(t *testing.T) {
	hostID := 12345
	name := "www"
	recType := "A"
	address := "1.2.3.4"
	ttl := 1800

	host := namecheap.DomainsDNSHostRecordDetailed{
		HostId:  &hostID,
		Name:    &name,
		Type:    &recType,
		Address: &address,
		TTL:     &ttl,
	}

	result := convFromNamecheapHost(host, "example.com")

	assert.Equal(t, "12345", result.ID)
	assert.Equal(t, "www", result.Name)
	assert.Equal(t, "A", result.Type)
	assert.Equal(t, "1.2.3.4", result.Content)
	assert.Equal(t, 1800, result.TTL)
	assert.False(t, result.Proxied)
}

func TestConvFromNamecheapHost_NilFields(t *testing.T) {
	host := namecheap.DomainsDNSHostRecordDetailed{}
	result := convFromNamecheapHost(host, "example.com")

	assert.Equal(t, "", result.ID)
	assert.Equal(t, "", result.Name)
	assert.Equal(t, "", result.Type)
	assert.Equal(t, "", result.Content)
	assert.Equal(t, 0, result.TTL)
}

// --- convFromNamecheapDomains ---

func TestConvFromNamecheapDomains(t *testing.T) {
	id1, name1 := "111", "example.com"
	id2, name2 := "222", "test.net"

	domains := []namecheap.Domain{
		{ID: &id1, Name: &name1},
		{ID: &id2, Name: &name2},
	}

	result := convFromNamecheapDomains(domains)

	require.Len(t, result, 2)
	assert.Equal(t, models.Zone{ID: "111", Name: "example.com"}, result[0])
	assert.Equal(t, models.Zone{ID: "222", Name: "test.net"}, result[1])
}

func TestConvFromNamecheapDomains_Empty(t *testing.T) {
	result := convFromNamecheapDomains([]namecheap.Domain{})
	assert.Empty(t, result)
}

// --- convParamsToNamecheapHost ---

func TestConvParamsToNamecheapHost(t *testing.T) {
	ttl := 3600
	params := models.CreateDNSRecordParams{
		ZoneName: "example.com",
		Name:     "mail.example.com",
		Type:     "A",
		Content:  "1.2.3.4",
		TTL:      ttl,
	}

	host := convParamsToNamecheapHost(params)

	require.NotNil(t, host.Name)
	require.NotNil(t, host.Type)
	require.NotNil(t, host.Address)
	require.NotNil(t, host.TTL)

	assert.Equal(t, "mail", *host.Name)
	assert.Equal(t, "A", *host.Type)
	assert.Equal(t, "1.2.3.4", *host.Address)
	assert.Equal(t, ttl, *host.TTL)
}

func TestConvParamsToNamecheapHost_Root(t *testing.T) {
	ttl := 1800
	params := models.CreateDNSRecordParams{
		ZoneName: "example.com",
		Name:     "example.com",
		Type:     "TXT",
		Content:  "v=spf1 ~all",
		TTL:      ttl,
	}

	host := convParamsToNamecheapHost(params)
	assert.Equal(t, "@", *host.Name)
}

// --- Repo: ListZones ---

func TestRepoNamecheap_ListZones(t *testing.T) {
	fakeResponse := namecheapXMLResponse(`
<DomainGetListResult>
  <Domain ID="101" Name="example.com" User="testuser" IsExpired="false" IsLocked="false" AutoRenew="true" WhoisGuard="ENABLED" IsPremium="false" IsOurDNS="true" />
  <Domain ID="102" Name="test.net" User="testuser" IsExpired="false" IsLocked="false" AutoRenew="false" WhoisGuard="NOTPRESENT" IsPremium="false" IsOurDNS="true" />
</DomainGetListResult>
<Paging><TotalItems>2</TotalItems><CurrentPage>1</CurrentPage><PageSize>20</PageSize></Paging>`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, fakeResponse)
	}))
	defer server.Close()

	repo := NewRepoNamecheap(newTestNamecheapClient(server.URL))
	zones, err := repo.ListZones(context.Background())

	require.NoError(t, err)
	require.Len(t, zones, 2)
	assert.Equal(t, "101", zones[0].ID)
	assert.Equal(t, "example.com", zones[0].Name)
	assert.Equal(t, "102", zones[1].ID)
	assert.Equal(t, "test.net", zones[1].Name)
}

func TestRepoNamecheap_ListZones_Empty(t *testing.T) {
	fakeResponse := namecheapXMLResponse(`
<DomainGetListResult />
<Paging><TotalItems>0</TotalItems><CurrentPage>1</CurrentPage><PageSize>20</PageSize></Paging>`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, fakeResponse)
	}))
	defer server.Close()

	repo := NewRepoNamecheap(newTestNamecheapClient(server.URL))
	zones, err := repo.ListZones(context.Background())

	require.NoError(t, err)
	assert.Empty(t, zones)
}

// --- Repo: ListDNSRecords ---

func TestRepoNamecheap_ListDNSRecords(t *testing.T) {
	fakeResponse := namecheapXMLResponse(`
<DomainDNSGetHostsResult Domain="example.com" EmailType="MX" IsUsingOurDNS="true">
  <host HostId="1001" Name="@" Type="A" Address="1.2.3.4" MXPref="10" TTL="1800" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
  <host HostId="1002" Name="www" Type="CNAME" Address="example.com" MXPref="10" TTL="3600" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
</DomainDNSGetHostsResult>`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, fakeResponse)
	}))
	defer server.Close()

	repo := NewRepoNamecheap(newTestNamecheapClient(server.URL))
	records, err := repo.ListDNSRecords(context.Background(), "example.com")

	require.NoError(t, err)
	require.Len(t, records, 2)

	assert.Equal(t, "1001", records[0].ID)
	assert.Equal(t, "@", records[0].Name)
	assert.Equal(t, "A", records[0].Type)
	assert.Equal(t, "1.2.3.4", records[0].Content)
	assert.Equal(t, 1800, records[0].TTL)
	assert.False(t, records[0].Proxied)

	assert.Equal(t, "1002", records[1].ID)
	assert.Equal(t, "www", records[1].Name)
	assert.Equal(t, "CNAME", records[1].Type)
}

// --- Repo: GetDNSRecord ---

func TestRepoNamecheap_GetDNSRecord_Found(t *testing.T) {
	fakeResponse := namecheapXMLResponse(`
<DomainDNSGetHostsResult Domain="example.com" EmailType="NONE" IsUsingOurDNS="true">
  <host HostId="2001" Name="sub" Type="TXT" Address="hello" MXPref="10" TTL="900" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
</DomainDNSGetHostsResult>`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, fakeResponse)
	}))
	defer server.Close()

	repo := NewRepoNamecheap(newTestNamecheapClient(server.URL))
	rec, err := repo.GetDNSRecord(context.Background(), "example.com", "2001")

	require.NoError(t, err)
	assert.Equal(t, "2001", rec.ID)
	assert.Equal(t, "sub", rec.Name)
	assert.Equal(t, "TXT", rec.Type)
	assert.Equal(t, "hello", rec.Content)
}

func TestRepoNamecheap_GetDNSRecord_NotFound(t *testing.T) {
	fakeResponse := namecheapXMLResponse(`
<DomainDNSGetHostsResult Domain="example.com" EmailType="NONE" IsUsingOurDNS="true">
  <host HostId="2001" Name="sub" Type="TXT" Address="hello" MXPref="10" TTL="900" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
</DomainDNSGetHostsResult>`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, fakeResponse)
	}))
	defer server.Close()

	repo := NewRepoNamecheap(newTestNamecheapClient(server.URL))
	_, err := repo.GetDNSRecord(context.Background(), "example.com", "9999")

	assert.ErrorContains(t, err, "not found")
}

// --- Repo: DeleteDNSRecord ---

func TestRepoNamecheap_DeleteDNSRecord(t *testing.T) {
	getHostsResponse := namecheapXMLResponse(`
<DomainDNSGetHostsResult Domain="example.com" EmailType="NONE" IsUsingOurDNS="true">
  <host HostId="3001" Name="keep" Type="A" Address="1.1.1.1" MXPref="10" TTL="1800" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
  <host HostId="3002" Name="delete-me" Type="A" Address="2.2.2.2" MXPref="10" TTL="1800" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
</DomainDNSGetHostsResult>`)

	setHostsResponse := namecheapXMLResponse(`
<DomainDNSSetHostsResult Domain="example.com" IsSuccess="true" />`)

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount == 1 {
			_, _ = io.WriteString(w, getHostsResponse)
		} else {
			body, _ := io.ReadAll(r.Body)
			// Verify deleted record is not in the setHosts request
			assert.NotContains(t, string(body), "delete-me")
			assert.Contains(t, string(body), "keep")
			_, _ = io.WriteString(w, setHostsResponse)
		}
	}))
	defer server.Close()

	repo := NewRepoNamecheap(newTestNamecheapClient(server.URL))
	err := repo.DeleteDNSRecord(context.Background(), "example.com", "3002")

	require.NoError(t, err)
	assert.Equal(t, 2, requestCount)
}

func TestRepoNamecheap_DeleteDNSRecord_NotFound(t *testing.T) {
	fakeResponse := namecheapXMLResponse(`
<DomainDNSGetHostsResult Domain="example.com" EmailType="NONE" IsUsingOurDNS="true">
  <host HostId="3001" Name="keep" Type="A" Address="1.1.1.1" MXPref="10" TTL="1800" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
</DomainDNSGetHostsResult>`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, fakeResponse)
	}))
	defer server.Close()

	repo := NewRepoNamecheap(newTestNamecheapClient(server.URL))
	err := repo.DeleteDNSRecord(context.Background(), "example.com", "9999")

	assert.ErrorContains(t, err, "not found")
}

// --- Repo: ZoneIDByName ---

func TestRepoNamecheap_ZoneIDByName_Found(t *testing.T) {
	fakeResponse := namecheapXMLResponse(`
<DomainGetListResult>
  <Domain ID="501" Name="example.com" User="testuser" IsExpired="false" IsLocked="false" AutoRenew="true" WhoisGuard="ENABLED" IsPremium="false" IsOurDNS="true" />
</DomainGetListResult>
<Paging><TotalItems>1</TotalItems><CurrentPage>1</CurrentPage><PageSize>20</PageSize></Paging>`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, fakeResponse)
	}))
	defer server.Close()

	repo := NewRepoNamecheap(newTestNamecheapClient(server.URL))
	id, err := repo.ZoneIDByName("example.com")

	require.NoError(t, err)
	assert.Equal(t, "501", id)
}

func TestRepoNamecheap_ZoneIDByName_NotFound(t *testing.T) {
	fakeResponse := namecheapXMLResponse(`
<DomainGetListResult />
<Paging><TotalItems>0</TotalItems><CurrentPage>1</CurrentPage><PageSize>20</PageSize></Paging>`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, fakeResponse)
	}))
	defer server.Close()

	repo := NewRepoNamecheap(newTestNamecheapClient(server.URL))
	_, err := repo.ZoneIDByName("missing.com")

	assert.ErrorContains(t, err, "not found")
}

// --- getHosts error paths ---

func TestRepoNamecheap_getHosts_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, namecheapXMLError("Domain is invalid", "2030166"))
	}))
	defer server.Close()

	repo := NewRepoNamecheap(newTestNamecheapClient(server.URL))
	_, err := repo.ListDNSRecords(context.Background(), "example.com")

	assert.Error(t, err)
}

func TestRepoNamecheap_getHosts_NilResult(t *testing.T) {
	// Response without CommandResponse → GetHosts returns nil, nil → triggers resp == nil
	noCommandResponse := `<?xml version="1.0" encoding="utf-8"?>
<ApiResponse Status="OK" xmlns="http://api.namecheap.com/xml.response">
  <Errors />
</ApiResponse>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, noCommandResponse)
	}))
	defer server.Close()

	repo := NewRepoNamecheap(newTestNamecheapClient(server.URL))
	_, err := repo.ListDNSRecords(context.Background(), "example.com")

	assert.ErrorContains(t, err, "empty response")
}

// --- ListDNSRecords: nil hosts ---

func TestRepoNamecheap_ListDNSRecords_NilHosts(t *testing.T) {
	// DomainDNSGetHostsResult with no <host> elements → Hosts == nil
	fakeResponse := namecheapXMLResponse(`
<DomainDNSGetHostsResult Domain="example.com" EmailType="NONE" IsUsingOurDNS="true" />`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, fakeResponse)
	}))
	defer server.Close()

	repo := NewRepoNamecheap(newTestNamecheapClient(server.URL))
	records, err := repo.ListDNSRecords(context.Background(), "example.com")

	require.NoError(t, err)
	assert.Empty(t, records)
}

// --- GetDNSRecord: error propagation ---

func TestRepoNamecheap_GetDNSRecord_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, namecheapXMLError("Domain is invalid", "2030166"))
	}))
	defer server.Close()

	repo := NewRepoNamecheap(newTestNamecheapClient(server.URL))
	_, err := repo.GetDNSRecord(context.Background(), "example.com", "1001")

	assert.Error(t, err)
}

// --- DeleteDNSRecord: error paths ---

func TestRepoNamecheap_DeleteDNSRecord_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, namecheapXMLError("Domain is invalid", "2030166"))
	}))
	defer server.Close()

	repo := NewRepoNamecheap(newTestNamecheapClient(server.URL))
	err := repo.DeleteDNSRecord(context.Background(), "example.com", "3001")

	assert.Error(t, err)
}

func TestRepoNamecheap_DeleteDNSRecord_NilHosts(t *testing.T) {
	fakeResponse := namecheapXMLResponse(`
<DomainDNSGetHostsResult Domain="example.com" EmailType="NONE" IsUsingOurDNS="true" />`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, fakeResponse)
	}))
	defer server.Close()

	repo := NewRepoNamecheap(newTestNamecheapClient(server.URL))
	err := repo.DeleteDNSRecord(context.Background(), "example.com", "3001")

	assert.ErrorContains(t, err, "not found")
}

// --- ListZones: error and nil paging ---

func TestRepoNamecheap_ListZones_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, namecheapXMLError("Unknown error", "9999"))
	}))
	defer server.Close()

	repo := NewRepoNamecheap(newTestNamecheapClient(server.URL))
	_, err := repo.ListZones(context.Background())

	assert.Error(t, err)
}

func TestRepoNamecheap_ListZones_NilPaging(t *testing.T) {
	// No <Paging> element → resp.Paging == nil → break immediately
	fakeResponse := namecheapXMLResponse(`
<DomainGetListResult>
  <Domain ID="201" Name="example.com" User="testuser" IsExpired="false" IsLocked="false" AutoRenew="true" WhoisGuard="ENABLED" IsPremium="false" IsOurDNS="true" />
</DomainGetListResult>`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, fakeResponse)
	}))
	defer server.Close()

	repo := NewRepoNamecheap(newTestNamecheapClient(server.URL))
	zones, err := repo.ListZones(context.Background())

	require.NoError(t, err)
	require.Len(t, zones, 1)
	assert.Equal(t, "example.com", zones[0].Name)
}

func TestRepoNamecheap_ListZones_WithFilter(t *testing.T) {
	fakeResponse := namecheapXMLResponse(`
<DomainGetListResult>
  <Domain ID="301" Name="example.com" User="testuser" IsExpired="false" IsLocked="false" AutoRenew="true" WhoisGuard="ENABLED" IsPremium="false" IsOurDNS="true" />
</DomainGetListResult>
<Paging><TotalItems>1</TotalItems><CurrentPage>1</CurrentPage><PageSize>100</PageSize></Paging>`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, fakeResponse)
	}))
	defer server.Close()

	repo := NewRepoNamecheap(newTestNamecheapClient(server.URL))
	zones, err := repo.ListZones(context.Background(), "example.com")

	require.NoError(t, err)
	require.Len(t, zones, 1)
	assert.Equal(t, "example.com", zones[0].Name)
}

// --- ZoneIDByName: error propagation ---

func TestRepoNamecheap_ZoneIDByName_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, namecheapXMLError("Unknown error", "9999"))
	}))
	defer server.Close()

	repo := NewRepoNamecheap(newTestNamecheapClient(server.URL))
	_, err := repo.ZoneIDByName("example.com")

	assert.Error(t, err)
}

// --- convParamsToNamecheapHost: ZoneID fallback ---

func TestConvParamsToNamecheapHost_ZoneIDFallback(t *testing.T) {
	ttl := 1800
	params := models.CreateDNSRecordParams{
		ZoneID:  "example.com",
		Name:    "www.example.com",
		Type:    "A",
		Content: "1.2.3.4",
		TTL:     ttl,
	}

	host := convParamsToNamecheapHost(params)

	require.NotNil(t, host.Name)
	assert.Equal(t, "www", *host.Name)
}

// --- CreateDNSRecord ---

func TestRepoNamecheap_CreateDNSRecord(t *testing.T) {
	existingHosts := namecheapXMLResponse(`
<DomainDNSGetHostsResult Domain="example.com" EmailType="NONE" IsUsingOurDNS="true">
  <host HostId="4001" Name="@" Type="A" Address="1.1.1.1" MXPref="10" TTL="1800" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
</DomainDNSGetHostsResult>`)

	afterCreateHosts := namecheapXMLResponse(`
<DomainDNSGetHostsResult Domain="example.com" EmailType="NONE" IsUsingOurDNS="true">
  <host HostId="4001" Name="@" Type="A" Address="1.1.1.1" MXPref="10" TTL="1800" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
  <host HostId="4002" Name="www" Type="A" Address="2.2.2.2" MXPref="10" TTL="3600" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
</DomainDNSGetHostsResult>`)

	server := httptest.NewServer(multiHandler([]string{
		existingHosts,
		namecheapXMLSetHostsOK("example.com"),
		afterCreateHosts,
	}))
	defer server.Close()

	repo := NewRepoNamecheap(newTestNamecheapClient(server.URL))
	rec, err := repo.CreateDNSRecord(context.Background(), models.CreateDNSRecordParams{
		ZoneName: "example.com",
		Name:     "www.example.com",
		Type:     "A",
		Content:  "2.2.2.2",
		TTL:      3600,
	})

	require.NoError(t, err)
	assert.Equal(t, "4002", rec.ID)
	assert.Equal(t, "www", rec.Name)
	assert.Equal(t, "A", rec.Type)
	assert.Equal(t, "2.2.2.2", rec.Content)
}

func TestRepoNamecheap_CreateDNSRecord_ZoneIDFallback(t *testing.T) {
	existingHosts := namecheapXMLResponse(`
<DomainDNSGetHostsResult Domain="example.com" EmailType="NONE" IsUsingOurDNS="true" />`)

	afterCreateHosts := namecheapXMLResponse(`
<DomainDNSGetHostsResult Domain="example.com" EmailType="NONE" IsUsingOurDNS="true">
  <host HostId="5001" Name="mail" Type="A" Address="3.3.3.3" MXPref="10" TTL="1800" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
</DomainDNSGetHostsResult>`)

	server := httptest.NewServer(multiHandler([]string{
		existingHosts,
		namecheapXMLSetHostsOK("example.com"),
		afterCreateHosts,
	}))
	defer server.Close()

	repo := NewRepoNamecheap(newTestNamecheapClient(server.URL))
	rec, err := repo.CreateDNSRecord(context.Background(), models.CreateDNSRecordParams{
		ZoneID:  "example.com",
		Name:    "mail",
		Type:    "A",
		Content: "3.3.3.3",
		TTL:     1800,
	})

	require.NoError(t, err)
	assert.Equal(t, "5001", rec.ID)
}

func TestRepoNamecheap_CreateDNSRecord_EmptyZone(t *testing.T) {
	repo := NewRepoNamecheap(newTestNamecheapClient("http://127.0.0.1:1"))
	_, err := repo.CreateDNSRecord(context.Background(), models.CreateDNSRecordParams{})

	assert.ErrorContains(t, err, "zone name or zone ID must be provided")
}

// --- UpdateDNSRecord ---

func TestRepoNamecheap_UpdateDNSRecord(t *testing.T) {
	existingHosts := namecheapXMLResponse(`
<DomainDNSGetHostsResult Domain="example.com" EmailType="NONE" IsUsingOurDNS="true">
  <host HostId="6001" Name="old" Type="A" Address="1.1.1.1" MXPref="10" TTL="1800" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
</DomainDNSGetHostsResult>`)

	afterUpdateHosts := namecheapXMLResponse(`
<DomainDNSGetHostsResult Domain="example.com" EmailType="NONE" IsUsingOurDNS="true">
  <host HostId="6001" Name="new" Type="A" Address="9.9.9.9" MXPref="10" TTL="300" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
</DomainDNSGetHostsResult>`)

	server := httptest.NewServer(multiHandler([]string{
		existingHosts,
		namecheapXMLSetHostsOK("example.com"),
		afterUpdateHosts,
	}))
	defer server.Close()

	repo := NewRepoNamecheap(newTestNamecheapClient(server.URL))
	rec, err := repo.UpdateDNSRecord(context.Background(), models.UpdateDNSRecordParams{
		ZoneName: "example.com",
		ID:       "6001",
		Name:     "new.example.com",
		Type:     "A",
		Content:  "9.9.9.9",
		TTL:      300,
	})

	require.NoError(t, err)
	assert.Equal(t, "6001", rec.ID)
	assert.Equal(t, "new", rec.Name)
	assert.Equal(t, "9.9.9.9", rec.Content)
}

func TestRepoNamecheap_UpdateDNSRecord_ZoneIDFallback(t *testing.T) {
	existingHosts := namecheapXMLResponse(`
<DomainDNSGetHostsResult Domain="example.com" EmailType="NONE" IsUsingOurDNS="true">
  <host HostId="7001" Name="sub" Type="TXT" Address="old-val" MXPref="10" TTL="900" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
</DomainDNSGetHostsResult>`)

	afterUpdateHosts := namecheapXMLResponse(`
<DomainDNSGetHostsResult Domain="example.com" EmailType="NONE" IsUsingOurDNS="true">
  <host HostId="7001" Name="sub" Type="TXT" Address="new-val" MXPref="10" TTL="900" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
</DomainDNSGetHostsResult>`)

	server := httptest.NewServer(multiHandler([]string{
		existingHosts,
		namecheapXMLSetHostsOK("example.com"),
		afterUpdateHosts,
	}))
	defer server.Close()

	repo := NewRepoNamecheap(newTestNamecheapClient(server.URL))
	_, err := repo.UpdateDNSRecord(context.Background(), models.UpdateDNSRecordParams{
		ZoneID:  "example.com",
		ID:      "7001",
		Name:    "sub",
		Type:    "TXT",
		Content: "new-val",
		TTL:     900,
	})

	require.NoError(t, err)
}

func TestRepoNamecheap_UpdateDNSRecord_EmptyZone(t *testing.T) {
	repo := NewRepoNamecheap(newTestNamecheapClient("http://127.0.0.1:1"))
	_, err := repo.UpdateDNSRecord(context.Background(), models.UpdateDNSRecordParams{})

	assert.ErrorContains(t, err, "zone name or zone ID must be provided")
}

func TestRepoNamecheap_UpdateDNSRecord_NotFound(t *testing.T) {
	fakeResponse := namecheapXMLResponse(`
<DomainDNSGetHostsResult Domain="example.com" EmailType="NONE" IsUsingOurDNS="true">
  <host HostId="8001" Name="sub" Type="A" Address="1.1.1.1" MXPref="10" TTL="1800" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
</DomainDNSGetHostsResult>`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, fakeResponse)
	}))
	defer server.Close()

	repo := NewRepoNamecheap(newTestNamecheapClient(server.URL))
	_, err := repo.UpdateDNSRecord(context.Background(), models.UpdateDNSRecordParams{
		ZoneName: "example.com",
		ID:       "9999",
		Name:     "sub",
		Type:     "A",
		Content:  "2.2.2.2",
		TTL:      1800,
	})

	assert.ErrorContains(t, err, "not found")
}

func TestRepoNamecheap_UpdateDNSRecord_NilHosts(t *testing.T) {
	fakeResponse := namecheapXMLResponse(`
<DomainDNSGetHostsResult Domain="example.com" EmailType="NONE" IsUsingOurDNS="true" />`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, fakeResponse)
	}))
	defer server.Close()

	repo := NewRepoNamecheap(newTestNamecheapClient(server.URL))
	_, err := repo.UpdateDNSRecord(context.Background(), models.UpdateDNSRecordParams{
		ZoneName: "example.com",
		ID:       "1001",
		Name:     "sub",
		Type:     "A",
		Content:  "1.1.1.1",
		TTL:      1800,
	})

	assert.ErrorContains(t, err, "not found")
}
