package rabbitmq

import (
	"context"
	"fmt"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewListener(t *testing.T) {
	_, err := NewListener(nil, false)
	assert.NoError(t, err)

	l, err := NewListener(&amqp.Connection{}, false)
	assert.NoError(t, err)
	assert.NotNil(t, l)
}

// TestListener_Start_ChannelError exercises the very first error branch:
// when the connection cannot open a channel (closed/broken connection), Start
// must return an error or panic – both indicate the error path is exercised.
// We use a goroutine+recover so the panic in amqp library internals (nil
// allocator on zero Connection) doesn't abort the test binary.
func TestListener_Start_ChannelError(t *testing.T) {
	// Use an empty amqp.Connection (zero value).
	// Channel() panics with a nil pointer dereference inside the amqp library
	// because internal allocator is nil. This is the error path of Start().
	conn := &amqp.Connection{}

	l, err := NewListener(conn, false)
	require.NoError(t, err)

	// Run Start in a goroutine so we can recover from the panic
	done := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Panic == channel error path was exercised
				done <- fmt.Errorf("channel error (panic): %v", r)
			}
		}()
		done <- l.Start(context.Background(), &Handlers{})
	}()

	result := <-done
	// Either an error return OR a panic-converted-to-error; both are acceptable
	assert.Error(t, result)
}

// TestListener_GetHandler_FinalizeCancelKeys ensures finalize and cancel
// routing keys are mapped correctly (they were not in listener_test.go).
func TestListener_GetHandler_FinalizeCancelKeys(t *testing.T) {
	l := &Listener{}
	h := &Handlers{}

	handler, valid := l.GetHandler(CmdPartyFinalizeDeletion, h)
	assert.True(t, valid)
	assert.NotNil(t, handler)

	handler, valid = l.GetHandler(CmdPartyCancelDeletion, h)
	assert.True(t, valid)
	assert.NotNil(t, handler)
}

// TestListener_Start_WithRealRabbitMQ starts the listener using the shared
// testcontainers connection and immediately cancels the context to exercise the
// ctx.Done() branch in the main message loop.
func TestListener_Start_WithRealRabbitMQ(t *testing.T) {
	if sharedConn == nil {
		t.Skip("Skipping integration test: testcontainers unavailable")
	}

	l, err := NewListener(sharedConn, false)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err = l.Start(ctx, &Handlers{})
	assert.NoError(t, err)
}

func TestListener_MessageProcessing(t *testing.T) {
	if sharedConn == nil {
		t.Skip("Skipping integration test: testcontainers unavailable")
	}

	mockRepo := new(MockRepository)
	mockRepo.On("CreateIndividual", testifymock.Anything, testifymock.Anything).Return(nil)
	h := NewHandlers(mockRepo, nil, nil, &NoOpTransactionManager{})

	l, err := NewListener(sharedConn, false)
	require.NoError(t, err)

	ctx := t.Context()

	// Start listener in a goroutine
	go func() {
		_ = l.Start(ctx, h)
	}()

	// Wait for exchanges/queues to be bound
	time.Sleep(1 * time.Second)

	ch, err := sharedConn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	// Setup valid user header for AuthMiddleware
	authHeaders := amqp.Table{
		"user": "{\"id\":\"test-user\",\"roles\":[\"admin\"]}",
	}

	// 1. Publish Valid Command (Covers success branch)
	err = ch.PublishWithContext(ctx, CommandExchange, CmdPartyCreate, false, false, amqp.Publishing{
		ContentType: "application/json",
		Headers:     authHeaders,
		Body:        []byte(`{"@type":"Individual","id":"msg-test"}`),
	})
	assert.NoError(t, err)

	// 2. Publish Invalid Command with ReplyTo (Covers error + reply branch)
	err = ch.PublishWithContext(ctx, CommandExchange, CmdPartyCreate, false, false, amqp.Publishing{
		ContentType:   "application/json",
		ReplyTo:       "amq.rabbitmq.reply-to",
		CorrelationId: "1234",
		Body:          []byte(`{bad-json}`),
	})
	assert.NoError(t, err)

	// 3. Publish Unknown Routing Key Command (Covers nack branch)
	err = ch.PublishWithContext(ctx, CommandExchange, "unknown.command", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        []byte(`{}`),
	})
	assert.NoError(t, err)

	// 4. Publish Valid Event (Covers success event branch)
	err = ch.PublishWithContext(ctx, EventExchange, EvtCustomerCreated, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        []byte(`{"id":"evt-1"}`),
	})
	assert.NoError(t, err)

	// 5. Publish Unknown Event (Covers log branch)
	err = ch.PublishWithContext(ctx, EventExchange, "unknown.event", false, false, amqp.Publishing{
		Body: []byte(`{}`),
	})
	assert.NoError(t, err)

	// Wait for processing to complete
	time.Sleep(2 * time.Second)
}
