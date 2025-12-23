package rabbitmq

import (
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_DLQ(t *testing.T) {
	// This test requires the shared connection from TestMain
	if sharedConn == nil {
		t.Skip("RabbitMQ not available")
	}

	ch, err := sharedConn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	// 1. Setup DLX/DLQ (this is normally done in Listener.Start, but we do it here to ensure it exists)
	err = ch.ExchangeDeclare(DeadLetterExchange, "fanout", true, false, false, false, nil)
	require.NoError(t, err)
	_, err = ch.QueueDeclare(DeadLetterQueue, true, false, false, false, nil)
	require.NoError(t, err)
	err = ch.QueueBind(DeadLetterQueue, "", DeadLetterExchange, false, nil)
	require.NoError(t, err)

	// 2. Setup a test queue with DLX
	testQueue := "test.dlq.source"
	_, err = ch.QueueDeclare(testQueue, false, false, false, false, amqp.Table{
		"x-dead-letter-exchange": DeadLetterExchange,
	})
	require.NoError(t, err)

	// Purge DLQ first
	_, _ = ch.QueuePurge(DeadLetterQueue, false)

	// 3. Publish a message
	err = ch.Publish("", testQueue, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte("fail-me"),
	})
	require.NoError(t, err)

	// 4. Consume and Nack(false, false)
	msgs, err := ch.Consume(testQueue, "", false, false, false, false, nil)
	require.NoError(t, err)

	d := <-msgs
	err = d.Nack(false, false) // Send to DLX
	require.NoError(t, err)

	// 5. Verify message is in DLQ
	// Wait a bit for routing
	time.Sleep(500 * time.Millisecond)

	dlqMsgs, err := ch.Consume(DeadLetterQueue, "", false, false, false, false, nil)
	require.NoError(t, err)

	select {
	case dlqMsg := <-dlqMsgs:
		assert.Equal(t, []byte("fail-me"), dlqMsg.Body)
		dlqMsg.Ack(false)
	case <-time.After(2 * time.Second):
		t.Fatal("Message did not arrive in DLQ")
	}
}
