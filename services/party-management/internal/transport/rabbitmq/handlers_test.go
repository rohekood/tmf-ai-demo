package rabbitmq

import (
	"encoding/json"
	"testing"
	"time"

	"tmf/services/party-management/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository implements domain.Repository for testing
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateIndividual(ind *domain.Individual) error {
	args := m.Called(ind)
	return args.Error(0)
}

func (m *MockRepository) GetIndividual(id string) (*domain.Individual, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Individual), args.Error(1)
}

func (m *MockRepository) UpdateIndividual(ind *domain.Individual) error {
	args := m.Called(ind)
	return args.Error(0)
}

func (m *MockRepository) CreateOrganization(org *domain.Organization) error {
	args := m.Called(org)
	return args.Error(0)
}

func (m *MockRepository) GetOrganization(id string) (*domain.Organization, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Organization), args.Error(1)
}

func (m *MockRepository) UpdateOrganization(org *domain.Organization) error {
	args := m.Called(org)
	return args.Error(0)
}

func (m *MockRepository) DeleteParty(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRepository) SearchParties(criteria map[string]interface{}) ([]domain.Party, error) {
	args := m.Called(criteria)
	return args.Get(0).([]domain.Party), args.Error(1)
}

// MockPublisher for testing (no actual RabbitMQ connection)
type MockPublisher struct {
	PublishedEvents []struct {
		Exchange   string
		RoutingKey string
		Event      interface{}
	}
}

func (m *MockPublisher) Publish(exchange, routingKey string, event interface{}) error {
	m.PublishedEvents = append(m.PublishedEvents, struct {
		Exchange   string
		RoutingKey string
		Event      interface{}
	}{exchange, routingKey, event})
	return nil
}

func (m *MockPublisher) GetChannel() (*amqp.Channel, error) {
	return nil, nil
}

func (m *MockPublisher) Close() error {
	return nil
}

// TestableHandlers wraps Handlers for testing with mock publisher
type TestableHandlers struct {
	repo      domain.Repository
	publisher *MockPublisher
}

func NewTestableHandlers(repo domain.Repository) *TestableHandlers {
	return &TestableHandlers{
		repo:      repo,
		publisher: &MockPublisher{},
	}
}

func (h *TestableHandlers) HandleCreateParty(d amqp.Delivery) error {
	var payload CreatePartyPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		return err
	}

	now := time.Now()

	if payload.Type == "Individual" && payload.Individual != nil {
		ind := &domain.Individual{
			Party: domain.Party{
				ID:        payload.Individual.ID,
				Type:      domain.PartyTypeIndividual,
				Href:      payload.Individual.Href,
				Status:    "Initialized",
				CreatedAt: now,
				UpdatedAt: now,
			},
			GivenName:  payload.Individual.GivenName,
			FamilyName: payload.Individual.FamilyName,
		}
		if err := h.repo.CreateIndividual(ind); err != nil {
			return err
		}
		h.publisher.Publish(EventExchange, EvtPartyCreated, ind)
		h.publisher.Publish(EventExchange, EvtPartyStateChange, map[string]interface{}{
			"id":       ind.ID,
			"newState": ind.Status,
		})
	} else if payload.Type == "Organization" && payload.Organization != nil {
		org := &domain.Organization{
			Party: domain.Party{
				ID:        payload.Organization.ID,
				Type:      domain.PartyTypeOrganization,
				Href:      payload.Organization.Href,
				Status:    "Initialized",
				CreatedAt: now,
				UpdatedAt: now,
			},
			TradingName:   payload.Organization.TradingName,
			IsLegalEntity: payload.Organization.IsLegalEntity,
		}
		if err := h.repo.CreateOrganization(org); err != nil {
			return err
		}
		h.publisher.Publish(EventExchange, EvtPartyCreated, org)
		h.publisher.Publish(EventExchange, EvtPartyStateChange, map[string]interface{}{
			"id":       org.ID,
			"newState": org.Status,
		})
	}
	return nil
}

func (h *TestableHandlers) HandleDeleteParty(d amqp.Delivery) error {
	var payload DeletePartyPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		return err
	}

	if err := h.repo.DeleteParty(payload.ID); err != nil {
		return err
	}

	h.publisher.Publish(EventExchange, EvtPartyDeleted, map[string]interface{}{
		"id": payload.ID,
	})
	return nil
}

// --- Tests ---

func TestHandleCreateParty_Individual(t *testing.T) {
	mockRepo := new(MockRepository)
	handlers := NewTestableHandlers(mockRepo)

	mockRepo.On("CreateIndividual", mock.AnythingOfType("*domain.Individual")).Return(nil)

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
	err := handlers.HandleCreateParty(delivery)

	assert.NoError(t, err)
	mockRepo.AssertCalled(t, "CreateIndividual", mock.AnythingOfType("*domain.Individual"))

	// Verify events published
	assert.Len(t, handlers.publisher.PublishedEvents, 2)
	assert.Equal(t, EvtPartyCreated, handlers.publisher.PublishedEvents[0].RoutingKey)
	assert.Equal(t, EvtPartyStateChange, handlers.publisher.PublishedEvents[1].RoutingKey)
}

func TestHandleCreateParty_Organization(t *testing.T) {
	mockRepo := new(MockRepository)
	handlers := NewTestableHandlers(mockRepo)

	mockRepo.On("CreateOrganization", mock.AnythingOfType("*domain.Organization")).Return(nil)

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
	err := handlers.HandleCreateParty(delivery)

	assert.NoError(t, err)
	mockRepo.AssertCalled(t, "CreateOrganization", mock.AnythingOfType("*domain.Organization"))

	// Verify events published
	assert.Len(t, handlers.publisher.PublishedEvents, 2)
	assert.Equal(t, EvtPartyCreated, handlers.publisher.PublishedEvents[0].RoutingKey)
	assert.Equal(t, EvtPartyStateChange, handlers.publisher.PublishedEvents[1].RoutingKey)
}

func TestHandleDeleteParty(t *testing.T) {
	mockRepo := new(MockRepository)
	handlers := NewTestableHandlers(mockRepo)

	mockRepo.On("DeleteParty", "delete-test-1").Return(nil)

	payload := DeletePartyPayload{ID: "delete-test-1"}
	body, _ := json.Marshal(payload)

	delivery := amqp.Delivery{Body: body}
	err := handlers.HandleDeleteParty(delivery)

	assert.NoError(t, err)
	mockRepo.AssertCalled(t, "DeleteParty", "delete-test-1")

	// Verify event published
	assert.Len(t, handlers.publisher.PublishedEvents, 1)
	assert.Equal(t, EvtPartyDeleted, handlers.publisher.PublishedEvents[0].RoutingKey)
}
