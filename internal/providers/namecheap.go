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
	"strconv"
	"strings"

	"github.com/mixanemca/cdnscli/internal/models"
	namecheap "github.com/namecheap/go-namecheap-sdk/v2/namecheap"
)

const (
	// namecheapListZonesPageSize is the largest page size domains.getList accepts.
	namecheapListZonesPageSize = 100

	// namecheapConflictRetries is how many read-modify-write-verify cycles a record
	// mutation attempts before giving up with ErrConcurrentModification.
	namecheapConflictRetries = 3

	// namecheapDefaultMXPref is the MX preference assigned to records created by
	// cdnscli. Namecheap reports MXPref="10" for every host record it returns,
	// including A, CNAME and TXT (see the official getHosts examples and the SDK's
	// captured fixtures). The SDK verifies a write by re-reading the zone and
	// comparing records MXPref-inclusive, so a record sent without MXPref would never
	// match its echo and every mutation would be reported as a concurrent
	// modification.
	namecheapDefaultMXPref uint8 = 10
)

type repoNamecheap struct {
	client *namecheap.Client
}

// NewRepoNamecheap creates a repository for Namecheap provider.
func NewRepoNamecheap(client *namecheap.Client) Repo {
	return &repoNamecheap{client: client}
}

func (r *repoNamecheap) GetDNSRecord(ctx context.Context, zoneID, recordID string) (models.DNSRecord, error) {
	records, err := r.ListDNSRecords(ctx, zoneID)
	if err != nil {
		return models.DNSRecord{}, err
	}

	for _, rec := range records {
		if rec.ID == recordID {
			return rec, nil
		}
	}

	return models.DNSRecord{}, fmt.Errorf("record with ID %s not found in zone %s", recordID, zoneID)
}

func (r *repoNamecheap) CreateDNSRecord(ctx context.Context, params models.CreateDNSRecordParams) (models.DNSRecord, error) {
	domain, err := namecheapDomain(params.ZoneName, params.ZoneID)
	if err != nil {
		return models.DNSRecord{}, err
	}

	record := convParamsToNamecheapRecord(params)
	records := []namecheap.DomainsDNSHostRecord{record}

	opts, err := r.mutationOptions(ctx, domain, namecheap.AddOp(record), isMailRecordType(params.Type))
	if err != nil {
		return models.DNSRecord{}, err
	}

	if _, err := r.client.DomainsDNS.AddRecordsWithContext(ctx, domain, records, opts...); err != nil {
		return models.DNSRecord{}, wrapNamecheapMutationError(domain, err)
	}

	return r.findRecord(ctx, domain, record)
}

func (r *repoNamecheap) DeleteDNSRecord(ctx context.Context, zoneID, recordID string) error {
	hosts, err := r.getHosts(ctx, zoneID)
	if err != nil {
		return err
	}

	target, err := findHostByID(hosts, zoneID, recordID)
	if err != nil {
		return err
	}

	selector := selectorForHost(target)
	mail := isMailRecordType(derefString(target.Type))

	opts, err := r.mutationOptions(ctx, zoneID, namecheap.DeleteOp(selector), mail)
	if err != nil {
		return err
	}

	_, err = r.client.DomainsDNS.DeleteRecordsWithContext(ctx, zoneID, selector, opts...)
	return wrapNamecheapMutationError(zoneID, err)
}

func (r *repoNamecheap) ListDNSRecords(ctx context.Context, id string) ([]models.DNSRecord, error) {
	result, err := r.getHosts(ctx, id)
	if err != nil {
		return nil, err
	}

	if result.Hosts == nil {
		return []models.DNSRecord{}, nil
	}

	records := make([]models.DNSRecord, 0, len(*result.Hosts))
	for _, h := range *result.Hosts {
		records = append(records, convFromNamecheapHost(h, id))
	}
	return records, nil
}

func (r *repoNamecheap) ListZones(ctx context.Context, z ...string) ([]models.Zone, error) {
	args := &namecheap.DomainsGetListArgs{
		PageSize: namecheap.Int(namecheapListZonesPageSize),
	}
	if len(z) > 0 && z[0] != "" {
		args.SearchTerm = namecheap.String(z[0])
	}

	domains, err := r.client.Domains.ListAllSlice(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}

	zones := convFromNamecheapDomains(domains)

	for i := range zones {
		resp, err := r.client.DomainsDNS.GetListWithContext(ctx, zones[i].Name)
		if err != nil {
			// Nameservers are informational: an API error for one zone does not
			// fail the listing, but a cancelled or expired context always does.
			if ctx.Err() != nil {
				return nil, err
			}
			continue
		}
		if resp != nil && resp.DomainDNSGetListResult != nil && resp.DomainDNSGetListResult.Nameservers != nil {
			zones[i].NameServers = *resp.DomainDNSGetListResult.Nameservers
		}
	}

	return zones, nil
}

