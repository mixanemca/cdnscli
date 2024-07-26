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

package zones

import (
	"context"

	"github.com/cloudflare/cloudflare-go"
)

type Client interface {
	// AddRR creates a new DNS resource record for a given zone.
	AddRR(ctx context.Context, zone string, params cloudflare.CreateDNSRecordParams) (cloudflare.DNSRecord, error)
	// DeleteRR deletes a DNS resource record from a given zone.
	DeleteRR(ctx context.Context, zone string, rr cloudflare.DNSRecord) error
	// GetRRByName returns a single DNS resource record for the given zone & record identifiers.
	GetRRByName(ctx context.Context, zone, name string) (cloudflare.DNSRecord, error)
	// List lists the zones on an account.
	List(ctx context.Context) ([]cloudflare.Zone, error)
	// ListByName lists the zone in an account using the zone name for filtering.
	ListByName(ctx context.Context, name string) ([]cloudflare.Zone, error)
	// ListRecords returns a slice of DNS records for the given zone identifier.
	ListRecords(ctx context.Context, zone string, params cloudflare.ListDNSRecordsParams) ([]cloudflare.DNSRecord, error)
}
