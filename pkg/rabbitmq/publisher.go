package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher interface to allow mocking
type Publisher interface {
	Publish(ctx context.Context, exchange, routingKey string, body any) error
	PublishToQueue(ctx context.Context, queueName string, correlationID string, body any) error
	DeclareTopicExchange(name string, durable, autoDelete, internal, noWait bool) error
	Close() error
}

type rabbitPublisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	mu   sync.Mutex
}

func NewPublisher(url string) (Publisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	pub, err := NewPublisherWithConnection(conn)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return pub, nil
}

// NewPublisherWithConnection creates a Publisher using an existing AMQP connection.
// The caller is responsible for closing the connection when done.
func NewPublisherWithConnection(conn *amqp.Connection) (Publisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &rabbitPublisher{
		conn: conn,
		ch:   ch,
	}, nil
}

func (p *rabbitPublisher) DeclareTopicExchange(name string, durable, autoDelete, internal, noWait bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.ch.ExchangeDeclare(
		name,       // name
		"topic",    // type
		durable,    // durable
		autoDelete, // auto-deleted
		internal,   // internal
		noWait,     // no-wait
		nil,        // arguments
	)
}

func (p *rabbitPublisher) Publish(ctx context.Context, exchange, routingKey string, body any) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	headers := injectContextHeaders(ctx)
	correlationID, _ := headers[HeaderCorrelationID].(string)

	return p.ch.PublishWithContext(ctx,
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			Headers:       headers,
			CorrelationId: correlationID,
			Body:          data,
		},
	)
}

func (p *rabbitPublisher) PublishToQueue(ctx context.Context, queueName string, correlationID string, body any) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	headers := injectContextHeaders(ctx)

	return p.ch.PublishWithContext(ctx,
		"",        // Default exchange
		queueName, // Routing Key = Queue Name
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			Headers:       headers,
			CorrelationId: correlationID,
			Body:          data,
		},
	)
}

func (p *rabbitPublisher) Close() error {
	_ = p.ch.Close()
	return p.conn.Close()
}
