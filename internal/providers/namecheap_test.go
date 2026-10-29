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
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/mixanemca/cdnscli/internal/models"
	namecheap "github.com/namecheap/go-namecheap-sdk/v2/namecheap"
	"github.com/namecheap/go-namecheap-sdk/v2/namecheaptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Namecheap API commands routed by the namecheaptest mock server.
const (
	ncCmdGetHosts    = "namecheap.domains.dns.getHosts"
	ncCmdSetHosts    = "namecheap.domains.dns.setHosts"
	ncCmdDomainsList = "namecheap.domains.getList"
	ncCmdDNSList     = "namecheap.domains.dns.getList"
)

// ncHost describes a host record the way Namecheap reports it. A zero mxPref is
// rendered as MXPref="10", which the real API echoes for every record type.
type ncHost struct {
	id      int
	name    string
	typ     string
	address string
	ttl     int
	mxPref  int
}

func ncHostXML(h ncHost) string {
	mxPref := h.mxPref
	if mxPref == 0 {
		mxPref = 10
	}
	return fmt.Sprintf(`<host HostId="%d" Name="%s" Type="%s" Address="%s" MXPref="%d" TTL="%d" `+
		`AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />`,
		h.id, h.name, h.typ, h.address, mxPref, h.ttl)
}

// ncEnvelope wraps a CommandResponse body in the standard ApiResponse envelope.
func ncEnvelope(commandResponse string) string {
	return `<?xml version="1.0" encoding="utf-8"?>
<ApiResponse Status="OK" xmlns="http://api.namecheap.com/xml.response">
  <Errors />
  <Warnings />
  <CommandResponse>` + commandResponse + `</CommandResponse>
</ApiResponse>`
}

// ncGetHostsXML renders a domains.dns.getHosts response.
func ncGetHostsXML(domain, emailType string, hosts ...ncHost) string {
	var b strings.Builder
	for _, h := range hosts {
		b.WriteString("\n    " + ncHostXML(h))
	}
	return ncEnvelope(fmt.Sprintf(`
  <DomainDNSGetHostsResult Domain="%s" EmailType="%s" IsUsingOurDNS="true">%s
  </DomainDNSGetHostsResult>`, domain, emailType, b.String()))
}

func ncDomainXML(id int, name string) string {
	return fmt.Sprintf(`<Domain ID="%d" Name="%s" User="testuser" IsExpired="false" IsLocked="false" `+
		`AutoRenew="true" WhoisGuard="ENABLED" IsPremium="false" IsOurDNS="true" />`, id, name)
}

// ncDomainsListXML renders a domains.getList page; total goes to Paging.TotalItems.
func ncDomainsListXML(total int, names ...string) string {
	var b strings.Builder
	for i, name := range names {
		b.WriteString("\n    " + ncDomainXML(100+i, name))
	}
	return ncEnvelope(fmt.Sprintf(`
  <DomainGetListResult>%s
  </DomainGetListResult>
  <Paging><TotalItems>%d</TotalItems><CurrentPage>1</CurrentPage><PageSize>100</PageSize></Paging>`,
		b.String(), total))
}

// ncDomainsListNoPagingXML renders a domains.getList page without a Paging element.
func ncDomainsListNoPagingXML(names ...string) string {
	var b strings.Builder
	for i, name := range names {
		b.WriteString("\n    " + ncDomainXML(200+i, name))
	}
	return ncEnvelope(fmt.Sprintf(`
  <DomainGetListResult>%s
  </DomainGetListResult>`, b.String()))
}

// ncDNSListXML renders a domains.dns.getList response with the given nameservers.
func ncDNSListXML(domain string, ns ...string) string {
	var b strings.Builder
	for _, n := range ns {
		b.WriteString("\n    <Nameserver>" + n + "</Nameserver>")
	}
	return ncEnvelope(fmt.Sprintf(`
  <DomainDNSGetListResult Domain="%s" IsUsingOurDNS="true" IsPremiumDNS="false" IsUsingFreeDNS="false">%s
  </DomainDNSGetListResult>`, domain, b.String()))
}

// newNamecheapTestRepo starts a mock Namecheap API and returns a repository
// wired to it.
func newNamecheapTestRepo(t *testing.T) (*namecheaptest.Server, Repo) {
	t.Helper()
	srv := namecheaptest.NewServer(t)
	srv.Stub(ncCmdSetHosts, namecheaptest.FixtureOK("domains_dns_setHosts"))
	return srv, NewRepoNamecheap(srv.Client())
}

// ncSetHostsValues returns every parameter value of the single setHosts call.
func ncSetHostsValues(t *testing.T, srv *namecheaptest.Server) []string {
	t.Helper()
	calls := srv.Calls(ncCmdSetHosts)
	require.Len(t, calls, 1)
	values := make([]string, 0, len(calls[0]))
	for _, v := range calls[0] {
		values = append(values, v)
	}
	return values
}

