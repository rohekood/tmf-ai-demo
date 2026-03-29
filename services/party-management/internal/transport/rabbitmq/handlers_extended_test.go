package rabbitmq

import (
	"context"
	"encoding/json"
	"testing"

	"tmf/services/party-management/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
)

// mockPublisher is needed to mock publishEvent calls inside handlers
type mockTestPublisher struct {
	testifymock.Mock
}

func (m *mockTestPublisher) Publish(ctx context.Context, exchange, routingKey string, body interface{}) error {
	args := m.Called(ctx, exchange, routingKey, body)
	return args.Error(0)
}
func (m *mockTestPublisher) PublishToQueue(ctx context.Context, exchange, queue string, body interface{}) error {
	return nil
}
func (m *mockTestPublisher) DeclareTopicExchange(name string, d, a, i, n bool) error {
	return nil
}
func (m *mockTestPublisher) Close() error { return nil }

func setupHandlerWithMocks() (*Handlers, *MockRepository, *mockTestPublisher) {
	repo := new(MockRepository)
	pub := new(mockTestPublisher)
	// We pass pub as both outboxPublisher and eventPublisher/rpcPublisher
	// The Handlers constructor takes (repo, outboxPublisher, rpcPublisher, transactionManager)
	// Wait, publishEvent uses outboxPublisher which is eventPublisher internally.
	h := NewHandlers(repo, pub, pub, &NoOpTransactionManager{})
	return h, repo, pub
}

func TestHandleUpdateParty_Individual(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	existingInd := &domain.Individual{
		Party:     domain.Party{ID: "update-ind", Type: domain.PartyTypeIndividual, Status: "Initialized"},
		GivenName: "OldName",
	}

	repo.On("GetParty", ctx, "update-ind").Return(&existingInd.Party, nil)
	repo.On("UpdateIndividual", ctx, testifymock.AnythingOfType("*domain.Individual")).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyUpdated, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyStateChange, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, "", "reply-queue", testifymock.Anything).Return(nil)

	payload := map[string]interface{}{
		"@type":     "Individual",
		"id":        "update-ind",
		"givenName": "NewName",
		"status":    "Active",
	}
	body, _ := json.Marshal(payload)
	d := amqp.Delivery{Body: body, ReplyTo: "reply-queue"}

	err := h.HandleUpdateParty(ctx, d)
	assert.NoError(t, err)

	repo.AssertCalled(t, "GetParty", ctx, "update-ind")
	repo.AssertCalled(t, "UpdateIndividual", ctx, testifymock.MatchedBy(func(i *domain.Individual) bool {
		return i.GivenName == "NewName" && i.Status == "Active"
	}))
	pub.AssertCalled(t, "Publish", ctx, EventExchange, EvtPartyUpdated, testifymock.Anything)
	pub.AssertCalled(t, "Publish", ctx, EventExchange, EvtPartyStateChange, testifymock.Anything)
}

func TestHandleUpdateParty_Organization(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	existingOrg := &domain.Organization{
		Party:       domain.Party{ID: "update-org", Type: domain.PartyTypeOrganization, Status: "Initialized"},
		TradingName: "OldCorp",
	}

	repo.On("GetParty", ctx, "update-org").Return(&existingOrg.Party, nil)
	repo.On("UpdateOrganization", ctx, testifymock.AnythingOfType("*domain.Organization")).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyUpdated, testifymock.Anything).Return(nil)

	payload := map[string]interface{}{
		"@type":       "Organization",
		"id":          "update-org",
		"tradingName": "NewCorp",
	}
	body, _ := json.Marshal(payload)
	d := amqp.Delivery{Body: body}

	err := h.HandleUpdateParty(ctx, d)
	assert.NoError(t, err)
	repo.AssertCalled(t, "UpdateOrganization", ctx, testifymock.MatchedBy(func(o *domain.Organization) bool {
		return o.TradingName == "NewCorp" && o.Status == "Initialized" // Implicit keep
	}))
}

