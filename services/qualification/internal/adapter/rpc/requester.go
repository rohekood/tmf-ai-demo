package rpc

import (
	"context"
)

type Requester interface {
	Request(ctx context.Context, routingKey string, request any) ([]byte, error)
}
