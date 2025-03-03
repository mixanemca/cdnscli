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

// Package providers holds DNS providers and interface for works with its.
package providers

import (
	"context"

	"github.com/mixanemca/cfdnscli/internal/models"
)

// Provider exposes methods for manage DNS.
type Provider interface {
	// AddRR creates a new DNS resource record for a given zone.
	AddRR(ctx context.Context, zone string, params models.CreateDNSRecordParams) (models.DNSRecord, error)
	// DeleteRR deletes a DNS resource record from a given zone.
	DeleteRR(ctx context.Context, zone string, rr models.DNSRecord) error
	// GetRRByName returns a single DNS resource record for the given zone & record identifiers.
	GetRRByName(ctx context.Context, zone, name string) (models.DNSRecord, error)
	// ListZones lists the zones on an account.
	ListZones(ctx context.Context) ([]models.Zone, error)
	// ListZonesByName lists the zone in an account using the zone name for filtering.
	ListZonesByName(ctx context.Context, name string) ([]models.Zone, error)
	// ListRecords returns a slice of DNS records for the given zone name.
	ListRecords(ctx context.Context, params models.ListDNSRecordsParams) ([]models.DNSRecord, error)
	// ListRecordsByZoneID returns a slice of DNS records for the given zone identifier.
	ListRecordsByZoneID(ctx context.Context, id string, params models.ListDNSRecordsParams) ([]models.DNSRecord, error)
	// UpdateRR updates and returns an existing DNS resource record.
	UpdateRR(ctx context.Context, zone string, rr models.DNSRecord) (models.DNSRecord, error)
}

type provider struct {
	repo Repo
}

// NewProvider creates a new provider.
func NewProvider(repo Repo) Provider {
	return &provider{
		repo: repo,
	}
}
