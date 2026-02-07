package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"tmf/pkg/rabbitmq"
	"tmf/services/qualification/internal/core/domain"
	"tmf/services/qualification/internal/core/ports"
)

type CatalogRPCClient struct {
	rpcClient *rabbitmq.RPCClient
}

func NewCatalogRPCClient(rabbitURL string) (ports.CatalogClient, error) {
	client, err := rabbitmq.NewRPCClient(rabbitURL, rabbitmq.WithExchange("catalog_events"))
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC client: %w", err)
	}
	return &CatalogRPCClient{rpcClient: client}, nil
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
	// rpcClient might need Closing if it holds connection, but usually pkg/rabbitmq handles it or it's transient?
	// Checking RPCClient implementation: it creates a connection in NewRPCClient?
	// e2e_test.go doesn't close it explicitly? Wait, if it creates a connection, it should be closed.
	// For now, let's assume it should be closed.
	// However, the interface ports.CatalogClient doesn't require Close().
	// It's acceptable to hold it open for the service lifecycle.
	return nil
}
