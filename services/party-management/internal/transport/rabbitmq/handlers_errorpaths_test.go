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

// ========== Create Party Publish Event Errors ==========

func TestHandleCreateParty_IndividualPublishCreatedError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("CreateIndividual", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyCreated, testifymock.Anything).Return(errors.New("pub error"))

	body, _ := json.Marshal(map[string]any{
		"@type": "Individual", "id": "pub-err-1", "givenName": "Test",
	})
	err := h.HandleCreateParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pub error")
}


func TestHandleCreateParty_OrgPublishError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("CreateOrganization", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyCreated, testifymock.Anything).Return(errors.New("pub error"))

	body, _ := json.Marshal(map[string]any{
		"@type": "Organization", "id": "pub-err-3", "tradingName": "Corp",
	})
	err := h.HandleCreateParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}


// ========== Update Party Error Paths ==========

func TestHandleUpdateParty_IndividualGetPartyError(t *testing.T) {
	h, repo, _ := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "upd-err").Return(nil, errors.New("not found"))

	body, _ := json.Marshal(map[string]any{
		"@type": "Individual", "id": "upd-err", "givenName": "Test",
	})
	err := h.HandleUpdateParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get existing party")
}

func TestHandleUpdateParty_IndividualUpdateError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "upd-fail").Return(&domain.Party{ID: "upd-fail", Status: "Active"}, nil)
	repo.On("UpdateIndividual", ctx, testifymock.Anything).Return(errors.New("update fail"))
	pub.On("Publish", ctx, testifymock.Anything, testifymock.Anything, testifymock.Anything).Return(nil)

	body, _ := json.Marshal(map[string]any{
		"@type": "Individual", "id": "upd-fail", "givenName": "Test",
	})
	err := h.HandleUpdateParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandleUpdateParty_IndividualPublishError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "upd-pub-err").Return(&domain.Party{ID: "upd-pub-err", Status: "Active"}, nil)
	repo.On("UpdateIndividual", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyUpdated, testifymock.Anything).Return(errors.New("pub fail"))

	body, _ := json.Marshal(map[string]any{
		"@type": "Individual", "id": "upd-pub-err", "givenName": "Test",
	})
	err := h.HandleUpdateParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandleUpdateParty_IndividualStateChangePublishError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "upd-sc-err").Return(&domain.Party{ID: "upd-sc-err", Status: "Initialized"}, nil)
	repo.On("UpdateIndividual", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyUpdated, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyStateChange, testifymock.Anything).Return(errors.New("sc fail"))

	body, _ := json.Marshal(map[string]any{
		"@type": "Individual", "id": "upd-sc-err", "givenName": "Test", "status": "Active",
	})
	err := h.HandleUpdateParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandleUpdateParty_OrgGetPartyError(t *testing.T) {
	h, repo, _ := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "upd-org-err").Return(nil, errors.New("not found"))

	body, _ := json.Marshal(map[string]any{
		"@type": "Organization", "id": "upd-org-err",
	})
	err := h.HandleUpdateParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandleUpdateParty_OrgUpdateError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "upd-org-fail").Return(&domain.Party{ID: "upd-org-fail", Status: "Active"}, nil)
	repo.On("UpdateOrganization", ctx, testifymock.Anything).Return(errors.New("update fail"))
	pub.On("Publish", ctx, testifymock.Anything, testifymock.Anything, testifymock.Anything).Return(nil)

	body, _ := json.Marshal(map[string]any{
		"@type": "Organization", "id": "upd-org-fail", "tradingName": "Fail",
	})
	err := h.HandleUpdateParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandleUpdateParty_OrgPublishError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "upd-org-pub").Return(&domain.Party{ID: "upd-org-pub", Status: "Active"}, nil)
	repo.On("UpdateOrganization", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyUpdated, testifymock.Anything).Return(errors.New("pub fail"))

	body, _ := json.Marshal(map[string]any{
		"@type": "Organization", "id": "upd-org-pub", "tradingName": "Fail",
	})
	err := h.HandleUpdateParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandleUpdateParty_OrgStateChangePublishError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "upd-org-sc").Return(&domain.Party{ID: "upd-org-sc", Status: "Init"}, nil)
	repo.On("UpdateOrganization", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyUpdated, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyStateChange, testifymock.Anything).Return(errors.New("sc fail"))

	body, _ := json.Marshal(map[string]any{
		"@type": "Organization", "id": "upd-org-sc", "tradingName": "Corp", "status": "Active",
	})
	err := h.HandleUpdateParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

