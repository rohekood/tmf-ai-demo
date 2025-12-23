package rabbitmq

import (
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_DLQ(t *testing.T) {
	if sharedConn == nil {
		t.Skip("RabbitMQ not available")
	}

	ch, err := sharedConn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	// Setup DLX/DLQ
	err = ch.ExchangeDeclare(DeadLetterExchange, "fanout", true, false, false, false, nil)
	require.NoError(t, err)
	_, err = ch.QueueDeclare(DeadLetterQueue, true, false, false, false, nil)
	require.NoError(t, err)
	err = ch.QueueBind(DeadLetterQueue, "", DeadLetterExchange, false, nil)
	require.NoError(t, err)

	// Setup source queue
	sourceQ := "party.test.dlq.source"
	_, err = ch.QueueDeclare(sourceQ, false, false, false, false, amqp.Table{
		"x-dead-letter-exchange": DeadLetterExchange,
	})
	require.NoError(t, err)

	_, _ = ch.QueuePurge(DeadLetterQueue, false)

	// Publish
	err = ch.Publish("", sourceQ, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte("party-fail"),
	})
	require.NoError(t, err)

	// Consume and Nack
	msgs, err := ch.Consume(sourceQ, "", false, false, false, false, nil)
	require.NoError(t, err)

	d := <-msgs
	err = d.Nack(false, false)
	require.NoError(t, err)

	// Verify
	time.Sleep(500 * time.Millisecond)
	dlqMsgs, err := ch.Consume(DeadLetterQueue, "", false, false, false, false, nil)
	require.NoError(t, err)

	select {
	case dlqMsg := <-dlqMsgs:
		assert.Equal(t, []byte("party-fail"), dlqMsg.Body)
		dlqMsg.Ack(false)
	case <-time.After(2 * time.Second):
		t.Fatal("Message did not arrive in Party DLQ")
	}
}
