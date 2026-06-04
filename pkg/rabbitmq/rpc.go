package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// DirectReplyToQueue is RabbitMQ's pseudo-queue for direct reply-to RPC.
const DirectReplyToQueue = "amq.rabbitmq.reply-to"

// ErrReconnecting is returned to in-flight RPC calls when the connection drops.
var ErrReconnecting = errors.New("rpc: connection lost, reconnecting")

type rpcReply struct {
	body []byte
	err  error
}

// channelPool is a fixed-size pool of AMQP channels for concurrent publishing.
type channelPool struct {
	pool chan *amqp.Channel
}

func newChannelPool(conn *amqp.Connection, size int) (*channelPool, error) {
	p := &channelPool{pool: make(chan *amqp.Channel, size)}
	for i := 0; i < size; i++ {
		ch, err := conn.Channel()
		if err != nil {
			// Close channels opened so far before returning the error.
			close(p.pool)
			for ch := range p.pool {
				_ = ch.Close()
			}
			return nil, fmt.Errorf("failed to open channel %d/%d: %w", i+1, size, err)
		}
		p.pool <- ch
	}
	return p, nil
}

// acquire returns a channel from the pool. Blocks until one is available or ctx is done.
func (p *channelPool) acquire(ctx context.Context) (*amqp.Channel, error) {
	select {
	case ch := <-p.pool:
		return ch, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// release returns a channel to the pool. Closes it if the pool is at capacity.
func (p *channelPool) release(ch *amqp.Channel) {
	select {
	case p.pool <- ch:
	default:
		_ = ch.Close()
	}
}

// drain closes and discards all channels currently in the pool.
func (p *channelPool) drain() {
	for {
		select {
		case ch := <-p.pool:
			_ = ch.Close()
		default:
			return
		}
	}
}

// RPCClient handles Request-Reply messaging over RabbitMQ with automatic reconnection.
type RPCClient struct {
	url        string
	replyQueue string
	exchange   string
	poolSize   int

	// connMu guards conn, pool, and consumerCh. Read-locked by publishers; write-locked during reconnect.
	connMu     sync.RWMutex
	conn       *amqp.Connection
	pool       *channelPool  // for fire-and-forget Publish only
	consumerCh *amqp.Channel // handles both reply consumption and RPC publish (Direct Reply-To requires same channel)
	rpcMu      sync.Mutex    // serialises RPC publishes through consumerCh

	pending sync.Map // map[string]chan rpcReply (correlationID → response channel)

	done chan struct{} // closed by Close()
	once sync.Once
}

// RPCClientOption configures the RPC client.
type RPCClientOption func(*RPCClient)

// WithExchange sets the exchange for RPC requests.
func WithExchange(exchange string) RPCClientOption {
	return func(c *RPCClient) { c.exchange = exchange }
}

// WithReplyQueue configures the reply target used by RPC requests.
//
// Supported values:
//   - A pre-created named queue (must already exist; no configure permission needed).
//   - DirectReplyToQueue (amq.rabbitmq.reply-to), which does not require queue declaration.
func WithReplyQueue(name string) RPCClientOption {
	return func(c *RPCClient) { c.replyQueue = name }
}

// WithPoolSize sets the number of AMQP channels in the publish pool (default: NumCPU*2).
func WithPoolSize(n int) RPCClientOption {
	return func(c *RPCClient) { c.poolSize = n }
}

// NewRPCClient creates a new RPC client and starts the reconnection monitor.
func NewRPCClient(url string, opts ...RPCClientOption) (*RPCClient, error) {
	client := &RPCClient{
		url:      url,
		exchange: "",
		done:     make(chan struct{}),
	}
	for _, opt := range opts {
		opt(client)
	}
	if client.replyQueue == "" {
		client.replyQueue = DirectReplyToQueue
	}
	if client.poolSize <= 0 {
		client.poolSize = runtime.NumCPU() * 2
	}

	if err := client.connect(); err != nil {
		return nil, err
	}
	go client.reconnectLoop()
	return client, nil
}

// connect dials, creates the pool and consumer channel, and installs them under connMu write lock.
func (c *RPCClient) connect() error {
	conn, err := amqp.Dial(c.url)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	pool, err := newChannelPool(conn, c.poolSize)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("failed to create channel pool: %w", err)
	}

	consumerCh, err := conn.Channel()
	if err != nil {
		pool.drain()
		_ = conn.Close()
		return fmt.Errorf("failed to open consumer channel: %w", err)
	}

	var msgs <-chan amqp.Delivery
	if c.replyQueue == DirectReplyToQueue {
		msgs, err = consumerCh.Consume(c.replyQueue, "", true, true, false, false, nil)
	} else {
		if _, err = consumerCh.QueueDeclarePassive(c.replyQueue, true, false, false, false, nil); err == nil {
			msgs, err = consumerCh.Consume(c.replyQueue, "", true, false, false, false, nil)
		}
	}
	if err != nil {
		_ = consumerCh.Close()
		pool.drain()
		_ = conn.Close()
		return fmt.Errorf("failed to start consumer: %w", err)
	}

	c.connMu.Lock()
	// If Close() was called while we were connecting, clean up and bail.
	select {
	case <-c.done:
		c.connMu.Unlock()
		_ = consumerCh.Close()
		pool.drain()
		_ = conn.Close()
		return errors.New("client closed")
	default:
	}
	c.conn = conn
	c.pool = pool
	c.consumerCh = consumerCh
	c.connMu.Unlock()

	go c.handleReplies(msgs)
	return nil
}

// reconnectLoop watches for connection loss and transparently rebuilds the connection.
func (c *RPCClient) reconnectLoop() {
	const (
		initialBackoff = 500 * time.Millisecond
		maxBackoff     = 30 * time.Second
	)
	backoff := initialBackoff

	for {
		// Register close listeners against the current connection and consumer channel.
		c.connMu.RLock()
		connClose := c.conn.NotifyClose(make(chan *amqp.Error, 1))
		chClose := c.consumerCh.NotifyClose(make(chan *amqp.Error, 1))
		c.connMu.RUnlock()

		select {
		case <-c.done:
			return
		case <-connClose:
		case <-chClose:
		}

		// Fail all in-flight callers immediately so they don't hang until their context expires.
		c.connMu.Lock()
		c.drainPending()
		if c.pool != nil {
			c.pool.drain()
		}
		c.connMu.Unlock()

		// Reconnect with exponential backoff.
		for {
			select {
			case <-c.done:
				return
			case <-time.After(backoff):
			}
			if err := c.connect(); err != nil {
				backoff = min(backoff*2, maxBackoff)
				continue
			}
			backoff = initialBackoff
			break
		}
	}
}

// drainPending notifies all in-flight RPC callers that the connection was lost.
// Must be called with connMu write-locked (or before any concurrent access begins).
func (c *RPCClient) drainPending() {
	c.pending.Range(func(key, value any) bool {
		value.(chan rpcReply) <- rpcReply{err: ErrReconnecting}
		c.pending.Delete(key)
		return true
	})
}

func (c *RPCClient) handleReplies(msgs <-chan amqp.Delivery) {
	for d := range msgs {
		if ch, ok := c.pending.Load(d.CorrelationId); ok {
			ch.(chan rpcReply) <- rpcReply{body: d.Body}
		}
	}
}

// Request sends a message and waits for a reply (uses default exchange).
// The context must carry a deadline; without one, the call blocks until the broker replies.
func (c *RPCClient) Request(ctx context.Context, routingKey string, payload any) ([]byte, error) {
	return c.RequestWithHeaders(ctx, c.exchange, routingKey, payload, nil)
}

// RequestToExchange sends a message to a specific exchange and waits for a reply.
func (c *RPCClient) RequestToExchange(ctx context.Context, exchange, routingKey string, payload any) ([]byte, error) {
	return c.RequestWithHeaders(ctx, exchange, routingKey, payload, nil)
}

// RequestWithHeaders sends a message with custom headers and waits for a reply.
func (c *RPCClient) RequestWithHeaders(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal: %w", err)
	}

	correlationID := uuid.New().String()
	responseChan := make(chan rpcReply, 1)
	c.pending.Store(correlationID, responseChan)
	defer c.pending.Delete(correlationID)

	amqpHeaders := injectContextHeaders(ctx)
	maps.Copy(amqpHeaders, headers)

	if err := c.publishRPC(ctx, exchange, routingKey, correlationID, amqpHeaders, body); err != nil {
		return nil, fmt.Errorf("failed to publish request: %w", err)
	}

	select {
	case reply := <-responseChan:
		return reply.body, reply.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Publish sends a message without waiting for a reply (fire-and-forget).
func (c *RPCClient) Publish(ctx context.Context, exchange, routingKey string, payload any) error {
	return c.PublishWithHeaders(ctx, exchange, routingKey, payload, nil)
}

// PublishWithHeaders sends a message with custom headers without waiting for a reply.
func (c *RPCClient) PublishWithHeaders(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	amqpHeaders := injectContextHeaders(ctx)
	maps.Copy(amqpHeaders, headers)

	return c.publishFireAndForget(ctx, exchange, routingKey, amqpHeaders, body)
}

// publishRPC publishes an RPC request through consumerCh so that Direct Reply-To
// routes the response back to the consumer registered on that same channel.
func (c *RPCClient) publishRPC(ctx context.Context, exchange, routingKey, correlationID string, headers amqp.Table, body []byte) error {
	c.connMu.RLock()
	ch := c.consumerCh
	c.connMu.RUnlock()

	c.rpcMu.Lock()
	err := ch.PublishWithContext(ctx, exchange, routingKey, false, false, amqp.Publishing{
		ContentType:   "application/json",
		CorrelationId: correlationID,
		ReplyTo:       c.replyQueue,
		Headers:       headers,
		Body:          body,
	})
	c.rpcMu.Unlock()
	return err
}

// publishFireAndForget acquires a pooled channel, publishes, and releases it.
func (c *RPCClient) publishFireAndForget(ctx context.Context, exchange, routingKey string, headers amqp.Table, body []byte) error {
	c.connMu.RLock()
	pool := c.pool
	c.connMu.RUnlock()

	ch, err := pool.acquire(ctx)
	if err != nil {
		return err
	}
	err = ch.PublishWithContext(ctx, exchange, routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Headers:     headers,
		Body:        body,
	})
	pool.release(ch)
	return err
}

// ReplyQueue returns the name of the reply queue for this client.
func (c *RPCClient) ReplyQueue() string { return c.replyQueue }

// Connection returns the underlying AMQP connection. The value may change after a reconnect.
func (c *RPCClient) Connection() *amqp.Connection {
	c.connMu.RLock()
	defer c.connMu.RUnlock()
	return c.conn
}

// Done returns a channel that is closed when Close() is called.
func (c *RPCClient) Done() <-chan struct{} { return c.done }

// Close shuts down the client, draining the pool and closing the connection.
func (c *RPCClient) Close() error {
	c.once.Do(func() { close(c.done) })
	c.connMu.Lock()
	defer c.connMu.Unlock()
	if c.pool != nil {
		c.pool.drain()
	}
	if c.consumerCh != nil {
		_ = c.consumerCh.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