// ========== Patch Party Error Paths ==========

func TestHandlePatchParty_GetPartyError(t *testing.T) {
	h, repo, _ := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "ptch-err").Return(nil, errors.New("not found"))

	body, _ := json.Marshal(PatchPartyPayload{ID: "ptch-err"})
	err := h.HandlePatchParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandlePatchParty_IndividualGetError(t *testing.T) {
	h, repo, _ := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "ptch-ind-err").Return(&domain.Party{ID: "ptch-ind-err", Type: domain.PartyTypeIndividual}, nil)
	repo.On("GetIndividual", ctx, "ptch-ind-err").Return(nil, errors.New("not found"))

	body, _ := json.Marshal(PatchPartyPayload{ID: "ptch-ind-err"})
	err := h.HandlePatchParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandlePatchParty_IndividualUpdateError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "ptch-upd-err").Return(&domain.Party{ID: "ptch-upd-err", Type: domain.PartyTypeIndividual}, nil)
	repo.On("GetIndividual", ctx, "ptch-upd-err").Return(&domain.Individual{
		Party: domain.Party{ID: "ptch-upd-err", Status: "Active"},
	}, nil)
	repo.On("UpdateIndividual", ctx, testifymock.Anything).Return(errors.New("upd fail"))
	pub.On("Publish", ctx, testifymock.Anything, testifymock.Anything, testifymock.Anything).Return(nil)

	nm := "new"
	body, _ := json.Marshal(PatchPartyPayload{ID: "ptch-upd-err", GivenName: &nm})
	err := h.HandlePatchParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandlePatchParty_IndividualPublishError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "ptch-pub-err").Return(&domain.Party{ID: "ptch-pub-err", Type: domain.PartyTypeIndividual}, nil)
	repo.On("GetIndividual", ctx, "ptch-pub-err").Return(&domain.Individual{
		Party: domain.Party{ID: "ptch-pub-err", Status: "Active"},
	}, nil)
	repo.On("UpdateIndividual", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyUpdated, testifymock.Anything).Return(errors.New("pub fail"))

	nm := "new"
	body, _ := json.Marshal(PatchPartyPayload{ID: "ptch-pub-err", GivenName: &nm})
	err := h.HandlePatchParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandlePatchParty_IndividualStateChangePublishError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "ptch-sc-err").Return(&domain.Party{ID: "ptch-sc-err", Type: domain.PartyTypeIndividual}, nil)
	repo.On("GetIndividual", ctx, "ptch-sc-err").Return(&domain.Individual{
		Party: domain.Party{ID: "ptch-sc-err", Status: "Init"},
	}, nil)
	repo.On("UpdateIndividual", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyUpdated, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyStateChange, testifymock.Anything).Return(errors.New("sc fail"))

	st := "Active"
	body, _ := json.Marshal(PatchPartyPayload{ID: "ptch-sc-err", Status: &st})
	err := h.HandlePatchParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandlePatchParty_OrgGetError(t *testing.T) {
	h, repo, _ := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "ptch-org-err").Return(&domain.Party{ID: "ptch-org-err", Type: domain.PartyTypeOrganization}, nil)
	repo.On("GetOrganization", ctx, "ptch-org-err").Return(nil, errors.New("not found"))

	body, _ := json.Marshal(PatchPartyPayload{ID: "ptch-org-err"})
	err := h.HandlePatchParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandlePatchParty_OrgUpdateError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "ptch-org-uer").Return(&domain.Party{ID: "ptch-org-uer", Type: domain.PartyTypeOrganization}, nil)
	repo.On("GetOrganization", ctx, "ptch-org-uer").Return(&domain.Organization{
		Party: domain.Party{ID: "ptch-org-uer", Status: "Active"},
	}, nil)
	repo.On("UpdateOrganization", ctx, testifymock.Anything).Return(errors.New("upd fail"))
	pub.On("Publish", ctx, testifymock.Anything, testifymock.Anything, testifymock.Anything).Return(nil)

	st := "Suspended"
	body, _ := json.Marshal(PatchPartyPayload{ID: "ptch-org-uer", Status: &st})
	err := h.HandlePatchParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandlePatchParty_OrgPublishError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "ptch-org-per").Return(&domain.Party{ID: "ptch-org-per", Type: domain.PartyTypeOrganization}, nil)
	repo.On("GetOrganization", ctx, "ptch-org-per").Return(&domain.Organization{
		Party: domain.Party{ID: "ptch-org-per", Status: "Active"},
	}, nil)
	repo.On("UpdateOrganization", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyUpdated, testifymock.Anything).Return(errors.New("pub fail"))

	st := "Suspended"
	body, _ := json.Marshal(PatchPartyPayload{ID: "ptch-org-per", Status: &st})
	err := h.HandlePatchParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandlePatchParty_OrgStateChangePublishError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "ptch-org-sc").Return(&domain.Party{ID: "ptch-org-sc", Type: domain.PartyTypeOrganization}, nil)
	repo.On("GetOrganization", ctx, "ptch-org-sc").Return(&domain.Organization{
		Party: domain.Party{ID: "ptch-org-sc", Status: "Active"},
	}, nil)
	repo.On("UpdateOrganization", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyUpdated, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyStateChange, testifymock.Anything).Return(errors.New("sc fail"))

	st := "Suspended"
	body, _ := json.Marshal(PatchPartyPayload{ID: "ptch-org-sc", Status: &st})
	err := h.HandlePatchParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandlePatchParty_UnknownType(t *testing.T) {
	h, repo, _ := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "ptch-unk").Return(&domain.Party{ID: "ptch-unk", Type: "UnknownType"}, nil)

	body, _ := json.Marshal(PatchPartyPayload{ID: "ptch-unk"})
	err := h.HandlePatchParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid party type")
}

