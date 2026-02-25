package rpc

import (
	"context"
)

type Requester interface {
	Request(ctx context.Context, routingKey string, request interface{}) ([]byte, error)
}
