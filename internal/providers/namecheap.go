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
	"strconv"
	"strings"

	"github.com/mixanemca/cdnscli/internal/models"
	namecheap "github.com/namecheap/go-namecheap-sdk/v2/namecheap"
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
	domain := params.ZoneName
	if domain == "" {
		domain = params.ZoneID
	}
	if domain == "" {
		return models.DNSRecord{}, fmt.Errorf("zone name or zone ID must be provided")
	}

	result, err := r.getHosts(domain)
	if err != nil {
		return models.DNSRecord{}, err
	}

	var existing []namecheap.DomainsDNSHostRecordDetailed
	if result.Hosts != nil {
		existing = *result.Hosts
	}

	newHost := convParamsToNamecheapHost(params)
	updated := append(existing, newHost)

	if err := r.setHosts(domain, updated, result.EmailType); err != nil {
		return models.DNSRecord{}, err
	}

	// Fetch fresh list to get the assigned HostId
	records, err := r.ListDNSRecords(ctx, domain)
	if err != nil {
		return models.DNSRecord{}, err
	}

	hostName := relativeHostName(params.Name, domain)
	for i := len(records) - 1; i >= 0; i-- {
		rec := records[i]
		if rec.Type == params.Type && rec.Content == params.Content &&
			(relativeHostName(rec.Name, domain) == hostName || rec.Name == hostName) {
			return rec, nil
		}
	}

	return models.DNSRecord{}, fmt.Errorf("created record not found after setHosts")
}

func (r *repoNamecheap) DeleteDNSRecord(ctx context.Context, zoneID, recordID string) error {
	result, err := r.getHosts(zoneID)
	if err != nil {
		return err
	}

	if result.Hosts == nil {
		return fmt.Errorf("record with ID %s not found in zone %s", recordID, zoneID)
	}

	filtered := make([]namecheap.DomainsDNSHostRecordDetailed, 0, len(*result.Hosts))
	found := false
	for _, h := range *result.Hosts {
		if h.HostId != nil && strconv.Itoa(*h.HostId) == recordID {
			found = true
			continue
		}
		filtered = append(filtered, h)
	}

	if !found {
		return fmt.Errorf("record with ID %s not found in zone %s", recordID, zoneID)
	}

	return r.setHosts(zoneID, filtered, result.EmailType)
}

func (r *repoNamecheap) ListDNSRecords(ctx context.Context, id string) ([]models.DNSRecord, error) {
	result, err := r.getHosts(id)
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
	pageSize := 100
	page := 1
	var allDomains []namecheap.Domain

	for {
		args := &namecheap.DomainsGetListArgs{
			PageSize: namecheap.Int(pageSize),
			Page:     namecheap.Int(page),
		}

		if len(z) > 0 && z[0] != "" {
			args.SearchTerm = namecheap.String(z[0])
		}

		resp, err := r.client.Domains.GetList(args)
		if err != nil {
			return nil, err
		}

		if resp.Domains == nil || len(*resp.Domains) == 0 {
			break
		}

		allDomains = append(allDomains, *resp.Domains...)

		if resp.Paging == nil || resp.Paging.TotalItems == nil {
			break
		}
		if page*pageSize >= *resp.Paging.TotalItems {
			break
		}
		page++
	}

	zones := convFromNamecheapDomains(allDomains)

	for i, zone := range zones {
		resp, err := r.client.DomainsDNS.GetList(zone.Name)
		if err == nil && resp.DomainDNSGetListResult != nil && resp.DomainDNSGetListResult.Nameservers != nil {
			zones[i].NameServers = *resp.DomainDNSGetListResult.Nameservers
		}
	}

	return zones, nil
}

