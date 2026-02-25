package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestClient_PumpsManual(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hub.ServeWs(w, r)
	}))
	defer s.Close()

	url := "ws" + strings.TrimPrefix(s.URL, "http")
	dummyConn, _, _ := websocket.DefaultDialer.Dial(url, nil)

	client := &Client{
		hub:  hub,
		conn: dummyConn,
		send: make(chan []byte, 256),
	}

	client.send <- []byte("msg1")
	client.send <- []byte("msg2")
	go client.writePump()

	time.Sleep(10 * time.Millisecond)

	close(client.send)
	time.Sleep(10 * time.Millisecond)

	dummyConn.Close()
	go client.readPump()
	time.Sleep(50 * time.Millisecond)
}
