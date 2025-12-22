package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"tmf/services/party-management/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewPublisher(conn *amqp.Connection) (*Publisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	return &Publisher{
		conn: conn,
		ch:   ch,
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, exchange, routingKey string, event interface{}) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	headers := amqp.Table{}
	if userID, ok := ctx.Value(domain.UserContextKey).(string); ok {
		headers["user"] = userID
	}

	err = p.ch.PublishWithContext(ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Headers:     headers,
			Body:        body,
		})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

func (p *Publisher) Close() error {
	return p.ch.Close()
}

func (p *Publisher) GetChannel() (*amqp.Channel, error) {
	return p.ch, nil
}
