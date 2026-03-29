package rabbitmq

import (
	"context"
	"fmt"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rmqContainer "github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

type rabbitSetupResult struct {
	connStr string
	err     error
}

func setupRabbitMQ(t *testing.T) string {
	t.Helper()
	ctx := context.Background()

	resultCh := make(chan rabbitSetupResult, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultCh <- rabbitSetupResult{err: fmt.Errorf("testcontainers panic: %v", r)}
			}
		}()
		container, err := rmqContainer.Run(ctx, "rabbitmq:3-management")
		if err != nil {
			resultCh <- rabbitSetupResult{err: err}
			return
		}
		t.Cleanup(func() {
			if err := container.Terminate(ctx); err != nil {
				t.Logf("failed to terminate container: %s", err)
			}
		})
		connStr, err := container.AmqpURL(ctx)
		resultCh <- rabbitSetupResult{connStr: connStr, err: err}
	}()

	result := <-resultCh
	if result.err != nil {
		t.Skipf("Skipping test: RabbitMQ container unavailable: %v", result.err)
		return ""
	}

	return result.connStr
}

func TestConnectionManager_Connect_Success(t *testing.T) {
	url := setupRabbitMQ(t)
	if url == "" {
		return
	}

	mgr := NewConnectionManager(url)
	err := mgr.Connect()
	assert.NoError(t, err)
	assert.NotNil(t, mgr.GetConnection())
	assert.NotNil(t, mgr.GetChannel())

	err = mgr.Close()
	assert.NoError(t, err)
}

func TestConnectionManager_Close_WithConnectionAndChannel(t *testing.T) {
	url := setupRabbitMQ(t)
	if url == "" {
		return
	}

	mgr := NewConnectionManager(url)
	require.NoError(t, mgr.Connect())

	assert.NotNil(t, mgr.GetConnection())
	assert.NotNil(t, mgr.GetChannel())

	err := mgr.Close()
	assert.NoError(t, err)
}

func TestConnectionManager_Connect_ChannelError(t *testing.T) {
	// Inject a dialFn that returns a closed connection to trigger channel error
	dialFn := func(url string) (*amqp.Connection, error) {
		conn, err := amqp.Dial(url)
		if err != nil {
			return nil, err
		}
		// Close the connection immediately so Channel() fails
		_ = conn.Close()
		return conn, nil
	}

	url := setupRabbitMQ(t)
	if url == "" {
		return
	}

	mgr := NewConnectionManager(url, dialFn)
	err := mgr.Connect()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open channel")
}

func TestPublisher_NewPublisher_Success(t *testing.T) {
	url := setupRabbitMQ(t)
	if url == "" {
		return
	}

	conn, err := amqp.Dial(url)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	pub, err := NewPublisher(conn)
	assert.NoError(t, err)
	assert.NotNil(t, pub)

	err = pub.Close()
	assert.NoError(t, err)
}

func TestPublisher_Publish_Success(t *testing.T) {
	url := setupRabbitMQ(t)
	if url == "" {
		return
	}

	conn, err := amqp.Dial(url)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	pub, err := NewPublisher(conn)
	require.NoError(t, err)
	defer func() { _ = pub.Close() }()

	// Declare exchange
	ch, err := pub.GetChannel()
	require.NoError(t, err)
	err = ch.ExchangeDeclare("test.exchange", "topic", false, true, false, false, nil)
	require.NoError(t, err)

	// Publish a message
	ctx := context.Background()
	err = pub.Publish(ctx, "test.exchange", "test.key", map[string]string{"key": "value"})
	assert.NoError(t, err)
}

func TestPublisher_Publish_WithContextHeaders(t *testing.T) {
	url := setupRabbitMQ(t)
	if url == "" {
		return
	}

	conn, err := amqp.Dial(url)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	pub, err := NewPublisher(conn)
	require.NoError(t, err)
	defer func() { _ = pub.Close() }()

	ch, err := pub.GetChannel()
	require.NoError(t, err)
	err = ch.ExchangeDeclare("test.exchange2", "topic", false, true, false, false, nil)
	require.NoError(t, err)

	ctx := context.Background()
	// Using domain context keys
	type contextKey string
	ctx = context.WithValue(ctx, contextKey("user"), "test-user")
	ctx = context.WithValue(ctx, contextKey("Authorization"), "Bearer token")

	err = pub.Publish(ctx, "test.exchange2", "test.key2", "hello")
	assert.NoError(t, err)
}

func TestPublisher_GetChannel_Success(t *testing.T) {
	url := setupRabbitMQ(t)
	if url == "" {
		return
	}

	conn, err := amqp.Dial(url)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	pub, err := NewPublisher(conn)
	require.NoError(t, err)
	defer func() { _ = pub.Close() }()

	ch, err := pub.GetChannel()
	assert.NoError(t, err)
	assert.NotNil(t, ch)
}

func TestConnectionManager_HandleReconnect_ContextCancel(t *testing.T) {
	url := setupRabbitMQ(t)
	if url == "" {
		return
	}

	mgr := NewConnectionManager(url)
	require.NoError(t, mgr.Connect())

	// Cancel context and wait for handleReconnect to exit
	mgr.cancel()
	time.Sleep(100 * time.Millisecond)

	err := mgr.Close()
	assert.NoError(t, err)
}

func TestPublisher_NewPublisher_Error_Integration(t *testing.T) {
	// Closed/empty connection will fail on conn.Channel()
	conn := &amqp.Connection{}

	done := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("panic: %v", r)
			}
		}()
		_, e := NewPublisher(conn)
		done <- e
	}()
	err := <-done
	assert.Error(t, err)
}

func TestPublisher_Publish_PublishError_Integration(t *testing.T) {
	url := setupRabbitMQ(t)
	if url == "" {
		return
	}
	conn, err := amqp.Dial(url)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	pub, err := NewPublisher(conn)
	require.NoError(t, err)

	// Close the channel to force PublishWithContext to fail
	_ = pub.Close()

	err = pub.Publish(context.Background(), "ex", "rk", map[string]string{"a": "b"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to publish message")
}

func TestConnectionManager_HandleReconnect_RetryLoop(t *testing.T) {
	url := setupRabbitMQ(t)
	if url == "" {
		return
	}

	mgr := NewConnectionManager(url)
	require.NoError(t, mgr.Connect())

	// Force the underlying connection to close, triggering the handleReconnect loop
	_ = mgr.conn.Close()

	// Wait for the reconnect loop to try at least once (sleeps 5s)
	time.Sleep(6 * time.Second)

	// Clean up
	_ = mgr.Close()
}