// --- relativeHostName ---

func TestRelativeHostName(t *testing.T) {
	tests := []struct {
		name   string
		domain string
		want   string
	}{
		{"www.example.com", "example.com", "www"},
		{"example.com", "example.com", "@"},
		{"", "example.com", "@"},
		{"www", "example.com", "www"},
		{"sub.www.example.com", "example.com", "sub.www"},
		{"www.example.com.", "example.com", "www"},
		{"other.org", "example.com", "other.org"},
	}
	for _, tt := range tests {
		t.Run(tt.name+"/"+tt.domain, func(t *testing.T) {
			assert.Equal(t, tt.want, relativeHostName(tt.name, tt.domain))
		})
	}
}

// --- conversions ---

func TestConvFromNamecheapHost(t *testing.T) {
	h := namecheap.DomainsDNSHostRecordDetailed{
		HostId:  namecheap.Int(1001),
		Name:    namecheap.String("www"),
		Type:    namecheap.String("A"),
		Address: namecheap.String("1.2.3.4"),
		TTL:     namecheap.Int(1800),
	}

	rec := convFromNamecheapHost(h, "example.com")

	assert.Equal(t, "1001", rec.ID)
	assert.Equal(t, "www", rec.Name)
	assert.Equal(t, "A", rec.Type)
	assert.Equal(t, "1.2.3.4", rec.Content)
	assert.Equal(t, 1800, rec.TTL)
	assert.False(t, rec.Proxied)
}

func TestConvFromNamecheapHost_NilFields(t *testing.T) {
	rec := convFromNamecheapHost(namecheap.DomainsDNSHostRecordDetailed{}, "example.com")

	assert.Empty(t, rec.ID)
	assert.Empty(t, rec.Name)
	assert.Empty(t, rec.Type)
	assert.Empty(t, rec.Content)
	assert.Zero(t, rec.TTL)
}

func TestConvFromNamecheapDomains(t *testing.T) {
	domains := []*namecheap.Domain{
		{ID: namecheap.String("1"), Name: namecheap.String("example.com")},
		nil,
		{ID: namecheap.String("2"), Name: namecheap.String("test.net")},
	}

	zones := convFromNamecheapDomains(domains)

	require.Len(t, zones, 2)
	assert.Equal(t, "example.com", zones[0].ID)
	assert.Equal(t, "example.com", zones[0].Name)
	assert.Equal(t, "test.net", zones[1].ID)
}

func TestConvFromNamecheapDomains_Empty(t *testing.T) {
	assert.Empty(t, convFromNamecheapDomains(nil))
}

func TestConvParamsToNamecheapRecord(t *testing.T) {
	rec := convParamsToNamecheapRecord(models.CreateDNSRecordParams{
		ZoneName: "example.com",
		Name:     "www.example.com",
		Type:     "A",
		Content:  "1.2.3.4",
		TTL:      3600,
	})

	require.NotNil(t, rec.HostName)
	assert.Equal(t, "www", *rec.HostName)
	assert.Equal(t, "A", *rec.RecordType)
	assert.Equal(t, "1.2.3.4", *rec.Address)
	assert.Equal(t, 3600, *rec.TTL)
	require.NotNil(t, rec.MXPref)
	assert.Equal(t, namecheapDefaultMXPref, *rec.MXPref)
}

func TestConvParamsToNamecheapRecord_Root(t *testing.T) {
	rec := convParamsToNamecheapRecord(models.CreateDNSRecordParams{
		ZoneName: "example.com",
		Name:     "example.com",
		Type:     "TXT",
		Content:  "v=spf1 -all",
		TTL:      1800,
	})

	assert.Equal(t, "@", *rec.HostName)
}

func TestConvParamsToNamecheapRecord_ZoneIDFallback(t *testing.T) {
	rec := convParamsToNamecheapRecord(models.CreateDNSRecordParams{
		ZoneID:  "example.com",
		Name:    "www.example.com",
		Type:    "A",
		Content: "1.2.3.4",
		TTL:     1800,
	})

	assert.Equal(t, "www", *rec.HostName)
}

