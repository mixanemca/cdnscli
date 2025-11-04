/*
Copyright © 2024 Michael Bruskov <mixanemca@yandex.ru>

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

	// TODO: this code need to refactoring
	for _, rec := range rrset {
		rr, err = p.repo.GetDNSRecord(ctx, zoneID, rec.ID)
		if err != nil {
			return rr, err
		}
	}

	return rr, nil
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

	return p.ListRecordsByZoneID(ctx, id, params)
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
