package rabbitmq

import (
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// DebugMessage represents a captured RabbitMQ message for the UI
type DebugMessage struct {
	ID            string         `json:"id"` // Unique ID for UI tracking
	Timestamp     time.Time      `json:"timestamp"`
	Type          string         `json:"type"` // cmd, evt, query
	Topic         string         `json:"topic"`
	Exchange      string         `json:"exchange,omitempty"`
	CorrelationID string         `json:"correlationId,omitempty"`
	ReplyTo       string         `json:"replyTo,omitempty"`
	Payload       map[string]any `json:"payload"`
	Headers       map[string]any `json:"headers,omitempty"`
	Service       string         `json:"service"` // derived from topic prefix
}

// Broadcaster interface for decoupling from HTTP package
type Broadcaster interface {
	Broadcast(msg any)
}

// DebugConsumer consumes all messages for debugging purposes
type DebugConsumer struct {
	client      *Client
	broadcaster Broadcaster
}

func NewDebugConsumer(client *Client, broadcaster Broadcaster) *DebugConsumer {
	return &DebugConsumer{
		client:      client,
		broadcaster: broadcaster,
	}
}

// debugExchanges is the list of exchanges the debug consumer binds to.
// Defined at package level so tests can assert the complete set without a live broker.
var debugExchanges = []string{
	"tmf.party",
	"tmf.customer",
	"ex.domain.market",   // Qualification service
	"ex.domain.commerce", // Shopping Cart service
	"ex.domain.order",    // POCV saga service
}

// StartSubscribing sets up a queue to listen to everything on the topic exchange
func (dc *DebugConsumer) StartSubscribing(exchangeName string) error {
	ch, err := dc.client.Connection().Channel()
	if err != nil {
		return fmt.Errorf("failed updates channel: %w", err)
	}

	// Declare a temporary exclusive queue for debugging
	q, err := ch.QueueDeclare(
		"",    // name - generated
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed declare debug queue: %w", err)
	}

	// Bind to all topics
	exchanges := debugExchanges

	for _, exchange := range exchanges {
		// Ensure exchange exists (should be declared by services, but safer here)
		err = ch.ExchangeDeclare(
			exchange,
			"topic",
			true,  // durable
			false, // auto-deleted
			false, // internal
			false, // no-wait
			nil,   // args
		)
		if err != nil {
			// Ignore error as exchange might already exist with different config or service not up.
			// Debug consumer will just fail to bind if exchange is missing.
			_ = err
		}

		err = ch.QueueBind(
			q.Name,
			"#",      // routing key - listen to everything
			exchange, // exchange
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to bind debug queue to %s: %w", exchange, err)
		}
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("failed register debug consumer: %w", err)
	}

	go dc.handleMessages(msgs)

	return nil
}

func (dc *DebugConsumer) handleMessages(msgs <-chan amqp.Delivery) {
	for d := range msgs {
		msgType := "unknown"
		if len(d.RoutingKey) > 4 {
			if d.RoutingKey[:3] == "cmd" {
				msgType = "command"
			} else if d.RoutingKey[:3] == "evt" {
				msgType = "event"
			} else if d.RoutingKey[:5] == "query" {
				msgType = "query"
			}
		}

		service := "unknown"
		if len(d.Exchange) > 4 && d.Exchange[:4] == "tmf." {
			service = d.Exchange[4:]
		} else {
			// Map ordering-related exchanges to "ordering"
			orderingExchanges := map[string]string{
				"ex.domain.market":   "ordering",
				"ex.domain.commerce": "ordering",
				"ex.domain.order":    "ordering",
			}
			if mapped, ok := orderingExchanges[d.Exchange]; ok {
				service = mapped
			}
		}

		var payload map[string]any
		_ = json.Unmarshal(d.Body, &payload) // Ignore error, payload might be raw string or empty

		// If payload failed to unmarshal, store as raw string if possible or empty
		if payload == nil {
			payload = map[string]any{
				"raw": string(d.Body),
			}
		}

		debugMsg := DebugMessage{
			ID:            fmt.Sprintf("%s-%d", d.MessageId, time.Now().UnixNano()), // Generate ID if missing
			Timestamp:     time.Now().UTC(),
			Type:          msgType,
			Topic:         d.RoutingKey,
			CorrelationID: d.CorrelationId,
			ReplyTo:       d.ReplyTo,
			Payload:       payload,
			Service:       service,
		}

		dc.broadcaster.Broadcast(debugMsg)
	}
}