// ========== Delete Party Error Paths ==========

func TestHandleDeleteParty_GetPartyError(t *testing.T) {
	h, repo, _ := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "del-err").Return(nil, errors.New("not found"))

	body, _ := json.Marshal(DeletePartyPayload{ID: "del-err"})
	err := h.HandleDeleteParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandleDeleteParty_IndividualUpdateError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "del-uf").Return(&domain.Party{ID: "del-uf", Type: domain.PartyTypeIndividual, Status: "Active"}, nil)
	repo.On("GetIndividual", ctx, "del-uf").Return(&domain.Individual{Party: domain.Party{ID: "del-uf"}}, nil)
	repo.On("UpdateIndividual", ctx, testifymock.Anything).Return(errors.New("upd fail"))
	pub.On("Publish", ctx, testifymock.Anything, testifymock.Anything, testifymock.Anything).Return(nil)

	body, _ := json.Marshal(DeletePartyPayload{ID: "del-uf"})
	err := h.HandleDeleteParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandleDeleteParty_OrgUpdateError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "del-ou").Return(&domain.Party{ID: "del-ou", Type: domain.PartyTypeOrganization, Status: "Active"}, nil)
	repo.On("GetOrganization", ctx, "del-ou").Return(&domain.Organization{Party: domain.Party{ID: "del-ou"}}, nil)
	repo.On("UpdateOrganization", ctx, testifymock.Anything).Return(errors.New("upd fail"))
	pub.On("Publish", ctx, testifymock.Anything, testifymock.Anything, testifymock.Anything).Return(nil)

	body, _ := json.Marshal(DeletePartyPayload{ID: "del-ou"})
	err := h.HandleDeleteParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandleDeleteParty_PublishDeletionInitiatedError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "del-pe").Return(&domain.Party{ID: "del-pe", Type: domain.PartyTypeIndividual, Status: "Active"}, nil)
	repo.On("GetIndividual", ctx, "del-pe").Return(&domain.Individual{Party: domain.Party{ID: "del-pe"}}, nil)
	repo.On("UpdateIndividual", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyDeletionInitiated, testifymock.Anything).Return(errors.New("pub fail"))

	body, _ := json.Marshal(DeletePartyPayload{ID: "del-pe"})
	err := h.HandleDeleteParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandleDeleteParty_PublishStateChangeError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "del-se").Return(&domain.Party{ID: "del-se", Type: domain.PartyTypeIndividual, Status: "Active"}, nil)
	repo.On("GetIndividual", ctx, "del-se").Return(&domain.Individual{Party: domain.Party{ID: "del-se"}}, nil)
	repo.On("UpdateIndividual", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyDeletionInitiated, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyStateChange, testifymock.Anything).Return(errors.New("sc fail"))

	body, _ := json.Marshal(DeletePartyPayload{ID: "del-se"})
	err := h.HandleDeleteParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandleDeleteParty_IndividualGetError(t *testing.T) {
	h, repo, _ := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "del-ige").Return(&domain.Party{ID: "del-ige", Type: domain.PartyTypeIndividual, Status: "Active"}, nil)
	repo.On("GetIndividual", ctx, "del-ige").Return(nil, errors.New("not found"))

	body, _ := json.Marshal(DeletePartyPayload{ID: "del-ige"})
	err := h.HandleDeleteParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandleDeleteParty_OrgGetError(t *testing.T) {
	h, repo, _ := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "del-oge").Return(&domain.Party{ID: "del-oge", Type: domain.PartyTypeOrganization, Status: "Active"}, nil)
	repo.On("GetOrganization", ctx, "del-oge").Return(nil, errors.New("not found"))

	body, _ := json.Marshal(DeletePartyPayload{ID: "del-oge"})
	err := h.HandleDeleteParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

