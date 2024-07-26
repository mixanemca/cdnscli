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
	"strings"

	"github.com/cloudflare/cloudflare-go"
)

// AddRR creates a new DNS resource record for a zone.
func (c *client) AddRR(ctx context.Context, zone string, params cloudflare.CreateDNSRecordParams) (cloudflare.DNSRecord, error) {
	var rr cloudflare.DNSRecord

	zoneID, err := c.api.ZoneIDByName(zone)
	if err != nil {
		return rr, err
	}

	// Create a ResourceContainer for the zone
	rc := cloudflare.ResourceContainer{
		Level:      cloudflare.ZoneRouteLevel,
		Identifier: zoneID,
		Type:       cloudflare.ZoneType,
	}

	rr, err = c.api.CreateDNSRecord(ctx, &rc, params)
	if err != nil {
		return rr, err
	}

	return rr, nil
}

// DeleteRR deletes a DNS resource record from a given zone.
func (c *client) DeleteRR(ctx context.Context, zone string, rr cloudflare.DNSRecord) error {
	zoneID, err := c.api.ZoneIDByName(zone)
	if err != nil {
		return err
	}

	// Create a ResourceContainer for the zone
	rc := cloudflare.ResourceContainer{
		Level:      cloudflare.ZoneRouteLevel,
		Identifier: zoneID,
		Type:       cloudflare.ZoneType,
	}

	err = c.api.DeleteDNSRecord(ctx, &rc, rr.ID)
	if err != nil {
		return err
	}

	return nil
}

// GetRRByName returns a single DNS record for the given zone & record identifiers.
func (c *client) GetRRByName(ctx context.Context, zone, name string) (cloudflare.DNSRecord, error) {
	var rr cloudflare.DNSRecord

	zoneID, err := c.api.ZoneIDByName(zone)
	if err != nil {
		return rr, err
	}

	// Create a ResourceContainer for the zone
	rc := cloudflare.ResourceContainer{
		Level:      cloudflare.ZoneRouteLevel,
		Identifier: zoneID,
		Type:       cloudflare.ZoneType,
	}

	recs, _, err := c.api.ListDNSRecords(context.Background(), cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{
		Name: strings.Join([]string{name, zone}, "."),
	})
	if err != nil {
		return rr, err
	}

	// TODO: this code need to refactoring
	for _, rec := range recs {
		rr, err = c.api.GetDNSRecord(ctx, &rc, rec.ID)
		if err != nil {
			return rr, err
		}
	}

	return rr, nil
}

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

// ListRecords returns a slice of DNS records for the given zone identifier and parameters.
func (c *client) ListRecords(ctx context.Context, zone string, params cloudflare.ListDNSRecordsParams) ([]cloudflare.DNSRecord, error) {
	zoneID, err := c.api.ZoneIDByName(zone)
	if err != nil {
		return nil, err
	}

	// Fetch all records for a zone
	recs, _, err := c.api.ListDNSRecords(context.Background(), cloudflare.ZoneIdentifier(zoneID), params)
	if err != nil {
		return nil, err
	}

	return recs, nil
}
