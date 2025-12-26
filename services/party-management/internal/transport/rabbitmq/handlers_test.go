package rabbitmq

import (
	"context"
	"encoding/json"
	"testing"

	"tmf/services/party-management/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository implements domain.Repository for testing
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetParty(ctx context.Context, id string) (*domain.Party, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Party), args.Error(1)
}

func (m *MockRepository) CreateIndividual(ctx context.Context, ind *domain.Individual) error {
	args := m.Called(ctx, ind)
	return args.Error(0)
}

func (m *MockRepository) GetIndividual(ctx context.Context, id string) (*domain.Individual, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Individual), args.Error(1)
}

func (m *MockRepository) UpdateIndividual(ctx context.Context, ind *domain.Individual) error {
	args := m.Called(ctx, ind)
	return args.Error(0)
}

func (m *MockRepository) CreateOrganization(ctx context.Context, org *domain.Organization) error {
	args := m.Called(ctx, org)
	return args.Error(0)
}

func (m *MockRepository) GetOrganization(ctx context.Context, id string) (*domain.Organization, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Organization), args.Error(1)
}

func (m *MockRepository) UpdateOrganization(ctx context.Context, org *domain.Organization) error {
	args := m.Called(ctx, org)
	return args.Error(0)
}

func (m *MockRepository) DeleteParty(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) SearchParties(ctx context.Context, criteria map[string]interface{}) ([]domain.Party, error) {
	args := m.Called(ctx, criteria)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Party), args.Error(1)
}

// --- Tests ---

func TestHandleCreateParty_Individual(t *testing.T) {
	mockRepo := new(MockRepository)
	h := NewHandlers(mockRepo, nil) // Publisher is NIL here, so publishEvent will slog error but not crash if we don't call it.
	// Actually we should probably mock the publisher too if we want to verify events.
	// But Handlers takes *infraRabbit.Publisher which is a concrete type.
	// For these tests, let's just focus on Repo calls if we can't easily mock the publisher without an interface.

	ctx := context.Background()
	mockRepo.On("CreateIndividual", ctx, mock.AnythingOfType("*domain.Individual")).Return(nil)

	payload := CreatePartyPayload{
		Type: "Individual",
		Individual: &CreateIndividualPayload{
			ID:         "test-ind-1",
			GivenName:  "Test",
			FamilyName: "User",
			Href:       "http://example.com/test-ind-1",
		},
	}
	body, _ := json.Marshal(payload)

	delivery := amqp.Delivery{Body: body}
	err := h.HandleCreateParty(ctx, delivery)

	assert.NoError(t, err)
	mockRepo.AssertCalled(t, "CreateIndividual", ctx, mock.AnythingOfType("*domain.Individual"))
}

func TestHandleCreateParty_Organization(t *testing.T) {
	mockRepo := new(MockRepository)
	h := NewHandlers(mockRepo, nil)

	ctx := context.Background()
	mockRepo.On("CreateOrganization", ctx, mock.AnythingOfType("*domain.Organization")).Return(nil)

	payload := CreatePartyPayload{
		Type: "Organization",
		Organization: &CreateOrganizationPayload{
			ID:            "test-org-1",
			TradingName:   "Test Corp",
			IsLegalEntity: true,
			Href:          "http://example.com/test-org-1",
		},
	}
	body, _ := json.Marshal(payload)

	delivery := amqp.Delivery{Body: body}
	err := h.HandleCreateParty(ctx, delivery)

	assert.NoError(t, err)
	mockRepo.AssertCalled(t, "CreateOrganization", ctx, mock.AnythingOfType("*domain.Organization"))
}

func TestHandleDeleteParty(t *testing.T) {
	mockRepo := new(MockRepository)
	h := NewHandlers(mockRepo, nil)

	ctx := context.Background()

	// 1. GetParty is called first
	mockRepo.On("GetParty", ctx, "delete-test-1").Return(&domain.Party{
		ID:     "delete-test-1",
		Type:   domain.PartyTypeIndividual,
		Status: "Active",
	}, nil)

	// 2. GetIndividual is called (since type is Individual)
	mockRepo.On("GetIndividual", ctx, "delete-test-1").Return(&domain.Individual{
		Party: domain.Party{ID: "delete-test-1", Type: domain.PartyTypeIndividual, Status: "Active"},
	}, nil)

	// 3. UpdateIndividual is called with "DeletionPending"
	mockRepo.On("UpdateIndividual", ctx, mock.MatchedBy(func(ind *domain.Individual) bool {
		return ind.Status == "DeletionPending"
	})).Return(nil)

	payload := DeletePartyPayload{ID: "delete-test-1"}
	body, _ := json.Marshal(payload)

	delivery := amqp.Delivery{Body: body}
	err := h.HandleDeleteParty(ctx, delivery)

	assert.NoError(t, err)
	mockRepo.AssertCalled(t, "GetParty", ctx, "delete-test-1")
	mockRepo.AssertCalled(t, "GetIndividual", ctx, "delete-test-1")
	mockRepo.AssertCalled(t, "UpdateIndividual", ctx, mock.Anything)
}

