package rpc

import (
	"context"
	"encoding/json"
	"fmt"

	"tmf/pkg/rabbitmq"
	"tmf/services/qualification/internal/core/ports"
)

type catalogPricingClient struct {
	rpc *rabbitmq.RPCClient
}

// NewCatalogPricingClient creates a new RPC client for product catalog pricing
func NewCatalogPricingClient(rpc *rabbitmq.RPCClient) ports.CatalogPricingClient {
	return &catalogPricingClient{rpc: rpc}
}

// GetOffering retrieves a product offering with base price
func (c *catalogPricingClient) GetOffering(ctx context.Context, offeringID string) (*ports.Offering, error) {
	request := map[string]string{
		"offeringId": offeringID,
	}

	respBytes, err := c.rpc.Request(ctx, "query.catalog.offering.get", request)
	if err != nil {
		return nil, fmt.Errorf("failed to get offering: %w", err)
	}

	var offering ports.Offering
	if err := json.Unmarshal(respBytes, &offering); err != nil {
		return nil, fmt.Errorf("failed to unmarshal offering: %w", err)
	}

	return &offering, nil
}
