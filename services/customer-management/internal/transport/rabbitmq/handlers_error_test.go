package rabbitmq

import (
	"context"
	"errors"
	"testing"
	"tmf/services/customer-management/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockEventPublisher struct {
	mock.Mock
}

func (m *mockEventPublisher) Publish(ctx context.Context, eventType string, payload interface{}) error {
	args := m.Called(ctx, eventType, payload)
	return args.Error(0)
}

func TestHandlers_ErrorPaths(t *testing.T) {
	mockRepo := new(MockRepository)
	mockPub := new(MockPublisher)
	mockTM := &mockTransactionManager{}
	mockEvtPub := new(mockEventPublisher)

	h := NewHandlers(mockRepo, mockPub, mockTM, mockEvtPub)
	ctx := context.Background()

	// 1. Invalid JSON
	invalidJSON := amqp.Delivery{Body: []byte(`{invalid}`)}

	assert.Error(t, h.HandleOnboardCustomer(ctx, invalidJSON))
	assert.Error(t, h.HandleUpdateCustomer(ctx, invalidJSON))
	assert.Error(t, h.HandleGetCustomer(ctx, invalidJSON))
	assert.Error(t, h.HandleSearchCustomer(ctx, invalidJSON))
	assert.Error(t, h.HandleDeleteCustomer(ctx, invalidJSON))
	assert.Error(t, h.HandleLogInteraction(ctx, invalidJSON))
	assert.Error(t, h.HandlePartyEvent(ctx, invalidJSON))

	// 2. Repository Error in Search
	mockRepo.On("SearchCustomers", mock.Anything, mock.Anything).Return([]domain.Customer(nil), errors.New("db error")).Once()
	assert.Error(t, h.HandleSearchCustomer(ctx, amqp.Delivery{Body: []byte(`{}`)}))

	// 3. handlePartyUpdated Repo Error (Search fails)
	mockRepo.On("SearchCustomers", mock.Anything, mock.Anything).Return([]domain.Customer(nil), errors.New("db error")).Once()
	assert.Error(t, h.handlePartyUpdated(ctx, PartyEventPayload{}))

	// 3b. handlePartyUpdated Repo Error (Patch fails)
	mockRepo.On("SearchCustomers", mock.Anything, mock.Anything).Return([]domain.Customer{{ID: "cust-1", Name: "Old"}}, nil).Once()
	mockRepo.On("PatchCustomer", mock.Anything, "cust-1", mock.Anything).Return(errors.New("patch error")).Once()
	// No error returned, just logged, but we hit the branch!
	h.handlePartyUpdated(ctx, PartyEventPayload{ID: "party-1", Type: "Individual", GivenName: "New"})

	// 4. handlePartyDeleted Repo Error
	mockRepo.On("SearchCustomers", mock.Anything, mock.Anything).Return([]domain.Customer(nil), errors.New("db error")).Once()
	assert.Error(t, h.handlePartyDeleted(ctx, PartyEventPayload{}))

	// 4b. handlePartyDeleted Repo Error (Patch fails)
	mockRepo.On("SearchCustomers", mock.Anything, mock.Anything).Return([]domain.Customer{{ID: "cust-1"}}, nil).Once()
	mockRepo.On("PatchCustomer", mock.Anything, "cust-1", mock.Anything).Return(errors.New("patch error")).Once()
	h.handlePartyDeleted(ctx, PartyEventPayload{ID: "party-1"})

	// 5. handlePartyDeletionInitiated Repo Error
	mockRepo.On("SearchCustomers", mock.Anything, mock.Anything).Return([]domain.Customer(nil), errors.New("db error")).Once()
	assert.Error(t, h.handlePartyDeletionInitiated(ctx, PartyEventPayload{ID: "party-1"}))

	// 6. DeleteCustomer Repo Error
	mockRepo.On("DeleteCustomer", mock.Anything, mock.Anything).Return(errors.New("db error")).Once()
	assert.Error(t, h.HandleDeleteCustomer(ctx, amqp.Delivery{Body: []byte(`{"id": "c-1"}`)}))

	// 7. UpdateCustomer Repo Error
	mockRepo.On("GetCustomer", mock.Anything, mock.Anything).Return(nil, errors.New("db error")).Once()
	assert.Error(t, h.HandleUpdateCustomer(ctx, amqp.Delivery{Body: []byte(`{"id": "c-1"}`)}))

	// 8. LogInteraction Repo Error
	mockRepo.On("AddInteraction", mock.Anything, mock.Anything).Return(errors.New("db error")).Once()
	assert.Error(t, h.HandleLogInteraction(ctx, amqp.Delivery{Body: []byte(`{"customerId": "c-1"}`)}))

	// 9. OnboardCustomer Repo Error
	mockRepo.On("CreateCustomer", mock.Anything, mock.Anything).Return(errors.New("db error")).Once()
	assert.Error(t, h.HandleOnboardCustomer(ctx, amqp.Delivery{Body: []byte(`{"id": "c-1", "name": "foo"}`)}))
}

type mockTransactionManager struct{}

func (m *mockTransactionManager) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
