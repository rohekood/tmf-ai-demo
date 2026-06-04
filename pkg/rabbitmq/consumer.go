package rabbitmq

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ConsumerHandler func(ctx context.Context, payload []byte) error

// ConsumerOption configures a Consumer.
type ConsumerOption func(*rabbitConsumer)

// WithMessageTimeout sets a per-message processing deadline. Every message handler
// receives a context that expires after d, preventing a hung downstream RPC call
// (or slow handler) from blocking the consumer indefinitely.
// A zero value (the default) means no deadline — use only when the handler context
// already carries a deadline from elsewhere.
func WithMessageTimeout(d time.Duration) ConsumerOption {
	return func(c *rabbitConsumer) { c.msgTimeout = d }
}

type Consumer interface {
	Subscribe(topic string, handler ConsumerHandler) error
	Close() error
}

type rabbitConsumer struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	exName    string
	queueName string

	msgTimeout time.Duration

	mu       sync.RWMutex
	handlers []subscription
	started  bool
}

type subscription struct {
	topic   string
	handler ConsumerHandler
}

func NewConsumer(url, exchangeName, queueName string, opts ...ConsumerOption) (Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	consumer, err := NewConsumerWithConnection(conn, exchangeName, queueName, opts...)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return consumer, nil
}

// NewConsumerWithConnection creates a Consumer using an existing AMQP connection.
// The caller is responsible for closing the connection when done.
func NewConsumerWithConnection(conn *amqp.Connection, exchangeName, queueName string, opts ...ConsumerOption) (Consumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	if os.Getenv("RABBITMQ_PASSIVE_DECLARE") == "true" {
		err = ch.ExchangeDeclarePassive(exchangeName, "topic", true, false, false, false, nil)
	} else {
		err = ch.ExchangeDeclare(exchangeName, "topic", true, false, false, false, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	var q amqp.Queue
	if os.Getenv("RABBITMQ_PASSIVE_DECLARE") == "true" {
		q, err = ch.QueueDeclarePassive(queueName, true, false, false, false, nil)
	} else {
		q, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	c := &rabbitConsumer{
		conn:      conn,
		channel:   ch,
		exName:    exchangeName,
		queueName: q.Name,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

func (c *rabbitConsumer) Subscribe(topic string, handler ConsumerHandler) error {
	slog.Info("[DEBUG] Binding queue to topic", "queue", c.queueName, "topic", topic, "exchange", c.exName)
	err := c.channel.QueueBind(
		c.queueName,
		topic,
		c.exName,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}
	slog.Info("[DEBUG] Queue binding successful", "queue", c.queueName, "topic", topic)

	c.mu.Lock()
	c.handlers = append(c.handlers, subscription{topic: topic, handler: handler})
	if c.started {
		c.mu.Unlock()
		return nil
	}
	c.started = true
	c.mu.Unlock()

	msgs, err := c.channel.Consume(
		c.queueName,
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}
	slog.Info("[DEBUG] Consumer started on queue", "queue", c.queueName, "topic", topic)

	// Use a ready channel to ensure goroutine has started before returning
	ready := make(chan struct{})

	go func() {
		slog.Info("[DEBUG] Consumer goroutine running", "queue", c.queueName, "topic", topic)
		close(ready)

		for d := range msgs {
			slog.Info("[DEBUG] Message received on queue", "queue", c.queueName, "routingKey", d.RoutingKey, "replyTo", d.ReplyTo, "correlationId", d.CorrelationId)
			handler, found := c.handlerForRoutingKey(d.RoutingKey)
			if !found {
				slog.Warn("No handler registered for routing key", "routingKey", d.RoutingKey, "queue", c.queueName)
				_ = d.Ack(false)
				continue
			}

			func(d amqp.Delivery, handler ConsumerHandler) {
				var cancel context.CancelFunc
				ctx := context.Background()
				if c.msgTimeout > 0 {
					ctx, cancel = context.WithTimeout(ctx, c.msgTimeout)
				} else {
					ctx, cancel = context.WithCancel(ctx)
				}
				defer cancel()

				if val, ok := d.Headers[HeaderCorrelationID].(string); ok {
					ctx = context.WithValue(ctx, ContextKeyCorrelationID, val)
				}
				if val, ok := d.Headers[HeaderUserID].(string); ok {
					ctx = context.WithValue(ctx, ContextKeyUserID, val)
				}

				if d.ReplyTo != "" {
					slog.Debug("Consumer received ReplyTo", "replyTo", d.ReplyTo)
					ctx = context.WithValue(ctx, ContextKeyReplyTo, d.ReplyTo)
				}

				ctx = context.WithValue(ctx, ContextKeyRoutingKey, d.RoutingKey)

				if d.CorrelationId != "" {
					if ctx.Value(ContextKeyCorrelationID) == nil {
						ctx = context.WithValue(ctx, ContextKeyCorrelationID, d.CorrelationId)
					}
					ctx = context.WithValue(ctx, ContextKeyAMQPCorrelationID, d.CorrelationId)
				}

				if err := handler(ctx, d.Body); err != nil {
					slog.Error("Failed to process message", "error", err, "topic", topic)
					_ = d.Nack(false, false)
				} else {
					_ = d.Ack(false)
				}
			}(d, handler)
		}
	}()

	<-ready

	return nil
}

func (c *rabbitConsumer) handlerForRoutingKey(routingKey string) (ConsumerHandler, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, sub := range c.handlers {
		if sub.topic == routingKey {
			return sub.handler, true
		}
	}

	for _, sub := range c.handlers {
		if topicMatches(sub.topic, routingKey) {
			return sub.handler, true
		}
	}

	return nil, false
}

func topicMatches(bindingKey, routingKey string) bool {
	bindingParts := strings.Split(bindingKey, ".")
	routingParts := strings.Split(routingKey, ".")

	return topicPartsMatch(bindingParts, routingParts)
}

func topicPartsMatch(bindingParts, routingParts []string) bool {
	for len(bindingParts) > 0 {
		part := bindingParts[0]
		bindingParts = bindingParts[1:]

		switch part {
		case "#":
			return true
		case "*":
			if len(routingParts) == 0 {
				return false
			}
			routingParts = routingParts[1:]
		default:
			if len(routingParts) == 0 || routingParts[0] != part {
				return false
			}
			routingParts = routingParts[1:]
		}
	}

	return len(routingParts) == 0
}

func (c *rabbitConsumer) Close() error {
	if err := c.channel.Close(); err != nil {
		return err
	}
	return c.conn.Close()
}
