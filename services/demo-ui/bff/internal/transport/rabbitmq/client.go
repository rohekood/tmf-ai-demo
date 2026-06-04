package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	pkgrmq "tmf/pkg/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Client wraps the shared RPC client with BFF-specific features like debug broadcasting
type Client struct {
	*pkgrmq.RPCClient
	broadcaster Broadcaster
}

var newRPCClientFunc = pkgrmq.NewRPCClient

// NewClient creates a new BFF RabbitMQ client using the shared library
func NewClient(url string) (*Client, error) {
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}

	rpcClient, err := newRPCClientFunc(url, pkgrmq.WithReplyQueue(pkgrmq.DirectReplyToQueue))
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC client: %w", err)
	}

	return &Client{
		RPCClient: rpcClient,
	}, nil
}

// SetBroadcaster sets the broadcaster for forwarding RPC replies to the debug console
func (c *Client) SetBroadcaster(b Broadcaster) {
	c.broadcaster = b
}

// CallRPC sends a request and waits for a response, with debug broadcasting support
func (c *Client) CallRPC(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
	// Broadcast request if broadcaster is set
	if c.broadcaster != nil {
		c.broadcastRequest(exchange, routingKey, payload, headers)
	}

	// Use the shared library for the actual RPC call
	resp, err := c.RequestWithHeaders(ctx, exchange, routingKey, payload, headers)

	// Broadcast reply if broadcaster is set
	if c.broadcaster != nil && err == nil {
		c.broadcastReply(routingKey, resp)
	}

	return resp, err
}

// PublishCommand sends a message without waiting for reply (kept for backward compatibility)
func (c *Client) PublishCommand(ctx context.Context, exchange, routingKey string, payload any) error {
	// Broadcast if broadcaster is set
	if c.broadcaster != nil {
		c.broadcastRequest(exchange, routingKey, payload, nil)
	}

	return c.Publish(ctx, exchange, routingKey, payload)
}

// Connection returns the underlying AMQP connection for advanced use cases
func (c *Client) Connection() *amqp.Connection {
	return c.RPCClient.Connection()
}

// Close closes the underlying RPC client
func (c *Client) Close() {
	if c.RPCClient != nil {
		_ = c.RPCClient.Close()
	}
}

func (c *Client) broadcastRequest(exchange, routingKey string, payload any, headers map[string]any) {
	var payloadMap map[string]any
	if data, err := json.Marshal(payload); err == nil {
		_ = json.Unmarshal(data, &payloadMap)
	}
	if payloadMap == nil {
		payloadMap = map[string]any{"data": payload}
	}

	debugMsg := DebugMessage{
		ID:        fmt.Sprintf("req-%s-%d", routingKey, time.Now().UnixNano()),
		Timestamp: time.Now().UTC(),
		Type:      "request",
		Topic:     routingKey,
		Exchange:  exchange,
		Payload:   payloadMap,
		Headers:   headers,
		Service:   "bff",
	}
	c.broadcaster.Broadcast(debugMsg)
}

func (c *Client) broadcastReply(routingKey string, body []byte) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		payload = map[string]any{
			"raw": string(body),
		}
	}

	debugMsg := DebugMessage{
		ID:        fmt.Sprintf("reply-%s-%d", routingKey, time.Now().UnixNano()),
		Timestamp: time.Now().UTC(),
		Type:      "reply",
		Topic:     "rpc.reply",
		Payload:   payload,
		Service:   "bff",
	}
	c.broadcaster.Broadcast(debugMsg)
}

// LogUnknownCorrelation logs when a reply is received for an unknown correlation ID
// This is handled internally by the shared library now, but we keep logging for debugging
func LogUnknownCorrelation(correlationID string) {
	log.Printf("Received reply for unknown correlation ID: %s", correlationID)
}
