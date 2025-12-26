package rabbitmq

import (
	"context"
	"encoding/json"
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

func (l *Listener) GetHandler(routingKey string, h *Handlers) (func(context.Context, amqp.Delivery) error, bool) {
	switch routingKey {
	case CmdPartyCreate:
		return h.HandleCreateParty, true
	case CmdPartyUpdate:
		return h.HandleUpdateParty, true
	case CmdPartyPatch:
		return h.HandlePatchParty, true
	case CmdPartyDelete:
		return h.HandleDeleteParty, true
	case QueryPartyGet:
		return h.HandleGetParty, true
	case QueryPartySearch:
		return h.HandleSearchParty, true
	default:
		return nil, false
	}
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
				targetHandler, valid := l.GetHandler(d.RoutingKey, h)
				if !valid {
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

					// If it's an RPC call (ReplyTo set), send error response
					if d.ReplyTo != "" {
						// Create a temporary handler to access replyTo helper or implement it here
						// Since we don't have access to handler instance easily here without passing it,
						// we can use the publisher from handlers if available, or just use the channel directly.
						// But wait, the handler 'h' IS available here in scope.

						// We need to construct a response with error.
						// Ideally we reuse h.replyTo but it's a method on Handlers.
						// h IS a *Handlers.

						// We need to publicize replyTo or duplicate logic.
						// Since replyTo is private, we can't call it easily unless we change visibility or duplicate.
						// Let's look at h definition. 'h' is *Handlers.

						// duplicating reply logic for now to avoid changing public API of Handlers if not needed,
						// OR better: Assume we can make replyTo public or add a method HandleError?

						// Let's try to send raw error response
						errResponse, _ := json.Marshal(map[string]string{"error": err.Error()})

						pubErr := ch.PublishWithContext(ctx,
							"",        // default exchange
							d.ReplyTo, // routing key = reply queue
							false,
							false,
							amqp.Publishing{
								ContentType:   "application/json",
								Headers:       amqp.Table{"user": d.Headers["user"]},
								CorrelationId: d.CorrelationId,
								Body:          errResponse,
							})
						if pubErr != nil {
							slog.Error("failed to publish error response", "error", pubErr)
						}

						d.Ack(false) // Ack because we handled it by reporting error
					} else {
						d.Nack(false, false) // Don't requeue, send to DLX
					}
				} else {
					d.Ack(false)
				}
			}(d)
		}
	}
}