// ========== Finalize Deletion Error Paths ==========

func TestHandleFinalizeDeletion_GetPartyError(t *testing.T) {
	h, repo, _ := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "fin-ge").Return(nil, errors.New("not found"))

	body, _ := json.Marshal(DeletePartyPayload{ID: "fin-ge"})
	err := h.HandleFinalizeDeletion(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandleFinalizeDeletion_IndividualUpdateError(t *testing.T) {
	h, repo, _ := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "fin-iue").Return(&domain.Party{ID: "fin-iue", Type: domain.PartyTypeIndividual, Status: string(domain.PartyStatusDeletionPending)}, nil)
	repo.On("GetIndividual", ctx, "fin-iue").Return(&domain.Individual{Party: domain.Party{ID: "fin-iue"}}, nil)
	repo.On("UpdateIndividual", ctx, testifymock.Anything).Return(errors.New("upd fail"))

	body, _ := json.Marshal(DeletePartyPayload{ID: "fin-iue"})
	err := h.HandleFinalizeDeletion(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandleFinalizeDeletion_OrgUpdateError(t *testing.T) {
	h, repo, _ := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "fin-oue").Return(&domain.Party{ID: "fin-oue", Type: domain.PartyTypeOrganization, Status: string(domain.PartyStatusDeletionPending)}, nil)
	repo.On("GetOrganization", ctx, "fin-oue").Return(&domain.Organization{Party: domain.Party{ID: "fin-oue"}}, nil)
	repo.On("UpdateOrganization", ctx, testifymock.Anything).Return(errors.New("upd fail"))

	body, _ := json.Marshal(DeletePartyPayload{ID: "fin-oue"})
	err := h.HandleFinalizeDeletion(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandleFinalizeDeletion_PublishDeletedError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "fin-pe").Return(&domain.Party{ID: "fin-pe", Type: domain.PartyTypeIndividual, Status: string(domain.PartyStatusDeletionPending)}, nil)
	repo.On("GetIndividual", ctx, "fin-pe").Return(&domain.Individual{Party: domain.Party{ID: "fin-pe"}}, nil)
	repo.On("UpdateIndividual", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyDeleted, testifymock.Anything).Return(errors.New("pub fail"))

	body, _ := json.Marshal(DeletePartyPayload{ID: "fin-pe"})
	err := h.HandleFinalizeDeletion(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandleFinalizeDeletion_PublishStateChangeError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "fin-sce").Return(&domain.Party{ID: "fin-sce", Type: domain.PartyTypeIndividual, Status: string(domain.PartyStatusDeletionPending)}, nil)
	repo.On("GetIndividual", ctx, "fin-sce").Return(&domain.Individual{Party: domain.Party{ID: "fin-sce"}}, nil)
	repo.On("UpdateIndividual", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyDeleted, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyStateChange, testifymock.Anything).Return(errors.New("sc fail"))

	body, _ := json.Marshal(DeletePartyPayload{ID: "fin-sce"})
	err := h.HandleFinalizeDeletion(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

// ========== Cancel Deletion Error Paths ==========

func TestHandleCancelDeletion_GetPartyError(t *testing.T) {
	h, repo, _ := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "can-ge").Return(nil, errors.New("not found"))

	body, _ := json.Marshal(DeletePartyPayload{ID: "can-ge"})
	err := h.HandleCancelDeletion(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandleCancelDeletion_IndividualUpdateError(t *testing.T) {
	h, repo, _ := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "can-iue").Return(&domain.Party{ID: "can-iue", Type: domain.PartyTypeIndividual, Status: string(domain.PartyStatusDeletionPending)}, nil)
	repo.On("GetIndividual", ctx, "can-iue").Return(&domain.Individual{Party: domain.Party{ID: "can-iue"}}, nil)
	repo.On("UpdateIndividual", ctx, testifymock.Anything).Return(errors.New("upd fail"))

	body, _ := json.Marshal(DeletePartyPayload{ID: "can-iue"})
	err := h.HandleCancelDeletion(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandleCancelDeletion_OrgUpdateError(t *testing.T) {
	h, repo, _ := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "can-oue").Return(&domain.Party{ID: "can-oue", Type: domain.PartyTypeOrganization, Status: string(domain.PartyStatusDeletionPending)}, nil)
	repo.On("GetOrganization", ctx, "can-oue").Return(&domain.Organization{Party: domain.Party{ID: "can-oue"}}, nil)
	repo.On("UpdateOrganization", ctx, testifymock.Anything).Return(errors.New("upd fail"))

	body, _ := json.Marshal(DeletePartyPayload{ID: "can-oue"})
	err := h.HandleCancelDeletion(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandleCancelDeletion_PublishStateChangeError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "can-sce").Return(&domain.Party{ID: "can-sce", Type: domain.PartyTypeIndividual, Status: string(domain.PartyStatusDeletionPending)}, nil)
	repo.On("GetIndividual", ctx, "can-sce").Return(&domain.Individual{Party: domain.Party{ID: "can-sce"}}, nil)
	repo.On("UpdateIndividual", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyStateChange, testifymock.Anything).Return(errors.New("sc fail"))

	body, _ := json.Marshal(DeletePartyPayload{ID: "can-sce"})
	err := h.HandleCancelDeletion(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

// ========== Customer Created Error Paths ==========

func TestHandleCustomerCreated_InvalidJSON(t *testing.T) {
	h, _, _ := setupHandlerWithMocks()
	err := h.HandleCustomerCreated(context.Background(), amqp.Delivery{Body: []byte("bad")})
	assert.Error(t, err)
}

func TestHandleCustomerCreated_IndividualUpdateError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "cc-iue").Return(&domain.Party{ID: "cc-iue", Type: domain.PartyTypeIndividual, Status: string(domain.PartyStatusDeletionPending)}, nil)
	repo.On("GetIndividual", ctx, "cc-iue").Return(&domain.Individual{Party: domain.Party{ID: "cc-iue"}}, nil)
	repo.On("UpdateIndividual", ctx, testifymock.Anything).Return(errors.New("upd fail"))
	pub.On("Publish", ctx, testifymock.Anything, testifymock.Anything, testifymock.Anything).Return(nil)

	body, _ := json.Marshal(map[string]string{"id": "c", "partyId": "cc-iue"})
	err := h.HandleCustomerCreated(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandleCustomerCreated_OrgUpdateError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "cc-oue").Return(&domain.Party{ID: "cc-oue", Type: domain.PartyTypeOrganization, Status: string(domain.PartyStatusDeletionPending)}, nil)
	repo.On("GetOrganization", ctx, "cc-oue").Return(&domain.Organization{Party: domain.Party{ID: "cc-oue"}}, nil)
	repo.On("UpdateOrganization", ctx, testifymock.Anything).Return(errors.New("upd fail"))
	pub.On("Publish", ctx, testifymock.Anything, testifymock.Anything, testifymock.Anything).Return(nil)

	body, _ := json.Marshal(map[string]string{"id": "c", "partyId": "cc-oue"})
	err := h.HandleCustomerCreated(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}

func TestHandleCustomerCreated_PublishStateChangeError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "cc-sce").Return(&domain.Party{ID: "cc-sce", Type: domain.PartyTypeIndividual, Status: string(domain.PartyStatusDeletionPending)}, nil)
	repo.On("GetIndividual", ctx, "cc-sce").Return(&domain.Individual{Party: domain.Party{ID: "cc-sce"}}, nil)
	repo.On("UpdateIndividual", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyStateChange, testifymock.Anything).Return(errors.New("sc fail"))

	body, _ := json.Marshal(map[string]string{"id": "c", "partyId": "cc-sce"})
	err := h.HandleCustomerCreated(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
}
