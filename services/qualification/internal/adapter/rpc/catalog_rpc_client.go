package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"tmf/services/qualification/internal/core/domain"
	"tmf/services/qualification/internal/core/ports"
)

type CatalogRPCClient struct {
	rpcClient Requester
}

func NewCatalogRPCClient(client Requester) ports.CatalogClient {
	return &CatalogRPCClient{rpcClient: client}
}

func (c *CatalogRPCClient) GetOffersByCategory(ctx context.Context, category string) ([]domain.EligibleCategory, error) {
	req := map[string]string{
		"category": category,
	}

	respBytes, err := c.rpcClient.Request(ctx, "query.catalog.offering.by_category", req)
	if err != nil {
		return nil, fmt.Errorf("rpc request failed: %w", err)
	}

	var offers []domain.EligibleCategory
	if err := json.Unmarshal(respBytes, &offers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return offers, nil
}

func (c *CatalogRPCClient) Close() error {
	return nil
}
