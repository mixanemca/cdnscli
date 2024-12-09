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
	"errors"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/mixanemca/cfdnscli/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClient mock client for zones.
type MockClient struct {
	mock.Mock
}

// ListRecordsByZoneID returns a slice of DNS records for the given zone identifier and parameters.
func (m *MockClient) ListRecordsByZoneID(ctx context.Context, id string, params models.ListDNSRecordsParams) ([]models.DNSRecord, error) {
	/*
		rrset, err := m.ListDNSRecords(ctx, id)
		if err != nil {
			return []models.DNSRecord{}, err
		}
		return rrset, nil
	*/
	args := m.Called(ctx, id, params)
	return args.Get(0).([]models.DNSRecord), args.Error(1)
}

// ListRecordsByZoneName returns a slice of DNS records for the given zone name and parameters.
func (m *MockClient) ListRecords(ctx context.Context, params models.ListDNSRecordsParams) ([]models.DNSRecord, error) {
	/*
		id, err := m.ZoneIDByName(params.ZoneName)
		if err != nil {
			return []models.DNSRecord{}, err
		}

		return m.ListRecordsByZoneID(ctx, id, params)
	*/
	args := m.Called(ctx, params)
	return args.Get(0).([]models.DNSRecord), args.Error(1)
}

func (m *MockClient) GetDNSRecord(ctx context.Context, zoneID, recordID string) (models.DNSRecord, error) {
	args := m.Called(ctx, zoneID, recordID)
	return args.Get(0).(models.DNSRecord), args.Error(1)
}

func (m *MockClient) CreateDNSRecord(ctx context.Context, params models.CreateDNSRecordParams) (models.DNSRecord, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(models.DNSRecord), args.Error(1)
}

func (m *MockClient) DeleteDNSRecord(ctx context.Context, zoneID, recordID string) error {
	args := m.Called(ctx, zoneID, recordID)
	return args.Error(0)
}

func (m *MockClient) ListDNSRecords(ctx context.Context, id string) ([]models.DNSRecord, error) {
	args := m.Called(ctx, id)
	return args.Get(0).([]models.DNSRecord), args.Error(1)
}

func (m *MockClient) ListZones(ctx context.Context, z ...string) ([]models.Zone, error) {
	args := m.Called(ctx, mock.Anything)
	return args.Get(0).([]models.Zone), args.Error(1)
}

func (m *MockClient) UpdateDNSRecord(ctx context.Context, params models.UpdateDNSRecordParams) (models.DNSRecord, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(models.DNSRecord), args.Error(1)
}

func (m *MockClient) ZoneIDByName(zoneName string) (string, error) {
	args := m.Called(zoneName)
	return args.Get(0).(string), args.Error(1)
}