func (r *repoNamecheap) UpdateDNSRecord(ctx context.Context, params models.UpdateDNSRecordParams) (models.DNSRecord, error) {
	domain, err := namecheapDomain(params.ZoneName, params.ZoneID)
	if err != nil {
		return models.DNSRecord{}, err
	}

	hosts, err := r.getHosts(ctx, domain)
	if err != nil {
		return models.DNSRecord{}, err
	}

	target, err := findHostByID(hosts, domain, params.ID)
	if err != nil {
		return models.DNSRecord{}, err
	}

	selector := selectorForHost(target)
	record := convUpdateParamsToNamecheapRecord(params, target)
	records := []namecheap.DomainsDNSHostRecord{record}
	mail := isMailRecordType(derefString(target.Type), params.Type)

	opts, err := r.mutationOptions(ctx, domain, namecheap.UpsertOp(selector, records), mail)
	if err != nil {
		return models.DNSRecord{}, err
	}

	if _, err := r.client.DomainsDNS.UpsertRecordsWithContext(ctx, domain, selector, records, opts...); err != nil {
		return models.DNSRecord{}, wrapNamecheapMutationError(domain, err)
	}

	return r.findRecord(ctx, domain, record)
}

func (r *repoNamecheap) ZoneIDByName(zoneName string) (string, error) {
	zones, err := r.ListZones(context.Background(), zoneName)
	if err != nil {
		return "", err
	}

	for _, z := range zones {
		if z.Name == zoneName {
			return z.ID, nil
		}
	}

	return "", fmt.Errorf("zone %s not found", zoneName)
}

// getHosts fetches the full DNS host record result for a domain.
func (r *repoNamecheap) getHosts(ctx context.Context, domain string) (*namecheap.DomainDNSGetHostsResult, error) {
	resp, err := r.client.DomainsDNS.GetHostsWithContext(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS hosts for %s: %w", domain, err)
	}
	if resp == nil || resp.DomainDNSGetHostsResult == nil {
		return nil, fmt.Errorf("empty response for domain %s", domain)
	}
	return resp.DomainDNSGetHostsResult, nil
}

// mutationOptions builds the options for a record mutation: automatic retry when a
// concurrent modification is detected and, for changes that touch mail records, the
// EmailType transition Namecheap requires. A zone accepts MX records only with
// EmailType=MX and has to drop back to NONE once the last MX record is gone, so the
// change is planned first and the required EmailType is written together with it.
func (r *repoNamecheap) mutationOptions(ctx context.Context, domain string, op namecheap.RecordOp, mail bool) ([]namecheap.RecordOption, error) {
	opts := []namecheap.RecordOption{namecheap.WithRetryOnConflict(namecheapConflictRetries)}
	if !mail {
		return opts, nil
	}

	diff, err := r.client.DomainsDNS.PlanWithContext(ctx, domain, op)
	if err != nil {
		return nil, fmt.Errorf("failed to plan DNS change for %s: %w", domain, err)
	}
	if !diff.Satisfiable {
		return nil, fmt.Errorf("DNS change for %s is not applicable: "+
			"a zone cannot hold both MX and MXE records or more than one MXE record", domain)
	}
	if diff.RequiredEmailType != "" {
		opts = append(opts, namecheap.WithEmailType(diff.RequiredEmailType))
	}
	return opts, nil
}

// findRecord re-reads the zone and returns the record matching want by host name,
// type and address. The SDK's write result carries no HostId, so this is the only
// way to learn the identifier Namecheap assigned.
func (r *repoNamecheap) findRecord(ctx context.Context, domain string, want namecheap.DomainsDNSHostRecord) (models.DNSRecord, error) {
	hosts, err := r.getHosts(ctx, domain)
	if err != nil {
		return models.DNSRecord{}, err
	}

	if hosts.Hosts != nil {
		list := *hosts.Hosts
		for i := len(list) - 1; i >= 0; i-- {
			if sameNamecheapRecord(namecheap.RecordFromDetailed(list[i]), want) {
				return convFromNamecheapHost(list[i], domain), nil
			}
		}
	}

	return models.DNSRecord{}, fmt.Errorf("record %s %s %s not found in zone %s after the change",
		derefString(want.HostName), derefString(want.RecordType), derefString(want.Address), domain)
}

// findHostByID returns the host record with the given HostId.
func findHostByID(hosts *namecheap.DomainDNSGetHostsResult, domain, recordID string) (namecheap.DomainsDNSHostRecordDetailed, error) {
	if hosts.Hosts != nil {
		for _, h := range *hosts.Hosts {
			if h.HostId != nil && strconv.Itoa(*h.HostId) == recordID {
				return h, nil
			}
		}
	}
	return namecheap.DomainsDNSHostRecordDetailed{},
		fmt.Errorf("record with ID %s not found in zone %s", recordID, domain)
}

// selectorForHost builds a selector that matches exactly the given host record.
func selectorForHost(h namecheap.DomainsDNSHostRecordDetailed) namecheap.RecordSelector {
	rec := namecheap.RecordFromDetailed(h)
	return namecheap.RecordSelector{
		HostName:   rec.HostName,
		RecordType: rec.RecordType,
		Address:    rec.Address,
		MXPref:     rec.MXPref,
	}
}

