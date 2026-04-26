package rpc_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"tmf/services/qualification/internal/adapter/rpc"
	"tmf/services/qualification/internal/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRPCRequester struct {
	mock.Mock
}

func (m *MockRPCRequester) Request(ctx context.Context, routingKey string, request any) ([]byte, error) {
	args := m.Called(ctx, routingKey, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func TestCatalogPricingClient(t *testing.T) {
	mockReq := new(MockRPCRequester)
	client := rpc.NewCatalogPricingClient(mockReq)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		resp := map[string]any{
			"id":        "off-1",
			"name":      "Offer 1",
			"basePrice": 10.0,
			"currency":  "USD",
		}
		respBytes, _ := json.Marshal(resp)
		mockReq.On("Request", ctx, "query.catalog.offering.get", mock.Anything).Return(respBytes, nil).Once()

		offering, err := client.GetOffering(ctx, "off-1")
		assert.NoError(t, err)
		assert.Equal(t, "off-1", offering.ID)
	})

	t.Run("Success with TMF structure", func(t *testing.T) {
		resp := map[string]any{
			"id":   "off-1",
			"name": "Offer 1",
			"productOfferingPrice": []map[string]any{
				{"price": map[string]any{"value": 15.0, "unit": "EUR"}},
			},
		}
		respBytes, _ := json.Marshal(resp)
		mockReq.On("Request", ctx, "query.catalog.offering.get", mock.Anything).Return(respBytes, nil).Once()

		offering, err := client.GetOffering(ctx, "off-1")
		assert.NoError(t, err)
		assert.Equal(t, "off-1", offering.ID)
		assert.Equal(t, 15.0, offering.BasePrice)
		assert.Equal(t, "EUR", offering.Currency)
	})

	t.Run("Error", func(t *testing.T) {
		mockReq.On("Request", ctx, "query.catalog.offering.get", mock.Anything).Return(nil, errors.New("rpc error")).Once()

		_, err := client.GetOffering(ctx, "off-1")
		assert.Error(t, err)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		mockReq.On("Request", ctx, "query.catalog.offering.get", mock.Anything).Return([]byte("{"), nil).Once()

		_, err := client.GetOffering(ctx, "off-1")
		assert.Error(t, err)
	})
}

func TestCustomerPricingClient(t *testing.T) {
	mockReq := new(MockRPCRequester)
	client := rpc.NewCustomerPricingClient(mockReq)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		resp := map[string]any{
			"id":           "cust-1",
			"status":       "Active",
			"customerTier": "Gold",
		}
		respBytes, _ := json.Marshal(resp)
		mockReq.On("Request", ctx, "query.customer.get", mock.Anything).Return(respBytes, nil).Once()

		customer, err := client.GetCustomer(ctx, "cust-1")
		assert.NoError(t, err)
		assert.Equal(t, "cust-1", customer.ID)
	})

	t.Run("Error", func(t *testing.T) {
		mockReq.On("Request", ctx, "query.customer.get", mock.Anything).Return(nil, errors.New("rpc error")).Once()

		_, err := client.GetCustomer(ctx, "cust-1")
		assert.Error(t, err)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		mockReq.On("Request", ctx, "query.customer.get", mock.Anything).Return([]byte("{"), nil).Once()

		_, err := client.GetCustomer(ctx, "cust-1")
		assert.Error(t, err)
	})
}

func TestGISClient(t *testing.T) {
	mockReq := new(MockRPCRequester)
	client := rpc.NewGISClient(mockReq)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		respBytes := []byte(`{"inFootprint": true}`)
		mockReq.On("Request", ctx, "query.gis.geography.check", mock.Anything).Return(respBytes, nil).Once()

		avail, err := client.CheckPolygon(ctx, domain.Address{})
		assert.NoError(t, err)
		assert.True(t, avail)
	})

	t.Run("Error", func(t *testing.T) {
		mockReq.On("Request", ctx, "query.gis.geography.check", mock.Anything).Return(nil, errors.New("rpc error")).Once()

		_, err := client.CheckPolygon(ctx, domain.Address{})
		assert.Error(t, err)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		mockReq.On("Request", ctx, "query.gis.geography.check", mock.Anything).Return([]byte("{"), nil).Once()

		_, err := client.CheckPolygon(ctx, domain.Address{})
		assert.Error(t, err)
	})
}

func TestInventoryClient(t *testing.T) {
	mockReq := new(MockRPCRequester)
	client := rpc.NewInventoryClient(mockReq)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		respBytes := []byte(`{"free": 100}`)
		mockReq.On("Request", ctx, "query.inventory.resource.capacity", mock.Anything).Return(respBytes, nil).Once()

		cap, err := client.GetPortCapacity(ctx, domain.Address{})
		assert.NoError(t, err)
		assert.Equal(t, 100, cap)
	})

	t.Run("Error", func(t *testing.T) {
		mockReq.On("Request", ctx, "query.inventory.resource.capacity", mock.Anything).Return(nil, errors.New("rpc error")).Once()

		_, err := client.GetPortCapacity(ctx, domain.Address{})
		assert.Error(t, err)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		mockReq.On("Request", ctx, "query.inventory.resource.capacity", mock.Anything).Return([]byte("{"), nil).Once()

		_, err := client.GetPortCapacity(ctx, domain.Address{})
		assert.Error(t, err)
	})
}

func TestCatalogRPCClient(t *testing.T) {
	mockReq := new(MockRPCRequester)
	client := rpc.NewCatalogRPCClient(mockReq)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		respBytes := []byte(`[{"id": "cat-1", "name": "Category 1", "offerings": []}]`)
		mockReq.On("Request", ctx, "query.catalog.offering.by_category", mock.Anything).Return(respBytes, nil).Once()

		offers, err := client.GetOffersByCategory(ctx, "cat-1")
		assert.NoError(t, err)
		assert.Len(t, offers, 1)

		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		mockReq.On("Request", ctx, "query.catalog.offering.by_category", mock.Anything).Return(nil, errors.New("rpc error")).Once()

		_, err := client.GetOffersByCategory(ctx, "cat-1")
		assert.Error(t, err)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		mockReq.On("Request", ctx, "query.catalog.offering.by_category", mock.Anything).Return([]byte("{"), nil).Once()

		_, err := client.GetOffersByCategory(ctx, "cat-1")
		assert.Error(t, err)
	})

	t.Run("Close", func(t *testing.T) {
		err := client.(*rpc.CatalogRPCClient).Close()
		assert.NoError(t, err)
	})
}
