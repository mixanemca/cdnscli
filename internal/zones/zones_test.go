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
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClient mock client for zones.
type MockClient struct {
	mock.Mock
}

// ListRecordsByZoneID returns a slice of DNS records for the given zone identifier and parameters.
func (m *MockClient) ListRecordsByZoneID(ctx context.Context, id string, params cloudflare.ListDNSRecordsParams) ([]cloudflare.DNSRecord, error) {
	recs, _, err := m.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(id), params)
	if err != nil {
		return nil, err
	}
	return recs, nil
}

func (m *MockClient) GetDNSRecord(ctx context.Context, rc *cloudflare.ResourceContainer, recordID string) (cloudflare.DNSRecord, error) {
	return cloudflare.DNSRecord{}, nil
}
func (m *MockClient) CreateDNSRecord(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateDNSRecordParams) (cloudflare.DNSRecord, error) {
	return cloudflare.DNSRecord{}, nil
}
func (m *MockClient) DeleteDNSRecord(ctx context.Context, rc *cloudflare.ResourceContainer, recordID string) error {
	return nil
}
func (m *MockClient) ListDNSRecords(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListDNSRecordsParams) ([]cloudflare.DNSRecord, *cloudflare.ResultInfo, error) {
	args := m.Called(ctx, rc, params)
	return args.Get(0).([]cloudflare.DNSRecord), args.Get(1).(*cloudflare.ResultInfo), args.Error(2)
}
func (m *MockClient) ListZones(ctx context.Context, z ...string) ([]cloudflare.Zone, error) {
	return nil, nil
}
func (m *MockClient) UpdateDNSRecord(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateDNSRecordParams) (cloudflare.DNSRecord, error) {
	return cloudflare.DNSRecord{}, nil
}
func (m *MockClient) ZoneIDByName(zoneName string) (string, error) {
	return "", nil
}

func TestListRecordsByZoneID(t *testing.T) {
	tests := []struct {
		name        string
		zone        string
		zoneID      string
		mockResp    []cloudflare.DNSRecord
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "successful retrieving set of DNS resource records",
			zone:   "example.com",
			zoneID: "12345",
			mockResp: []cloudflare.DNSRecord{
				{
					Name:    "test.example.com",
					Type:    "A",
					Content: "192.0.2.1",
					TTL:     3600,
				},
			},
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name:        "empty set of DNS resource records",
			zone:        "empty.com",
			zoneID:      "12345",
			mockResp:    []cloudflare.DNSRecord{},
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name:        "missing zone ID",
			zone:        "noexists.com",
			zoneID:      "",
			mockResp:    nil,
			wantErr:     true,
			expectedErr: cloudflare.ErrMissingZoneID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)

			mockClient.On("ListDNSRecords", mock.Anything, cloudflare.ZoneIdentifier(tt.zoneID), mock.Anything).
				Return(tt.mockResp, &cloudflare.ResultInfo{}, tt.expectedErr)

			ctx := context.Background()

			client := New(mockClient)

			result, err := client.ListRecordsByZoneID(ctx, tt.zoneID, cloudflare.ListDNSRecordsParams{
				Name: tt.zone,
			})

			if tt.wantErr {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.mockResp, result)

			mockClient.AssertExpectations(t)
		})
	}
}
