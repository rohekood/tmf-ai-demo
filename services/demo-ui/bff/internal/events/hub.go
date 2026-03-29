package events

import (
	"context"
	"log"
	"sync"
	"tmf/pkg/rabbitmq"

	"github.com/gorilla/websocket"
)

type Hub struct {
	// Map[CorrelationID] -> Connection
	// For simple checkout, we map order Correlation ID to the WS Conn
	clients map[string]*websocket.Conn
	mu      sync.RWMutex
}

var fatalf = log.Fatalf

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*websocket.Conn),
	}
}

func (h *Hub) Register(correlationID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[correlationID] = conn
	log.Printf("WS Client registered: %s", correlationID)
}

func (h *Hub) Unregister(correlationID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[correlationID]; ok {
		delete(h.clients, correlationID)
		log.Printf("WS Client unregistered: %s", correlationID)
	}
}

func (h *Hub) StartConsumer(rmq rabbitmq.Consumer) {
	// Subscribe to ALL events needed for UI
	// In production, might need specific queues. wildcards on topic exchange.
	err := rmq.Subscribe("evt.#", h.HandleEvent)
	if err != nil {
		fatalf("Failed to start BFF consumer: %v", err)
	}
}

func (h *Hub) HandleEvent(ctx context.Context, payload []byte) error {
	// 1. Extract CorrelationID from Context
	var correlationID string
	if val, ok := ctx.Value(rabbitmq.ContextKeyCorrelationID).(string); ok {
		correlationID = val
	} else if val, ok := ctx.Value(rabbitmq.Key(rabbitmq.HeaderCorrelationID)).(string); ok {
		correlationID = val
	} else {
		// Fallback: Try to parse payload if ID is inside body (not ideal but fallback)
		// Or if headers missing.
		// For now, if no correlation ID, we can't route to specific user. Broadcast? No.
		// Check "user" header?
		if user, ok := ctx.Value(rabbitmq.ContextKeyUser).(string); ok {
			correlationID = user // Strategy: Route by UserID if CorrelationID missing
		} else if user, ok := ctx.Value(rabbitmq.Key(rabbitmq.HeaderUser)).(string); ok {
			correlationID = user // Strategy: Route by UserID if CorrelationID missing
		} else {
			return nil // Drop unroutable message
		}
	}

	// 2. Find WebSocket Connection
	h.mu.RLock()
	conn, ok := h.clients[correlationID]
	h.mu.RUnlock()

	if !ok {
		// User disconnected or not on this instance
		return nil
	}

	// 3. Forward Message
	// We might want to wrap it: { "type": "evt.order.created", "payload": ... }
	// But raw forwarding is fine for now
	return conn.WriteMessage(websocket.TextMessage, payload)
}
