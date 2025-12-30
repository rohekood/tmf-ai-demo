package rabbitmq

import (
	"context"
	"testing"
	"tmf/services/party-management/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

func TestHandlers_ExtractUser(t *testing.T) {
	h := &Handlers{}
	ctx := context.Background()

	testCases := []struct {
		name     string
		headers  amqp.Table
		expected string
	}{
		{
			name:     "Valid user header",
			headers:  amqp.Table{"user": "party-admin"},
			expected: "party-admin",
		},
		{
			name:     "Missing user header",
			headers:  amqp.Table{},
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := amqp.Delivery{Headers: tc.headers}
			newCtx := h.extractUser(ctx, d)

			userID, _ := newCtx.Value(domain.UserContextKey).(string)
			assert.Equal(t, tc.expected, userID)
		})
	}
}
