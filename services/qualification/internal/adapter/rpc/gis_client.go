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

type rabbitGISClient struct {
	rpc Requester
}

func NewGISClient(rpc Requester) ports.GISClient {
	return &rabbitGISClient{rpc: rpc}
}

func (c *rabbitGISClient) CheckPolygon(ctx context.Context, address domain.Address) (bool, error) {
	reqID := uuid.New().String()
	payload := map[string]interface{}{
		"address":   address,
		"requestId": reqID,
	}

	slog.InfoContext(ctx, "RPC Call: CheckPolygon", "reqID", reqID)

	responseBytes, err := c.rpc.Request(ctx, "query.gis.geography.check", payload)
	if err != nil {
		return false, fmt.Errorf("rpc failed: %w", err)
	}

	var resp struct {
		InFootprint bool `json:"inFootprint"`
	}
	if err := json.Unmarshal(responseBytes, &resp); err != nil {
		return false, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return resp.InFootprint, nil
}
