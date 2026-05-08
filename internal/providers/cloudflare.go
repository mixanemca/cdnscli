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
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/mixanemca/cdnscli/internal/models"
)

// AddRR creates a new DNS resource record for a zone.
func (p *provider) AddRR(ctx context.Context, zone string, params models.CreateDNSRecordParams) (models.DNSRecord, error) {
	var rr models.DNSRecord

	zoneID, err := p.repo.ZoneIDByName(zone)
	if err != nil {
		return rr, err
	}
	params.ZoneID = zoneID

	rr, err = p.repo.CreateDNSRecord(ctx, params)
	if err != nil {
		return rr, err
	}

	return rr, nil
}

// DeleteRR deletes a DNS resource record from a given zone.
func (p *provider) DeleteRR(ctx context.Context, zone string, rr models.DNSRecord) error {
	zoneID, err := p.repo.ZoneIDByName(zone)
	if err != nil {
		return err
	}

	if rr.ID == "" {
		if regRepo, ok := p.repo.(*repoRegRu); ok {
			return regRepo.deleteDNSRecordByData(ctx, zone, rr)
		}
		return fmt.Errorf("required DNS record ID missing")
	}

	err = p.repo.DeleteDNSRecord(ctx, zoneID, rr.ID)
	if err != nil {
		return err
	}

	return nil
}

// UpdateRR updates an existing DNS resource record
func (p *provider) UpdateRR(ctx context.Context, zone string, rr models.DNSRecord) (models.DNSRecord, error) {
	zoneID, err := p.repo.ZoneIDByName(zone)
	if err != nil {
		return models.DNSRecord{}, err
	}

	updateParams := models.UpdateDNSRecordParams{
		Content: rr.Content,
		ID:      rr.ID,
		Name:    rr.Name,
		Proxied: rr.Proxied,
		TTL:     rr.TTL,
		Type:    rr.Type,
		ZoneID:  zoneID,
	}

	return p.repo.UpdateDNSRecord(ctx, updateParams)
}

// GetRRByName returns a single DNS record for the given zone & record identifiers.
func (p *provider) GetRRByName(ctx context.Context, zone, name string) (models.DNSRecord, error) {
	var rr models.DNSRecord

	zoneID, err := p.repo.ZoneIDByName(zone)
	if err != nil {
		return rr, err
	}

	rrset, err := p.repo.ListDNSRecords(ctx, zoneID)
	if err != nil {
		return rr, err
	}

	zoneName := normalizeDNSName(zone)
	requestedName := normalizeDNSName(name)
	expectedFQDN := buildFQDN(zoneName, requestedName)
	hasName := requestedName != ""

	for _, rec := range rrset {
		if hasName {
			recName := normalizeDNSName(rec.Name)
			if recName != requestedName && recName != expectedFQDN {
				continue
			}
		}

		if rec.ID == "" {
			return rr, fmt.Errorf("required DNS record ID missing")
		}

		rr, err = p.repo.GetDNSRecord(ctx, zoneID, rec.ID)
		if err != nil {
			return rr, err
		}
		return rr, nil
	}

	return rr, fmt.Errorf("record %q not found in zone %q", name, zone)
}

func normalizeDNSName(value string) string {
	return strings.TrimSuffix(strings.TrimSpace(value), ".")
}

func buildFQDN(zone, name string) string {
	if name == "" {
		return zone
	}

	if name == "@" {
		return zone
	}

	if strings.HasSuffix(name, "."+zone) || name == zone {
		return name
	}

	return strings.Join([]string{name, zone}, ".")
}

func filterDNSRecords(records []models.DNSRecord, params models.ListDNSRecordsParams) []models.DNSRecord {
	filtered := make([]models.DNSRecord, 0, len(records))
	zoneName := normalizeDNSName(params.ZoneName)
	requestedName := normalizeDNSName(params.Name)
	expectedFQDN := buildFQDN(zoneName, requestedName)
	hasName := requestedName != ""

	for _, rec := range records {
		if params.ID != "" && rec.ID != params.ID {
			continue
		}
		if params.Type != "" && !strings.EqualFold(rec.Type, params.Type) {
			continue
		}
		if params.Content != "" && rec.Content != params.Content {
			continue
		}
		if params.TTL != 0 && rec.TTL != params.TTL {
			continue
		}
		if hasName {
			recName := normalizeDNSName(rec.Name)
			relativeRequested := strings.TrimSuffix(requestedName, "."+zoneName)
			if recName != requestedName && recName != expectedFQDN && recName != relativeRequested {
				continue
			}
		}

		filtered = append(filtered, rec)
	}

	return filtered
}

