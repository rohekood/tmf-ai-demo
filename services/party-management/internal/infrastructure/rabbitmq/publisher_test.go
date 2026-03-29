package rabbitmq

import (
	"context"
	"testing"
	"tmf/services/party-management/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestPublisher_Publish_MarshalError(t *testing.T) {
	pub := &Publisher{
		conn: nil,
		ch:   nil,
	}

	// Unmarshalable channel context (e.g. passing a chan into json.Marshal)
	unmarshalableEvent := make(chan int)
	err := pub.Publish(context.Background(), "test.exchange", "test.routing.key", unmarshalableEvent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal event")
}

func TestPublisher_GetChannel(t *testing.T) {
	pub := &Publisher{
		conn: nil,
		ch:   nil,
	}

	ch, err := pub.GetChannel()
	assert.NoError(t, err)
	assert.Nil(t, ch)
}

func TestPublisher_Publish_WithHeaders(t *testing.T) {
	// Not practically testable without an interface for amqp.Channel because PublishWithContext relies on concrete *amqp.Channel methods.
	// But we can test that the context extraction logic wouldn't panic.
	ctx := context.WithValue(context.Background(), domain.UserContextKey, "user-123")
	ctx = context.WithValue(ctx, domain.AuthContextKey, "Bearer abc")

	pub := &Publisher{
		conn: nil,
		ch:   nil,
	}

	// This will panic when it tries to p.ch.PublishWithContext because p.ch is nil.
	// defer func() { recover() }()
	// We'll skip invoking because amqp.Channel is a concrete type and not mockable easily without wrappers.
	_ = ctx
	_ = pub
}
