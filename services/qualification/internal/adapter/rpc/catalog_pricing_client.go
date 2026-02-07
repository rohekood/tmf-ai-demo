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

	// DTO to handle both simplified (RPC Handler) and standard (RabbitMQ Handler) responses
	var dto struct {
		ID                   string  `json:"id"`
		Name                 string  `json:"name"`
		BasePrice            float64 `json:"basePrice"`
		Currency             string  `json:"currency"`
		ProductOfferingPrice []struct {
			Price struct {
				Value float64 `json:"value"`
				Unit  string  `json:"unit"`
			} `json:"price"`
		} `json:"productOfferingPrice"`
	}

	if err := json.Unmarshal(respBytes, &dto); err != nil {
		return nil, fmt.Errorf("failed to unmarshal offering: %w", err)
	}

	// Map to domain port
	offering := &ports.Offering{
		ID:        dto.ID,
		Name:      dto.Name,
		BasePrice: dto.BasePrice,
		Currency:  dto.Currency,
	}

	// Fallback: If BasePrice is 0, try to extract from TMF structure
	if offering.BasePrice == 0 && len(dto.ProductOfferingPrice) > 0 {
		offering.BasePrice = dto.ProductOfferingPrice[0].Price.Value
		offering.Currency = dto.ProductOfferingPrice[0].Price.Unit
	}

	return offering, nil
}
