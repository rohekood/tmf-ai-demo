package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"tmf/services/qualification/internal/core/domain"
	"tmf/services/qualification/internal/core/ports"

	"github.com/google/uuid"
)

type RPCCaller interface {
	Request(ctx context.Context, topic string, payload interface{}) ([]byte, error)
}

type rabbitInventoryClient struct {
	rpc RPCCaller
}

func NewInventoryClient(rpc RPCCaller) ports.InventoryClient {
	return &rabbitInventoryClient{rpc: rpc}
}

func (c *rabbitInventoryClient) GetPortCapacity(ctx context.Context, address domain.Address) (int, error) {
	reqID := uuid.New().String()
	payload := map[string]interface{}{
		"locationId":   "CABINET_BERLIN_05",
		"resourceType": "OLT_PORT",
		"requestId":    reqID,
	}

	slog.InfoContext(ctx, "RPC Call: GetPortCapacity", "reqID", reqID)

	responseBytes, err := c.rpc.Request(ctx, "query.inventory.resource.capacity", payload)
	if err != nil {
		return 0, fmt.Errorf("rpc failed: %w", err)
	}

	var resp struct {
		Free int `json:"free"`
	}
	if err := json.Unmarshal(responseBytes, &resp); err != nil {
		return 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return resp.Free, nil
}
