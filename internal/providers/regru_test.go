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
	"testing"

	"github.com/mixanemca/cdnscli/internal/models"
	"github.com/mixanemca/regru-go"
	"github.com/stretchr/testify/assert"
)

func TestConvFromRegRuDNSRecord(t *testing.T) {
	tests := []struct {
		name     string
		input    regru.DNSRecord
		expected models.DNSRecord
	}{
		{
			name: "Valid input",
			input: regru.DNSRecord{
				ID:      "record-id",
				Name:    "example.com",
				TTL:     3600,
				Type:    "A",
				Content: "192.168.0.1",
			},
			expected: models.DNSRecord{
				ID:      "record-id",
				Name:    "example.com",
				TTL:     3600,
				Type:    "A",
				Content: "192.168.0.1",
				Proxied: false, // RegRu doesn't support proxying
			},
		},
		{
			name:  "Empty input",
			input: regru.DNSRecord{},
			expected: models.DNSRecord{
				ID:      "",
				Name:    "",
				TTL:     0,
				Type:    "",
				Content: "",
				Proxied: false,
			},
		},
		{
			name: "CNAME record",
			input: regru.DNSRecord{
				ID:      "cname-id",
				Name:    "www.example.com",
				TTL:     7200,
				Type:    "CNAME",
				Content: "example.com",
			},
			expected: models.DNSRecord{
				ID:      "cname-id",
				Name:    "www.example.com",
				TTL:     7200,
				Type:    "CNAME",
				Content: "example.com",
				Proxied: false,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := convFromRegRuDNSRecord(test.input)
			assert.Equal(t, test.expected, result, "Test %s failed", test.name)
		})
	}
}

func TestConvFromRegRuDNSRecords(t *testing.T) {
	tests := []struct {
		name     string
		input    []regru.DNSRecord
		expected []models.DNSRecord
	}{
		{
			name: "Valid input with multiple records",
			input: []regru.DNSRecord{
				{
					ID:      "record-id-1",
					Name:    "example.com",
					TTL:     3600,
					Type:    "A",
					Content: "192.168.0.1",
				},
				{
					ID:      "record-id-2",
					Name:    "www.example.com",
					TTL:     7200,
					Type:    "CNAME",
					Content: "example.com",
				},
			},
			expected: []models.DNSRecord{
				{
					ID:      "record-id-1",
					Name:    "example.com",
					TTL:     3600,
					Type:    "A",
					Content: "192.168.0.1",
					Proxied: false,
				},
				{
					ID:      "record-id-2",
					Name:    "www.example.com",
					TTL:     7200,
					Type:    "CNAME",
					Content: "example.com",
					Proxied: false,
				},
			},
		},
		{
			name:     "Empty input",
			input:    []regru.DNSRecord{},
			expected: []models.DNSRecord{},
		},
		{
			name: "Single record",
			input: []regru.DNSRecord{
				{
					ID:      "record-id",
					Name:    "example.com",
					TTL:     3600,
					Type:    "A",
					Content: "192.168.0.1",
				},
			},
			expected: []models.DNSRecord{
				{
					ID:      "record-id",
					Name:    "example.com",
					TTL:     3600,
					Type:    "A",
					Content: "192.168.0.1",
					Proxied: false,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := convFromRegRuDNSRecords(test.input)
			assert.Equal(t, test.expected, result, "Test %s failed", test.name)
		})
	}
}

func TestConvFromRegRuZones(t *testing.T) {
	tests := []struct {
		name     string
		input    []regru.Zone
		expected []models.Zone
	}{
		{
			name: "Valid input with multiple zones",
			input: []regru.Zone{
				{
					ID:          "zone-id-1",
					Name:        "example.com",
					NameServers: []string{"ns1.example.com", "ns2.example.com"},
					Status:      "active",
				},
				{
					ID:          "zone-id-2",
					Name:        "test.com",
					NameServers: []string{"ns1.test.com"},
					Status:      "active",
				},
			},
			expected: []models.Zone{
				{
					ID:          "zone-id-1",
					Name:        "example.com",
					NameServers: []string{"ns1.example.com", "ns2.example.com"},
					Status:      "active",
				},
				{
					ID:          "zone-id-2",
					Name:        "test.com",
					NameServers: []string{"ns1.test.com"},
					Status:      "active",
				},
			},
		},
		{
			name:     "Empty input",
			input:    []regru.Zone{},
			expected: []models.Zone{},
		},
		{
			name: "Single zone",
			input: []regru.Zone{
				{
					ID:          "zone-id",
					Name:        "example.com",
					NameServers: []string{"ns1.example.com", "ns2.example.com"},
					Status:      "active",
				},
			},
			expected: []models.Zone{
				{
					ID:          "zone-id",
					Name:        "example.com",
					NameServers: []string{"ns1.example.com", "ns2.example.com"},
					Status:      "active",
				},
			},
		},
		{
			name: "Zone without name servers",
			input: []regru.Zone{
				{
					ID:          "zone-id",
					Name:        "example.com",
					NameServers: []string{},
					Status:      "pending",
				},
			},
			expected: []models.Zone{
				{
					ID:          "zone-id",
					Name:        "example.com",
					NameServers: []string{},
					Status:      "pending",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := convFromRegRuZones(test.input)
			assert.Equal(t, test.expected, result, "Test %s failed", test.name)
		})
	}
}

func TestConvToRegRuDNSRecord(t *testing.T) {
	tests := []struct {
		name     string
		input    models.DNSRecord
		expected regru.DNSRecord
	}{
		{
			name: "Valid input",
			input: models.DNSRecord{
				ID:      "record-id",
				Name:    "example.com",
				TTL:     3600,
				Type:    "A",
				Content: "192.168.0.1",
				Proxied: true, // Should be ignored in RegRu
			},
			expected: regru.DNSRecord{
				ID:      "record-id",
				Name:    "example.com",
				Type:    "A",
				Content: "192.168.0.1",
				TTL:     3600,
			},
		},
		{
			name:  "Empty input",
			input: models.DNSRecord{},
			expected: regru.DNSRecord{
				ID:      "",
				Name:    "",
				Type:    "",
				Content: "",
				TTL:     0,
			},
		},
		{
			name: "Name with trailing dot",
			input: models.DNSRecord{
				ID:      "record-id",
				Name:    "example.com.",
				TTL:     3600,
				Type:    "A",
				Content: "192.168.0.1",
			},
			expected: regru.DNSRecord{
				ID:      "record-id",
				Name:    "example.com", // Trailing dot should be removed
				Type:    "A",
				Content: "192.168.0.1",
				TTL:     3600,
			},
		},
		{
			name: "Name without trailing dot",
			input: models.DNSRecord{
				ID:      "record-id",
				Name:    "example.com",
				TTL:     3600,
				Type:    "A",
				Content: "192.168.0.1",
			},
			expected: regru.DNSRecord{
				ID:      "record-id",
				Name:    "example.com",
				Type:    "A",
				Content: "192.168.0.1",
				TTL:     3600,
			},
		},
		{
			name: "CNAME record",
			input: models.DNSRecord{
				ID:      "cname-id",
				Name:    "www.example.com",
				TTL:     7200,
				Type:    "CNAME",
				Content: "example.com",
			},
			expected: regru.DNSRecord{
				ID:      "cname-id",
				Name:    "www.example.com",
				Type:    "CNAME",
				Content: "example.com",
				TTL:     7200,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := convToRegRuDNSRecord(test.input)
			assert.Equal(t, test.expected, result, "Test %s failed", test.name)
		})
	}
}
