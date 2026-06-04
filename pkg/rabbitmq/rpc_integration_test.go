package rabbitmq_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
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

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = client.RequestWithHeaders(ctx, "", "non.existent.queue", map[string]string{}, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestRPCClient_StringResponseNotUnwrapped(t *testing.T) {
	// Regression: a response body that is a JSON string (starts with '"') must be
	// returned verbatim, not unwrapped. Verifies the double-marshal workaround is gone.
	client, err := rabbitmq.NewRPCClient(amqpURL)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	conn, err := amqp.Dial(amqpURL)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()

	q, err := ch.QueueDeclare("q.test.string.response", false, true, false, false, nil)
	require.NoError(t, err)

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	require.NoError(t, err)

	go func() {
		for d := range msgs {
			// Reply with a plain JSON string value (not an object).
			body, _ := json.Marshal("hello")
			if err := ch.PublishWithContext(context.Background(),
				"", d.ReplyTo, false, false,
				amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: d.CorrelationId,
					Body:          body,
				}); err != nil {
				log.Printf("failed to publish: %v", err)
			}
		}
	}()

	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := client.RequestWithHeaders(ctx, "", q.Name, map[string]string{}, nil)
	require.NoError(t, err)

	var decoded string
	require.NoError(t, json.Unmarshal(resp, &decoded))
	assert.Equal(t, "hello", decoded)
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

func TestRPCClient_SingleClientHighConcurrency(t *testing.T) {
	client, err := rabbitmq.NewRPCClient(amqpURL)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	conn, err := amqp.Dial(amqpURL)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()

	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	require.NoError(t, err)

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	require.NoError(t, err)

	go func() {
		for d := range msgs {
			var req struct {
				ID int `json:"id"`
			}
			if err := json.Unmarshal(d.Body, &req); err != nil {
				log.Printf("failed to unmarshal request: %v", err)
				continue
			}

			respBody, _ := json.Marshal(map[string]int{"id": req.ID})
			if err := ch.PublishWithContext(context.Background(),
				"", d.ReplyTo, false, false,
				amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: d.CorrelationId,
					Body:          respBody,
				}); err != nil {
				log.Printf("failed to publish response: %v", err)
			}
		}
	}()

	const requests = 200
	errCh := make(chan error, requests)
	var wg sync.WaitGroup

	for i := 0; i < requests; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			resp, err := client.RequestWithHeaders(ctx, "", q.Name, map[string]int{"id": i}, nil)
			if err != nil {
				errCh <- err
				return
			}

			var decoded struct {
				ID int `json:"id"`
			}
			if err := json.Unmarshal(resp, &decoded); err != nil {
				errCh <- err
				return
			}
			if decoded.ID != i {
				errCh <- fmt.Errorf("id mismatch: got %d want %d", decoded.ID, i)
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for callErr := range errCh {
		require.NoError(t, callErr)
	}
}

func TestRPCClient_WithPoolSize(t *testing.T) {
	client, err := rabbitmq.NewRPCClient(amqpURL, rabbitmq.WithPoolSize(4))
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	conn, err := amqp.Dial(amqpURL)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()

	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	require.NoError(t, err)

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	require.NoError(t, err)

	go func() {
		for d := range msgs {
			if err := ch.PublishWithContext(context.Background(),
				"", d.ReplyTo, false, false,
				amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: d.CorrelationId,
					Body:          []byte(`{"ok":true}`),
				}); err != nil {
				log.Printf("failed to publish: %v", err)
			}
		}
	}()

	const concurrency = 8
	var wg sync.WaitGroup
	errCh2 := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, err := client.RequestWithHeaders(ctx, "", q.Name, map[string]string{}, nil)
			if err != nil {
				errCh2 <- err
			}
		}()
	}

	wg.Wait()
	close(errCh2)
	for e := range errCh2 {
		require.NoError(t, e)
	}
}

func TestRPCClient_ReconnectDrainsInFlight(t *testing.T) {
	// Verifies that in-flight RPC calls receive ErrReconnecting when the connection drops.
	client, err := rabbitmq.NewRPCClient(amqpURL)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	// Start a request to a non-existent queue (it will never reply).
	callDone := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, err := client.Request(ctx, "q.never.replies", map[string]string{})
		callDone <- err
	}()

	// Give the goroutine time to register its pending entry.
	time.Sleep(100 * time.Millisecond)

	// Force-close the underlying connection to simulate a broker drop.
	client.Connection().Close() //nolint:errcheck

	select {
	case err := <-callDone:
		require.Error(t, err)
		assert.ErrorIs(t, err, rabbitmq.ErrReconnecting)
	case <-time.After(3 * time.Second):
		t.Fatal("in-flight call did not receive error after connection drop")
	}
}