func TestListRecordsByZoneID(t *testing.T) {
	tests := []struct {
		name        string
		zone        string
		mockParams  models.ListDNSRecordsParams
		mockResp    []models.DNSRecord
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful retrieving set of DNS resource records",
			zone: "example.com",
			mockParams: models.ListDNSRecordsParams{
				ZoneID:   "12345",
				ZoneName: "example.com",
			},
			mockResp: []models.DNSRecord{
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
			name: "empty set of DNS resource records",
			zone: "empty.com",
			mockParams: models.ListDNSRecordsParams{
				ZoneID:   "12345",
				ZoneName: "example.com",
			},
			mockResp:    []models.DNSRecord{},
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name: "missing zone ID",
			zone: "noexists.com",
			mockParams: models.ListDNSRecordsParams{
				ZoneID:   "",
				ZoneName: "example.com",
			},
			mockResp:    []models.DNSRecord{},
			wantErr:     true,
			expectedErr: errors.New("required missing zone ID"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)

			mockClient.On("ListDNSRecords", mock.Anything, tt.mockParams.ZoneID).
				Return(tt.mockResp, tt.expectedErr)

			ctx := context.Background()

			client := NewProvider(mockClient)

			result, err := client.ListRecordsByZoneID(ctx, tt.mockParams.ZoneID, tt.mockParams)

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

func TestListRecords(t *testing.T) {
	tests := []struct {
		name        string
		mockParams  models.ListDNSRecordsParams
		mockResp    []models.DNSRecord
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful retrieving set of DNS resource records",
			mockParams: models.ListDNSRecordsParams{
				ZoneID:   "12345",
				ZoneName: "example.com",
			},
			mockResp: []models.DNSRecord{
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
			name: "empty set of DNS resource records",
			mockParams: models.ListDNSRecordsParams{
				ZoneID:   "12345",
				ZoneName: "example.com",
			},
			mockResp:    []models.DNSRecord{},
			wantErr:     false,
			expectedErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)

			mockClient.On("ZoneIDByName", tt.mockParams.ZoneName).
				Return(tt.mockParams.ZoneID, nil)
			mockClient.On("ListDNSRecords", mock.Anything, tt.mockParams.ZoneID).
				Return(tt.mockResp, nil)

			ctx := context.Background()

			client := NewProvider(mockClient)

			result, err := client.ListRecords(ctx, tt.mockParams)

			if tt.wantErr {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.mockResp, result)

			mockClient.AssertExpectations(t)
		})
	}

	testsErrors := []struct {
		name        string
		zone        string
		zoneID      string
		mockParams  models.ListDNSRecordsParams
		mockResp    []models.DNSRecord
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "missing zone name",
			zone:        "",
			zoneID:      "",
			mockParams:  models.ListDNSRecordsParams{},
			mockResp:    []models.DNSRecord{},
			wantErr:     true,
			expectedErr: errors.New("zone could not be found"),
		},
	}
	for _, tt := range testsErrors {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)

			mockClient.On("ZoneIDByName", tt.zone).
				Return(tt.zoneID, tt.expectedErr)

			ctx := context.Background()

			provider := NewProvider(mockClient)

			result, err := provider.ListRecords(ctx, tt.mockParams)

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

func TestListZones(t *testing.T) {
	tests := []struct {
		name        string
		zone        string
		zoneID      string
		mockResp    []models.Zone
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "successful retrieving list of DNS zones",
			zone:   "example.com",
			zoneID: "12345",
			mockResp: []models.Zone{
				{
					ID:   "12345",
					Name: "example.com",
					NameServers: []string{
						"ns1.example.com",
						"ns2.example.com",
					},
					Status: "active",
				},
			},
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name:        "empty zone name",
			zone:        "",
			zoneID:      "",
			mockResp:    []models.Zone{},
			wantErr:     true,
			expectedErr: errors.New(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)

			mockClient.On("ListZones", mock.Anything).
				Return(tt.mockResp, tt.expectedErr)

			ctx := context.Background()
			provider := NewProvider(mockClient)

			result, err := provider.ListZones(ctx)
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

func TestListZonesByName(t *testing.T) {
	tests := []struct {
		name        string
		zone        string
		zoneID      string
		mockResp    []models.Zone
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "successful retrieving list of DNS zones",
			zone:   "example.com",
			zoneID: "12345",
			mockResp: []models.Zone{
				{
					ID:   "12345",
					Name: "example.com",
					NameServers: []string{
						"ns1.example.com",
						"ns2.example.com",
					},
					Status: "active",
				},
			},
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name:        "empty zone name",
			zone:        "",
			zoneID:      "",
			mockResp:    []models.Zone{},
			wantErr:     true,
			expectedErr: errors.New(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)

			mockClient.On("ListZones", mock.Anything, tt.name).
				Return(tt.mockResp, tt.expectedErr)

			ctx := context.Background()
			provider := NewProvider(mockClient)

			result, err := provider.ListZonesByName(ctx, tt.name)
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

func TestGetRRByName(t *testing.T) {
	tests := []struct {
		name          string
		zone          string
		zoneID        string
		record        string
		recordID      string
		mockResp      models.DNSRecord
		mockRespRRSet []models.DNSRecord
		wantErr       bool
		expectedErr   error
	}{
		{
			name:     "successful retrieving list of DNS zones",
			zone:     "example.com",
			zoneID:   "12345",
			record:   "test.example.com",
			recordID: "67890",
			mockResp: models.DNSRecord{
				ID:   "67890",
				Name: "test.example.com",
			},
			mockRespRRSet: []models.DNSRecord{
				{
					ID:   "67890",
					Name: "test.example.com",
				},
			},
			wantErr:     false,
			expectedErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)

			mockClient.On("ZoneIDByName", tt.zone).
				Return(tt.zoneID, tt.expectedErr)
			mockClient.On("ListDNSRecords", mock.Anything, tt.zoneID).
				Return(tt.mockRespRRSet, tt.expectedErr)
			mockClient.On("GetDNSRecord", mock.Anything, tt.zoneID, tt.recordID).
				Return(tt.mockResp, tt.expectedErr)

			ctx := context.Background()
			provider := NewProvider(mockClient)

			result, err := provider.GetRRByName(ctx, tt.zone, tt.record)
			if tt.wantErr {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.mockResp, result)

			mockClient.AssertExpectations(t)
		})
	}

	testsErrors := []struct {
		name          string
		zone          string
		zoneID        string
		record        string
		recordID      string
		mockResp      models.DNSRecord
		mockRespRRSet []models.DNSRecord
		wantErr       bool
		expectedErr   error
	}{
		{
			name:          "empty zone name",
			zone:          "",
			zoneID:        "12345",
			recordID:      "67890",
			mockResp:      models.DNSRecord{},
			mockRespRRSet: []models.DNSRecord{},
			wantErr:       true,
			expectedErr:   errors.New("zone could not be found"),
		},
	}

	for _, tt := range testsErrors {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)

			mockClient.On("ZoneIDByName", tt.zone).
				Return(tt.zoneID, tt.expectedErr)

			ctx := context.Background()
			provider := NewProvider(mockClient)

			result, err := provider.GetRRByName(ctx, tt.zone, tt.record)
			if tt.wantErr {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.mockResp, result)

			mockClient.AssertExpectations(t)
		})
	}

	testsErrors = []struct {
		name          string
		zone          string
		zoneID        string
		record        string
		recordID      string
		mockResp      models.DNSRecord
		mockRespRRSet []models.DNSRecord
		wantErr       bool
		expectedErr   error
	}{
		{
			name:          "empty zone id",
			zone:          "example.com",
			zoneID:        "12345",
			recordID:      "67890",
			mockResp:      models.DNSRecord{},
			mockRespRRSet: []models.DNSRecord{},
			wantErr:       true,
			expectedErr:   errors.New("required missing zone ID"),
		},
	}

	for _, tt := range testsErrors {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)

			mockClient.On("ZoneIDByName", tt.zone).
				Return(tt.zoneID, nil)
			mockClient.On("ListDNSRecords", mock.Anything, mock.Anything).
				Return(tt.mockRespRRSet, tt.expectedErr)

			ctx := context.Background()
			provider := NewProvider(mockClient)

			result, err := provider.GetRRByName(ctx, tt.zone, tt.record)
			if tt.wantErr {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.mockResp, result)

			mockClient.AssertExpectations(t)
		})
	}

	testsErrors = []struct {
		name          string
		zone          string
		zoneID        string
		record        string
		recordID      string
		mockResp      models.DNSRecord
		mockRespRRSet []models.DNSRecord
		wantErr       bool
		expectedErr   error
	}{
		{
			name:     "empty zone id",
			zone:     "example.com",
			zoneID:   "12345",
			recordID: "67890",
			mockResp: models.DNSRecord{},
			mockRespRRSet: []models.DNSRecord{
				{
					ID:   "67890",
					Name: "test.example.com",
				},
			},
			wantErr:     true,
			expectedErr: errors.New("required DNS record ID missing"),
		},
	}

	for _, tt := range testsErrors {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)

			mockClient.On("ZoneIDByName", tt.zone).
				Return(tt.zoneID, nil)
			mockClient.On("ListDNSRecords", mock.Anything, tt.zoneID).
				Return(tt.mockRespRRSet, nil)
			mockClient.On("GetDNSRecord", mock.Anything, tt.zoneID, mock.Anything).
				Return(tt.mockResp, tt.expectedErr)

			ctx := context.Background()
			provider := NewProvider(mockClient)

			result, err := provider.GetRRByName(ctx, tt.zone, tt.record)
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

