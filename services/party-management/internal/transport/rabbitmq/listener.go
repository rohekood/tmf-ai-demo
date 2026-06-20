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
	// autoDeclare creates exchanges/queues on startup (dev/local). When false
	// (production) the topology must already exist and is only verified/bound.
	autoDeclare bool
}

func NewListener(conn *amqp.Connection, autoDeclare bool) (*Listener, error) {
	return &Listener{conn: conn, autoDeclare: autoDeclare}, nil
}

// declareQueue creates the queue (durable) when autoDeclare is on, otherwise
// verifies it exists passively. Returns the resolved queue name.
func (l *Listener) declareQueue(ch *amqp.Channel, name string, args amqp.Table) (string, error) {
	if l.autoDeclare {
		q, err := ch.QueueDeclare(name, true, false, false, false, args)
		if err != nil {
			return "", err
		}
		return q.Name, nil
	}
	q, err := ch.QueueDeclarePassive(name, true, false, false, false, nil)
	if err != nil {
		return "", err
	}
	return q.Name, nil
}

// declareExchanges creates the exchanges this service uses. No-op unless
// autoDeclare is on (in production they are provisioned out-of-band).
func (l *Listener) declareExchanges(ch *amqp.Channel) error {
	if !l.autoDeclare {
		return nil
	}
	exchanges := []struct {
		name string
		kind string
	}{
		{CommandExchange, "topic"},
		{EventExchange, "topic"},
		{DeadLetterExchange, "fanout"},
	}
	for _, e := range exchanges {
		if err := ch.ExchangeDeclare(e.name, e.kind, true, false, false, false, nil); err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", e.name, err)
		}
	}
	return nil
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
	case CmdPartyFinalizeDeletion:
		return h.HandleFinalizeDeletion, true
	case CmdPartyCancelDeletion:
		return h.HandleCancelDeletion, true
	case CmdPartyPurge:
		return h.HandlePurgeParty, true
	default:
		return nil, false
	}
}

func (l *Listener) Start(ctx context.Context, h *Handlers) error {
	ch, err := l.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer func() { _ = ch.Close() }()

	if err := l.declareExchanges(ch); err != nil {
		return err
	}

	if _, err = l.declareQueue(ch, DeadLetterQueue, nil); err != nil {
		return fmt.Errorf("failed to declare DLQ: %w", err)
	}

	err = ch.QueueBind(DeadLetterQueue, "", DeadLetterExchange, false, nil)
	if err != nil {
		slog.Warn("failed to bind DLQ to DLX (ignoring)", "error", err)
	}

	// Main command queue dead-letters to the DLX when auto-declared.
	var partyQueueArgs amqp.Table
	if l.autoDeclare {
		partyQueueArgs = amqp.Table{"x-dead-letter-exchange": DeadLetterExchange}
	}
	partyQueueName, err := l.declareQueue(ch, PartyQueue, partyQueueArgs)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}
	q := amqp.Queue{Name: partyQueueName}

	routingKeys := []string{
		CmdPartyCreate,
		CmdPartyUpdate,
		CmdPartyPatch,
		CmdPartyDelete,
		CmdPartyFinalizeDeletion,
		CmdPartyCancelDeletion,
		CmdPartyPurge,
		QueryPartyGet,
		QueryPartySearch,
	}

	for _, rk := range routingKeys {
		err = ch.QueueBind(q.Name, rk, CommandExchange, false, nil)
		if err != nil {
			return fmt.Errorf("failed to bind command queue: %w", err)
		}
	}

	// Bind to Event Queue
	eventQueueName := "party.events"
	if _, err = l.declareQueue(ch, eventQueueName, nil); err != nil {
		return fmt.Errorf("failed to declare event queue: %w", err)
	}

	// Bind to existing Event Exchange
	err = ch.QueueBind(eventQueueName, EvtCustomerCreated, EventExchange, false, nil)
	if err != nil {
		return fmt.Errorf("failed to bind event queue: %w", err)
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
		eventQueueName,
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

	slog.Info("Party Management listener started...")

	for {
		select {
		case <-ctx.Done():
			return nil
		case d := <-eventMsgs:
			go func(d amqp.Delivery) {
				// Simple event handling
				var err error
				switch d.RoutingKey {
				case EvtCustomerCreated:
					err = h.HandleCustomerCreated(ctx, d)
				default:
					slog.Warn("ignoring unknown event", "routingKey", d.RoutingKey)
				}

				if err != nil {
					slog.Error("failed to handle event", "routingKey", d.RoutingKey, "error", err)
					if ackErr := d.Nack(false, false); ackErr != nil {
						slog.Error("failed to nack message", "error", ackErr)
					}
				} else {
					if ackErr := d.Ack(false); ackErr != nil {
						slog.Error("failed to ack message", "error", ackErr)
					}
				}
			}(d)
		case d := <-msgs:
			go func(d amqp.Delivery) {
				var targetHandler func(context.Context, amqp.Delivery) error
				targetHandler, valid := l.GetHandler(d.RoutingKey, h)
				if !valid {
					slog.Warn("unknown routing key", "routing_key", d.RoutingKey)
					if ackErr := d.Nack(false, false); ackErr != nil {
						slog.Error("failed to nack message", "error", ackErr)
					}
					return
				}

				// Wrap with middlewares
				wrappedHandler := Chain(targetHandler,
					// TracingMiddleware("party-management"),
					AuthMiddleware(),
					JWTMiddleware())

				err := wrappedHandler(ctx, d)
				if err != nil {
					slog.Error("error handling message", "routing_key", d.RoutingKey, "error", err)

					// If it's an RPC call (ReplyTo set), send error response
					if d.ReplyTo != "" {
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

						if ackErr := d.Ack(false); ackErr != nil {
							slog.Error("failed to ack message (error reported)", "error", ackErr)
						}
					} else {
						if ackErr := d.Nack(false, false); ackErr != nil {
							slog.Error("failed to nack message (dlx)", "error", ackErr)
						}
					}
				} else {
					if ackErr := d.Ack(false); ackErr != nil {
						slog.Error("failed to ack message", "error", ackErr)
					}
				}
			}(d)
		}
	}
}