func TestRPCClient_ReconnectFullCycle(t *testing.T) {
	// Full cycle: connection drop → ErrReconnecting for in-flight → automatic reconnect → calls succeed.
	// The broker (testcontainer) stays running; only the client's TCP connection is force-closed.
	// reconnectLoop redials the same URL and recovers transparently.
	client, err := rabbitmq.NewRPCClient(amqpURL)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	// Dedicated responder connection (independent of the client, so it survives the client reconnect).
	respConn, err := amqp.Dial(amqpURL)
	require.NoError(t, err)
	defer func() { _ = respConn.Close() }()

	respCh, err := respConn.Channel()
	require.NoError(t, err)
	defer func() { _ = respCh.Close() }()

	q, err := respCh.QueueDeclare("q.test.reconnect.cycle", false, true, false, false, nil)
	require.NoError(t, err)

	msgs, err := respCh.Consume(q.Name, "", true, false, false, false, nil)
	require.NoError(t, err)

	go func() {
		for d := range msgs {
			_ = respCh.PublishWithContext(context.Background(), "", d.ReplyTo, false, false,
				amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: d.CorrelationId,
					Body:          []byte(`{"ok":true}`),
				})
		}
	}()

	// Phase 1 — verify the client works before the drop.
	ctx1, cancel1 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel1()
	_, err = client.Request(ctx1, q.Name, map[string]string{})
	require.NoError(t, err, "pre-drop request must succeed")

	// Phase 2 — drop the connection; an in-flight call must get ErrReconnecting immediately.
	callDone := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, err := client.Request(ctx, "q.never.replies.cycle", map[string]string{})
		callDone <- err
	}()
	time.Sleep(80 * time.Millisecond)
	client.Connection().Close() //nolint:errcheck

	select {
	case err := <-callDone:
		require.ErrorIs(t, err, rabbitmq.ErrReconnecting)
	case <-time.After(3 * time.Second):
		t.Fatal("in-flight call did not error after connection drop")
	}

	// Phase 3 — reconnectLoop redials; poll until a fresh RPC call succeeds.
	// Reconnect fires after the initial 500 ms backoff, so 10 s is generous.
	require.Eventually(t, func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		_, err := client.Request(ctx, q.Name, map[string]string{})
		return err == nil
	}, 10*time.Second, 500*time.Millisecond, "client did not recover within 10 s after reconnect")
}

// startEchoResponder declares a transient queue, starts a goroutine that echoes every
// request back, and returns the queue name. Caller owns closing conn and ch.
func startEchoResponder(t testing.TB) (queueName string, conn *amqp.Connection, ch *amqp.Channel) {
	t.Helper()
	var err error
	conn, err = amqp.Dial(amqpURL)
	require.NoError(t, err)
	ch, err = conn.Channel()
	require.NoError(t, err)

	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	require.NoError(t, err)

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	require.NoError(t, err)

	go func() {
		for d := range msgs {
			_ = ch.PublishWithContext(context.Background(), "", d.ReplyTo, false, false,
				amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: d.CorrelationId,
					Body:          d.Body,
				})
		}
	}()
	return q.Name, conn, ch
}

func p99(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	idx := int(float64(len(sorted)) * 0.99)
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

// TestRPCClient_ConcurrentP99 asserts that the p99 latency under 10-goroutine concurrency
// is no worse than 5× the sequential p99 (generous bound — avoids CI flakiness while
// catching gross regressions like a serialising lock held for the full round-trip).
func TestRPCClient_ConcurrentP99(t *testing.T) {
	const (
		sequentialN = 100
		concurrency = 10
		perWorker   = 100
		factor      = 5 // p99_concurrent must be ≤ factor × p99_sequential
	)

	client, err := rabbitmq.NewRPCClient(amqpURL)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	qName, conn, ch := startEchoResponder(t)
	defer func() { _ = conn.Close(); _ = ch.Close() }()

	payload := map[string]string{"bench": "data"}

	// Sequential baseline.
	seqLatencies := make([]time.Duration, 0, sequentialN)
	for i := 0; i < sequentialN; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		start := time.Now()
		_, err := client.Request(ctx, qName, payload)
		seqLatencies = append(seqLatencies, time.Since(start))
		cancel()
		require.NoError(t, err)
	}
	p99Seq := p99(seqLatencies)

	// Concurrent load.
	concLatencies := make([]time.Duration, concurrency*perWorker)
	var wg sync.WaitGroup
	for w := 0; w < concurrency; w++ {
		w := w
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perWorker; i++ {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				start := time.Now()
				_, callErr := client.Request(ctx, qName, payload)
				concLatencies[w*perWorker+i] = time.Since(start)
				cancel()
				require.NoError(t, callErr)
			}
		}()
	}
	wg.Wait()

	p99Conc := p99(concLatencies)
	t.Logf("p99 sequential=%v  p99 concurrent(10×%d)=%v  ratio=%.1f×",
		p99Seq, perWorker, p99Conc, float64(p99Conc)/float64(p99Seq))

	assert.LessOrEqual(t, p99Conc, time.Duration(factor)*p99Seq,
		"p99 concurrent latency regressed beyond %d× sequential baseline", factor)
}

// BenchmarkRPCClient_Sequential measures single-goroutine RPC round-trip throughput.
// Run with: go test -bench=BenchmarkRPCClient -benchtime=5s ./rabbitmq/...
func BenchmarkRPCClient_Sequential(b *testing.B) {
	client, err := rabbitmq.NewRPCClient(amqpURL)
	require.NoError(b, err)
	defer func() { _ = client.Close() }()

	qName, conn, ch := startEchoResponder(b)
	defer func() { _ = conn.Close(); _ = ch.Close() }()

	payload := map[string]string{"bench": "data"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, _ = client.Request(ctx, qName, payload)
		cancel()
	}
}

// BenchmarkRPCClient_Concurrent10 measures throughput with 10 concurrent goroutines.
// Run with: go test -bench=BenchmarkRPCClient -benchtime=5s -cpu=1 ./rabbitmq/...
func BenchmarkRPCClient_Concurrent10(b *testing.B) {
	client, err := rabbitmq.NewRPCClient(amqpURL)
	require.NoError(b, err)
	defer func() { _ = client.Close() }()

	qName, conn, ch := startEchoResponder(b)
	defer func() { _ = conn.Close(); _ = ch.Close() }()

	payload := map[string]string{"bench": "data"}
	b.SetParallelism(10)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_, _ = client.Request(ctx, qName, payload)
			cancel()
		}
	})
}