func TestCreateDNSRecord(t *testing.T) {
	tests := []struct {
		name        string
		zone        string
		zoneID      string
		mockParams  models.CreateDNSRecordParams
		mockResp    models.DNSRecord
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "create a DNS resource record successful",
			zone:   "example.com",
			zoneID: "12345",
			mockParams: models.CreateDNSRecordParams{
				Content:  "192.0.2.1",
				Name:     "test.example.com",
				Proxied:  true,
				TTL:      60,
				Type:     "A",
				ZoneName: "example.com",
				ZoneID:   "12345",
			},
			mockResp: models.DNSRecord{
				Name:    "test.example.com",
				Proxied: true,
				TTL:     60,
				Type:    "A",
				Content: "192.0.2.1",
			},
			wantErr:     false,
			expectedErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)

			mockClient.On("ZoneIDByName", tt.zone).
				Return(tt.zoneID, nil)
			mockClient.On("CreateDNSRecord", mock.Anything, tt.mockParams).
				Return(tt.mockResp, nil)

			ctx := context.Background()
			provider := NewProvider(mockClient)

			result, err := provider.AddRR(ctx, tt.zone, tt.mockParams)
			if tt.wantErr {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.mockResp, result)

			mockClient.AssertExpectations(t)
		})
	}
	testsErrors := []struct {
		name        string
		zone        string
		zoneID      string
		mockParams  models.CreateDNSRecordParams
		mockResp    models.DNSRecord
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "create a DNS resource record without zone name error",
			zone:   "",
			zoneID: "12345",
			mockParams: models.CreateDNSRecordParams{
				Content:  "192.0.2.1",
				Name:     "test.example.com",
				Proxied:  true,
				TTL:      60,
				Type:     "A",
				ZoneName: "example.com",
				ZoneID:   "12345",
			},
			mockResp:    models.DNSRecord{},
			wantErr:     true,
			expectedErr: errors.New("zone could not be found"),
		},
	}
	for _, tt := range testsErrors {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)

			mockClient.On("ZoneIDByName", tt.zone).
				Return(tt.zoneID, tt.expectedErr)

			ctx := context.Background()
			provider := NewProvider(mockClient)

			result, err := provider.AddRR(ctx, tt.zone, tt.mockParams)
			if tt.wantErr {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.mockResp, result)

			mockClient.AssertExpectations(t)
		})
	}

	testsErrors = []struct {
		name        string
		zone        string
		zoneID      string
		mockParams  models.CreateDNSRecordParams
		mockResp    models.DNSRecord
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "create a DNS resource record without zone name error",
			zone:   "",
			zoneID: "12345",
			mockParams: models.CreateDNSRecordParams{
				Content:  "192.0.2.1",
				Name:     "test.example.com",
				Proxied:  true,
				TTL:      60,
				Type:     "A",
				ZoneName: "example.com",
				ZoneID:   "12345",
			},
			mockResp:    models.DNSRecord{},
			wantErr:     true,
			expectedErr: errors.New(""),
		},
	}
	for _, tt := range testsErrors {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)

			mockClient.On("ZoneIDByName", tt.zone).
				Return(tt.zoneID, nil)
			mockClient.On("CreateDNSRecord", mock.Anything, tt.mockParams).
				Return(tt.mockResp, tt.expectedErr)

			ctx := context.Background()
			provider := NewProvider(mockClient)

			result, err := provider.AddRR(ctx, tt.zone, tt.mockParams)
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

