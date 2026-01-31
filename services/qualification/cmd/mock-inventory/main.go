package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"tmf/pkg/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
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
		"q.mock.inventory.rpc", // name
		false,                  // durable
		false,                  // delete when unused
		false,                  // exclusive
		false,                  // no-wait
		nil,                    // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	// Bind
	err = ch.QueueBind(q.Name, rabbitmq.QueryInventoryResourceCapacity, exchange, false, nil)
	if err != nil {
		log.Fatalf("Failed to bind queue: %v", err)
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to consume: %v", err)
	}

	logger.Info("Mock Inventory Started. Waiting for RPC queries...")

	go func() {
		for d := range msgs {
			logger.Info("Received Query", "correlation_id", d.CorrelationId, "reply_to", d.ReplyTo)

			// Logic: If locationId is "FULL_CABINET", return 0 free ports
			// Request payload: { "locationId": "..." }
			var req struct {
				LocationID string `json:"locationId"`
			}
			_ = json.Unmarshal(d.Body, &req)

			free := 5
			if req.LocationID == "FULL_CABINET" {
				free = 0
			}

			response := map[string]interface{}{
				"total":    16,
				"used":     11,
				"reserved": 0,
				"free":     free,
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

	// Handle Inventory Reservation (Order Domain)
	exchangeOrder := "ex.domain.order"
	_ = ch.ExchangeDeclare(exchangeOrder, "topic", true, false, false, false, nil)

	qOrder, _ := ch.QueueDeclare("q.mock.inventory.order", false, false, false, false, nil)
	_ = ch.QueueBind(qOrder.Name, rabbitmq.CmdInventoryResourceReserve, exchangeOrder, false, nil)

	msgsOrder, _ := ch.Consume(qOrder.Name, "", false, false, false, false, nil)

	go func() {
		for d := range msgsOrder {
			logger.Info("Received Inventory Reserve Command", "routing_key", d.RoutingKey)

			var req struct {
				SagaID string `json:"sagaId"`
			}
			_ = json.Unmarshal(d.Body, &req)

			// Always succeed
			resp := map[string]string{
				"orderId": req.SagaID,
			}
			body, _ := json.Marshal(resp)

			err := ch.PublishWithContext(context.Background(),
				exchangeOrder,
				rabbitmq.EvtInventoryResourceReserved,
				false,
				false,
				amqp.Publishing{
					ContentType: "application/json",
					Body:        body,
				})

			if err != nil {
				logger.Error("Failed to publish reservation event", "error", err)
			}
			if err := d.Ack(false); err != nil {
				logger.Error("Failed to Ack order command", "error", err)
			}
		}
	}()

	// Wait for signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
