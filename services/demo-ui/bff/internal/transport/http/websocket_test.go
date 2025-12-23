package http

import (
	"testing"
	"time"
)

func TestHub_Run(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// 1. Test Register
	client1 := &Client{
		hub:  hub,
		send: make(chan []byte, 256),
	}
	hub.register <- client1

	// Give it a moment to register
	time.Sleep(10 * time.Millisecond)

	if len(hub.clients) != 1 {
		t.Errorf("Expected 1 client, got %d", len(hub.clients))
	}

	// 2. Test Broadcast
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

	// 3. Test Unregister
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

// Mocking websocket connection is hard without a real server,
// so we skip testing ServeWs directly unless calls for it.
// Hub logic is the core part here.
