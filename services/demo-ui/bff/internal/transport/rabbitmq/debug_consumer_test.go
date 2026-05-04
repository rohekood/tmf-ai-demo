package rabbitmq

import (
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// MockBroadcaster for testing
type MockBroadcaster struct {
	LastMsg any
	MsgChan chan any
}

func (m *MockBroadcaster) Broadcast(msg any) {
	m.LastMsg = msg
	if m.MsgChan != nil {
		m.MsgChan <- msg
	}
}

func TestDebugConsumer_HandleMessages(t *testing.T) {
	mockBroadcaster := &MockBroadcaster{
		MsgChan: make(chan any, 1),
	}

	// We don't need a real Client for this test as we only test handleMessages logic
	consumer := &DebugConsumer{
		client:      nil,
		broadcaster: mockBroadcaster,
	}

	msgs := make(chan amqp.Delivery, 1)

	// Run handleMessages in a goroutine
	go consumer.handleMessages(msgs)

	// Send a test message
	timestamp := time.Now()
	body := []byte(`{"foo":"bar"}`)
	msgs <- amqp.Delivery{
		MessageId:     "msg-1",
		CorrelationId: "corr-1",
		RoutingKey:    "cmd.customer.create",
		Exchange:      "tmf.customer",
		Body:          body,
		ReplyTo:       "reply-queue",
		Timestamp:     timestamp,
	}

	select {
	case receivedFunc := <-mockBroadcaster.MsgChan:
		// receivedFunc is interface{}, assert to DebugMessage
		debugMsg, ok := receivedFunc.(DebugMessage)
		if !ok {
			t.Fatalf("Expected DebugMessage, got %T", receivedFunc)
		}

		if debugMsg.ID == "" {
			t.Error("Expected ID to be set")
		}
		if debugMsg.Type != "command" {
			t.Errorf("Expected Type 'command', got %s", debugMsg.Type)
		}
		if debugMsg.Service != "customer" {
			t.Errorf("Expected Service 'customer', got %s", debugMsg.Service)
		}
		if debugMsg.Topic != "cmd.customer.create" {
			t.Errorf("Expected Topic 'cmd.customer.create', got %s", debugMsg.Topic)
		}

		payload := debugMsg.Payload
		if payload["foo"] != "bar" {
			t.Errorf("Expected payload foo=bar, got %v", payload["foo"])
		}

	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for broadcast")
	}

	close(msgs)
}

func TestDebugConsumer_HandleMessages_Raw(t *testing.T) {
	mockBroadcaster := &MockBroadcaster{
		MsgChan: make(chan any, 1),
	}
	consumer := &DebugConsumer{
		broadcaster: mockBroadcaster,
	}

	msgs := make(chan amqp.Delivery, 1)
	go consumer.handleMessages(msgs)

	// Send invalid JSON
	msgs <- amqp.Delivery{
		RoutingKey: "evt.something",
		Body:       []byte(`invalid-json`),
	}

	select {
	case receivedFunc := <-mockBroadcaster.MsgChan:
		debugMsg := receivedFunc.(DebugMessage)

		// Should have "raw" payload
		payload := debugMsg.Payload
		if payload["raw"] != "invalid-json" {
			t.Errorf("Expected raw payload, got %v", payload["raw"])
		}

	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for broadcast")
	}
	close(msgs)
}

func TestDebugConsumer_HandleMessages_OrderingExchanges(t *testing.T) {
	tests := []struct {
		name            string
		exchange        string
		routingKey      string
		expectedService string
		expectedType    string
	}{
		{
			name:            "qualification command on ex.domain.market",
			exchange:        "ex.domain.market",
			routingKey:      "cmd.qual.eligibility.check",
			expectedService: "ordering",
			expectedType:    "command",
		},
		{
			name:            "qualification event on ex.domain.market",
			exchange:        "ex.domain.market",
			routingKey:      "evt.qual.checked",
			expectedService: "ordering",
			expectedType:    "event",
		},
		{
			name:            "cart command on ex.domain.commerce",
			exchange:        "ex.domain.commerce",
			routingKey:      "cmd.cart.item.add",
			expectedService: "ordering",
			expectedType:    "command",
		},
		{
			name:            "cart event on ex.domain.commerce",
			exchange:        "ex.domain.commerce",
			routingKey:      "evt.cart.session.updated",
			expectedService: "ordering",
			expectedType:    "event",
		},
		{
			name:            "order command on ex.domain.order",
			exchange:        "ex.domain.order",
			routingKey:      "cmd.order.checkout.submit",
			expectedService: "ordering",
			expectedType:    "command",
		},
		{
			name:            "saga query on ex.domain.order",
			exchange:        "ex.domain.order",
			routingKey:      "query.pocv.saga.get",
			expectedService: "ordering",
			expectedType:    "query",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockBroadcaster := &MockBroadcaster{
				MsgChan: make(chan any, 1),
			}
			consumer := &DebugConsumer{
				broadcaster: mockBroadcaster,
			}

			msgs := make(chan amqp.Delivery, 1)
			go consumer.handleMessages(msgs)

			msgs <- amqp.Delivery{
				MessageId:  "msg-order",
				Exchange:   tc.exchange,
				RoutingKey: tc.routingKey,
				Body:       []byte(`{"sagaId":"abc123"}`),
			}

			select {
			case received := <-mockBroadcaster.MsgChan:
				debugMsg, ok := received.(DebugMessage)
				if !ok {
					t.Fatalf("Expected DebugMessage, got %T", received)
				}
				if debugMsg.Service != tc.expectedService {
					t.Errorf("Expected Service %q, got %q", tc.expectedService, debugMsg.Service)
				}
				if debugMsg.Type != tc.expectedType {
					t.Errorf("Expected Type %q, got %q", tc.expectedType, debugMsg.Type)
				}
				if debugMsg.Topic != tc.routingKey {
					t.Errorf("Expected Topic %q, got %q", tc.routingKey, debugMsg.Topic)
				}
			case <-time.After(100 * time.Millisecond):
				t.Error("Timeout waiting for broadcast")
			}

			close(msgs)
		})
	}
}

func TestDebugConsumer_NewDebugConsumer(t *testing.T) {
	mockBroadcaster := &MockBroadcaster{}
	consumer := NewDebugConsumer(nil, mockBroadcaster)
	if consumer == nil {
		t.Errorf("Expected NewDebugConsumer to return a valid instance")
	}
}

func TestDebugConsumer_StartSubscribing_ErrorNoClient(t *testing.T) {
	mockBroadcaster := &MockBroadcaster{}

	// Create client with no connection will fail
	// But since `StartSubscribing` calls `dc.client.Connection().Channel()`, it will panic if client is nil.
	// Let's create a dummy client or just recover.
	consumer := &DebugConsumer{
		client:      &Client{},
		broadcaster: mockBroadcaster,
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic due to nil connection in dummy client")
		}
	}()
	if err := consumer.StartSubscribing("test-exchange"); err == nil {
		t.Fatal("expected panic before nil client could return an error")
	}
}
