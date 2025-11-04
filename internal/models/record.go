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

// Package models holds an internal structs for DNS zones and records and also queries params.
package models

// DNSRecord represents a DNS record in a zone.
type DNSRecord struct {
	Content string `json:"content,omitempty"`
	ID      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Proxied bool   `json:"proxied,omitempty"`
	TTL     int    `json:"ttl,omitempty"`
	Type    string `json:"type,omitempty"`
}

// CreateDNSRecordParams params for creating DNS record.
type CreateDNSRecordParams struct {
	Content  string `json:"content,omitempty"`
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Proxied  bool   `json:"proxied,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	Type     string `json:"type,omitempty"`
	ZoneID   string `json:"zone_id,omitempty"`
	ZoneName string `json:"zone_name,omitempty"`
}

// UpdateDNSRecordParams params for updating DNS record.
type UpdateDNSRecordParams struct {
	Content  string `json:"content,omitempty"`
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Proxied  bool   `json:"proxied,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	Type     string `json:"type,omitempty"`
	ZoneID   string `json:"zone_id,omitempty"`
	ZoneName string `json:"zone_name,omitempty"`
}

// ListDNSRecordsParams params for list DNS records.
type ListDNSRecordsParams struct {
	Content  string `json:"content,omitempty"`
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Proxied  bool   `json:"proxied,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	Type     string `json:"type,omitempty"`
	ZoneID   string `json:"zone_id,omitempty"`
	ZoneName string `json:"zone_name,omitempty"`
}
