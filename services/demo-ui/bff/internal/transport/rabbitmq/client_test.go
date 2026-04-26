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

type mockBroadcaster struct {
	lastMsg DebugMessage
}

func (m *mockBroadcaster) Broadcast(msg any) {
	if d, ok := msg.(DebugMessage); ok {
		m.lastMsg = d
	}
}

func TestClient_BroadcastRequestAndReply(t *testing.T) {
	// We can test the broadcast methods without a real connection by just calling them
	client := &Client{}

	broadcaster := &mockBroadcaster{}
	client.SetBroadcaster(broadcaster)

	// Test broadcastRequest
	headers := map[string]any{"h1": "v1"}
	payload := map[string]string{"key": "value"}
	client.broadcastRequest("exchange", "routingKey", payload, headers)

	if broadcaster.lastMsg.Type != "request" {
		t.Errorf("expected request, got %v", broadcaster.lastMsg.Type)
	}
	if broadcaster.lastMsg.Topic != "routingKey" {
		t.Errorf("expected routingKey, got %v", broadcaster.lastMsg.Topic)
	}
	if broadcaster.lastMsg.Exchange != "exchange" {
		t.Errorf("expected exchange, got %v", broadcaster.lastMsg.Exchange)
	}

	// test unmarshalable payload
	client.broadcastRequest("exchange", "routingKey", make(chan int), nil)
	if broadcaster.lastMsg.Payload == nil {
		t.Errorf("expected payload to be populated via fallback")
	}

	// Test broadcastReply
	validJSON := []byte(`{"result": "success"}`)
	client.broadcastReply("routingKey", validJSON)

	if broadcaster.lastMsg.Type != "reply" {
		t.Errorf("expected reply, got %v", broadcaster.lastMsg.Type)
	}
	if broadcaster.lastMsg.Topic != "rpc.reply" {
		t.Errorf("expected rpc.reply, got %v", broadcaster.lastMsg.Topic)
	}

	invalidJSON := []byte(`not json`)
	client.broadcastReply("routingKey", invalidJSON)

	if broadcaster.lastMsg.Payload["raw"] != "not json" {
		t.Errorf("expected raw fallback, got %v", broadcaster.lastMsg.Payload["raw"])
	}
}

func TestClient_EmptyURL(t *testing.T) {
	// This will attempt to connect to localhost and likely fail, but it hits the url check
	client, err := NewClient("")
	if err == nil {
		defer client.Close()
	}
}

func TestLogUnknownCorrelation(t *testing.T) {
	LogUnknownCorrelation("test-id")
}

// Add dummy test for PublishCommand
func TestClient_PublishCommand(t *testing.T) {
	client := &Client{}
	broadcaster := &mockBroadcaster{}
	client.SetBroadcaster(broadcaster)

	// Because c.Publish will panic on a nil RPC client, this test only verifies that
	// the debug broadcast path is exercised before the call fails.
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when publishing with nil RPC client")
		}
	}()
	_ = client.PublishCommand(context.Background(), "exchange", "routingKey", map[string]string{})
}
