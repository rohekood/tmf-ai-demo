package rabbitmq

import (
	"context"
	"errors"
	"testing"

	"tmf/services/customer-management/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
)

// mockPartyChecker satisfies the PartyChecker interface.
type mockPartyChecker struct {
	err error
}

func (m *mockPartyChecker) CheckParty(_ context.Context, _ string) error {
	return m.err
}

// MockEventPublisher satisfies domain.EventPublisher.
type MockEventPublisher struct {
	testifymock.Mock
}

func (m *MockEventPublisher) Publish(ctx context.Context, routingKey string, payload any) error {
	args := m.Called(ctx, routingKey, payload)
	return args.Error(0)
}

// MockTransactionManager satisfies domain.TransactionManager.
type MockTransactionManager struct{}

func (m *MockTransactionManager) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func TestWithPartyChecker_SetsField(t *testing.T) {
	h := NewHandlers(nil, nil, nil, nil)
	checker := &mockPartyChecker{}
	result := h.WithPartyChecker(checker)
	assert.Same(t, h, result)
	assert.Equal(t, checker, h.partyChecker)
}

func TestHandleOnboardCustomer_PartyCheckerRejectsDeleted(t *testing.T) {
	mockRepo := new(MockRepository)
	h := NewHandlers(mockRepo, nil, nil, nil)
	h.WithPartyChecker(&mockPartyChecker{err: errors.New("party p1 has been deleted")})

	ctx := context.Background()
	payload := `{"id":"c1","name":"Test","partyId":"p1","partyType":"Individual"}`
	delivery := amqp.Delivery{Body: []byte(payload)}

	err := h.HandleOnboardCustomer(ctx, delivery)
	// Error is replied inline; handler returns nil
	assert.NoError(t, err)
	// Repo should never be called
	mockRepo.AssertNotCalled(t, "CreateCustomer", testifymock.Anything, testifymock.Anything)
}

func TestHandleOnboardCustomer_PartyCheckerPasses(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEventPub := new(MockEventPublisher)
	tm := &MockTransactionManager{}
	h := NewHandlers(mockRepo, nil, tm, mockEventPub)
	h.WithPartyChecker(&mockPartyChecker{err: nil}) // party is valid

	ctx := context.Background()

	mockRepo.On("CreateCustomer", testifymock.Anything, testifymock.AnythingOfType("*domain.Customer")).Return(nil)
	mockEventPub.On("Publish", testifymock.Anything, EvtCustomerCreated, testifymock.Anything).Return(nil)

	payload := `{"id":"c2","name":"Valid Customer","partyId":"p2","partyType":"Individual"}`
	delivery := amqp.Delivery{Body: []byte(payload)}

	err := h.HandleOnboardCustomer(ctx, delivery)
	assert.NoError(t, err)
	mockRepo.AssertCalled(t, "CreateCustomer", testifymock.Anything, testifymock.AnythingOfType("*domain.Customer"))
}

func TestHandleOnboardCustomer_NoPartyChecker_Proceeds(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEventPub := new(MockEventPublisher)
	tm := &MockTransactionManager{}
	h := NewHandlers(mockRepo, nil, tm, mockEventPub)

	ctx := context.Background()
	mockRepo.On("CreateCustomer", testifymock.Anything, testifymock.AnythingOfType("*domain.Customer")).Return(nil)
	mockEventPub.On("Publish", testifymock.Anything, EvtCustomerCreated, testifymock.Anything).Return(nil)

	payload := `{"id":"c3","name":"Customer","partyId":"p3","partyType":"Individual"}`
	delivery := amqp.Delivery{Body: []byte(payload)}

	err := h.HandleOnboardCustomer(ctx, delivery)
	assert.NoError(t, err)
	mockRepo.AssertCalled(t, "CreateCustomer", testifymock.Anything, testifymock.AnythingOfType("*domain.Customer"))
}

func TestHandleOnboardCustomer_EmptyPartyID_SkipsCheck(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEventPub := new(MockEventPublisher)
	tm := &MockTransactionManager{}
	h := NewHandlers(mockRepo, nil, tm, mockEventPub)
	// Checker would fail if called — but empty partyId means it must be skipped
	h.WithPartyChecker(&mockPartyChecker{err: errors.New("must not be called")})

	ctx := context.Background()
	mockRepo.On("CreateCustomer", testifymock.Anything, testifymock.AnythingOfType("*domain.Customer")).Return(nil)
	mockEventPub.On("Publish", testifymock.Anything, EvtCustomerCreated, testifymock.Anything).Return(nil)

	payload := `{"id":"c4","name":"No Party"}`
	delivery := amqp.Delivery{Body: []byte(payload)}

	err := h.HandleOnboardCustomer(ctx, delivery)
	assert.NoError(t, err)
}

// compile-time check: MockTransactionManager satisfies domain.TransactionManager
var _ domain.TransactionManager = (*MockTransactionManager)(nil)