func TestDeleteDNSRecord(t *testing.T) {
	tests := []struct {
		name        string
		zone        string
		zoneID      string
		mockRecord  models.DNSRecord
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "delete a DNS resource record successful",
			zone:   "example.com",
			zoneID: "12345",
			mockRecord: models.DNSRecord{
				Name:    "test.example.com",
				Proxied: true,
				TTL:     60,
				Type:    "A",
				Content: "192.0.2.1",
			},
			wantErr:     false,
			expectedErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)

			mockClient.On("ZoneIDByName", tt.zone).
				Return(tt.zoneID, nil)
			mockClient.On("DeleteDNSRecord", mock.Anything, tt.zoneID, tt.mockRecord.ID).
				Return(tt.expectedErr)

			ctx := context.Background()
			provider := NewProvider(mockClient)

			err := provider.DeleteRR(ctx, tt.zone, tt.mockRecord)
			if tt.wantErr {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}

	testsErrors := []struct {
		name        string
		zone        string
		zoneID      string
		mockRecord  models.DNSRecord
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "delete a DNS resource record without zone name error",
			zone:        "",
			zoneID:      "12345",
			mockRecord:  models.DNSRecord{},
			wantErr:     true,
			expectedErr: errors.New("zone could not be found"),
		},
	}
	for _, tt := range testsErrors {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)

			mockClient.On("ZoneIDByName", tt.zone).
				Return(tt.zoneID, tt.expectedErr)

			ctx := context.Background()
			provider := NewProvider(mockClient)

			err := provider.DeleteRR(ctx, tt.zone, tt.mockRecord)
			if tt.wantErr {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}

	testsErrors = []struct {
		name        string
		zone        string
		zoneID      string
		mockRecord  models.DNSRecord
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "delete a DNS resource record without record ID error",
			zone:   "example.com",
			zoneID: "12345",
			mockRecord: models.DNSRecord{
				ID:      "",
				Name:    "test.example.com",
				Proxied: true,
				TTL:     60,
				Type:    "A",
				Content: "192.0.2.1",
			},
			wantErr:     true,
			expectedErr: errors.New("required DNS record ID missing"),
		},
	}
	for _, tt := range testsErrors {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)

			mockClient.On("ZoneIDByName", tt.zone).
				Return(tt.zoneID, nil)
			mockClient.On("DeleteDNSRecord", mock.Anything, tt.zoneID, tt.mockRecord.ID).
				Return(tt.expectedErr)

			ctx := context.Background()
			provider := NewProvider(mockClient)

			err := provider.DeleteRR(ctx, tt.zone, tt.mockRecord)
			if tt.wantErr {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestUpdateteDNSRecord(t *testing.T) {
	tests := []struct {
		name        string
		zone        string
		zoneID      string
		mockParams  models.UpdateDNSRecordParams
		mockResp    models.DNSRecord
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "update a DNS resource record successful",
			zone:   "example.com",
			zoneID: "12345",
			mockParams: models.UpdateDNSRecordParams{
				Name:    "test.example.com",
				Proxied: true,
				TTL:     60,
				Type:    "A",
				Content: "192.0.2.1",
				ZoneID:  "12345",
			},
			mockResp: models.DNSRecord{
				Name:    "test.example.com",
				Proxied: true,
				TTL:     60,
				Type:    "A",
				Content: "192.0.2.1",
			},
			wantErr:     false,
			expectedErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)

			mockClient.On("ZoneIDByName", tt.zone).
				Return(tt.zoneID, nil)
			mockClient.On("UpdateDNSRecord", mock.Anything, tt.mockParams).
				Return(tt.mockResp, tt.expectedErr)

			ctx := context.Background()
			provider := NewProvider(mockClient)

			result, err := provider.UpdateRR(ctx, tt.zone, tt.mockResp)
			if tt.wantErr {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.mockResp, result)

			mockClient.AssertExpectations(t)
		})
	}

	testsErrors := []struct {
		name        string
		zone        string
		zoneID      string
		mockParams  models.UpdateDNSRecordParams
		mockResp    models.DNSRecord
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "update a DNS resource record without zone name error",
			zone:        "",
			zoneID:      "12345",
			mockParams:  models.UpdateDNSRecordParams{},
			mockResp:    models.DNSRecord{},
			wantErr:     true,
			expectedErr: errors.New("zone could not be found"),
		},
	}
	for _, tt := range testsErrors {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)

			mockClient.On("ZoneIDByName", tt.zone).
				Return(tt.zoneID, tt.expectedErr)

			ctx := context.Background()
			provider := NewProvider(mockClient)

			result, err := provider.UpdateRR(ctx, tt.zone, tt.mockResp)
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

