package rabbitmq

import (
	"context"
	"encoding/json"
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

func (l *Listener) GetHandler(routingKey string, h *Handlers) (func(context.Context, amqp.Delivery) error, bool) {
	switch routingKey {
	case "cmd.customer.onboard":
		return h.HandleOnboardCustomer, true
	case "cmd.customer.update":
		return h.HandleUpdateCustomer, true
	// case "query.customer.get":
	// 	return h.HandleGetCustomer, true
	case "query.customer.search":
		return h.HandleSearchCustomer, true
	case "cmd.customer.delete":
		return h.HandleDeleteCustomer, true
	case CmdCustomerLogInteraction:
		return h.HandleLogInteraction, true
	default:
		return nil, false
	}
}

func (l *Listener) Start(ctx context.Context, h *Handlers) error {
	ch, err := l.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer func() {
		if err := ch.Close(); err != nil {
			slog.Error("failed to close channel", "error", err)
		}
	}()

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
		return fmt.Errorf("failed to bind DLQ to DLX: %w", err)
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
		// "query.customer.get", // Removed: Handled by RPCHandler on dedicated queue
		"query.customer.search",
		"cmd.customer.delete",
		CmdCustomerLogInteraction,
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
		"evt.party.deletion_initiated",
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
			// TracingMiddleware("customer-management"),
			AuthMiddleware(),
			JWTMiddleware())

		for d := range eventMsgs {
			err := wrappedHandler(ctx, d)
			if err != nil {
				slog.Error("error handling event", "routing_key", d.RoutingKey, "error", err)
				if nackErr := d.Nack(false, false); nackErr != nil {
					slog.Error("failed to nack message", "error", nackErr)
				} // Don't requeue, send to DLX
			} else {
				if ackErr := d.Ack(false); ackErr != nil {
					slog.Error("failed to ack message", "error", ackErr)
				}
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
				targetHandler, valid := l.GetHandler(d.RoutingKey, h)
				if !valid {
					slog.Warn("unknown routing key", "routing_key", d.RoutingKey)
					if nackErr := d.Nack(false, false); nackErr != nil {
						slog.Error("failed to nack message", "error", nackErr)
					}
					return
				}

				// Wrap with middlewares
				wrappedHandler := Chain(targetHandler,
					// TracingMiddleware("customer-management"),
					AuthMiddleware(),
					JWTMiddleware())

				err := wrappedHandler(ctx, d)
				if err != nil {
					slog.Error("error handling message", "routing_key", d.RoutingKey, "error", err)

					// Attempt to reply with error
					if d.ReplyTo != "" {
						errResponse := map[string]string{"error": err.Error()}
						errBody, _ := json.Marshal(errResponse)

						ch, chErr := l.conn.Channel()
						if chErr == nil {
							defer func() {
								if err := ch.Close(); err != nil {
									slog.Error("failed to close error reply channel", "error", err)
								}
							}()
							if pubErr := ch.PublishWithContext(ctx,
								"",        // exchange
								d.ReplyTo, // routing key
								false,     // mandatory
								false,     // immediate
								amqp.Publishing{
									ContentType:   "application/json",
									CorrelationId: d.CorrelationId,
									Body:          errBody,
								}); pubErr != nil {
								slog.Error("failed to publish error reply", "error", pubErr)
							}
						} else {
							slog.Error("failed to open channel to send error reply", "error", chErr)
						}
					}

					if nackErr := d.Nack(false, false); nackErr != nil {
						slog.Error("failed to nack message", "error", nackErr)
					} // Don't requeue, send to DLX
				} else {
					if ackErr := d.Ack(false); ackErr != nil {
						slog.Error("failed to ack message", "error", ackErr)
					}
				}
			}(d)
		}
	}
}
