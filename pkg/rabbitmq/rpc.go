package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// DefaultRPCTimeout is the default timeout for RPC requests
const DefaultRPCTimeout = 30 * time.Second

// RPCClient handles Request-Reply messaging
type RPCClient struct {
	conn       *amqp.Connection
	channel    *amqp.Channel
	replyQueue string
	exchange   string
	pending    sync.Map // map[string]chan []byte (correlationID -> response channel)
}

// RPCClientOption configures the RPC client
type RPCClientOption func(*RPCClient)

// WithExchange sets the exchange for RPC requests
func WithExchange(exchange string) RPCClientOption {
	return func(c *RPCClient) {
		c.exchange = exchange
	}
}

// NewRPCClient creates a new RPC client with an exclusive reply queue
func NewRPCClient(url string, opts ...RPCClientOption) (*RPCClient, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	q, err := ch.QueueDeclare(
		"",    // name (empty = generated)
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("failed to declare reply queue: %w", err)
	}

	client := &RPCClient{
		conn:       conn,
		channel:    ch,
		replyQueue: q.Name,
		exchange:   "", // default to default exchange
	}

	for _, opt := range opts {
		opt(client)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer tag
		true,   // auto-ack
		true,   // exclusive - ensures only this client reads from its reply queue
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to consume replies: %w", err)
	}

	go client.handleReplies(msgs)

	return client, nil
}

func (c *RPCClient) handleReplies(msgs <-chan amqp.Delivery) {
	for d := range msgs {
		corrID := d.CorrelationId
		if ch, ok := c.pending.Load(corrID); ok {
			ch.(chan []byte) <- d.Body
		}
	}
}

// Request sends a message and waits for a reply (uses default exchange)
func (c *RPCClient) Request(ctx context.Context, routingKey string, payload any) ([]byte, error) {
	return c.RequestWithHeaders(ctx, c.exchange, routingKey, payload, nil)
}

// RequestToExchange sends a message to a specific exchange and waits for a reply
func (c *RPCClient) RequestToExchange(ctx context.Context, exchange, routingKey string, payload any) ([]byte, error) {
	return c.RequestWithHeaders(ctx, exchange, routingKey, payload, nil)
}

// RequestWithHeaders sends a message with custom headers and waits for a reply
func (c *RPCClient) RequestWithHeaders(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal: %w", err)
	}

	correlationID := uuid.New().String()
	responseChan := make(chan []byte, 1)
	c.pending.Store(correlationID, responseChan)
	defer c.pending.Delete(correlationID)

	amqpHeaders := injectContextHeaders(ctx)
	maps.Copy(amqpHeaders, headers)

	fmt.Printf("[RPC] Publishing to Ex: %s, Key: %s, ReplyTo: %s, CorrID: %s\n", exchange, routingKey, c.replyQueue, correlationID)
	err = c.channel.PublishWithContext(ctx,
		exchange,   // Exchange
		routingKey, // Routing Key
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: correlationID,
			ReplyTo:       c.replyQueue,
			Headers:       amqpHeaders,
			Body:          body,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to publish request: %w", err)
	}

	select {
	case response := <-responseChan:
		// Workaround for double-marshaling/Base64 issue
		if len(response) > 0 && response[0] == '"' {
			var decoded string
			if err := json.Unmarshal(response, &decoded); err == nil {
				// If it was a valid JSON string, use the decoded content (which is the actual JSON object)
				// If it wasn't a JSON string (e.g. just quote), Unmarshal would fail or return string.
				// We assume it's the double-marshaled JSON.
				return []byte(decoded), nil
			}
		}
		return response, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(DefaultRPCTimeout):
		return nil, errors.New("rpc timeout")
	}
}

// ReplyQueue returns the name of the reply queue for this client
func (c *RPCClient) ReplyQueue() string {
	return c.replyQueue
}

// Connection returns the underlying AMQP connection for advanced use cases
// (e.g., creating additional channels for separate consumers)
func (c *RPCClient) Connection() *amqp.Connection {
	return c.conn
}

// Publish sends a message without waiting for a reply (fire-and-forget)
func (c *RPCClient) Publish(ctx context.Context, exchange, routingKey string, payload any) error {
	return c.PublishWithHeaders(ctx, exchange, routingKey, payload, nil)
}

// PublishWithHeaders sends a message with custom headers without waiting for reply
func (c *RPCClient) PublishWithHeaders(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	amqpHeaders := injectContextHeaders(ctx)
	maps.Copy(amqpHeaders, headers)

	return c.channel.PublishWithContext(ctx,
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Headers:     amqpHeaders,
			Body:        body,
		})
}

func (c *RPCClient) Close() error {
	_ = c.channel.Close()
	return c.conn.Close()
}
