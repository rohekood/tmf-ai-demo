package rabbitmq

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestClient_Integration(t *testing.T) {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}

	// Try to connect to check if RabbitMQ is up
	conn, err := amqp.Dial(url)
	if err != nil {
		t.Skipf("RabbitMQ not available at %s, skipping integration test: %v", url, err)
	}
	_ = conn.Close()

	// 1. Create Client
	client, err := NewClient(url)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// 2. Setup a "Server" that responds to RPC
	exchange := "test.exchange"
	routingKey := "test.rpc"

	ch, err := client.Connection().Channel()
	if err != nil {
		t.Fatalf("Failed to open channel: %v", err)
	}
	defer func() { _ = ch.Close() }()

	err = ch.ExchangeDeclare(exchange, "topic", false, true, false, false, nil)
	if err != nil {
		t.Fatalf("Failed to declare exchange: %v", err)
	}

	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		t.Fatalf("Failed to declare queue: %v", err)
	}

	err = ch.QueueBind(q.Name, routingKey, exchange, false, nil)
	if err != nil {
		t.Fatalf("Failed to bind: %v", err)
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		t.Fatalf("Failed to consume: %v", err)
	}

	// Server Loop
	go func() {
		for d := range msgs {
			var req map[string]string
			_ = json.Unmarshal(d.Body, &req)

			res := map[string]string{"result": "echo-" + req["data"]}
			resBytes, _ := json.Marshal(res)

			_ = ch.Publish(
				"",        // default exchange
				d.ReplyTo, // routing key
				false,
				false,
				amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: d.CorrelationId,
					Body:          resBytes,
				})
		}
	}()

	// 3. Test CallRPC
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := map[string]string{"data": "test"}
	resBytes, err := client.CallRPC(ctx, exchange, routingKey, req, nil)
	if err != nil {
		t.Fatalf("CallRPC failed: %v", err)
	}

	var res map[string]string
	err = json.Unmarshal(resBytes, &res)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if res["result"] != "echo-test" {
		t.Errorf("Expected echo-test, got %s", res["result"])
	}
}
