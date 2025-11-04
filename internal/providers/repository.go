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

// Repo repository of DNS zones and resource records.
type Repo interface {
	GetDNSRecord(ctx context.Context, zoneID, recordID string) (models.DNSRecord, error)
	CreateDNSRecord(ctx context.Context, params models.CreateDNSRecordParams) (models.DNSRecord, error)
	DeleteDNSRecord(ctx context.Context, zoneID, recordID string) error
	ListDNSRecords(ctx context.Context, id string) ([]models.DNSRecord, error)
	ListZones(ctx context.Context, z ...string) ([]models.Zone, error)
	UpdateDNSRecord(ctx context.Context, params models.UpdateDNSRecordParams) (models.DNSRecord, error)
	ZoneIDByName(zoneName string) (string, error)
}

type repoCloudFlare struct {
	api *cloudflare.API
}

// NewRepoCloudFlare creates a repository for CloudFlare provider.
func NewRepoCloudFlare(api *cloudflare.API) Repo {
	return &repoCloudFlare{
		api: api,
	}
}

func (r *repoCloudFlare) GetDNSRecord(ctx context.Context, zoneID, recordID string) (models.DNSRecord, error) {
	// Create a ResourceContainer for the zone
	rc := cloudflare.ResourceContainer{
		Level:      cloudflare.ZoneRouteLevel,
		Identifier: zoneID,
		Type:       cloudflare.ZoneType,
	}

	record, err := r.api.GetDNSRecord(ctx, &rc, recordID)
	if err != nil {
		return models.DNSRecord{}, err
	}

	return convFromDNSRecord(record), nil
}

func (r *repoCloudFlare) CreateDNSRecord(ctx context.Context, params models.CreateDNSRecordParams) (models.DNSRecord, error) {
	// Create a ResourceContainer for the zone
	rc := cloudflare.ResourceContainer{
		Level:      cloudflare.ZoneRouteLevel,
		Identifier: params.ZoneID,
		Type:       cloudflare.ZoneType,
	}

	rr, err := r.api.CreateDNSRecord(ctx, &rc, convToCreateDNSRecordParams(params))
	if err != nil {
		return models.DNSRecord{}, err
	}

	return convFromDNSRecord(rr), nil
}

func (r *repoCloudFlare) DeleteDNSRecord(ctx context.Context, zoneID, recordID string) error {
	// Create a ResourceContainer for the zone
	rc := cloudflare.ResourceContainer{
		Level:      cloudflare.ZoneRouteLevel,
		Identifier: zoneID,
		Type:       cloudflare.ZoneType,
	}

	return r.api.DeleteDNSRecord(ctx, &rc, recordID)
}

func (r *repoCloudFlare) ListDNSRecords(ctx context.Context, id string) ([]models.DNSRecord, error) {
	rrset, _, err := r.api.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(id), cloudflare.ListDNSRecordsParams{})
	if err != nil {
		return []models.DNSRecord{}, err
	}

	return convFromDNSRecords(rrset), nil
}

func (r *repoCloudFlare) ListZones(ctx context.Context, z ...string) ([]models.Zone, error) {
	zones, err := r.api.ListZones(ctx, z...)
	if err != nil {
		return []models.Zone{}, err
	}

	return convFromDNSZones(zones), nil
}

func (r *repoCloudFlare) UpdateDNSRecord(ctx context.Context, params models.UpdateDNSRecordParams) (models.DNSRecord, error) {
	// Create a ResourceContainer
	rc := cloudflare.ResourceContainer{
		Level:      cloudflare.ZoneRouteLevel,
		Identifier: params.ZoneID,
		Type:       cloudflare.ZoneType,
	}

	rr, err := r.api.UpdateDNSRecord(ctx, &rc, convToUpdateDNSRecordParams(params))
	if err != nil {
		return models.DNSRecord{}, err
	}

	return convFromDNSRecord(rr), nil
}

func (r *repoCloudFlare) ZoneIDByName(zoneName string) (string, error) {
	return r.api.ZoneIDByName(zoneName)
}
