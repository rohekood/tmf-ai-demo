package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"tmf/services/customer-management/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	mu   sync.Mutex
}

func NewPublisher(conn *amqp.Connection) (*Publisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &Publisher{
		conn: conn,
		ch:   ch,
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, exchange, routingKey string, body interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	headers := amqp.Table{}
	if userID, ok := ctx.Value(domain.UserContextKey).(string); ok {
		headers["user"] = userID
	}

	return p.ch.PublishWithContext(ctx,
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Headers:     headers,
			Body:        data,
		},
	)
}

func (p *Publisher) Close() error {
	return p.ch.Close()
}

func (p *Publisher) GetChannel() (*amqp.Channel, error) {
	return p.ch, nil
}