func TestHandlePatchParty_Individual(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "patch-ind").Return(&domain.Party{ID: "patch-ind", Type: domain.PartyTypeIndividual}, nil)
	repo.On("GetIndividual", ctx, "patch-ind").Return(&domain.Individual{
		Party:     domain.Party{ID: "patch-ind", Type: domain.PartyTypeIndividual, Status: "Active"},
		GivenName: "Old",
	}, nil)
	repo.On("UpdateIndividual", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyUpdated, testifymock.Anything).Return(nil)

	newName := "Patched"
	payload := PatchPartyPayload{ID: "patch-ind", GivenName: &newName}
	body, _ := json.Marshal(payload)

	err := h.HandlePatchParty(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err)
	repo.AssertCalled(t, "UpdateIndividual", ctx, testifymock.MatchedBy(func(i *domain.Individual) bool {
		return i.GivenName == "Patched"
	}))
}

func TestHandlePatchParty_Organization(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "patch-org").Return(&domain.Party{ID: "patch-org", Type: domain.PartyTypeOrganization}, nil)
	repo.On("GetOrganization", ctx, "patch-org").Return(&domain.Organization{
		Party: domain.Party{ID: "patch-org", Type: domain.PartyTypeOrganization, Status: "Active"},
	}, nil)
	repo.On("UpdateOrganization", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyUpdated, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyStateChange, testifymock.Anything).Return(nil)

	newStatus := "Validated"
	payload := PatchPartyPayload{ID: "patch-org", Status: &newStatus}
	body, _ := json.Marshal(payload)

	err := h.HandlePatchParty(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err)
}

func TestHandleFinalizeDeletion(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "del-fin").Return(&domain.Party{ID: "del-fin", Type: domain.PartyTypeIndividual, Status: string(domain.PartyStatusDeletionPending)}, nil)
	repo.On("GetIndividual", ctx, "del-fin").Return(&domain.Individual{Party: domain.Party{ID: "del-fin"}}, nil)
	repo.On("UpdateIndividual", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyDeleted, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyStateChange, testifymock.Anything).Return(nil)

	payload := DeletePartyPayload{ID: "del-fin"}
	body, _ := json.Marshal(payload)
	err := h.HandleFinalizeDeletion(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err)
}

func TestHandleCancelDeletion(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "del-can").Return(&domain.Party{ID: "del-can", Type: domain.PartyTypeOrganization, Status: string(domain.PartyStatusDeletionPending)}, nil)
	repo.On("GetOrganization", ctx, "del-can").Return(&domain.Organization{Party: domain.Party{ID: "del-can"}}, nil)
	repo.On("UpdateOrganization", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyStateChange, testifymock.Anything).Return(nil)

	payload := DeletePartyPayload{ID: "del-can"}
	body, _ := json.Marshal(payload)
	err := h.HandleCancelDeletion(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err)
}

func TestHandleCustomerCreated(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "party-123").Return(&domain.Party{ID: "party-123", Type: domain.PartyTypeIndividual, Status: string(domain.PartyStatusDeletionPending)}, nil)
	repo.On("GetIndividual", ctx, "party-123").Return(&domain.Individual{Party: domain.Party{ID: "party-123"}}, nil)
	repo.On("UpdateIndividual", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyStateChange, testifymock.Anything).Return(nil)

	payload := map[string]string{"id": "cust-1", "partyId": "party-123"}
	body, _ := json.Marshal(payload)
	err := h.HandleCustomerCreated(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err)
}

func TestHandleGetParty_IndividualData(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "get-party").Return(&domain.Party{ID: "get-party", Type: domain.PartyTypeIndividual}, nil)
	repo.On("GetIndividual", ctx, "get-party").Return(&domain.Individual{Party: domain.Party{ID: "get-party", Type: domain.PartyTypeIndividual}}, nil)
	pub.On("Publish", ctx, "", "rpc-queue", testifymock.Anything).Return(nil)

	payload := map[string]string{"id": "get-party"}
	body, _ := json.Marshal(payload)
	err := h.HandleGetParty(ctx, amqp.Delivery{Body: body, ReplyTo: "rpc-queue"})
	assert.NoError(t, err)
	pub.AssertCalled(t, "Publish", ctx, "", "rpc-queue", testifymock.Anything)
}
