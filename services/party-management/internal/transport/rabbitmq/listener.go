package rabbitmq

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Listener struct {
	conn *amqp.Connection
}

func NewListener(conn *amqp.Connection) (*Listener, error) {
	return &Listener{conn: conn}, nil
}

type MessageHandler func(d amqp.Delivery) error

func (l *Listener) Listen(queueName string, handler MessageHandler) error {
	ch, err := l.conn.Channel()
	if err != nil {
		return err
	}
	// Don't close channel here as we need it for consuming

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
		false,  // auto-ack (we will manual ack)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	go func() {
		defer ch.Close()
		for d := range msgs {
			if err := handler(d); err != nil {
				log.Printf("Error handling message: %v", err)
				d.Nack(false, true) // Requeue
			} else {
				d.Ack(false)
			}
		}
	}()

	log.Printf("Listening on queue %s", queueName)
	return nil
}