func TestConvUpdateParamsToNamecheapRecord_KeepsMXPref(t *testing.T) {
	old := namecheap.DomainsDNSHostRecordDetailed{
		Name:    namecheap.String("@"),
		Type:    namecheap.String("MX"),
		Address: namecheap.String("mail1.example.com"),
		MXPref:  namecheap.Int(20),
		TTL:     namecheap.Int(1800),
	}

	rec := convUpdateParamsToNamecheapRecord(models.UpdateDNSRecordParams{
		ZoneID:  "example.com",
		Name:    "example.com",
		Type:    "MX",
		Content: "mail2.example.com",
		TTL:     900,
	}, old)

	assert.Equal(t, "@", *rec.HostName)
	assert.Equal(t, "mail2.example.com", *rec.Address)
	assert.Equal(t, 900, *rec.TTL)
	require.NotNil(t, rec.MXPref)
	assert.Equal(t, uint8(20), *rec.MXPref)
}

func TestSelectorForHost(t *testing.T) {
	sel := selectorForHost(namecheap.DomainsDNSHostRecordDetailed{
		Name:    namecheap.String("www"),
		Type:    namecheap.String("A"),
		Address: namecheap.String("1.2.3.4"),
		MXPref:  namecheap.Int(10),
	})

	assert.Equal(t, "www", *sel.HostName)
	assert.Equal(t, "A", *sel.RecordType)
	assert.Equal(t, "1.2.3.4", *sel.Address)
	require.NotNil(t, sel.MXPref)
	assert.Equal(t, uint8(10), *sel.MXPref)
	assert.False(t, sel.MatchAll)
}

func TestSameNamecheapRecord(t *testing.T) {
	a := namecheap.DomainsDNSHostRecord{
		HostName:   namecheap.String("WWW"),
		RecordType: namecheap.String("a"),
		Address:    namecheap.String("host.example.com."),
		TTL:        namecheap.Int(300),
	}
	b := namecheap.DomainsDNSHostRecord{
		HostName:   namecheap.String("www"),
		RecordType: namecheap.String("A"),
		Address:    namecheap.String("host.example.com"),
		TTL:        namecheap.Int(1800),
		MXPref:     namecheap.UInt8(10),
	}
	c := namecheap.DomainsDNSHostRecord{
		HostName:   namecheap.String("www"),
		RecordType: namecheap.String("A"),
		Address:    namecheap.String("other.example.com"),
	}

	assert.True(t, sameNamecheapRecord(a, b))
	assert.False(t, sameNamecheapRecord(a, c))
}

func TestIsMailRecordType(t *testing.T) {
	assert.True(t, isMailRecordType("MX"))
	assert.True(t, isMailRecordType("A", " mxe "))
	assert.False(t, isMailRecordType("A", "CNAME", "TXT"))
	assert.False(t, isMailRecordType())
}

func TestWrapNamecheapMutationError(t *testing.T) {
	assert.NoError(t, wrapNamecheapMutationError("example.com", nil))

	conflict := wrapNamecheapMutationError("example.com", namecheap.ErrConcurrentModification)
	assert.ErrorIs(t, conflict, namecheap.ErrConcurrentModification)
	assert.ErrorContains(t, conflict, "changed concurrently")

	other := wrapNamecheapMutationError("example.com", errors.New("boom"))
	assert.ErrorContains(t, other, "example.com")
	assert.ErrorContains(t, other, "boom")
}

// --- ListZones ---

func TestRepoNamecheap_ListZones(t *testing.T) {
	srv, repo := newNamecheapTestRepo(t)
	srv.Stub(ncCmdDomainsList, ncDomainsListXML(2, "example.com", "test.net"))
	srv.Stub(ncCmdDNSList, ncDNSListXML("example.com", "dns1.namecheaphosting.com", "dns2.namecheaphosting.com"))

	zones, err := repo.ListZones(context.Background())

	require.NoError(t, err)
	require.Len(t, zones, 2)
	assert.Equal(t, "example.com", zones[0].ID)
	assert.Equal(t, "example.com", zones[0].Name)
	assert.Equal(t, "test.net", zones[1].ID)
	assert.Equal(t, []string{"dns1.namecheaphosting.com", "dns2.namecheaphosting.com"}, zones[0].NameServers)
	srv.AssertCalled(t, ncCmdDomainsList, map[string]string{"PageSize": "100", "Page": "1"})
}

func TestRepoNamecheap_ListZones_Empty(t *testing.T) {
	srv, repo := newNamecheapTestRepo(t)
	srv.Stub(ncCmdDomainsList, ncDomainsListXML(0))

	zones, err := repo.ListZones(context.Background())

	require.NoError(t, err)
	assert.Empty(t, zones)
}

