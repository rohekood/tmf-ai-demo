package rabbitmq

import (
	"context"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Listener manages the RabbitMQ connection and message dispatching
type Listener struct {
	conn *amqp.Connection
}

func NewListener(conn *amqp.Connection) (*Listener, error) {
	return &Listener{conn: conn}, nil
}

func (l *Listener) Start(ctx context.Context, h *Handlers) error {
	ch, err := l.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	// Declare exchange
	err = ch.ExchangeDeclare(
		CommandExchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare and bind Command Queue
	q, err := ch.QueueDeclare(
		CustomerQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	routingKeys := []string{
		"cmd.customer.onboard",
		"cmd.customer.update",
		"cmd.customer.patch",
		"query.customer.get",
	}

	for _, rk := range routingKeys {
		err = ch.QueueBind(q.Name, rk, CommandExchange, false, nil)
		if err != nil {
			return fmt.Errorf("failed to bind command queue: %w", err)
		}
	}

	// Declare and bind Event Queue for Party events
	eq, err := ch.QueueDeclare(
		EventQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare event queue: %w", err)
	}

	err = ch.ExchangeDeclare(EventExchange, "topic", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare event exchange: %w", err)
	}

	eventRoutingKeys := []string{
		"evt.party.updated",
		"evt.party.deleted",
	}

	for _, rk := range eventRoutingKeys {
		err = ch.QueueBind(eq.Name, rk, EventExchange, false, nil)
		if err != nil {
			return fmt.Errorf("failed to bind event queue: %w", err)
		}
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		false, // auto-ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to consume commands: %w", err)
	}

	eventMsgs, err := ch.Consume(
		eq.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to consume events: %w", err)
	}

	log.Printf("Customer Management listener started...")

	// Launch event consumer
	go func() {
		for d := range eventMsgs {
			err := h.HandlePartyEvent(ctx, d)
			if err != nil {
				log.Printf("error handling event %s: %v", d.RoutingKey, err)
				d.Nack(false, true)
			} else {
				d.Ack(false)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case d := <-msgs:
			go func(d amqp.Delivery) {
				var err error
				switch d.RoutingKey {
				case "cmd.customer.onboard":
					err = h.HandleOnboardCustomer(ctx, d)
				case "cmd.customer.update":
					err = h.HandleUpdateCustomer(ctx, d)
				case "query.customer.get":
					err = h.HandleGetCustomer(ctx, d)
				default:
					log.Printf("unknown routing key: %s", d.RoutingKey)
				}

				if err != nil {
					log.Printf("error handling message %s: %v", d.RoutingKey, err)
					d.Nack(false, true) // requeue
				} else {
					d.Ack(false)
				}
			}(d)
		}
	}
}
