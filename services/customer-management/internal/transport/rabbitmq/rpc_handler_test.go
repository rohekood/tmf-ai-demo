package rabbitmq

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"tmf/pkg/rabbitmq"
	"tmf/services/customer-management/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRPCHandler_HandleGetCustomer(t *testing.T) {
	mockRepo := new(MockRepository)
	mockPub := new(MockPublisher)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	h := NewRPCHandler(mockRepo, mockPub, logger)
	ctx := context.Background()

	// Setup Context with ReplyTo/CorrelationID
	ctx = context.WithValue(ctx, rabbitmq.ContextKeyReplyTo, "reply-queue")
	ctx = context.WithValue(ctx, rabbitmq.ContextKeyAMQPCorrelationID, "corr-1")

	// Case 1: Success
	reqPayload := `{"customerId": "cust-1"}`
	cust := &domain.Customer{
		ID:     "cust-1",
		Status: domain.CustomerStatusActive,
		Characteristics: []domain.CustomerCharacteristic{
			{Name: "tier", Value: "Gold"},
		},
		MarketSegments: []domain.MarketSegment{
			{Name: "Consumer"},
		},
	}

	mockRepo.On("GetCustomer", mock.Anything, "cust-1").Return(cust, nil).Once()
	mockPub.On("PublishToQueue", mock.Anything, "reply-queue", "corr-1", mock.MatchedBy(func(resp map[string]interface{}) bool {
		return resp["id"] == "cust-1" &&
			resp["tier"] == "Gold" &&
			resp["segment"] == "Consumer" &&
			resp["status"] == domain.CustomerStatusActive
	})).Return(nil).Once()

	err := h.HandleGetCustomer(ctx, []byte(reqPayload))
	assert.NoError(t, err)

	// Case 2: Customer Not Found
	mockRepo.On("GetCustomer", mock.Anything, "missing").Return(nil, errors.New("not found")).Once()
	mockPub.On("PublishToQueue", mock.Anything, "reply-queue", "corr-1", mock.MatchedBy(func(resp map[string]string) bool {
		return resp["error"] == "not found"
	})).Return(nil).Once()

	reqPayloadMissing := `{"customerId": "missing"}`
	err = h.HandleGetCustomer(ctx, []byte(reqPayloadMissing))
	assert.NoError(t, err) // Should handle error gracefully by replying

	mockRepo.AssertExpectations(t)
	mockPub.AssertExpectations(t)
}

func TestRPCHandler_ExtractHelpers(t *testing.T) {
	// Test extractTier
	c1 := &domain.Customer{Characteristics: []domain.CustomerCharacteristic{{Name: "tier", Value: "Platinum"}}}
	assert.Equal(t, "Platinum", extractTier(c1))

	c2 := &domain.Customer{}
	assert.Equal(t, "Standard", extractTier(c2)) // Default

	// Test extractSegment
	c3 := &domain.Customer{MarketSegments: []domain.MarketSegment{{Name: "B2B"}}}
	assert.Equal(t, "B2B", extractSegment(c3))

	c4 := &domain.Customer{}
	assert.Equal(t, "Residential", extractSegment(c4)) // Default
}