func TestRepoNamecheap_ListZones_Paging(t *testing.T) {
	page1 := make([]string, 0, namecheapListZonesPageSize)
	for i := range namecheapListZonesPageSize {
		page1 = append(page1, fmt.Sprintf("zone%03d.com", i))
	}
	total := namecheapListZonesPageSize + 1

	srv, repo := newNamecheapTestRepo(t)
	srv.StubSequence(ncCmdDomainsList,
		ncDomainsListXML(total, page1...),
		ncDomainsListXML(total, "last.com"))
	srv.StubFixture(ncCmdDNSList, "domains_dns_getList")

	zones, err := repo.ListZones(context.Background())

	require.NoError(t, err)
	assert.Len(t, zones, total)
	assert.Equal(t, "last.com", zones[total-1].Name)

	calls := srv.Calls(ncCmdDomainsList)
	require.Len(t, calls, 2)
	assert.Equal(t, "2", calls[1]["Page"])
}

func TestRepoNamecheap_ListZones_NilPaging(t *testing.T) {
	srv, repo := newNamecheapTestRepo(t)
	srv.StubSequence(ncCmdDomainsList,
		ncDomainsListNoPagingXML("example.com"),
		ncDomainsListNoPagingXML())
	srv.Stub(ncCmdDNSList, ncDNSListXML("example.com"))

	zones, err := repo.ListZones(context.Background())

	require.NoError(t, err)
	require.Len(t, zones, 1)
	assert.Equal(t, "example.com", zones[0].Name)
}

func TestRepoNamecheap_ListZones_WithFilter(t *testing.T) {
	srv, repo := newNamecheapTestRepo(t)
	srv.Stub(ncCmdDomainsList, ncDomainsListXML(1, "example.com"))
	srv.Stub(ncCmdDNSList, ncDNSListXML("example.com"))

	zones, err := repo.ListZones(context.Background(), "example.com")

	require.NoError(t, err)
	require.Len(t, zones, 1)
	srv.AssertCalled(t, ncCmdDomainsList, map[string]string{"SearchTerm": "example.com"})
}

func TestRepoNamecheap_ListZones_NameserversErrorIgnored(t *testing.T) {
	srv, repo := newNamecheapTestRepo(t)
	srv.Stub(ncCmdDomainsList, ncDomainsListXML(1, "example.com"))
	srv.StubError(ncCmdDNSList, 2030166, "Domain is invalid")

	zones, err := repo.ListZones(context.Background())

	require.NoError(t, err)
	require.Len(t, zones, 1)
	assert.Empty(t, zones[0].NameServers)
}

func TestRepoNamecheap_ListZones_APIError(t *testing.T) {
	srv, repo := newNamecheapTestRepo(t)
	srv.StubError(ncCmdDomainsList, 4011103, "Access denied")

	_, err := repo.ListZones(context.Background())

	assert.ErrorIs(t, err, namecheap.ErrAccessDenied)
}

// --- ZoneIDByName ---

func TestRepoNamecheap_ZoneIDByName_Found(t *testing.T) {
	srv, repo := newNamecheapTestRepo(t)
	srv.Stub(ncCmdDomainsList, ncDomainsListXML(1, "example.com"))
	srv.Stub(ncCmdDNSList, ncDNSListXML("example.com"))

	id, err := repo.ZoneIDByName("example.com")

	require.NoError(t, err)
	assert.Equal(t, "example.com", id)
}

func TestRepoNamecheap_ZoneIDByName_NotFound(t *testing.T) {
	srv, repo := newNamecheapTestRepo(t)
	srv.Stub(ncCmdDomainsList, ncDomainsListXML(0))

	_, err := repo.ZoneIDByName("missing.com")

	assert.ErrorContains(t, err, "not found")
}

func TestRepoNamecheap_ZoneIDByName_APIError(t *testing.T) {
	srv, repo := newNamecheapTestRepo(t)
	srv.StubError(ncCmdDomainsList, 9999, "Unknown error")

	_, err := repo.ZoneIDByName("example.com")

	assert.Error(t, err)
}

// --- ListDNSRecords / GetDNSRecord ---

func TestRepoNamecheap_ListDNSRecords(t *testing.T) {
	srv, repo := newNamecheapTestRepo(t)
	srv.Stub(ncCmdGetHosts, ncGetHostsXML("example.com", "MX",
		ncHost{id: 1001, name: "@", typ: "A", address: "1.2.3.4", ttl: 1800},
		ncHost{id: 1002, name: "www", typ: "CNAME", address: "example.com", ttl: 3600}))

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
	assert.Equal(t, "CNAME", records[1].Type)
	srv.AssertCalled(t, ncCmdGetHosts, map[string]string{"SLD": "example", "TLD": "com"})
}

func TestRepoNamecheap_ListDNSRecords_NilHosts(t *testing.T) {
	srv, repo := newNamecheapTestRepo(t)
	srv.Stub(ncCmdGetHosts, ncGetHostsXML("example.com", "NONE"))

	records, err := repo.ListDNSRecords(context.Background(), "example.com")

	require.NoError(t, err)
	assert.Empty(t, records)
}