func TestConvFromDNSRecord(t *testing.T) {
	tests := []struct {
		name     string
		input    cloudflare.DNSRecord
		expected models.DNSRecord
	}{
		{
			name: "Valid input",
			input: cloudflare.DNSRecord{
				ID:      "record-id",
				Name:    "example.com",
				TTL:     3600,
				Type:    "A",
				Proxied: cloudflare.BoolPtr(true),
				Content: "192.168.0.1",
			},
			expected: models.DNSRecord{
				ID:      "record-id",
				Name:    "example.com",
				TTL:     3600,
				Type:    "A",
				Proxied: true,
				Content: "192.168.0.1",
			},
		},
		{
			name:  "Empty input",
			input: cloudflare.DNSRecord{},
			expected: models.DNSRecord{
				ID:      "",
				Name:    "",
				TTL:     0,
				Type:    "",
				Proxied: false,
				Content: "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := convFromDNSRecord(test.input)
			assert.Equal(t, test.expected, result, "Test %s failed", test.name)
		})
	}
}

func TestConvFromDNSRecords(t *testing.T) {
	tests := []struct {
		name     string
		input    []cloudflare.DNSRecord
		expected []models.DNSRecord
	}{
		{
			name: "Valid input",
			input: []cloudflare.DNSRecord{
				{
					ID:      "record-id",
					Name:    "example.com",
					TTL:     3600,
					Type:    "A",
					Proxied: cloudflare.BoolPtr(true),
					Content: "192.168.0.1",
				},
			},
			expected: []models.DNSRecord{
				{
					ID:      "record-id",
					Name:    "example.com",
					TTL:     3600,
					Type:    "A",
					Proxied: true,
					Content: "192.168.0.1",
				},
			},
		},
		{
			name:     "Empty input",
			input:    []cloudflare.DNSRecord{},
			expected: []models.DNSRecord{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := convFromDNSRecords(test.input)
			assert.Equal(t, test.expected, result, "Test %s failed", test.name)
		})
	}
}

