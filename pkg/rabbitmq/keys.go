package rabbitmq

type contextKey string

const (
	ContextKeyCorrelationID     contextKey = "correlation_id"
	ContextKeyUserID            contextKey = "user_id"
	ContextKeyReplyTo           contextKey = "reply_to"
	ContextKeyRoutingKey        contextKey = "routing_key"
	ContextKeyAMQPCorrelationID contextKey = "amqp_correlation_id"
	ContextKeyUser              contextKey = "user"
)

// Helper to cast string keys if needed
func Key(k string) contextKey {
	return contextKey(k)
}

func (k contextKey) String() string {
	return string(k)
}
