package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pkgrmq "tmf/pkg/rabbitmq"
)

const (
	customerExchange    = "tmf.customer"
	queryCustomerSearch = "query.customer.search"
	customerRPCTimeout  = 5 * time.Second
)

// rpcRequester is a narrow interface satisfied by *pkgrmq.RPCClient.
type rpcRequester interface {
	RequestToExchange(ctx context.Context, exchange, routingKey string, payload any) ([]byte, error)
}

// CustomerCheckerRPC implements CustomerChecker via a synchronous RPC call to customer-management.
type CustomerCheckerRPC struct {
	rpcClient rpcRequester
}

func NewCustomerCheckerRPC(rpcClient *pkgrmq.RPCClient) *CustomerCheckerRPC {
	return &CustomerCheckerRPC{rpcClient: rpcClient}
}

// HasCustomers returns true if customer-management reports at least one customer for the given partyID.
func (c *CustomerCheckerRPC) HasCustomers(ctx context.Context, partyID string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, customerRPCTimeout)
	defer cancel()

	payload := map[string]string{"partyId": partyID}
	responseBytes, err := c.rpcClient.RequestToExchange(ctx, customerExchange, queryCustomerSearch, payload)
	if err != nil {
		return false, fmt.Errorf("customer check RPC failed: %w", err)
	}

	var result []any
	if err := json.Unmarshal(responseBytes, &result); err != nil {
		// Non-array response (e.g. error object) — treat as no customers found
		return false, nil
	}

	return len(result) > 0, nil
}
