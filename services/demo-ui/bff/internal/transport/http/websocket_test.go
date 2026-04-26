package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestHub_Run(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client1 := &Client{
		hub:  hub,
		send: make(chan []byte, 256),
	}
	hub.register <- client1

	time.Sleep(10 * time.Millisecond)

	if len(hub.clients) != 1 {
		t.Errorf("Expected 1 client, got %d", len(hub.clients))
	}

	msg := []byte("test message")
	hub.broadcast <- msg

	select {
	case received := <-client1.send:
		if string(received) != string(msg) {
			t.Errorf("Expected message %s, got %s", msg, received)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for broadcast message")
	}

	hub.unregister <- client1
	time.Sleep(10 * time.Millisecond)

	if len(hub.clients) != 0 {
		t.Errorf("Expected 0 clients, got %d", len(hub.clients))
	}
}

func TestHub_Broadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{
		hub:  hub,
		send: make(chan []byte, 256),
	}
	hub.register <- client
	time.Sleep(10 * time.Millisecond)

	payload := map[string]string{"foo": "bar"}
	hub.Broadcast(payload)

	select {
	case received := <-client.send:
		expected := `{"foo":"bar"}`
		if string(received) != expected {
			t.Errorf("Expected %s, got %s", expected, received)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for broadcast message")
	}
}

type mockValidator struct {
	valid bool
}

func (m *mockValidator) ValidateToken(ctx context.Context, tokenString string) (any, error) {
	if m.valid {
		return "valid", nil
	}
	return nil, nil
}

func TestHub_ServeWs(t *testing.T) {
	hub := NewHub()
	hub.SetTokenValidator(&mockValidator{valid: true})
	go hub.Run()

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hub.ServeWs(w, r)
	}))
	defer s.Close()

	url := "ws" + strings.TrimPrefix(s.URL, "http")

	// Missing token -> 401
	_, resp, err := websocket.DefaultDialer.Dial(url, nil)
	if err == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %v", resp)
	}

	header := http.Header{}
	header.Add("Sec-WebSocket-Protocol", "access_token.valid-token")
	conn, _, err := websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer func() { _ = conn.Close() }()

	time.Sleep(50 * time.Millisecond)

	hub.Broadcast("test-msg")

	if err := conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read broadcast msg: %v", err)
	}
	if string(msg) != `"test-msg"` {
		t.Errorf("Expected \"test-msg\", got %s", string(msg))
	}

	// Write ping/pong/message to hit readPump
	err = conn.WriteMessage(websocket.TextMessage, []byte("ignored text message"))
	if err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
}

func TestHub_Run_WithBuffer(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Put messages in buffer
	hub.addToBuffer([]byte("bufmsg"))

	client := &Client{
		hub:  hub,
		send: make(chan []byte, 1), // Only 1 capacity to hit default case when blocked
	}
	hub.register <- client

	// Give time to register and process buffer
	time.Sleep(10 * time.Millisecond)

	select {
	case msg := <-client.send:
		if string(msg) != "bufmsg" {
			t.Errorf("Expected bufmsg, got %s", msg)
		}
	default:
		t.Errorf("Expected message in client send")
	}

	// Now register a blocked client
	hub.addToBuffer([]byte("bufmsg2"))
	hub.addToBuffer([]byte("bufmsg3"))

	clientBlocked := &Client{
		hub:  hub,
		send: make(chan []byte, 1),
	}
	// Fill the single slot so the second buffer message drops (default case)
	clientBlocked.send <- []byte("initial")
	hub.register <- clientBlocked

	time.Sleep(10 * time.Millisecond)
}

func TestHub_Broadcast_Error(t *testing.T) {
	hub := NewHub()
	// Pass unmarshalable type to hit error
	hub.Broadcast(make(chan int))
}

func TestHub_AddToBuffer_Overflow(t *testing.T) {
	hub := NewHub()
	for range 60 { // buffer size is 50 usually in code
		hub.addToBuffer([]byte("msg"))
	}
	if len(hub.buffer) > 50 { // assuming 50 is max
		t.Errorf("Buffer should not exceed size limit")
	}
}
