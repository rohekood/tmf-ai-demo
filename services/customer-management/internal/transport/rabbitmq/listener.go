package rabbitmq

import (
	"context"
	"fmt"
	"log/slog"

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

	// Declare DLX and DLQ
	err = ch.ExchangeDeclare(DeadLetterExchange, "fanout", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare DLX: %w", err)
	}

	_, err = ch.QueueDeclare(DeadLetterQueue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare DLQ: %w", err)
	}

	err = ch.QueueBind(DeadLetterQueue, "", DeadLetterExchange, false, nil)
	if err != nil {
		// Ignore bind error if routing key mismatch, but direct exchange with # is not standard.
		// Using empty routing key for direct DLX or topic DLX with #.
		// Let's use fanout for DLX or bind with specific keys.
		// To keep it simple, bind everything to DLQ.
	}

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
		amqp.Table{
			"x-dead-letter-exchange": DeadLetterExchange,
		},
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
		amqp.Table{
			"x-dead-letter-exchange": DeadLetterExchange,
		},
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

	slog.Info("Customer Management listener started...")

	// Launch event consumer
	go func() {
		// Wrap event handler with middlewares
		wrappedHandler := Chain(h.HandlePartyEvent,
			TracingMiddleware("customer-management"),
			AuthMiddleware())

		for d := range eventMsgs {
			err := wrappedHandler(ctx, d)
			if err != nil {
				slog.Error("error handling event", "routing_key", d.RoutingKey, "error", err)
				d.Nack(false, false) // Don't requeue, send to DLX
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
				var targetHandler func(context.Context, amqp.Delivery) error
				switch d.RoutingKey {
				case "cmd.customer.onboard":
					targetHandler = h.HandleOnboardCustomer
				case "cmd.customer.update":
					targetHandler = h.HandleUpdateCustomer
				case "query.customer.get":
					targetHandler = h.HandleGetCustomer
				default:
					slog.Warn("unknown routing key", "routing_key", d.RoutingKey)
					d.Nack(false, false)
					return
				}

				// Wrap with middlewares
				wrappedHandler := Chain(targetHandler,
					TracingMiddleware("customer-management"),
					AuthMiddleware())

				err := wrappedHandler(ctx, d)
				if err != nil {
					slog.Error("error handling message", "routing_key", d.RoutingKey, "error", err)
					d.Nack(false, false) // Don't requeue, send to DLX
				} else {
					d.Ack(false)
				}
			}(d)
		}
	}
}
