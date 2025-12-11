/*
Copyright © 2024-2025 Michael Bruskov <mixanemca@yandex.ru>

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

	"github.com/mixanemca/cdnscli/internal/models"
	"github.com/mixanemca/regru-go"
)

type repoRegRu struct {
	client *regru.Client
}

// NewRepoRegRu creates a repository for RegRu provider.
func NewRepoRegRu(client *regru.Client) Repo {
	return &repoRegRu{
		client: client,
	}
}

func (r *repoRegRu) GetDNSRecord(ctx context.Context, zoneID, recordID string) (models.DNSRecord, error) {
	// Get zone name by ID
	zones, err := r.client.ListZones(ctx)
	if err != nil {
		return models.DNSRecord{}, err
	}

	var zoneName string
	for _, zone := range zones {
		if zone.ID == zoneID {
			zoneName = zone.Name
			break
		}
	}

	if zoneName == "" {
		return models.DNSRecord{}, fmt.Errorf("zone with ID %s not found", zoneID)
	}

	// List all records and find the one with matching ID
	params := regru.ListDNSRecordsParams{
		ZoneName: zoneName,
	}
	records, err := r.client.ListRecords(ctx, params)
	if err != nil {
		return models.DNSRecord{}, err
	}

	for _, rec := range records {
		if rec.ID == recordID {
			return convFromRegRuDNSRecord(rec), nil
		}
	}

	return models.DNSRecord{}, fmt.Errorf("record with ID %s not found", recordID)
}

func (r *repoRegRu) CreateDNSRecord(ctx context.Context, params models.CreateDNSRecordParams) (models.DNSRecord, error) {
	zoneName := params.ZoneName
	if zoneName == "" {
		// Try to get zone name from zone ID
		zones, err := r.client.ListZones(ctx)
		if err != nil {
			return models.DNSRecord{}, err
		}
		for _, zone := range zones {
			if zone.ID == params.ZoneID {
				zoneName = zone.Name
				break
			}
		}
		if zoneName == "" {
			return models.DNSRecord{}, fmt.Errorf("zone name or zone ID must be provided")
		}
	}

	createParams := regru.CreateDNSRecordParams{
		Name:    params.Name,
		Type:    params.Type,
		Content: params.Content,
		TTL:     params.TTL,
	}

	record, err := r.client.AddRR(ctx, zoneName, createParams)
	if err != nil {
		return models.DNSRecord{}, err
	}

	return convFromRegRuDNSRecord(record), nil
}

func (r *repoRegRu) DeleteDNSRecord(ctx context.Context, zoneID, recordID string) error {
	// Get zone name by ID
	zones, err := r.client.ListZones(ctx)
	if err != nil {
		return err
	}

	var zoneName string
	for _, zone := range zones {
		if zone.ID == zoneID {
			zoneName = zone.Name
			break
		}
	}

	if zoneName == "" {
		return fmt.Errorf("zone with ID %s not found", zoneID)
	}

	// Get record to delete
	record, err := r.GetDNSRecord(ctx, zoneID, recordID)
	if err != nil {
		return err
	}

	// Convert to regru DNSRecord format
	regruRecord := convToRegRuDNSRecord(record)

	return r.client.DeleteRR(ctx, zoneName, regruRecord)
}

func (r *repoRegRu) ListDNSRecords(ctx context.Context, id string) ([]models.DNSRecord, error) {
	// Get zone name by ID
	zones, err := r.client.ListZones(ctx)
	if err != nil {
		return []models.DNSRecord{}, err
	}

	var zoneName string
	for _, zone := range zones {
		if zone.ID == id {
			zoneName = zone.Name
			break
		}
	}

	if zoneName == "" {
		return []models.DNSRecord{}, fmt.Errorf("zone with ID %s not found", id)
	}

	params := regru.ListDNSRecordsParams{
		ZoneName: zoneName,
	}
	records, err := r.client.ListRecords(ctx, params)
	if err != nil {
		return []models.DNSRecord{}, err
	}

	return convFromRegRuDNSRecords(records), nil
}

func (r *repoRegRu) ListZones(ctx context.Context, z ...string) ([]models.Zone, error) {
	var zones []regru.Zone
	var err error

	if len(z) > 0 && z[0] != "" {
		// Filter by zone name
		zones, err = r.client.ListZonesByName(ctx, z[0])
	} else {
		// List all zones
		zones, err = r.client.ListZones(ctx)
	}

	if err != nil {
		return []models.Zone{}, err
	}

	return convFromRegRuZones(zones), nil
}

func (r *repoRegRu) UpdateDNSRecord(ctx context.Context, params models.UpdateDNSRecordParams) (models.DNSRecord, error) {
	zoneName := params.ZoneName
	if zoneName == "" {
		// Try to get zone name from zone ID
		zones, err := r.client.ListZones(ctx)
		if err != nil {
			return models.DNSRecord{}, err
		}
		for _, zone := range zones {
			if zone.ID == params.ZoneID {
				zoneName = zone.Name
				break
			}
		}
		if zoneName == "" {
			return models.DNSRecord{}, fmt.Errorf("zone name or zone ID must be provided")
		}
	}

	// Convert to regru DNSRecord format
	regruRecord := regru.DNSRecord{
		ID:      params.ID,
		Name:    params.Name,
		Type:    params.Type,
		Content: params.Content,
		TTL:     params.TTL,
	}

	record, err := r.client.UpdateRR(ctx, zoneName, regruRecord)
	if err != nil {
		return models.DNSRecord{}, err
	}

	return convFromRegRuDNSRecord(record), nil
}

func (r *repoRegRu) ZoneIDByName(zoneName string) (string, error) {
	ctx := context.Background()
	zones, err := r.client.ListZonesByName(ctx, zoneName)
	if err != nil {
		return "", err
	}

	if len(zones) == 0 {
		return "", fmt.Errorf("zone %s not found", zoneName)
	}

	// Return the first matching zone ID
	return zones[0].ID, nil
}

// Conversion functions

func convFromRegRuDNSRecord(rr regru.DNSRecord) models.DNSRecord {
	return models.DNSRecord{
		ID:      rr.ID,
		Name:    rr.Name,
		TTL:     rr.TTL,
		Type:    rr.Type,
		Content: rr.Content,
		Proxied: false, // RegRu doesn't support proxying
	}
}

func convFromRegRuDNSRecords(rrset []regru.DNSRecord) []models.DNSRecord {
	records := make([]models.DNSRecord, 0, len(rrset))
	for _, rr := range rrset {
		record := models.DNSRecord{
			ID:      rr.ID,
			Name:    rr.Name,
			TTL:     rr.TTL,
			Type:    rr.Type,
			Content: rr.Content,
			Proxied: false, // RegRu doesn't support proxying
		}
		records = append(records, record)
	}
	return records
}

func convFromRegRuZones(zones []regru.Zone) []models.Zone {
	result := make([]models.Zone, 0, len(zones))
	for _, z := range zones {
		zone := models.Zone{
			ID:          z.ID,
			Name:        z.Name,
			NameServers: z.NameServers,
			Status:      z.Status,
		}
		result = append(result, zone)
	}
	return result
}

func convToRegRuDNSRecord(rr models.DNSRecord) regru.DNSRecord {
	// Clean up name - remove trailing dot if present
	name := strings.TrimSuffix(rr.Name, ".")

	return regru.DNSRecord{
		ID:      rr.ID,
		Name:    name,
		Type:    rr.Type,
		Content: rr.Content,
		TTL:     rr.TTL,
	}
}