func TestConvFromDNSZones(t *testing.T) {
	tests := []struct {
		name     string
		input    []cloudflare.Zone
		expected []models.Zone
	}{
		{
			name: "Valid input",
			input: []cloudflare.Zone{
				{
					ID:          "zone-id",
					Name:        "example-zone",
					NameServers: []string{"ns1.example.com", "ns2.example.com"},
					Status:      "active",
				},
			},
			expected: []models.Zone{
				{
					ID:          "zone-id",
					Name:        "example-zone",
					NameServers: []string{"ns1.example.com", "ns2.example.com"},
					Status:      "active",
				},
			},
		},
		{
			name:     "Empty input",
			input:    []cloudflare.Zone{},
			expected: []models.Zone{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := convFromDNSZones(test.input)
			assert.Equal(t, test.expected, result, "Test %s failed", test.name)
		})
	}
}

func TestConvToCreateDNSRecordParams(t *testing.T) {
	tests := []struct {
		name     string
		input    models.CreateDNSRecordParams
		expected cloudflare.CreateDNSRecordParams
	}{
		{
			name: "Valid input",
			input: models.CreateDNSRecordParams{
				Content:  "192.168.0.1",
				Name:     "example.com",
				Proxied:  true,
				TTL:      3600,
				Type:     "A",
				ZoneName: "example-zone",
				ZoneID:   "zone-id",
			},
			expected: cloudflare.CreateDNSRecordParams{
				Content:  "192.168.0.1",
				Name:     "example.com",
				Proxied:  cloudflare.BoolPtr(true),
				TTL:      3600,
				Type:     "A",
				ZoneName: "example-zone",
				ZoneID:   "zone-id",
			},
		},
		{
			name:  "Empty input",
			input: models.CreateDNSRecordParams{},
			expected: cloudflare.CreateDNSRecordParams{
				Content:  "",
				Name:     "",
				Proxied:  cloudflare.BoolPtr(false),
				TTL:      0,
				Type:     "",
				ZoneName: "",
				ZoneID:   "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := convToCreateDNSRecordParams(test.input)
			assert.Equal(t, test.expected, result, "Test %s failed", test.name)
		})
	}
}

