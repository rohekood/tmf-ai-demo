package rabbitmq

import (
	"context"
	"testing"
	"tmf/services/customer-management/internal/domain"

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
			headers:  amqp.Table{"user": "admin-1"},
			expected: "admin-1",
		},
		{
			name:     "Missing user header",
			headers:  amqp.Table{},
			expected: "",
		},
		{
			name:     "Empty user header",
			headers:  amqp.Table{"user": ""},
			expected: "",
		},
		{
			name:     "Wrong type user header",
			headers:  amqp.Table{"user": 123},
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
