package rpc

import (
	"context"
	"encoding/json"
	"fmt"

	"tmf/services/pocv/internal/core/ports"
)

// RPCRequestor is an interface to allow mocking of rabbitmq.RPCClient
type RPCRequestor interface {
	Request(ctx context.Context, routingKey string, payload interface{}) ([]byte, error)
}

type cartClient struct {
	rpc RPCRequestor
}

func NewCartClient(rpc RPCRequestor) ports.CartClient {
	return &cartClient{rpc: rpc}
}

func (c *cartClient) GetCart(ctx context.Context, cartID string) (map[string]interface{}, error) {
	// Send RPC request to Shopping Cart
	// Topic: query.cart.session.get (We need to define this topic in pkg if not exists)

	payload := map[string]string{"cartId": cartID}
	respBytes, err := c.rpc.Request(ctx, "query.cart.session.get", payload)

	if err != nil {
		return nil, fmt.Errorf("rpc cart fetch failed: %w", err)
	}

	var cart map[string]interface{}
	if err := json.Unmarshal(respBytes, &cart); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cart response: %w", err)
	}
	return cart, nil
}