// ListZones return lists zones on an account.
func (p *provider) ListZones(ctx context.Context) ([]models.Zone, error) {
	zones, err := p.repo.ListZones(ctx)
	if err != nil {
		return []models.Zone{}, err
	}

	return zones, nil
}

// ListZonesByName return lists zones on an account using the zone name for filtering.
func (p *provider) ListZonesByName(ctx context.Context, name string) ([]models.Zone, error) {
	zones, err := p.repo.ListZones(ctx, name)
	if err != nil {
		return []models.Zone{}, err
	}

	return zones, nil
}

// ListRecordsByZoneID returns a slice of DNS records for the given zone identifier and parameters.
func (p *provider) ListRecordsByZoneID(ctx context.Context, id string, params models.ListDNSRecordsParams) ([]models.DNSRecord, error) {
	// Fetch all records for a zone
	rrset, err := p.repo.ListDNSRecords(context.Background(), id)
	if err != nil {
		return []models.DNSRecord{}, err
	}

	return rrset, nil
}

// ListRecords returns a slice of DNS records for the given zone name.
func (p *provider) ListRecords(ctx context.Context, params models.ListDNSRecordsParams) ([]models.DNSRecord, error) {
	id, err := p.repo.ZoneIDByName(params.ZoneName)
	if err != nil {
		return []models.DNSRecord{}, err
	}

	rrset, err := p.ListRecordsByZoneID(ctx, id, params)
	if err != nil {
		return []models.DNSRecord{}, err
	}

	return filterDNSRecords(rrset, params), nil
}

func convFromDNSRecord(cfrr cloudflare.DNSRecord) models.DNSRecord {
	return models.DNSRecord{
		ID:      cfrr.ID,
		Name:    cfrr.Name,
		TTL:     cfrr.TTL,
		Type:    cfrr.Type,
		Proxied: cloudflare.Bool(cfrr.Proxied),
		Content: cfrr.Content,
	}
}

func convFromDNSRecords(cfrrset []cloudflare.DNSRecord) []models.DNSRecord {
	rrset := make([]models.DNSRecord, 0, len(cfrrset))
	for _, cfrr := range cfrrset {
		rr := models.DNSRecord{
			ID:      cfrr.ID,
			Name:    cfrr.Name,
			TTL:     cfrr.TTL,
			Type:    cfrr.Type,
			Proxied: cloudflare.Bool(cfrr.Proxied),
			Content: cfrr.Content,
		}
		rrset = append(rrset, rr)
	}

	return rrset
}

func convFromDNSZones(cfzones []cloudflare.Zone) []models.Zone {
	zones := make([]models.Zone, 0, len(cfzones))
	for _, z := range cfzones {
		zone := models.Zone{
			ID:          z.ID,
			Name:        z.Name,
			NameServers: z.NameServers,
			Status:      z.Status,
		}
		zones = append(zones, zone)
	}

	return zones
}

func convToCreateDNSRecordParams(p models.CreateDNSRecordParams) cloudflare.CreateDNSRecordParams {
	return cloudflare.CreateDNSRecordParams{
		Content:  p.Content,
		Name:     p.Name,
		Proxied:  cloudflare.BoolPtr(p.Proxied),
		TTL:      p.TTL,
		Type:     p.Type,
		ZoneName: p.ZoneName,
		ZoneID:   p.ZoneID,
	}
}

func convFromCreateDNSRecordParams(p cloudflare.CreateDNSRecordParams) models.CreateDNSRecordParams {
	return models.CreateDNSRecordParams{
		Content:  p.Content,
		Name:     p.Name,
		Proxied:  cloudflare.Bool(p.Proxied),
		TTL:      p.TTL,
		Type:     p.Type,
		ZoneName: p.ZoneName,
		ZoneID:   p.ZoneID,
	}
}

func convToUpdateDNSRecordParams(p models.UpdateDNSRecordParams) cloudflare.UpdateDNSRecordParams {
	return cloudflare.UpdateDNSRecordParams{
		Content: p.Content,
		ID:      p.ID,
		Name:    p.Name,
		Proxied: cloudflare.BoolPtr(p.Proxied),
		TTL:     p.TTL,
		Type:    p.Type,
	}
}