func TestConvFromCreateDNSRecordParams(t *testing.T) {
	tests := []struct {
		name     string
		input    cloudflare.CreateDNSRecordParams
		expected models.CreateDNSRecordParams
	}{
		{
			name: "Valid input",
			input: cloudflare.CreateDNSRecordParams{
				Content:  "192.168.0.1",
				Name:     "example.com",
				Proxied:  cloudflare.BoolPtr(true),
				TTL:      3600,
				Type:     "A",
				ZoneName: "example-zone",
				ZoneID:   "zone-id",
			},
			expected: models.CreateDNSRecordParams{
				Content:  "192.168.0.1",
				Name:     "example.com",
				Proxied:  true,
				TTL:      3600,
				Type:     "A",
				ZoneName: "example-zone",
				ZoneID:   "zone-id",
			},
		},
		{
			name:  "Empty input",
			input: cloudflare.CreateDNSRecordParams{},
			expected: models.CreateDNSRecordParams{
				Content:  "",
				Name:     "",
				Proxied:  false,
				TTL:      0,
				Type:     "",
				ZoneName: "",
				ZoneID:   "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := convFromCreateDNSRecordParams(test.input)
			assert.Equal(t, test.expected, result, "Test %s failed", test.name)
		})
	}
}

func TestConvToUpdateDNSRecordParams(t *testing.T) {
	tests := []struct {
		name     string
		input    models.UpdateDNSRecordParams
		expected cloudflare.UpdateDNSRecordParams
	}{
		{
			name: "Valid input",
			input: models.UpdateDNSRecordParams{
				Content: "192.168.0.1",
				ID:      "record-id",
				Name:    "example.com",
				Proxied: true,
				TTL:     3600,
				Type:    "A",
			},
			expected: cloudflare.UpdateDNSRecordParams{
				Content: "192.168.0.1",
				ID:      "record-id",
				Name:    "example.com",
				Proxied: cloudflare.BoolPtr(true),
				TTL:     3600,
				Type:    "A",
			},
		},
		{
			name:  "Empty input",
			input: models.UpdateDNSRecordParams{},
			expected: cloudflare.UpdateDNSRecordParams{
				Content: "",
				ID:      "",
				Name:    "",
				Proxied: cloudflare.BoolPtr(false),
				TTL:     0,
				Type:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convToUpdateDNSRecordParams(tt.input)
			assert.Equal(t, tt.expected, result, "Test %s failed", tt.name)
		})
	}
}
