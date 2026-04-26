package rabbitmq

import (
	"context"
	"testing"
	"tmf/services/customer-management/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository for unit tests
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateCustomer(ctx context.Context, c *domain.Customer) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}
func (m *MockRepository) GetCustomer(ctx context.Context, id string) (*domain.Customer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Customer), args.Error(1)
}
func (m *MockRepository) UpdateCustomer(ctx context.Context, c *domain.Customer) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}
func (m *MockRepository) PatchCustomer(ctx context.Context, id string, updates map[string]any) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}
func (m *MockRepository) DeleteCustomer(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockRepository) SearchCustomers(ctx context.Context, criteria map[string]any) ([]domain.Customer, error) {
	args := m.Called(ctx, criteria)
	return args.Get(0).([]domain.Customer), args.Error(1)
}
func (m *MockRepository) AddInteraction(ctx context.Context, i *domain.CustomerInteraction) error {
	args := m.Called(ctx, i)
	return args.Error(0)
}

// MockPublisher for unit tests
type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(ctx context.Context, exchange, routingKey string, body any) error {
	args := m.Called(ctx, exchange, routingKey, body)
	return args.Error(0)
}
func (m *MockPublisher) PublishToQueue(ctx context.Context, queue, correlationID string, body any) error {
	args := m.Called(ctx, queue, correlationID, body)
	return args.Error(0)
}
func (m *MockPublisher) Close() error {
	return nil
}
func (m *MockPublisher) GetChannel() *amqp.Channel {
	return nil
}
func (m *MockPublisher) DeclareTopicExchange(name string, durable, autoDelete, internal, noWait bool) error {
	return nil
}

func TestHandlers_HandleLogInteraction(t *testing.T) {
	mockRepo := new(MockRepository)
	mockPub := new(MockPublisher)
	h := NewHandlers(mockRepo, mockPub, nil, nil) // TM and EventPub not needed for this

	ctx := context.Background()
	payload := `{"customerId": "cust-1", "channel": "Web", "type": "Call", "description": "Support call", "agentId": "agent-1"}`
	delivery := amqp.Delivery{
		Body:    []byte(payload),
		ReplyTo: "reply-queue",
	}

	mockRepo.On("AddInteraction", mock.Anything, mock.MatchedBy(func(i *domain.CustomerInteraction) bool {
		return i.CustomerID == "cust-1" && i.Channel == "Web"
	})).Return(nil)

	mockPub.On("Publish", mock.Anything, "", "reply-queue", mock.MatchedBy(func(resp map[string]string) bool {
		return resp["status"] == "logged" && resp["id"] != ""
	})).Return(nil)

	err := h.HandleLogInteraction(ctx, delivery)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockPub.AssertExpectations(t)
}

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
		{
			name:     "Valid Authorization header",
			headers:  amqp.Table{"Authorization": "Bearer token"},
			expected: "", // We only test userID extraction here, but this covers the Auth lines in extractUser
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
