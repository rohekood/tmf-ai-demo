package rabbitmq_test

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcRabbitmq "github.com/testcontainers/testcontainers-go/modules/rabbitmq"

	"tmf/pkg/rabbitmq"
)

var (
	amqpURL         string
	rabbitContainer *tcRabbitmq.RabbitMQContainer
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Start RabbitMQ container
	rmq, err := tcRabbitmq.Run(ctx,
		"rabbitmq:3.12-management",
		tcRabbitmq.WithAdminPassword("guest"),
		tcRabbitmq.WithAdminUsername("guest"),
	)
	if err != nil {
		log.Fatalf("failed to start rabbitmq: %v", err)
	}
	rabbitContainer = rmq

	amqpURL, err = rmq.AmqpURL(ctx)
	if err != nil {
		log.Fatalf("failed to get amqp url: %v", err)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	if err := rmq.Terminate(ctx); err != nil {
		log.Printf("failed to terminate rabbitmq: %v", err)
	}

	os.Exit(code)
}

func TestRPCClient_DefaultDirectReplyTo(t *testing.T) {
	// Default behavior uses RabbitMQ direct reply-to pseudo-queue.
	client1, err := rabbitmq.NewRPCClient(amqpURL)
	require.NoError(t, err)
	defer func() { _ = client1.Close() }()

	client2, err := rabbitmq.NewRPCClient(amqpURL)
	require.NoError(t, err)
	defer func() { _ = client2.Close() }()

	assert.Equal(t, rabbitmq.DirectReplyToQueue, client1.ReplyQueue())
	assert.Equal(t, rabbitmq.DirectReplyToQueue, client2.ReplyQueue())
}

func TestRPCClient_RequestReply(t *testing.T) {
	// Create RPC client with custom exchange
	client, err := rabbitmq.NewRPCClient(amqpURL, rabbitmq.WithExchange("ex.test.rpc"))
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	// Setup a mock responder
	conn, err := amqp.Dial(amqpURL)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()

	// Declare the exchange
	err = ch.ExchangeDeclare("ex.test.rpc", "topic", true, false, false, false, nil)
	require.NoError(t, err)

	// Declare and bind request queue
	q, err := ch.QueueDeclare("q.test.rpc.requests", false, true, false, false, nil)
	require.NoError(t, err)

	err = ch.QueueBind(q.Name, "test.ping", "ex.test.rpc", false, nil)
	require.NoError(t, err)

	// Consume requests and respond
	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	require.NoError(t, err)

	// Start responder goroutine
	go func() {
		for d := range msgs {
			// Echo back the request with "pong" response
			var req map[string]string
			if err := json.Unmarshal(d.Body, &req); err != nil {
				log.Printf("failed to unmarshal request: %v", err)
				continue
			}

			response := map[string]string{
				"response": "pong",
				"echo":     req["message"],
			}
			respBody, _ := json.Marshal(response)

			if err := ch.PublishWithContext(context.Background(),
				"",        // default exchange
				d.ReplyTo, // reply to the client's queue
				false, false,
				amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: d.CorrelationId,
					Body:          respBody,
				}); err != nil {
				log.Printf("failed to publish response: %v", err)
			}
		}
	}()

	// Give responder time to start
	time.Sleep(100 * time.Millisecond)

	// Send RPC request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	payload := map[string]string{"message": "hello"}
	respBytes, err := client.Request(ctx, "test.ping", payload)
	require.NoError(t, err)

	var resp map[string]string
	err = json.Unmarshal(respBytes, &resp)
	require.NoError(t, err)

	assert.Equal(t, "pong", resp["response"])
	assert.Equal(t, "hello", resp["echo"])
}

