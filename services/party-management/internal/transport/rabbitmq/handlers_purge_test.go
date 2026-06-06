package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"tmf/services/party-management/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
)

// mockCustomerChecker satisfies the CustomerChecker interface.
type mockCustomerChecker struct {
	hasCustomers bool
	err          error
}

func (m *mockCustomerChecker) HasCustomers(_ context.Context, _ string) (bool, error) {
	return m.hasCustomers, m.err
}

func TestWithCustomerChecker_SetsField(t *testing.T) {
	h := NewHandlers(nil, nil, nil, nil)
	checker := &mockCustomerChecker{}
	result := h.WithCustomerChecker(checker)
	assert.Same(t, h, result)
	assert.Equal(t, checker, h.customerChecker)
}

// --- HandlePurgeParty ---

func TestHandlePurgeParty_Success(t *testing.T) {
	repo := new(MockRepository)
	h := NewHandlers(repo, nil, nil, &NoOpTransactionManager{})

	ctx := context.Background()
	repo.On("GetParty", ctx, "party-1").Return(&domain.Party{
		ID:     "party-1",
		Status: string(domain.PartyStatusDeleted),
	}, nil)
	repo.On("DeleteParty", ctx, "party-1").Return(nil)

	body, _ := json.Marshal(DeletePartyPayload{ID: "party-1"})
	err := h.HandlePurgeParty(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err)
	repo.AssertCalled(t, "DeleteParty", ctx, "party-1")
}

func TestHandlePurgeParty_NotDeleted(t *testing.T) {
	repo := new(MockRepository)
	h := NewHandlers(repo, nil, nil, &NoOpTransactionManager{})

	ctx := context.Background()
	repo.On("GetParty", ctx, "party-2").Return(&domain.Party{
		ID:     "party-2",
		Status: string(domain.PartyStatusActive),
	}, nil)

	body, _ := json.Marshal(DeletePartyPayload{ID: "party-2"})
	err := h.HandlePurgeParty(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err) // error is replied to caller, handler returns nil
	repo.AssertNotCalled(t, "DeleteParty", testifymock.Anything, testifymock.Anything)
}

func TestHandlePurgeParty_PartyNotFound(t *testing.T) {
	repo := new(MockRepository)
	h := NewHandlers(repo, nil, nil, &NoOpTransactionManager{})

	ctx := context.Background()
	repo.On("GetParty", ctx, "party-x").Return(nil, domain.ErrNotFound)

	body, _ := json.Marshal(DeletePartyPayload{ID: "party-x"})
	err := h.HandlePurgeParty(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err)
	repo.AssertNotCalled(t, "DeleteParty", testifymock.Anything, testifymock.Anything)
}

func TestHandlePurgeParty_MissingID(t *testing.T) {
	h := NewHandlers(nil, nil, nil, nil)
	body, _ := json.Marshal(DeletePartyPayload{ID: ""})
	err := h.HandlePurgeParty(context.Background(), amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandlePurgeParty_InvalidJSON(t *testing.T) {
	h := NewHandlers(nil, nil, nil, nil)
	err := h.HandlePurgeParty(context.Background(), amqp.Delivery{Body: []byte("not json")})
	assert.Error(t, err)
}

// --- HandleDeleteParty with customer pre-check ---

func TestHandleDeleteParty_CustomerCheckerRejects(t *testing.T) {
	repo := new(MockRepository)
	h := NewHandlers(repo, nil, nil, &NoOpTransactionManager{})
	h.WithCustomerChecker(&mockCustomerChecker{hasCustomers: true})

	ctx := context.Background()
	body, _ := json.Marshal(DeletePartyPayload{ID: "party-3"})
	err := h.HandleDeleteParty(ctx, amqp.Delivery{Body: body})
	// No error returned to caller (replied inline), GetParty never called
	assert.NoError(t, err)
	repo.AssertNotCalled(t, "GetParty", testifymock.Anything, testifymock.Anything)
}

func TestHandleDeleteParty_CustomerCheckerError_Proceeds(t *testing.T) {
	repo := new(MockRepository)
	h := NewHandlers(repo, nil, nil, &NoOpTransactionManager{})
	h.WithCustomerChecker(&mockCustomerChecker{err: errors.New("RPC down")})

	ctx := context.Background()
	// Checker fails → warn and proceed with saga → GetParty is called
	repo.On("GetParty", ctx, "party-4").Return(&domain.Party{
		ID:     "party-4",
		Type:   domain.PartyTypeIndividual,
		Status: string(domain.PartyStatusActive),
	}, nil)
	repo.On("GetIndividual", ctx, "party-4").Return(&domain.Individual{
		Party: domain.Party{ID: "party-4", Status: string(domain.PartyStatusActive)},
	}, nil)
	repo.On("UpdateIndividual", ctx, testifymock.MatchedBy(func(ind *domain.Individual) bool {
		return ind.Status == string(domain.PartyStatusDeletionPending)
	})).Return(nil)

	body, _ := json.Marshal(DeletePartyPayload{ID: "party-4"})
	err := h.HandleDeleteParty(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err)
	repo.AssertCalled(t, "GetParty", ctx, "party-4")
}

func TestHandleDeleteParty_NoCustomerChecker_Proceeds(t *testing.T) {
	repo := new(MockRepository)
	h := NewHandlers(repo, nil, nil, &NoOpTransactionManager{})
	// No checker set — should proceed normally

	ctx := context.Background()
	repo.On("GetParty", ctx, "party-5").Return(&domain.Party{
		ID:     "party-5",
		Type:   domain.PartyTypeOrganization,
		Status: string(domain.PartyStatusActive),
	}, nil)
	repo.On("GetOrganization", ctx, "party-5").Return(&domain.Organization{
		Party: domain.Party{ID: "party-5", Status: string(domain.PartyStatusActive)},
	}, nil)
	repo.On("UpdateOrganization", ctx, testifymock.MatchedBy(func(org *domain.Organization) bool {
		return org.Status == string(domain.PartyStatusDeletionPending)
	})).Return(nil)

	body, _ := json.Marshal(DeletePartyPayload{ID: "party-5"})
	err := h.HandleDeleteParty(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err)
	repo.AssertCalled(t, "GetParty", ctx, "party-5")
}
