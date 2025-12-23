package http

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for demo purposes
	},
}

// Client represents a connected WebSocket client
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients
	clients map[*Client]bool
	// Inbound messages from the system to be broadcast
	broadcast chan []byte
	// Register requests from the clients
	register chan *Client
	// Unregister requests from clients
	unregister chan *Client
	// Buffered recent messages for new clients
	buffer      [][]byte
	bufferMutex sync.RWMutex
	bufferSize  int
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		buffer:     make([][]byte, 0, 50), // Keep last 50 messages
		bufferSize: 50,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			// Send buffered messages to new client
			h.bufferMutex.RLock()
			for _, msg := range h.buffer {
				select {
				case client.send <- msg:
				default:
				}
			}
			h.bufferMutex.RUnlock()

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}

		case message := <-h.broadcast:
			// Add to buffer
			h.addToBuffer(message)

			// Broadcast to all clients
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// addToBuffer adds a message to the circular buffer
func (h *Hub) addToBuffer(msg []byte) {
	h.bufferMutex.Lock()
	defer h.bufferMutex.Unlock()

	if len(h.buffer) >= h.bufferSize {
		// Remove oldest
		h.buffer = h.buffer[1:]
	}
	h.buffer = append(h.buffer, msg)
}

// Broadcast sends a message to all connected clients
func (h *Hub) Broadcast(msg interface{}) {
	bytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling broadcast message: %v", err)
		return
	}
	h.broadcast <- bytes
}

// ServeWs handles websocket requests from peers
func (h *Hub) ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{hub: h, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the websocket connection to the hub.
// For the debug view, clients mostly just listen, but we keep this for pong handling.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(60 * time.Second)); return nil })

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
