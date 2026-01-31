package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"tmf/services/qualification/internal/core/domain"

	"github.com/stretchr/testify/assert"
)

func TestGetPortCapacity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Mock setup
		expectedCapacity := 5
		mockCaller := &MockRPCCaller{
			OnRequest: func(ctx context.Context, topic string, payload interface{}) ([]byte, error) {
				assert.Equal(t, "query.inventory.resource.capacity", topic)

				resp := map[string]interface{}{
					"free": expectedCapacity,
				}
				return json.Marshal(resp)
			},
		}

		client := NewInventoryClient(mockCaller)
		capacity, err := client.GetPortCapacity(context.Background(), domain.Address{City: "Berlin"})

		assert.NoError(t, err)
		assert.Equal(t, expectedCapacity, capacity)
	})

	t.Run("Failure RPC Error", func(t *testing.T) {
		mockCaller := &MockRPCCaller{
			OnRequest: func(ctx context.Context, topic string, payload interface{}) ([]byte, error) {
				return nil, errors.New("rpc error")
			},
		}

		client := NewInventoryClient(mockCaller)
		_, err := client.GetPortCapacity(context.Background(), domain.Address{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rpc failed")
	})

	t.Run("Failure JSON Unmarshal", func(t *testing.T) {
		mockCaller := &MockRPCCaller{
			OnRequest: func(ctx context.Context, topic string, payload interface{}) ([]byte, error) {
				return []byte("invalid json"), nil
			},
		}

		client := NewInventoryClient(mockCaller)
		_, err := client.GetPortCapacity(context.Background(), domain.Address{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal response")
	})
}