// sameNamecheapRecord reports whether two records denote the same host name, type
// and address in normalized form, ignoring TTL and MX preference.
func sameNamecheapRecord(a, b namecheap.DomainsDNSHostRecord) bool {
	na := namecheap.NormalizeRecord(a)
	nb := namecheap.NormalizeRecord(b)
	return *na.HostName == *nb.HostName &&
		*na.RecordType == *nb.RecordType &&
		*na.Address == *nb.Address
}

// isMailRecordType reports whether any of the types is an MX or MXE record, whose
// presence is coupled to the zone EmailType.
func isMailRecordType(types ...string) bool {
	for _, t := range types {
		switch strings.ToUpper(strings.TrimSpace(t)) {
		case namecheap.RecordTypeMX, namecheap.RecordTypeMXE:
			return true
		}
	}
	return false
}

// wrapNamecheapMutationError adds the zone to a record mutation error and explains a
// concurrent modification that survived the retries.
func wrapNamecheapMutationError(domain string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, namecheap.ErrConcurrentModification) {
		return fmt.Errorf("DNS zone %s changed concurrently, gave up after %d attempts: %w",
			domain, namecheapConflictRetries, err)
	}
	return fmt.Errorf("failed to change DNS records of %s: %w", domain, err)
}

// namecheapDomain picks the zone name from the params, falling back to the zone ID
// (which equals the domain name for this provider).
func namecheapDomain(zoneName, zoneID string) (string, error) {
	domain := strings.TrimSpace(zoneName)
	if domain == "" {
		domain = strings.TrimSpace(zoneID)
	}
	if domain == "" {
		return "", fmt.Errorf("zone name or zone ID must be provided")
	}
	return domain, nil
}

// relativeHostName strips the domain suffix from a FQDN to get a relative host name.
// Returns "@" if the name equals the domain itself.
func relativeHostName(name, domain string) string {
	name = strings.TrimSuffix(name, ".")
	if name == domain || name == "" {
		return "@"
	}
	suffix := "." + domain
	if strings.HasSuffix(name, suffix) {
		return strings.TrimSuffix(name, suffix)
	}
	return name
}

func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// Conversion functions

func convFromNamecheapHost(h namecheap.DomainsDNSHostRecordDetailed, domain string) models.DNSRecord {
	var id string
	if h.HostId != nil {
		id = strconv.Itoa(*h.HostId)
	}

	var ttl int
	if h.TTL != nil {
		ttl = *h.TTL
	}

	return models.DNSRecord{
		ID:      id,
		Name:    derefString(h.Name),
		TTL:     ttl,
		Type:    derefString(h.Type),
		Content: derefString(h.Address),
		Proxied: false,
	}
}

func convFromNamecheapDomains(domains []*namecheap.Domain) []models.Zone {
	result := make([]models.Zone, 0, len(domains))
	for _, d := range domains {
		if d == nil {
			continue
		}
		name := derefString(d.Name)
		// Zone ID equals domain name: all DNS operations (getHosts/setHosts) require the domain
		// name, not Namecheap's internal numeric ID.
		result = append(result, models.Zone{
			ID:   name,
			Name: name,
		})
	}
	return result
}

// newNamecheapRecord builds a setHosts record from cdnscli values. The host name is
// made relative to the domain; a nil mxPref gets the cdnscli default.
func newNamecheapRecord(domain, name, recordType, content string, ttl int, mxPref *uint8) namecheap.DomainsDNSHostRecord {
	pref := namecheapDefaultMXPref
	if mxPref != nil {
		pref = *mxPref
	}
	return namecheap.DomainsDNSHostRecord{
		HostName:   namecheap.String(relativeHostName(name, domain)),
		RecordType: namecheap.String(recordType),
		Address:    namecheap.String(content),
		TTL:        namecheap.Int(ttl),
		MXPref:     namecheap.UInt8(pref),
	}
}

func convParamsToNamecheapRecord(params models.CreateDNSRecordParams) namecheap.DomainsDNSHostRecord {
	domain := params.ZoneName
	if domain == "" {
		domain = params.ZoneID
	}
	return newNamecheapRecord(domain, params.Name, params.Type, params.Content, params.TTL, nil)
}

// convUpdateParamsToNamecheapRecord builds the replacement record for an update,
// keeping the MX preference of the record being replaced.
func convUpdateParamsToNamecheapRecord(params models.UpdateDNSRecordParams, old namecheap.DomainsDNSHostRecordDetailed) namecheap.DomainsDNSHostRecord {
	domain := params.ZoneName
	if domain == "" {
		domain = params.ZoneID
	}
	mxPref := namecheap.RecordFromDetailed(old).MXPref
	return newNamecheapRecord(domain, params.Name, params.Type, params.Content, params.TTL, mxPref)
}