func TestRPCClient_RequestWithHeaders(t *testing.T) {
	client, err := rabbitmq.NewRPCClient(amqpURL)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	conn, err := amqp.Dial(amqpURL)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()

	// Use default exchange with direct routing to queue
	q, err := ch.QueueDeclare("q.test.headers", false, true, false, false, nil)
	require.NoError(t, err)

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	require.NoError(t, err)

	receivedHeaders := make(chan amqp.Table, 1)

	go func() {
		for d := range msgs {
			receivedHeaders <- d.Headers

			// Send response
			if err := ch.PublishWithContext(context.Background(),
				"", d.ReplyTo, false, false,
				amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: d.CorrelationId,
					Body:          []byte(`{"ok": true}`),
				}); err != nil {
				log.Printf("failed to publish response: %v", err)
			}
		}
	}()

	time.Sleep(100 * time.Millisecond)

	// Create context with values
	ctx := context.WithValue(context.Background(), rabbitmq.Key(rabbitmq.HeaderCorrelationID), "test-corr-123")
	ctx = context.WithValue(ctx, rabbitmq.Key(rabbitmq.HeaderUser), "test-user")

	// Send request with additional headers
	customHeaders := map[string]any{
		"X-Custom-Header": "custom-value",
	}

	_, err = client.RequestWithHeaders(ctx, "", q.Name, map[string]string{"test": "data"}, customHeaders)
	require.NoError(t, err)

	// Check received headers
	select {
	case headers := <-receivedHeaders:
		assert.Equal(t, "test-corr-123", headers["X-Correlation-ID"])
		assert.Equal(t, "test-user", headers["user"])
		assert.Equal(t, "custom-value", headers["X-Custom-Header"])
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for headers")
	}
}

func TestRPCClient_Timeout(t *testing.T) {
	client, err := rabbitmq.NewRPCClient(amqpURL)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	// Request to non-existent queue - should timeout
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	_, err = client.RequestWithHeaders(ctx, "", "non.existent.queue", map[string]string{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestRPCClient_ContextCancellation(t *testing.T) {
	client, err := rabbitmq.NewRPCClient(amqpURL)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	_, err = client.RequestWithHeaders(ctx, "", "some.queue", map[string]string{}, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestRPCClient_MultipleClientsIsolation(t *testing.T) {
	// Verify that replies go to the correct client even with multiple clients
	client1, err := rabbitmq.NewRPCClient(amqpURL)
	require.NoError(t, err)
	defer func() { _ = client1.Close() }()

	client2, err := rabbitmq.NewRPCClient(amqpURL)
	require.NoError(t, err)
	defer func() { _ = client2.Close() }()

	// Setup responder that echoes client identifier
	conn, err := amqp.Dial(amqpURL)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()

	q, err := ch.QueueDeclare("q.test.isolation", false, true, false, false, nil)
	require.NoError(t, err)

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	require.NoError(t, err)

	go func() {
		for d := range msgs {
			var req map[string]string
			if err := json.Unmarshal(d.Body, &req); err != nil {
				log.Printf("failed to unmarshal: %v", err)
				continue
			}

			// Echo back the client ID
			response := map[string]string{"clientId": req["clientId"]}
			respBody, _ := json.Marshal(response)

			if err := ch.PublishWithContext(context.Background(),
				"", d.ReplyTo, false, false,
				amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: d.CorrelationId,
					Body:          respBody,
				}); err != nil {
				log.Printf("failed to publish: %v", err)
			}
		}
	}()

	time.Sleep(100 * time.Millisecond)

	ctx := context.Background()

	// Send concurrent requests from both clients
	done := make(chan struct{}, 2)

	go func() {
		resp, err := client1.RequestWithHeaders(ctx, "", q.Name, map[string]string{"clientId": "client1"}, nil)
		require.NoError(t, err)
		var r map[string]string
		_ = json.Unmarshal(resp, &r)
		assert.Equal(t, "client1", r["clientId"])
		done <- struct{}{}
	}()

	go func() {
		resp, err := client2.RequestWithHeaders(ctx, "", q.Name, map[string]string{"clientId": "client2"}, nil)
		require.NoError(t, err)
		var r map[string]string
		_ = json.Unmarshal(resp, &r)
		assert.Equal(t, "client2", r["clientId"])
		done <- struct{}{}
	}()

	// Wait for both to complete
	<-done
	<-done
}