func TestRepoNamecheap_ListDNSRecords_EmptyResponse(t *testing.T) {
	srv, repo := newNamecheapTestRepo(t)
	srv.Stub(ncCmdGetHosts, ncEnvelope(""))

	_, err := repo.ListDNSRecords(context.Background(), "example.com")

	assert.ErrorContains(t, err, "empty response")
}

func TestRepoNamecheap_ListDNSRecords_APIError(t *testing.T) {
	srv, repo := newNamecheapTestRepo(t)
	srv.StubError(ncCmdGetHosts, 2030166, "Domain is invalid")

	_, err := repo.ListDNSRecords(context.Background(), "example.com")

	assert.ErrorIs(t, err, namecheap.ErrDomainInvalid)
	assert.ErrorContains(t, err, "example.com")
}

func TestRepoNamecheap_GetDNSRecord_Found(t *testing.T) {
	srv, repo := newNamecheapTestRepo(t)
	srv.Stub(ncCmdGetHosts, ncGetHostsXML("example.com", "NONE",
		ncHost{id: 2001, name: "sub", typ: "TXT", address: "hello", ttl: 900}))

	rec, err := repo.GetDNSRecord(context.Background(), "example.com", "2001")

	require.NoError(t, err)
	assert.Equal(t, "2001", rec.ID)
	assert.Equal(t, "sub", rec.Name)
	assert.Equal(t, "TXT", rec.Type)
	assert.Equal(t, "hello", rec.Content)
}

func TestRepoNamecheap_GetDNSRecord_NotFound(t *testing.T) {
	srv, repo := newNamecheapTestRepo(t)
	srv.Stub(ncCmdGetHosts, ncGetHostsXML("example.com", "NONE",
		ncHost{id: 2001, name: "sub", typ: "TXT", address: "hello", ttl: 900}))

	_, err := repo.GetDNSRecord(context.Background(), "example.com", "9999")

	assert.ErrorContains(t, err, "not found")
}

func TestRepoNamecheap_GetDNSRecord_APIError(t *testing.T) {
	srv, repo := newNamecheapTestRepo(t)
	srv.StubError(ncCmdGetHosts, 2030166, "Domain is invalid")

	_, err := repo.GetDNSRecord(context.Background(), "example.com", "1001")

	assert.Error(t, err)
}

// --- CreateDNSRecord ---

func TestRepoNamecheap_CreateDNSRecord(t *testing.T) {
	existing := ncHost{id: 4001, name: "@", typ: "A", address: "1.1.1.1", ttl: 1800}
	created := ncHost{id: 4002, name: "www", typ: "A", address: "2.2.2.2", ttl: 3600}

	srv, repo := newNamecheapTestRepo(t)
	// Read before the write, the verifying re-read, and the lookup of the new HostId.
	srv.StubSequence(ncCmdGetHosts,
		ncGetHostsXML("example.com", "NONE", existing),
		ncGetHostsXML("example.com", "NONE", existing, created))

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

	srv.AssertCalled(t, ncCmdSetHosts, map[string]string{
		"EmailType":   "NONE",
		"HostName1":   "@",
		"RecordType1": "A",
		"Address1":    "1.1.1.1",
		"TTL1":        "1800",
		"MXPref1":     "10",
		"HostName2":   "www",
		"RecordType2": "A",
		"Address2":    "2.2.2.2",
		"TTL2":        "3600",
		"MXPref2":     "10",
	})
	assert.Len(t, srv.Calls(ncCmdSetHosts), 1)
}

