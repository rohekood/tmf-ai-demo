package rpc

import (
	"context"
	"encoding/json"
	"fmt"

	"tmf/services/qualification/internal/core/ports"
)

type customerPricingClient struct {
	rpc Requester
}

// NewCustomerPricingClient creates a new RPC client for customer management pricing
func NewCustomerPricingClient(rpc Requester) ports.CustomerPricingClient {
	return &customerPricingClient{rpc: rpc}
}

// GetCustomer retrieves customer information including tier
func (c *customerPricingClient) GetCustomer(ctx context.Context, customerID string) (*ports.Customer, error) {
	request := map[string]string{
		"customerId": customerID,
	}

	respBytes, err := c.rpc.Request(ctx, "query.customer.get", request)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	var customer ports.Customer
	if err := json.Unmarshal(respBytes, &customer); err != nil {
		return nil, fmt.Errorf("failed to unmarshal customer: %w", err)
	}

	return &customer, nil
}
