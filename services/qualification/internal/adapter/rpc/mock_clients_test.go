package rpc_test

import (
	"context"
	"testing"

	"tmf/services/qualification/internal/adapter/rpc"
	"tmf/services/qualification/internal/core/domain"

	"github.com/stretchr/testify/assert"
)

func TestMockGISClient(t *testing.T) {
	client := rpc.NewMockGISClient()

	res, err := client.CheckPolygon(context.Background(), domain.Address{City: "Berlin"})
	assert.NoError(t, err)
	assert.True(t, res)

	res, err = client.CheckPolygon(context.Background(), domain.Address{City: "Other"})
	assert.NoError(t, err)
	assert.False(t, res)
}

func TestMockInventoryClient(t *testing.T) {
	client := rpc.NewMockInventoryClient()

	res, err := client.GetPortCapacity(context.Background(), domain.Address{Zip: "12345"})
	assert.NoError(t, err)
	assert.Equal(t, 10, res)

	res, err = client.GetPortCapacity(context.Background(), domain.Address{Zip: "12344"})
	assert.NoError(t, err)
	assert.Equal(t, 0, res)
}

func TestMockCatalogClient(t *testing.T) {
	client := rpc.NewMockCatalogClient()

	res, err := client.GetOffersByCategory(context.Background(), "Fiber")
	assert.NoError(t, err)
	assert.Len(t, res, 1)

	res, err = client.GetOffersByCategory(context.Background(), "Other")
	assert.NoError(t, err)
	assert.Len(t, res, 0)
}

func TestMockRPCCaller(t *testing.T) {
	caller := &rpc.MockRPCCaller{
		OnRequest: func(ctx context.Context, topic string, payload any) ([]byte, error) {
			return []byte("test"), nil
		},
	}

	res, err := caller.Request(context.Background(), "test", nil)
	assert.NoError(t, err)
	assert.Equal(t, []byte("test"), res)

	caller.OnRequest = nil
	res, err = caller.Request(context.Background(), "test", nil)
	assert.NoError(t, err)
	assert.Nil(t, res)
}
