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

	"github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/dns"
	"github.com/cloudflare/cloudflare-go/v7/zones"
	"github.com/mixanemca/cdnscli/internal/models"
	"golang.org/x/net/idna"
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
	client *cloudflare.Client
}

// NewRepoCloudFlare creates a repository for CloudFlare provider.
func NewRepoCloudFlare(client *cloudflare.Client) Repo {
	return &repoCloudFlare{
		client: client,
	}
}

func (r *repoCloudFlare) GetDNSRecord(ctx context.Context, zoneID, recordID string) (models.DNSRecord, error) {
	record, err := r.client.DNS.Records.Get(ctx, recordID, dns.RecordGetParams{
		ZoneID: cloudflare.F(zoneID),
	})
	if err != nil {
		return models.DNSRecord{}, err
	}

	return convFromDNSRecord(*record), nil
}

func (r *repoCloudFlare) CreateDNSRecord(ctx context.Context, params models.CreateDNSRecordParams) (models.DNSRecord, error) {
	rr, err := r.client.DNS.Records.New(ctx, convToCreateDNSRecordParams(params))
	if err != nil {
		return models.DNSRecord{}, err
	}

	return convFromDNSRecord(*rr), nil
}

func (r *repoCloudFlare) DeleteDNSRecord(ctx context.Context, zoneID, recordID string) error {
	_, err := r.client.DNS.Records.Delete(ctx, recordID, dns.RecordDeleteParams{
		ZoneID: cloudflare.F(zoneID),
	})

	return err
}

func (r *repoCloudFlare) ListDNSRecords(ctx context.Context, id string) ([]models.DNSRecord, error) {
	iter := r.client.DNS.Records.ListAutoPaging(ctx, dns.RecordListParams{
		ZoneID: cloudflare.F(id),
	})

	var rrset []dns.RecordResponse
	for iter.Next() {
		rrset = append(rrset, iter.Current())
	}
	if err := iter.Err(); err != nil {
		return []models.DNSRecord{}, err
	}

	return convFromDNSRecords(rrset), nil
}

func (r *repoCloudFlare) ListZones(ctx context.Context, z ...string) ([]models.Zone, error) {
	if len(z) == 0 {
		cfzones, err := r.listZones(ctx, zones.ZoneListParams{})
		if err != nil {
			return []models.Zone{}, err
		}

		return convFromDNSZones(cfzones), nil
	}

	// A name filter matches at most a handful of zones, so a single page per
	// name is enough (as in the v0 SDK) and avoids fetching an empty next page.
	var cfzones []zones.Zone
	for _, name := range z {
		page, err := r.client.Zones.List(ctx, zones.ZoneListParams{
			Name: cloudflare.F(normalizeZoneName(name)),
		})
		if err != nil {
			return []models.Zone{}, err
		}
		cfzones = append(cfzones, page.Result...)
	}

	return convFromDNSZones(cfzones), nil
}

func (r *repoCloudFlare) UpdateDNSRecord(ctx context.Context, params models.UpdateDNSRecordParams) (models.DNSRecord, error) {
	rr, err := r.client.DNS.Records.Update(
		ctx, params.ID, convToUpdateDNSRecordParams(params),
	)
	if err != nil {
		return models.DNSRecord{}, err
	}

	return convFromDNSRecord(*rr), nil
}

func (r *repoCloudFlare) ZoneIDByName(zoneName string) (string, error) {
	// Single page for the same reason as in ListZones with a name filter.
	page, err := r.client.Zones.List(context.Background(), zones.ZoneListParams{
		Name: cloudflare.F(normalizeZoneName(zoneName)),
	})
	if err != nil {
		return "", fmt.Errorf("zones list command failed: %w", err)
	}

	switch len(page.Result) {
	case 0:
		return "", errors.New("zone could not be found")
	case 1:
		return page.Result[0].ID, nil
	default:
		return "", errors.New("ambiguous zone name; an account ID might help")
	}
}

// listZones collects zones from all result pages for the given list parameters.
func (r *repoCloudFlare) listZones(
	ctx context.Context, params zones.ZoneListParams,
) ([]zones.Zone, error) {
	iter := r.client.Zones.ListAutoPaging(ctx, params)

	var result []zones.Zone
	for iter.Next() {
		result = append(result, iter.Current())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// normalizeZoneName converts an IDN zone name from Punycode to its Unicode
// form, which is how Cloudflare stores zone names. Names that are not valid
// Punycode are returned unchanged.
func normalizeZoneName(name string) string {
	if n, err := idna.ToUnicode(name); err == nil {
		return n
	}

	return name
}
