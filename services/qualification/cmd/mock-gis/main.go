package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
	"tmf/pkg/rabbitmq"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close connection: %v", err)
		}
	}()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}
	defer func() {
		if err := ch.Close(); err != nil {
			log.Printf("Failed to close channel: %v", err)
		}
	}()

	// Declare Exchange
	exchange := "ex.domain.market"
	err = ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare exchange: %v", err)
	}

	// Declare Queue
	q, err := ch.QueueDeclare(
		"q.mock.gis.rpc", // name
		false,            // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	// Bind
	err = ch.QueueBind(q.Name, rabbitmq.QueryGISGeographyCheck, exchange, false, nil)
	if err != nil {
		log.Fatalf("Failed to bind queue: %v", err)
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to consume: %v", err)
	}

	logger.Info("Mock GIS Started. Waiting for RPC queries...")

	go func() {
		for d := range msgs {
			logger.Info("Received Query",
				"correlation_id", d.CorrelationId,
				"reply_to", d.ReplyTo,
				"routing_key", d.RoutingKey,
				"exchange", d.Exchange,
				"body", string(d.Body))

			// Logic: If City is "Nowhere", return false
			inFootprint := true
			// We need to parse the incoming request to check criteria
			// Request payload: { "address": { "city": "..." }, ... }
			var req struct {
				Address struct {
					City string `json:"city"`
				} `json:"address"`
			}
			_ = json.Unmarshal(d.Body, &req)

			if req.Address.City == "Nowhere" {
				inFootprint = false
			}

			response := map[string]interface{}{
				"inFootprint": inFootprint,
				"zoneId":      "ZONE_A",
				"technology":  "GPON",
			}
			body, _ := json.Marshal(response)

			// Send Reply
			err = ch.PublishWithContext(context.Background(),
				"",        // default exchange
				d.ReplyTo, // routing key = reply queue
				false,
				false,
				amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: d.CorrelationId,
					Body:          body,
				})

			if err != nil {
				logger.Error("Failed to reply", "error", err)
				if err := d.Nack(false, false); err != nil {
					logger.Error("Failed to Nack", "error", err)
				}
			} else {
				if err := d.Ack(false); err != nil {
					logger.Error("Failed to Ack", "error", err)
				}
			}
		}
	}()

	// Wait for signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
