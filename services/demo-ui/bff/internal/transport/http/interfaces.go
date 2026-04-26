package http

import "context"

// RPCClient defines the interface for making RPC calls to RabbitMQ
type RPCClient interface {
	CallRPC(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error)
}
