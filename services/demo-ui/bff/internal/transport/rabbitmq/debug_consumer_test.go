package rabbitmq

import (
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// MockBroadcaster for testing
type MockBroadcaster struct {
	LastMsg interface{}
	MsgChan chan interface{}
}

func (m *MockBroadcaster) Broadcast(msg interface{}) {
	m.LastMsg = msg
	if m.MsgChan != nil {
		m.MsgChan <- msg
	}
}

func TestDebugConsumer_HandleMessages(t *testing.T) {
	mockBroadcaster := &MockBroadcaster{
		MsgChan: make(chan interface{}, 1),
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
		MsgChan: make(chan interface{}, 1),
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
	consumer.StartSubscribing("test-exchange")
}
