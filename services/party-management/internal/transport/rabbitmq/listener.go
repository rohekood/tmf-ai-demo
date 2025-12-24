package rabbitmq

import (
	"context"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

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
		// Ignore bind error
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
		PartyQueue,
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
		CmdPartyCreate,
		CmdPartyUpdate,
		CmdPartyPatch,
		CmdPartyDelete,
		QueryPartyGet,
		QueryPartySearch,
	}

	for _, rk := range routingKeys {
		err = ch.QueueBind(q.Name, rk, CommandExchange, false, nil)
		if err != nil {
			return fmt.Errorf("failed to bind command queue: %w", err)
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

	slog.Info("Party Management listener started...")

	for {
		select {
		case <-ctx.Done():
			return nil
		case d := <-msgs:
			go func(d amqp.Delivery) {
				var targetHandler func(context.Context, amqp.Delivery) error
				switch d.RoutingKey {
				case CmdPartyCreate:
					targetHandler = h.HandleCreateParty
				case CmdPartyUpdate:
					targetHandler = h.HandleUpdateParty
				case CmdPartyPatch:
					targetHandler = h.HandlePatchParty
				case CmdPartyDelete:
					targetHandler = h.HandleDeleteParty
				case QueryPartyGet:
					targetHandler = h.HandleGetParty
				case QueryPartySearch:
					targetHandler = h.HandleSearchParty
				default:
					slog.Warn("unknown routing key", "routing_key", d.RoutingKey)
					d.Nack(false, false)
					return
				}

				// Wrap with middlewares
				wrappedHandler := Chain(targetHandler,
					TracingMiddleware("party-management"),
					AuthMiddleware(),
					JWTMiddleware())

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
