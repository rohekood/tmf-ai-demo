package rabbitmq

import (
	"context"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ConsumerHandler func(ctx context.Context, payload []byte) error

type Consumer interface {
	Subscribe(topic string, handler ConsumerHandler) error
	Close() error
}

type rabbitConsumer struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	exName    string
	queueName string
}

func NewConsumer(url, exchangeName, queueName string) (Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	consumer, err := NewConsumerWithConnection(conn, exchangeName, queueName)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return consumer, nil
}

// The caller is responsible for closing the connection when done.
func NewConsumerWithConnection(conn *amqp.Connection, exchangeName, queueName string) (Consumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	err = ch.ExchangeDeclare(exchangeName, "topic", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &rabbitConsumer{
		conn:      conn,
		channel:   ch,
		exName:    exchangeName,
		queueName: q.Name,
	}, nil
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
			ctx := context.Background()
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
		}
	}()

	<-ready

	return nil
}

func (c *rabbitConsumer) Close() error {
	if err := c.channel.Close(); err != nil {
		return err
	}
	return c.conn.Close()
}
