package rpc

import (
	"context"
	"time"
	"tmf/services/qualification/internal/core/domain"
)

// MockGISClient simulates a GIS lookup
type MockGISClient struct{}

func NewMockGISClient() *MockGISClient {
	return &MockGISClient{}
}

func (m *MockGISClient) CheckPolygon(ctx context.Context, addr domain.Address) (bool, error) {
	// Simulate Network Latency
	time.Sleep(100 * time.Millisecond)
	// Logic: If City is Berlin, we return true. Else false.
	return addr.City == "Berlin", nil
}

// MockInventoryClient simulates an Inventory lookup
type MockInventoryClient struct{}

func NewMockInventoryClient() *MockInventoryClient {
	return &MockInventoryClient{}
}

func (m *MockInventoryClient) GetPortCapacity(ctx context.Context, addr domain.Address) (int, error) {
	time.Sleep(150 * time.Millisecond)
	// Logic: If Zip ends in 5, we have ports.
	if len(addr.Zip) > 0 && addr.Zip[len(addr.Zip)-1] == '5' {
		return 10, nil
	}
	return 0, nil
}

// MockCatalogClient simulates a Product Catalog lookup
type MockCatalogClient struct{}

func NewMockCatalogClient() *MockCatalogClient {
	return &MockCatalogClient{}
}

func (m *MockCatalogClient) GetOffersByCategory(ctx context.Context, category string) ([]domain.EligibleCategory, error) {
	time.Sleep(50 * time.Millisecond)

	if category == "Fiber" || category == "Internet" {
		return []domain.EligibleCategory{
			{
				ID:   "OFFERING_INTERNET_1",
				Name: "Fiber Internet 1000",
				Characteristics: map[string]string{
					"MaxSpeed":   "1000Mbps",
					"Technology": "GPON",
				},
			},
		}, nil
	}

	return []domain.EligibleCategory{}, nil
}

// MockRPCCaller simulates RPC calls
type MockRPCCaller struct {
	OnRequest func(ctx context.Context, topic string, payload interface{}) ([]byte, error)
}

func (m *MockRPCCaller) Request(ctx context.Context, topic string, payload interface{}) ([]byte, error) {
	if m.OnRequest != nil {
		return m.OnRequest(ctx, topic, payload)
	}
	return nil, nil
}
