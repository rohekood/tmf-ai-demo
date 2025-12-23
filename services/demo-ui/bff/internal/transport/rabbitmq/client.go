package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn       *amqp.Connection
	channel    *amqp.Channel
	replyQueue amqp.Queue
	mu         sync.Mutex
	callbacks  map[string]chan<- []byte
}

func NewClient(url string) (*Client, error) {
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed dial rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed open channel: %w", err)
	}

	// Declare exclusive reply queue
	q, err := ch.QueueDeclare(
		"",    // name (empty = generated)
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed declare reply queue: %w", err)
	}

	client := &Client{
		conn:       conn,
		channel:    ch,
		replyQueue: q,
		callbacks:  make(map[string]chan<- []byte),
	}

	// Start consuming replies
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed consume reply queue: %w", err)
	}

	go client.handleReplies(msgs)

	return client, nil
}

func (c *Client) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *Client) handleReplies(msgs <-chan amqp.Delivery) {
	for d := range msgs {
		c.mu.Lock()
		callback, ok := c.callbacks[d.CorrelationId]
		delete(c.callbacks, d.CorrelationId)
		c.mu.Unlock()

		if ok {
			callback <- d.Body
		} else {
			log.Printf("Received reply for unknown correlation ID: %s", d.CorrelationId)
		}
	}
}

// CallRPC sends a request and waits for a response (RPC pattern)
func (c *Client) CallRPC(ctx context.Context, exchange, routingKey string, payload interface{}) ([]byte, error) {
	corrId := uuid.New().String()
	replyChan := make(chan []byte, 1)

	c.mu.Lock()
	c.callbacks[corrId] = replyChan
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.callbacks, corrId)
		c.mu.Unlock()
	}()

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	err = c.channel.PublishWithContext(ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrId,
			ReplyTo:       c.replyQueue.Name,
			Body:          body,
		})
	if err != nil {
		return nil, fmt.Errorf("failed publish: %w", err)
	}

	select {
	case res := <-replyChan:
		return res, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(10 * time.Second): // Default timeout
		return nil, errors.New("RPC timeout")
	}
}

// PublishCommand sends a message without waiting for reply
func (c *Client) PublishCommand(ctx context.Context, exchange, routingKey string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	return c.channel.PublishWithContext(ctx,
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}
