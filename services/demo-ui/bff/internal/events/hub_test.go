package events

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"tmf/pkg/rabbitmq"

	"github.com/gorilla/websocket"
)

type mockConsumer struct {
	topic        string
	handler      rabbitmq.ConsumerHandler
	subscribeErr error
}

func (m *mockConsumer) Subscribe(topic string, handler rabbitmq.ConsumerHandler) error {
	m.topic = topic
	m.handler = handler
	return m.subscribeErr
}

func (m *mockConsumer) Close() error {
	return nil
}

func TestHub(t *testing.T) {
	h := NewHub()

	var upgrader = websocket.Upgrader{}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = c.Close() }()
		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				break
			}
		}
	}))
	defer s.Close()

	url := "ws" + strings.TrimPrefix(s.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial err: %v", err)
	}
	defer func() { _ = conn.Close() }()

	h.Register("test-correlation-id", conn)
	h.mu.RLock()
	if _, ok := h.clients["test-correlation-id"]; !ok {
		t.Error("expected client to be registered")
	}
	h.mu.RUnlock()

	h.Unregister("test-correlation-id")
	h.mu.RLock()
	if _, ok := h.clients["test-correlation-id"]; ok {
		t.Error("expected client to be unregistered")
	}
	h.mu.RUnlock()
	h.Unregister("test-correlation-id")

	h.Register("test-correlation-id", conn)
	ctx := context.WithValue(context.Background(), rabbitmq.ContextKeyCorrelationID, "test-correlation-id")
	err = h.HandleEvent(ctx, []byte("test message"))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	ctxUser := context.WithValue(context.Background(), rabbitmq.ContextKeyUser, "test-user-id")
	h.Register("test-user-id", conn)
	err = h.HandleEvent(ctxUser, []byte("test message user"))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	ctxEmpty := context.Background()
	err = h.HandleEvent(ctxEmpty, []byte("test message empty"))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	ctxUnregistered := context.WithValue(context.Background(), rabbitmq.ContextKeyCorrelationID, "unregistered-id")
	err = h.HandleEvent(ctxUnregistered, []byte("test message unregistered"))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	mockCons := &mockConsumer{}
	h.StartConsumer(mockCons)
	if mockCons.topic != "evt.#" {
		t.Errorf("expected topic evt.#, got %v", mockCons.topic)
	}
}

func TestHub_HandleEvent_HeaderKeyFallbacks(t *testing.T) {
	h := NewHub()

	var upgrader = websocket.Upgrader{}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = c.Close() }()
		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				break
			}
		}
	}))
	defer s.Close()

	url := "ws" + strings.TrimPrefix(s.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial err: %v", err)
	}
	defer func() { _ = conn.Close() }()

	h.Register("header-correlation-id", conn)
	ctx := context.WithValue(context.Background(), rabbitmq.Key(rabbitmq.HeaderCorrelationID), "header-correlation-id")
	if err := h.HandleEvent(ctx, []byte("header correlation payload")); err != nil {
		t.Fatalf("expected no error with header correlation key, got %v", err)
	}

	h.Register("header-user-id", conn)
	ctx = context.WithValue(context.Background(), rabbitmq.Key(rabbitmq.HeaderUser), "header-user-id")
	if err := h.HandleEvent(ctx, []byte("header user payload")); err != nil {
		t.Fatalf("expected no error with header user key, got %v", err)
	}
}

func TestHub_StartConsumer_SubscribeError(t *testing.T) {
	h := NewHub()
	consumer := &mockConsumer{subscribeErr: fmt.Errorf("subscribe failed")}

	called := false
	originalFatalf := fatalf
	fatalf = func(format string, args ...interface{}) {
		called = true
	}
	defer func() {
		fatalf = originalFatalf
	}()

	h.StartConsumer(consumer)

	if !called {
		t.Fatal("expected fatalf to be called on subscribe error")
	}
	if consumer.topic != "evt.#" {
		t.Fatalf("expected topic evt.#, got %q", consumer.topic)
	}
}