func TestHandleSearchParty_ReturnsCompleteIndividualData(t *testing.T) {
	mockRepo := new(MockRepository)
	h := NewHandlers(mockRepo, nil)

	ctx := context.Background()

	// Mock SearchParties to return base Party objects
	baseParties := []domain.Party{
		{ID: "ind-1", Type: domain.PartyTypeIndividual, Status: "Active"},
		{ID: "ind-2", Type: domain.PartyTypeIndividual, Status: "Active"},
	}
	mockRepo.On("SearchParties", ctx, mock.Anything).Return(baseParties, nil)

	// Mock GetIndividual to return full Individual objects with givenName/familyName
	mockRepo.On("GetIndividual", ctx, "ind-1").Return(&domain.Individual{
		Party:      domain.Party{ID: "ind-1", Type: domain.PartyTypeIndividual, Status: "Active"},
		GivenName:  "John",
		FamilyName: "Doe",
	}, nil)
	mockRepo.On("GetIndividual", ctx, "ind-2").Return(&domain.Individual{
		Party:      domain.Party{ID: "ind-2", Type: domain.PartyTypeIndividual, Status: "Active"},
		GivenName:  "Jane",
		FamilyName: "Smith",
	}, nil)

	payload := SearchPartyPayload{}
	body, _ := json.Marshal(payload)

	delivery := amqp.Delivery{Body: body}
	err := h.HandleSearchParty(ctx, delivery)

	assert.NoError(t, err)
	// Verify that GetIndividual was called for each party to fetch full details
	mockRepo.AssertCalled(t, "GetIndividual", ctx, "ind-1")
	mockRepo.AssertCalled(t, "GetIndividual", ctx, "ind-2")
}

func TestHandleSearchParty_ReturnsCompleteOrganizationData(t *testing.T) {
	mockRepo := new(MockRepository)
	h := NewHandlers(mockRepo, nil)

	ctx := context.Background()

	// Mock SearchParties to return base Party objects
	baseParties := []domain.Party{
		{ID: "org-1", Type: domain.PartyTypeOrganization, Status: "Active"},
	}
	mockRepo.On("SearchParties", ctx, mock.Anything).Return(baseParties, nil)

	// Mock GetOrganization to return full Organization object with tradingName
	mockRepo.On("GetOrganization", ctx, "org-1").Return(&domain.Organization{
		Party:         domain.Party{ID: "org-1", Type: domain.PartyTypeOrganization, Status: "Active"},
		TradingName:   "Acme Corp",
		IsLegalEntity: true,
	}, nil)

	payload := SearchPartyPayload{}
	body, _ := json.Marshal(payload)

	delivery := amqp.Delivery{Body: body}
	err := h.HandleSearchParty(ctx, delivery)

	assert.NoError(t, err)
	// Verify that GetOrganization was called to fetch full details
	mockRepo.AssertCalled(t, "GetOrganization", ctx, "org-1")
}

func TestHandleSearchParty_MixedTypes(t *testing.T) {
	mockRepo := new(MockRepository)
	h := NewHandlers(mockRepo, nil)

	ctx := context.Background()

	// Mock SearchParties to return mixed party types
	baseParties := []domain.Party{
		{ID: "ind-1", Type: domain.PartyTypeIndividual, Status: "Active"},
		{ID: "org-1", Type: domain.PartyTypeOrganization, Status: "Active"},
	}
	mockRepo.On("SearchParties", ctx, mock.Anything).Return(baseParties, nil)

	// Mock GetIndividual and GetOrganization
	mockRepo.On("GetIndividual", ctx, "ind-1").Return(&domain.Individual{
		Party:      domain.Party{ID: "ind-1", Type: domain.PartyTypeIndividual, Status: "Active"},
		GivenName:  "John",
		FamilyName: "Doe",
	}, nil)
	mockRepo.On("GetOrganization", ctx, "org-1").Return(&domain.Organization{
		Party:         domain.Party{ID: "org-1", Type: domain.PartyTypeOrganization, Status: "Active"},
		TradingName:   "Acme Corp",
		IsLegalEntity: true,
	}, nil)

	payload := SearchPartyPayload{}
	body, _ := json.Marshal(payload)

	delivery := amqp.Delivery{Body: body}
	err := h.HandleSearchParty(ctx, delivery)

	assert.NoError(t, err)
	mockRepo.AssertCalled(t, "GetIndividual", ctx, "ind-1")
	mockRepo.AssertCalled(t, "GetOrganization", ctx, "org-1")
}
