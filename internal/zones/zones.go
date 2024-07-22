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

// List return lists zones on an account.
func (c *client) List(ctx context.Context) ([]cloudflare.Zone, error) {
	zones, err := c.api.ListZones(ctx)
	if err != nil {
		return nil, err
	}

	return zones, nil
}

// ListByName return lists zones on an account using the zone name for filtering.
func (c *client) ListByName(ctx context.Context, name string) ([]cloudflare.Zone, error) {
	zones, err := c.api.ListZones(ctx, name)
	if err != nil {
		return nil, err
	}

	return zones, nil
}

// ListRecords returns a slice of DNS records for the given zone identifier.
func (c *client) ListRecords(ctx context.Context, zone string) ([]cloudflare.DNSRecord, error) {
	zoneID, err := c.api.ZoneIDByName(zone)
	if err != nil {
		return nil, err
	}

	// Fetch all records for a zone
	recs, _, err := c.api.ListDNSRecords(context.Background(), cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{})
	if err != nil {
		return nil, err
	}

	return recs, nil
}