func TestRepoNamecheap_CreateDNSRecord_ZoneIDFallback(t *testing.T) {
	created := ncHost{id: 5001, name: "mail", typ: "A", address: "3.3.3.3", ttl: 1800}

	srv, repo := newNamecheapTestRepo(t)
	srv.StubSequence(ncCmdGetHosts,
		ncGetHostsXML("example.com", "NONE"),
		ncGetHostsXML("example.com", "NONE", created))

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

func TestRepoNamecheap_CreateDNSRecord_FirstMXSwitchesEmailType(t *testing.T) {
	existing := ncHost{id: 1, name: "@", typ: "A", address: "1.1.1.1", ttl: 1800}
	mx := ncHost{id: 2, name: "@", typ: "MX", address: "mail.example.com", ttl: 1800}

	srv, repo := newNamecheapTestRepo(t)
	// Plan read, read before the write, then the re-reads of the written zone.
	srv.StubSequence(ncCmdGetHosts,
		ncGetHostsXML("example.com", "NONE", existing),
		ncGetHostsXML("example.com", "NONE", existing),
		ncGetHostsXML("example.com", "MX", existing, mx))

	rec, err := repo.CreateDNSRecord(context.Background(), models.CreateDNSRecordParams{
		ZoneName: "example.com",
		Name:     "example.com",
		Type:     "MX",
		Content:  "mail.example.com",
		TTL:      1800,
	})

	require.NoError(t, err)
	assert.Equal(t, "2", rec.ID)
	srv.AssertCalled(t, ncCmdSetHosts, map[string]string{
		"EmailType":   "MX",
		"HostName2":   "@",
		"RecordType2": "MX",
		"Address2":    "mail.example.com",
		"MXPref2":     "10",
	})
}

func TestRepoNamecheap_CreateDNSRecord_Conflict(t *testing.T) {
	existing := ncHost{id: 1, name: "@", typ: "A", address: "1.1.1.1", ttl: 1800}

	srv, repo := newNamecheapTestRepo(t)
	// The zone never shows the new record, so every verification fails.
	srv.Stub(ncCmdGetHosts, ncGetHostsXML("example.com", "NONE", existing))

	_, err := repo.CreateDNSRecord(context.Background(), models.CreateDNSRecordParams{
		ZoneName: "example.com",
		Name:     "www",
		Type:     "A",
		Content:  "2.2.2.2",
		TTL:      1800,
	})

	assert.ErrorIs(t, err, namecheap.ErrConcurrentModification)
	assert.ErrorContains(t, err, "example.com")
	assert.Len(t, srv.Calls(ncCmdSetHosts), namecheapConflictRetries)
}

func TestRepoNamecheap_CreateDNSRecord_NotFoundAfterWrite(t *testing.T) {
	existing := ncHost{id: 1, name: "@", typ: "A", address: "1.1.1.1", ttl: 1800}
	// The verifying re-read shows the new record, but it is gone again by the time
	// the HostId lookup runs.
	srv, repo := newNamecheapTestRepo(t)
	srv.StubSequence(ncCmdGetHosts,
		ncGetHostsXML("example.com", "NONE", existing),
		ncGetHostsXML("example.com", "NONE", existing,
			ncHost{id: 2, name: "www", typ: "A", address: "2.2.2.2", ttl: 1800}),
		ncGetHostsXML("example.com", "NONE", existing))

	_, err := repo.CreateDNSRecord(context.Background(), models.CreateDNSRecordParams{
		ZoneName: "example.com",
		Name:     "www",
		Type:     "A",
		Content:  "2.2.2.2",
		TTL:      1800,
	})

	assert.ErrorContains(t, err, "not found in zone example.com after the change")
}

func TestRepoNamecheap_CreateDNSRecord_APIError(t *testing.T) {
	srv, repo := newNamecheapTestRepo(t)
	srv.StubError(ncCmdGetHosts, 2030166, "Domain is invalid")

	_, err := repo.CreateDNSRecord(context.Background(), models.CreateDNSRecordParams{
		ZoneName: "example.com",
		Name:     "www",
		Type:     "A",
		Content:  "2.2.2.2",
		TTL:      1800,
	})

	assert.ErrorIs(t, err, namecheap.ErrDomainInvalid)
}

func TestRepoNamecheap_CreateDNSRecord_EmptyZone(t *testing.T) {
	_, repo := newNamecheapTestRepo(t)

	_, err := repo.CreateDNSRecord(context.Background(), models.CreateDNSRecordParams{})

	assert.ErrorContains(t, err, "zone name or zone ID must be provided")
}

// --- DeleteDNSRecord ---

func TestRepoNamecheap_DeleteDNSRecord(t *testing.T) {
	keep := ncHost{id: 3001, name: "keep", typ: "A", address: "1.1.1.1", ttl: 1800}
	gone := ncHost{id: 3002, name: "delete-me", typ: "A", address: "2.2.2.2", ttl: 1800}

	srv, repo := newNamecheapTestRepo(t)
	// HostId lookup, read before the write, then the verifying re-read.
	srv.StubSequence(ncCmdGetHosts,
		ncGetHostsXML("example.com", "NONE", keep, gone),
		ncGetHostsXML("example.com", "NONE", keep, gone),
		ncGetHostsXML("example.com", "NONE", keep))

	err := repo.DeleteDNSRecord(context.Background(), "example.com", "3002")

	require.NoError(t, err)
	srv.AssertCalled(t, ncCmdSetHosts, map[string]string{
		"EmailType":   "NONE",
		"HostName1":   "keep",
		"RecordType1": "A",
		"Address1":    "1.1.1.1",
	})
	values := ncSetHostsValues(t, srv)
	assert.NotContains(t, values, "delete-me")
	assert.NotContains(t, values, "2.2.2.2")
}

func TestRepoNamecheap_DeleteDNSRecord_LastMXSwitchesEmailTypeOff(t *testing.T) {
	keep := ncHost{id: 1, name: "@", typ: "A", address: "1.1.1.1", ttl: 1800}
	mx := ncHost{id: 2, name: "@", typ: "MX", address: "mail.example.com", ttl: 1800}

	srv, repo := newNamecheapTestRepo(t)
	// HostId lookup, plan read, read before the write, then the verifying re-read.
	srv.StubSequence(ncCmdGetHosts,
		ncGetHostsXML("example.com", "MX", keep, mx),
		ncGetHostsXML("example.com", "MX", keep, mx),
		ncGetHostsXML("example.com", "MX", keep, mx),
		ncGetHostsXML("example.com", "NONE", keep))

	err := repo.DeleteDNSRecord(context.Background(), "example.com", "2")

	require.NoError(t, err)
	srv.AssertCalled(t, ncCmdSetHosts, map[string]string{
		"EmailType":   "NONE",
		"HostName1":   "@",
		"RecordType1": "A",
	})
	assert.NotContains(t, ncSetHostsValues(t, srv), "mail.example.com")
}

func TestRepoNamecheap_DeleteDNSRecord_Conflict(t *testing.T) {
	keep := ncHost{id: 1, name: "keep", typ: "A", address: "1.1.1.1", ttl: 1800}
	gone := ncHost{id: 2, name: "gone", typ: "A", address: "2.2.2.2", ttl: 1800}

	srv, repo := newNamecheapTestRepo(t)
	// The deleted record keeps coming back, so every verification fails.
	srv.Stub(ncCmdGetHosts, ncGetHostsXML("example.com", "NONE", keep, gone))

	err := repo.DeleteDNSRecord(context.Background(), "example.com", "2")

	assert.ErrorIs(t, err, namecheap.ErrConcurrentModification)
	assert.Len(t, srv.Calls(ncCmdSetHosts), namecheapConflictRetries)
}

func TestRepoNamecheap_DeleteDNSRecord_NotFound(t *testing.T) {
	srv, repo := newNamecheapTestRepo(t)
	srv.Stub(ncCmdGetHosts, ncGetHostsXML("example.com", "NONE",
		ncHost{id: 3001, name: "keep", typ: "A", address: "1.1.1.1", ttl: 1800}))

	err := repo.DeleteDNSRecord(context.Background(), "example.com", "9999")

	assert.ErrorContains(t, err, "not found")
	assert.Empty(t, srv.Calls(ncCmdSetHosts))
}

func TestRepoNamecheap_DeleteDNSRecord_NilHosts(t *testing.T) {
	srv, repo := newNamecheapTestRepo(t)
	srv.Stub(ncCmdGetHosts, ncGetHostsXML("example.com", "NONE"))

	err := repo.DeleteDNSRecord(context.Background(), "example.com", "3001")

	assert.ErrorContains(t, err, "not found")
}

func TestRepoNamecheap_DeleteDNSRecord_APIError(t *testing.T) {
	srv, repo := newNamecheapTestRepo(t)
	srv.StubError(ncCmdGetHosts, 2030166, "Domain is invalid")

	err := repo.DeleteDNSRecord(context.Background(), "example.com", "3001")

	assert.ErrorIs(t, err, namecheap.ErrDomainInvalid)
}

// --- UpdateDNSRecord ---

func TestRepoNamecheap_UpdateDNSRecord(t *testing.T) {
	old := ncHost{id: 6001, name: "old", typ: "A", address: "1.1.1.1", ttl: 1800}
	updated := ncHost{id: 6002, name: "new", typ: "A", address: "9.9.9.9", ttl: 300}

	srv, repo := newNamecheapTestRepo(t)
	// HostId lookup, read before the write, then the re-reads of the written zone.
	srv.StubSequence(ncCmdGetHosts,
		ncGetHostsXML("example.com", "NONE", old),
		ncGetHostsXML("example.com", "NONE", old),
		ncGetHostsXML("example.com", "NONE", updated))

	rec, err := repo.UpdateDNSRecord(context.Background(), models.UpdateDNSRecordParams{
		ZoneName: "example.com",
		ID:       "6001",
		Name:     "new.example.com",
		Type:     "A",
		Content:  "9.9.9.9",
		TTL:      300,
	})

	require.NoError(t, err)
	assert.Equal(t, "6002", rec.ID)
	assert.Equal(t, "new", rec.Name)
	assert.Equal(t, "9.9.9.9", rec.Content)
	assert.Equal(t, 300, rec.TTL)

	srv.AssertCalled(t, ncCmdSetHosts, map[string]string{
		"EmailType":   "NONE",
		"HostName1":   "new",
		"RecordType1": "A",
		"Address1":    "9.9.9.9",
		"TTL1":        "300",
		"MXPref1":     "10",
	})
	values := ncSetHostsValues(t, srv)
	assert.NotContains(t, values, "old")
	assert.NotContains(t, values, "1.1.1.1")
}

func TestRepoNamecheap_UpdateDNSRecord_KeepsMXPref(t *testing.T) {
	old := ncHost{id: 1, name: "@", typ: "MX", address: "mail1.example.com", ttl: 1800, mxPref: 20}
	updated := ncHost{id: 1, name: "@", typ: "MX", address: "mail2.example.com", ttl: 1800, mxPref: 20}

	srv, repo := newNamecheapTestRepo(t)
	// HostId lookup, plan read, read before the write, then the re-reads.
	srv.StubSequence(ncCmdGetHosts,
		ncGetHostsXML("example.com", "MX", old),
		ncGetHostsXML("example.com", "MX", old),
		ncGetHostsXML("example.com", "MX", old),
		ncGetHostsXML("example.com", "MX", updated))

	rec, err := repo.UpdateDNSRecord(context.Background(), models.UpdateDNSRecordParams{
		ZoneName: "example.com",
		ID:       "1",
		Name:     "@",
		Type:     "MX",
		Content:  "mail2.example.com",
		TTL:      1800,
	})

	require.NoError(t, err)
	assert.Equal(t, "mail2.example.com", rec.Content)
	srv.AssertCalled(t, ncCmdSetHosts, map[string]string{
		"EmailType":   "MX",
		"HostName1":   "@",
		"RecordType1": "MX",
		"Address1":    "mail2.example.com",
		"MXPref1":     "20",
	})
}

func TestRepoNamecheap_UpdateDNSRecord_ZoneIDFallback(t *testing.T) {
	old := ncHost{id: 7001, name: "sub", typ: "TXT", address: "old-val", ttl: 900}
	updated := ncHost{id: 7001, name: "sub", typ: "TXT", address: "new-val", ttl: 900}

	srv, repo := newNamecheapTestRepo(t)
	srv.StubSequence(ncCmdGetHosts,
		ncGetHostsXML("example.com", "NONE", old),
		ncGetHostsXML("example.com", "NONE", old),
		ncGetHostsXML("example.com", "NONE", updated))

	rec, err := repo.UpdateDNSRecord(context.Background(), models.UpdateDNSRecordParams{
		ZoneID:  "example.com",
		ID:      "7001",
		Name:    "sub",
		Type:    "TXT",
		Content: "new-val",
		TTL:     900,
	})

	require.NoError(t, err)
	assert.Equal(t, "7001", rec.ID)
	assert.Equal(t, "new-val", rec.Content)
}

func TestRepoNamecheap_UpdateDNSRecord_EmptyZone(t *testing.T) {
	_, repo := newNamecheapTestRepo(t)

	_, err := repo.UpdateDNSRecord(context.Background(), models.UpdateDNSRecordParams{})

	assert.ErrorContains(t, err, "zone name or zone ID must be provided")
}

func TestRepoNamecheap_UpdateDNSRecord_NotFound(t *testing.T) {
	srv, repo := newNamecheapTestRepo(t)
	srv.Stub(ncCmdGetHosts, ncGetHostsXML("example.com", "NONE",
		ncHost{id: 8001, name: "sub", typ: "A", address: "1.1.1.1", ttl: 1800}))

	_, err := repo.UpdateDNSRecord(context.Background(), models.UpdateDNSRecordParams{
		ZoneName: "example.com",
		ID:       "9999",
		Name:     "sub",
		Type:     "A",
		Content:  "2.2.2.2",
		TTL:      1800,
	})

	assert.ErrorContains(t, err, "not found")
	assert.Empty(t, srv.Calls(ncCmdSetHosts))
}

func TestRepoNamecheap_UpdateDNSRecord_NilHosts(t *testing.T) {
	srv, repo := newNamecheapTestRepo(t)
	srv.Stub(ncCmdGetHosts, ncGetHostsXML("example.com", "NONE"))

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

func TestRepoNamecheap_UpdateDNSRecord_Conflict(t *testing.T) {
	old := ncHost{id: 1, name: "old", typ: "A", address: "1.1.1.1", ttl: 1800}

	srv, repo := newNamecheapTestRepo(t)
	srv.Stub(ncCmdGetHosts, ncGetHostsXML("example.com", "NONE", old))

	_, err := repo.UpdateDNSRecord(context.Background(), models.UpdateDNSRecordParams{
		ZoneName: "example.com",
		ID:       "1",
		Name:     "new",
		Type:     "A",
		Content:  "9.9.9.9",
		TTL:      1800,
	})

	assert.ErrorIs(t, err, namecheap.ErrConcurrentModification)
	assert.Len(t, srv.Calls(ncCmdSetHosts), namecheapConflictRetries)
}
