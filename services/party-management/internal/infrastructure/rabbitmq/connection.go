package rabbitmq

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type amqpDialFn func(url string) (*amqp.Connection, error)

type ConnectionManager struct {
	url           string
	conn          *amqp.Connection
	ch            *amqp.Channel
	mu            sync.RWMutex
	reconnectChan chan struct{}
	closed        bool
	ctx           context.Context
	cancel        context.CancelFunc
	dialFn        amqpDialFn
}

func NewConnectionManager(url string, dialFn ...amqpDialFn) *ConnectionManager {
	ctx, cancel := context.WithCancel(context.Background())
	df := amqp.Dial
	if len(dialFn) > 0 {
		df = dialFn[0]
	}
	return &ConnectionManager{
		url:           url,
		reconnectChan: make(chan struct{}, 1),
		ctx:           ctx,
		cancel:        cancel,
		dialFn:        df,
	}
}

func (m *ConnectionManager) Connect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, err := m.dialFn(m.url)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	m.conn = conn
	m.ch = ch

	go m.handleReconnect()

	return nil
}

func (m *ConnectionManager) handleReconnect() {
	notifyClose := m.conn.NotifyClose(make(chan *amqp.Error))

	select {
	case <-m.ctx.Done():
		return
	case err := <-notifyClose:
		if err != nil {
			slog.Error("rabbitmq connection closed, attempting reconnect", "error", err)
			for {
				if m.closed {
					return
				}
				time.Sleep(5 * time.Second)
				if err := m.Connect(); err == nil {
					slog.Info("rabbitmq reconnected successfully")
					select {
					case m.reconnectChan <- struct{}{}:
					default:
					}
					return
				}
				slog.Error("rabbitmq reconnect failed, retrying...", "error", err)
			}
		}
	}
}

func (m *ConnectionManager) GetConnection() *amqp.Connection {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.conn
}

func (m *ConnectionManager) GetChannel() *amqp.Channel {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ch
}

func (m *ConnectionManager) ReconnectNotify() <-chan struct{} {
	return m.reconnectChan
}

func (m *ConnectionManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	m.cancel()
	if m.ch != nil {
		_ = m.ch.Close()
	}
	if m.conn != nil {
		return m.conn.Close()
	}
	return nil
}