func (r *repoNamecheap) UpdateDNSRecord(ctx context.Context, params models.UpdateDNSRecordParams) (models.DNSRecord, error) {
	domain := params.ZoneName
	if domain == "" {
		domain = params.ZoneID
	}
	if domain == "" {
		return models.DNSRecord{}, fmt.Errorf("zone name or zone ID must be provided")
	}

	result, err := r.getHosts(domain)
	if err != nil {
		return models.DNSRecord{}, err
	}

	if result.Hosts == nil {
		return models.DNSRecord{}, fmt.Errorf("record with ID %s not found in zone %s", params.ID, domain)
	}

	updated := make([]namecheap.DomainsDNSHostRecordDetailed, 0, len(*result.Hosts))
	found := false
	for _, h := range *result.Hosts {
		if h.HostId != nil && strconv.Itoa(*h.HostId) == params.ID {
			hostName := relativeHostName(params.Name, domain)
			ttl := params.TTL
			h.Name = namecheap.String(hostName)
			h.Type = namecheap.String(params.Type)
			h.Address = namecheap.String(params.Content)
			h.TTL = &ttl
			found = true
		}
		updated = append(updated, h)
	}

	if !found {
		return models.DNSRecord{}, fmt.Errorf("record with ID %s not found in zone %s", params.ID, domain)
	}

	if err := r.setHosts(domain, updated, result.EmailType); err != nil {
		return models.DNSRecord{}, err
	}

	return r.GetDNSRecord(ctx, domain, params.ID)
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
func (r *repoNamecheap) getHosts(domain string) (*namecheap.DomainDNSGetHostsResult, error) {
	resp, err := r.client.DomainsDNS.GetHosts(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS hosts for %s: %w", domain, err)
	}
	if resp == nil || resp.DomainDNSGetHostsResult == nil {
		return nil, fmt.Errorf("empty response for domain %s", domain)
	}
	return resp.DomainDNSGetHostsResult, nil
}

// setHosts replaces all DNS records for a domain, preserving EmailType.
func (r *repoNamecheap) setHosts(domain string, hosts []namecheap.DomainsDNSHostRecordDetailed, emailType *string) error {
	records := make([]namecheap.DomainsDNSHostRecord, 0, len(hosts))
	for _, h := range hosts {
		rec := namecheap.DomainsDNSHostRecord{
			HostName:   h.Name,
			RecordType: h.Type,
			Address:    h.Address,
			TTL:        h.TTL,
		}
		if h.MXPref != nil {
			mxPref := uint8(*h.MXPref)
			rec.MXPref = &mxPref
		}
		records = append(records, rec)
	}

	args := &namecheap.DomainsDNSSetHostsArgs{
		Domain:    namecheap.String(domain),
		Records:   &records,
		EmailType: emailType,
	}

	_, err := r.client.DomainsDNS.SetHosts(args)
	return err
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

// Conversion functions

func convFromNamecheapHost(h namecheap.DomainsDNSHostRecordDetailed, domain string) models.DNSRecord {
	var id string
	if h.HostId != nil {
		id = strconv.Itoa(*h.HostId)
	}

	var name string
	if h.Name != nil {
		name = *h.Name
	}

	var ttl int
	if h.TTL != nil {
		ttl = *h.TTL
	}

	var recType string
	if h.Type != nil {
		recType = *h.Type
	}

	var address string
	if h.Address != nil {
		address = *h.Address
	}

	return models.DNSRecord{
		ID:      id,
		Name:    name,
		TTL:     ttl,
		Type:    recType,
		Content: address,
		Proxied: false,
	}
}

func convFromNamecheapDomains(domains []namecheap.Domain) []models.Zone {
	result := make([]models.Zone, 0, len(domains))
	for _, d := range domains {
		var name string
		if d.Name != nil {
			name = *d.Name
		}
		// Zone ID equals domain name: all DNS operations (getHosts/setHosts) require the domain
		// name, not Namecheap's internal numeric ID.
		result = append(result, models.Zone{
			ID:   name,
			Name: name,
		})
	}
	return result
}

func convParamsToNamecheapHost(params models.CreateDNSRecordParams) namecheap.DomainsDNSHostRecordDetailed {
	domain := params.ZoneName
	if domain == "" {
		domain = params.ZoneID
	}
	hostName := relativeHostName(params.Name, domain)
	ttl := params.TTL

	return namecheap.DomainsDNSHostRecordDetailed{
		Name:    namecheap.String(hostName),
		Type:    namecheap.String(params.Type),
		Address: namecheap.String(params.Content),
		TTL:     &ttl,
	}
}
