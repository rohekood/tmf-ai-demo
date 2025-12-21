package rabbitmq

import (
	"context"
	"log/slog"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Listener struct {
	conn     *amqp.Connection
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	mu       sync.Mutex
	channels []*amqp.Channel
}

func NewListener(conn *amqp.Connection) (*Listener, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &Listener{
		conn:   conn,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

type MessageHandler func(ctx context.Context, d amqp.Delivery) error

func (l *Listener) Listen(queueName string, handler MessageHandler) error {
	ch, err := l.conn.Channel()
	if err != nil {
		return err
	}

	l.mu.Lock()
	l.channels = append(l.channels, ch)
	l.mu.Unlock()

	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		defer ch.Close()

		for {
			select {
			case <-l.ctx.Done():
				slog.Info("stopping listener", "queue", queueName)
				return
			case d, ok := <-msgs:
				if !ok {
					return
				}

				// Handle message with context
				if err := handler(l.ctx, d); err != nil {
					slog.Error("error handling message", "queue", queueName, "error", err)
					d.Nack(false, true) // Requeue
				} else {
					d.Ack(false)
				}
			}
		}
	}()

	slog.Info("listening on queue", "queue", queueName)
	return nil
}

func (l *Listener) Close() {
	l.cancel()
	l.wg.Wait()

	l.mu.Lock()
	for _, ch := range l.channels {
		ch.Close()
	}
	l.mu.Unlock()
}
