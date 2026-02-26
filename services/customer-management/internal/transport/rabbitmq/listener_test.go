package rabbitmq

import (
	"context"
	"os"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListener_GetHandler(t *testing.T) {
	l := &Listener{}
	h := &Handlers{}

	tests := []struct {
		name       string
		routingKey string
		wantValid  bool
	}{
		{"Onboard Customer", "cmd.customer.onboard", true},
		{"Update Customer", "cmd.customer.update", true},
		{"Get Customer", "query.customer.get", false},
		{"Search Customer", "query.customer.search", true},
		{"Delete Customer", "cmd.customer.delete", true},
		{"Unknown Key", "unknown.key", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, valid := l.GetHandler(tt.routingKey, h)
			assert.Equal(t, tt.wantValid, valid)
			if tt.wantValid {
				assert.NotNil(t, handler)
			} else {
				assert.Nil(t, handler)
			}
		})
	}
}

func TestListener_Start(t *testing.T) {
	if sharedConn == nil {
		t.Skip("Skipping integration test: sharedConn not initialized")
	}

	l, err := NewListener(sharedConn)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	// Start in background
	errChan := make(chan error)

	// We need a proper handler to mock things
	mockRepo := new(MockRepository)
	mockRepo.On("DeleteCustomer", mock.Anything, mock.Anything).Return(assert.AnError)
	mockTM := &mockTransactionManager{}
	h := NewHandlers(mockRepo, nil, mockTM, nil)
	go func() {
		errChan <- l.Start(ctx, h)
	}()

	// Give it a moment to start
	time.Sleep(500 * time.Millisecond)

	// Publish an event message
	ch, err := sharedConn.Channel()
	assert.NoError(t, err)

	err = ch.PublishWithContext(context.Background(),
		EventExchange,
		"evt.party.updated",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(`{}`),
		})
	assert.NoError(t, err)

	// Publish an unknown command
	err = ch.PublishWithContext(context.Background(),
		CommandExchange,
		"unknown.key",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(`{}`),
		})
	assert.NoError(t, err)

	// Publish a known command with replyTo to trigger an error and the error path
	err = ch.PublishWithContext(context.Background(),
		CommandExchange,
		"cmd.customer.delete", // will panic or return error because h.HandleDeleteCustomer will fail due to nil repo
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(`{"id":"some-id"}`),
			ReplyTo:     "some.reply.queue",
		})
	assert.NoError(t, err)

	// Wait a moment for messages to be consumed
	time.Sleep(1 * time.Second)

	// Stop
	cancel()
	err = <-errChan
	assert.NoError(t, err)
	ch.Close()
}

func TestListener_Start_ChannelError(t *testing.T) {
	if sharedConn == nil {
		t.Skip("Skipping integration test: sharedConn not initialized")
	}

	// Create a new connection and immediately close it so Channel() fails
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}
	conn, err := amqp.Dial(url)
	require.NoError(t, err)
	conn.Close()

	l, _ := NewListener(conn)
	ctx := context.Background()

	err = l.Start(ctx, &Handlers{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open channel")
}
