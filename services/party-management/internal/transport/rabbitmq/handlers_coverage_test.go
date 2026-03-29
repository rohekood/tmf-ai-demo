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

// ========== Payload Validation Tests ==========

func TestCreateIndividualPayload_Validate(t *testing.T) {
	p := &CreateIndividualPayload{ID: ""}
	assert.ErrorIs(t, p.Validate(), domain.ErrIDRequired)

	p.ID = "valid"
	assert.NoError(t, p.Validate())
}

func TestCreateOrganizationPayload_Validate(t *testing.T) {
	p := &CreateOrganizationPayload{ID: ""}
	assert.ErrorIs(t, p.Validate(), domain.ErrIDRequired)

	p.ID = "valid"
	assert.NoError(t, p.Validate())
}

func TestCreatePartyPayload_Validate(t *testing.T) {
	t.Run("Invalid type", func(t *testing.T) {
		p := &CreatePartyPayload{Type: "Unknown"}
		assert.ErrorIs(t, p.Validate(), domain.ErrInvalidType)
	})
	t.Run("Individual nil payload", func(t *testing.T) {
		p := &CreatePartyPayload{Type: "Individual", Individual: nil}
		assert.Error(t, p.Validate())
	})
	t.Run("Individual valid", func(t *testing.T) {
		p := &CreatePartyPayload{Type: "Individual", Individual: &CreateIndividualPayload{ID: "x"}}
		assert.NoError(t, p.Validate())
	})
	t.Run("Organization nil payload", func(t *testing.T) {
		p := &CreatePartyPayload{Type: "Organization", Organization: nil}
		assert.Error(t, p.Validate())
	})
	t.Run("Organization valid", func(t *testing.T) {
		p := &CreatePartyPayload{Type: "Organization", Organization: &CreateOrganizationPayload{ID: "x"}}
		assert.NoError(t, p.Validate())
	})
}

func TestUpdatePartyPayload_Validate(t *testing.T) {
	t.Run("Missing ID", func(t *testing.T) {
		p := &UpdatePartyPayload{ID: "", Type: "Individual"}
		assert.ErrorIs(t, p.Validate(), domain.ErrIDRequired)
	})
	t.Run("Invalid type", func(t *testing.T) {
		p := &UpdatePartyPayload{ID: "x", Type: "Unknown"}
		assert.ErrorIs(t, p.Validate(), domain.ErrInvalidType)
	})
	t.Run("Individual nil", func(t *testing.T) {
		p := &UpdatePartyPayload{ID: "x", Type: "Individual", Individual: nil}
		assert.Error(t, p.Validate())
	})
	t.Run("Organization nil", func(t *testing.T) {
		p := &UpdatePartyPayload{ID: "x", Type: "Organization", Organization: nil}
		assert.Error(t, p.Validate())
	})
	t.Run("Individual valid", func(t *testing.T) {
		p := &UpdatePartyPayload{ID: "x", Type: "Individual", Individual: &CreateIndividualPayload{ID: "x"}}
		assert.NoError(t, p.Validate())
	})
	t.Run("Organization valid", func(t *testing.T) {
		p := &UpdatePartyPayload{ID: "x", Type: "Organization", Organization: &CreateOrganizationPayload{ID: "x"}}
		assert.NoError(t, p.Validate())
	})
}

func TestPatchPartyPayload_Validate(t *testing.T) {
	p := &PatchPartyPayload{ID: ""}
	assert.ErrorIs(t, p.Validate(), domain.ErrIDRequired)

	p.ID = "valid"
	assert.NoError(t, p.Validate())
}

// ========== Handler Error Path Tests ==========

func TestHandleCreateParty_InvalidJSON(t *testing.T) {
	h, _, _ := setupHandlerWithMocks()
	err := h.HandleCreateParty(context.Background(), amqp.Delivery{Body: []byte("not-json")})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal type info")
}

func TestHandleCreateParty_InvalidType(t *testing.T) {
	h, _, _ := setupHandlerWithMocks()
	body, _ := json.Marshal(map[string]string{"@type": "Unknown"})
	err := h.HandleCreateParty(context.Background(), amqp.Delivery{Body: body})
	assert.ErrorIs(t, err, domain.ErrInvalidType)
}

func TestHandleCreateParty_IndividualRepoError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("CreateIndividual", ctx, testifymock.Anything).Return(errors.New("db error"))
	pub.On("Publish", ctx, testifymock.Anything, testifymock.Anything, testifymock.Anything).Return(nil)

	body, _ := json.Marshal(map[string]interface{}{
		"@type": "Individual", "id": "fail-1", "givenName": "Test",
	})
	err := h.HandleCreateParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create individual")
}

func TestHandleCreateParty_OrganizationRepoError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("CreateOrganization", ctx, testifymock.Anything).Return(errors.New("db error"))
	pub.On("Publish", ctx, testifymock.Anything, testifymock.Anything, testifymock.Anything).Return(nil)

	body, _ := json.Marshal(map[string]interface{}{
		"@type": "Organization", "id": "fail-2", "tradingName": "Corp",
	})
	err := h.HandleCreateParty(ctx, amqp.Delivery{Body: body})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create organization")
}

func TestHandleUpdateParty_InvalidJSON(t *testing.T) {
	h, _, _ := setupHandlerWithMocks()
	err := h.HandleUpdateParty(context.Background(), amqp.Delivery{Body: []byte("bad")})
	assert.Error(t, err)
}

func TestHandleUpdateParty_InvalidType(t *testing.T) {
	h, _, _ := setupHandlerWithMocks()
	body, _ := json.Marshal(map[string]string{"@type": "Unknown"})
	err := h.HandleUpdateParty(context.Background(), amqp.Delivery{Body: body})
	assert.ErrorIs(t, err, domain.ErrInvalidType)
}

func TestHandleUpdateParty_MissingID(t *testing.T) {
	h, _, _ := setupHandlerWithMocks()
	body, _ := json.Marshal(map[string]string{"@type": "Individual"})
	err := h.HandleUpdateParty(context.Background(), amqp.Delivery{Body: body})
	assert.ErrorIs(t, err, domain.ErrIDRequired)
}

func TestHandleUpdateParty_OrgMissingID(t *testing.T) {
	h, _, _ := setupHandlerWithMocks()
	body, _ := json.Marshal(map[string]string{"@type": "Organization"})
	err := h.HandleUpdateParty(context.Background(), amqp.Delivery{Body: body})
	assert.ErrorIs(t, err, domain.ErrIDRequired)
}

func TestHandlePatchParty_InvalidJSON(t *testing.T) {
	h, _, _ := setupHandlerWithMocks()
	err := h.HandlePatchParty(context.Background(), amqp.Delivery{Body: []byte("bad")})
	assert.Error(t, err)
}

func TestHandlePatchParty_MissingID(t *testing.T) {
	h, _, _ := setupHandlerWithMocks()
	body, _ := json.Marshal(PatchPartyPayload{})
	err := h.HandlePatchParty(context.Background(), amqp.Delivery{Body: body})
	assert.ErrorIs(t, err, domain.ErrIDRequired)
}

func TestHandleDeleteParty_InvalidJSON(t *testing.T) {
	h, _, _ := setupHandlerWithMocks()
	err := h.HandleDeleteParty(context.Background(), amqp.Delivery{Body: []byte("bad")})
	assert.Error(t, err)
}

func TestHandleDeleteParty_MissingID(t *testing.T) {
	h, _, _ := setupHandlerWithMocks()
	body, _ := json.Marshal(DeletePartyPayload{})
	err := h.HandleDeleteParty(context.Background(), amqp.Delivery{Body: body})
	assert.ErrorIs(t, err, domain.ErrIDRequired)
}

func TestHandleDeleteParty_Organization(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "del-org").Return(&domain.Party{
		ID: "del-org", Type: domain.PartyTypeOrganization, Status: "Active",
	}, nil)
	repo.On("GetOrganization", ctx, "del-org").Return(&domain.Organization{
		Party: domain.Party{ID: "del-org"},
	}, nil)
	repo.On("UpdateOrganization", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, testifymock.Anything, testifymock.Anything).Return(nil)

	body, _ := json.Marshal(DeletePartyPayload{ID: "del-org"})
	err := h.HandleDeleteParty(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err)
}

func TestHandleDeleteParty_AlreadyPending(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "del-dup").Return(&domain.Party{
		ID: "del-dup", Type: domain.PartyTypeIndividual, Status: string(domain.PartyStatusDeletionPending),
	}, nil)
	pub.On("Publish", ctx, testifymock.Anything, testifymock.Anything, testifymock.Anything).Return(nil)

	body, _ := json.Marshal(DeletePartyPayload{ID: "del-dup"})
	err := h.HandleDeleteParty(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err)
}

func TestHandleFinalizeDeletion_InvalidJSON(t *testing.T) {
	h, _, _ := setupHandlerWithMocks()
	err := h.HandleFinalizeDeletion(context.Background(), amqp.Delivery{Body: []byte("bad")})
	assert.Error(t, err)
}

func TestHandleFinalizeDeletion_NotPending(t *testing.T) {
	h, repo, _ := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "fin-skip").Return(&domain.Party{
		ID: "fin-skip", Type: domain.PartyTypeIndividual, Status: "Active",
	}, nil)

	body, _ := json.Marshal(DeletePartyPayload{ID: "fin-skip"})
	err := h.HandleFinalizeDeletion(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err) // skipped gracefully
}

func TestHandleFinalizeDeletion_Organization(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "fin-org").Return(&domain.Party{
		ID: "fin-org", Type: domain.PartyTypeOrganization, Status: string(domain.PartyStatusDeletionPending),
	}, nil)
	repo.On("GetOrganization", ctx, "fin-org").Return(&domain.Organization{Party: domain.Party{ID: "fin-org"}}, nil)
	repo.On("UpdateOrganization", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, testifymock.Anything, testifymock.Anything).Return(nil)

	body, _ := json.Marshal(DeletePartyPayload{ID: "fin-org"})
	err := h.HandleFinalizeDeletion(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err)
}

func TestHandleCancelDeletion_InvalidJSON(t *testing.T) {
	h, _, _ := setupHandlerWithMocks()
	err := h.HandleCancelDeletion(context.Background(), amqp.Delivery{Body: []byte("bad")})
	assert.Error(t, err)
}

func TestHandleCancelDeletion_NotPending(t *testing.T) {
	h, repo, _ := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "can-skip").Return(&domain.Party{
		ID: "can-skip", Status: "Active",
	}, nil)

	body, _ := json.Marshal(DeletePartyPayload{ID: "can-skip"})
	err := h.HandleCancelDeletion(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err)
}

func TestHandleCancelDeletion_Individual(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "can-ind").Return(&domain.Party{
		ID: "can-ind", Type: domain.PartyTypeIndividual, Status: string(domain.PartyStatusDeletionPending),
	}, nil)
	repo.On("GetIndividual", ctx, "can-ind").Return(&domain.Individual{Party: domain.Party{ID: "can-ind"}}, nil)
	repo.On("UpdateIndividual", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, testifymock.Anything, testifymock.Anything).Return(nil)

	body, _ := json.Marshal(DeletePartyPayload{ID: "can-ind"})
	err := h.HandleCancelDeletion(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err)
}

func TestHandleCustomerCreated_EmptyPartyID(t *testing.T) {
	h, _, _ := setupHandlerWithMocks()
	body, _ := json.Marshal(map[string]string{"id": "c1", "partyId": ""})
	err := h.HandleCustomerCreated(context.Background(), amqp.Delivery{Body: body})
	assert.NoError(t, err) // Should return nil early
}

func TestHandleCustomerCreated_PartyNotDeletionPending(t *testing.T) {
	h, repo, _ := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "p-ok").Return(&domain.Party{
		ID: "p-ok", Status: "Active",
	}, nil)

	body, _ := json.Marshal(map[string]string{"id": "c2", "partyId": "p-ok"})
	err := h.HandleCustomerCreated(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err)
}

func TestHandleCustomerCreated_PartyNotFound(t *testing.T) {
	h, repo, _ := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "p-missing").Return(nil, errors.New("not found"))

	body, _ := json.Marshal(map[string]string{"id": "c3", "partyId": "p-missing"})
	err := h.HandleCustomerCreated(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err) // Should return nil (ignore not found)
}

func TestHandleCustomerCreated_Organization(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "p-org").Return(&domain.Party{
		ID: "p-org", Type: domain.PartyTypeOrganization, Status: string(domain.PartyStatusDeletionPending),
	}, nil)
	repo.On("GetOrganization", ctx, "p-org").Return(&domain.Organization{Party: domain.Party{ID: "p-org"}}, nil)
	repo.On("UpdateOrganization", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, testifymock.Anything, testifymock.Anything).Return(nil)

	body, _ := json.Marshal(map[string]string{"id": "c4", "partyId": "p-org"})
	err := h.HandleCustomerCreated(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err)
}

func TestHandleGetParty_InvalidJSON(t *testing.T) {
	h, _, _ := setupHandlerWithMocks()
	err := h.HandleGetParty(context.Background(), amqp.Delivery{Body: []byte("bad")})
	assert.Error(t, err)
}

func TestHandleGetParty_MissingID(t *testing.T) {
	h, _, pub := setupHandlerWithMocks()
	ctx := context.Background()
	pub.On("Publish", ctx, "", "reply", testifymock.Anything).Return(nil)

	body, _ := json.Marshal(map[string]string{"id": ""})
	err := h.HandleGetParty(ctx, amqp.Delivery{Body: body, ReplyTo: "reply"})
	assert.NoError(t, err) // replies with error payload
}

func TestHandleGetParty_RepoError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "err-id").Return(nil, errors.New("db error"))
	pub.On("Publish", ctx, "", "reply", testifymock.Anything).Return(nil)

	body, _ := json.Marshal(map[string]string{"id": "err-id"})
	err := h.HandleGetParty(ctx, amqp.Delivery{Body: body, ReplyTo: "reply"})
	assert.NoError(t, err) // replies with error payload
}

func TestHandleGetParty_Organization(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "org-get").Return(&domain.Party{
		ID: "org-get", Type: domain.PartyTypeOrganization,
	}, nil)
	repo.On("GetOrganization", ctx, "org-get").Return(&domain.Organization{
		Party: domain.Party{ID: "org-get"},
	}, nil)
	pub.On("Publish", ctx, "", "reply", testifymock.Anything).Return(nil)

	body, _ := json.Marshal(map[string]string{"id": "org-get"})
	err := h.HandleGetParty(ctx, amqp.Delivery{Body: body, ReplyTo: "reply"})
	assert.NoError(t, err)
}

func TestHandleSearchParty_InvalidJSON(t *testing.T) {
	h, _, _ := setupHandlerWithMocks()
	err := h.HandleSearchParty(context.Background(), amqp.Delivery{Body: []byte("bad")})
	assert.Error(t, err)
}

func TestHandleSearchParty_RepoError(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("SearchParties", ctx, testifymock.Anything).Return(nil, errors.New("search error"))
	pub.On("Publish", ctx, "", "reply", testifymock.Anything).Return(nil)

	body, _ := json.Marshal(SearchPartyPayload{})
	err := h.HandleSearchParty(ctx, amqp.Delivery{Body: body, ReplyTo: "reply"})
	assert.NoError(t, err) // replies with error
}

func TestHandleSearchParty_WithAllCriteria(t *testing.T) {
	h, repo, _ := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("SearchParties", ctx, testifymock.Anything).Return([]domain.Party{}, nil)

	search := "q"
	name := "n"
	given := "g"
	family := "f"
	trading := "t"
	tp := "Individual"
	extRef := "ext"
	payload := SearchPartyPayload{
		Search:            &search,
		Name:              &name,
		GivenName:         &given,
		FamilyName:        &family,
		TradingName:       &trading,
		Type:              &tp,
		ExternalReference: &extRef,
	}
	body, _ := json.Marshal(payload)
	err := h.HandleSearchParty(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err)
}

func TestHandleSearchParty_IndividualFetchError(t *testing.T) {
	h, repo, _ := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("SearchParties", ctx, testifymock.Anything).Return([]domain.Party{
		{ID: "err-ind", Type: domain.PartyTypeIndividual},
	}, nil)
	repo.On("GetIndividual", ctx, "err-ind").Return(nil, errors.New("fetch error"))

	body, _ := json.Marshal(SearchPartyPayload{})
	err := h.HandleSearchParty(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err) // fallback to base party
}

func TestHandleSearchParty_OrgFetchError(t *testing.T) {
	h, repo, _ := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("SearchParties", ctx, testifymock.Anything).Return([]domain.Party{
		{ID: "err-org", Type: domain.PartyTypeOrganization},
	}, nil)
	repo.On("GetOrganization", ctx, "err-org").Return(nil, errors.New("fetch error"))

	body, _ := json.Marshal(SearchPartyPayload{})
	err := h.HandleSearchParty(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err) // fallback to base party
}

func TestHandleSearchParty_UnknownType(t *testing.T) {
	h, repo, _ := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("SearchParties", ctx, testifymock.Anything).Return([]domain.Party{
		{ID: "unk", Type: "UnknownType"},
	}, nil)

	body, _ := json.Marshal(SearchPartyPayload{})
	err := h.HandleSearchParty(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err) // appends base party
}

// ========== Mapping Helper Tests ==========

func TestMapContactMediums_WithAndWithoutID(t *testing.T) {
	h := &Handlers{}

	dtos := []ContactMediumDTO{
		{ID: "existing-id", MediumType: "email", Value: "test@test.com"},
		{ID: "", MediumType: "phone", Value: "123"},
	}
	result := h.mapContactMediums(dtos, "party-1")
	assert.Len(t, result, 2)
	assert.Equal(t, "existing-id", result[0].ID)
	assert.NotEmpty(t, result[1].ID) // auto-generated
	assert.Equal(t, "party-1", result[0].PartyID)
	assert.Equal(t, "party-1", result[1].PartyID)
	assert.Equal(t, "email", result[0].MediumType)
	assert.Equal(t, "phone", result[1].MediumType)
}

func TestMapIdentifications_WithAndWithoutID(t *testing.T) {
	h := &Handlers{}

	dtos := []IdentificationDTO{
		{ID: "id-1", IdentificationType: "passport", IssuingDate: "2024-01-01T00:00:00Z"},
		{ID: "", IdentificationType: "id-card", IssuingDate: ""},           // no date, auto-gen ID
		{ID: "id-2", IdentificationType: "driver", IssuingDate: "invalid"}, // invalid date
	}
	result := h.mapIdentifications(dtos, "party-2")
	assert.Len(t, result, 3)
	assert.Equal(t, "id-1", result[0].ID)
	assert.NotEmpty(t, result[1].ID)
	assert.Equal(t, "party-2", result[0].PartyID)
	assert.False(t, result[0].IssuingDate.IsZero())
	assert.True(t, result[1].IssuingDate.IsZero())
}

func TestMapCharacteristics_WithID(t *testing.T) {
	h := &Handlers{}

	dtos := []CharacteristicDTO{
		{ID: "c1", Name: "key", Value: "val", ValueType: "string"},
		{ID: "", Name: "key2", Value: "val2", ValueType: "int"},
	}
	result := h.mapCharacteristics(dtos, "party-3")
	assert.Len(t, result, 2)
	assert.Equal(t, "c1", result[0].ID)
	assert.Equal(t, "party-3", result[0].PartyID)
}

func TestMapExternalReferences_AutoID(t *testing.T) {
	h := &Handlers{}

	dtos := []ExternalReferenceDTO{
		{ID: "", ExternalSystemID: "sys1", ExternalReferenceID: "ref1"},
		{ID: "pre-set", ExternalSystemID: "sys2", ExternalReferenceID: "ref2"},
	}
	result := h.mapExternalReferences(dtos, "party-4")
	assert.Len(t, result, 2)
	assert.NotEmpty(t, result[0].ID)
	assert.Equal(t, "pre-set", result[1].ID)
}

func TestMapTaxExemptions_WithDates(t *testing.T) {
	h := &Handlers{}

	dtos := []TaxExemptionDTO{
		{ID: "", CertificateNumber: "cert1", ValidForStart: "2024-01-01T00:00:00Z", ValidForEnd: "2025-01-01T00:00:00Z"},
		{ID: "t2", CertificateNumber: "cert2", ValidForStart: "", ValidForEnd: ""},
		{ID: "", CertificateNumber: "cert3", ValidForStart: "invalid", ValidForEnd: "invalid"},
	}
	result := h.mapTaxExemptions(dtos, "party-5")
	assert.Len(t, result, 3)
	assert.NotEmpty(t, result[0].ID)
	assert.NotNil(t, result[0].ValidForEnd)
	assert.Nil(t, result[1].ValidForEnd)
}

func TestMapAttachments_AllBranches(t *testing.T) {
	h := &Handlers{}

	dtos := []AttachmentDTO{
		{ID: "", Name: "file1", URL: "https://s3.example.com/file", RefType: "", RefID: ""}, // S3 fallback
		{ID: "a2", Name: "file2", RefType: "External", RefID: "ref-2"},                      // pre-set
		{ID: "", Name: "file3", URL: "", RefType: ""},                                       // Internal fallback
		{ID: "", Name: "file4", Content: []byte("data"), RefType: "", RefID: ""},            // internal with content
	}
	result := h.mapAttachments(dtos, "party-6")
	assert.Len(t, result, 4)
	assert.Equal(t, "S3", result[0].RefType)
	assert.Equal(t, "https://s3.example.com/file", result[0].RefID)
	assert.Equal(t, "External", result[1].RefType)
	assert.Equal(t, "Internal", result[2].RefType)
	assert.Equal(t, "Internal", result[3].RefType)
}

// ========== Helper function tests ==========

func TestPublishEvent_NilPublisher(t *testing.T) {
	h := NewHandlers(new(MockRepository), nil, nil, &NoOpTransactionManager{})
	err := h.publishEvent(context.Background(), "test.key", map[string]string{"k": "v"})
	assert.NoError(t, err) // warns but no error
}

func TestPublishEvent_PublishError(t *testing.T) {
	pub := new(mockTestPublisher)
	h := NewHandlers(new(MockRepository), pub, nil, &NoOpTransactionManager{})

	pub.On("Publish", testifymock.Anything, testifymock.Anything, testifymock.Anything, testifymock.Anything).Return(errors.New("publish failed"))

	err := h.publishEvent(context.Background(), "test.key", "data")
	assert.Error(t, err)
}

func TestReplyTo_EmptyReplyTo(t *testing.T) {
	h := NewHandlers(new(MockRepository), nil, nil, &NoOpTransactionManager{})
	err := h.replyTo(context.Background(), amqp.Delivery{}, "response")
	assert.NoError(t, err) // no ReplyTo, just returns nil
}

func TestReplyTo_NilRpcPublisher(t *testing.T) {
	h := NewHandlers(new(MockRepository), nil, nil, &NoOpTransactionManager{})
	err := h.replyTo(context.Background(), amqp.Delivery{ReplyTo: "queue"}, "response")
	assert.NoError(t, err) // warns but no error
}

func TestReplyTo_WithCorrelationID(t *testing.T) {
	pub := new(mockTestPublisher)
	h := NewHandlers(new(MockRepository), nil, pub, &NoOpTransactionManager{})

	pub.On("Publish", testifymock.Anything, "", "reply-q", testifymock.Anything).Return(nil)

	err := h.replyTo(context.Background(), amqp.Delivery{ReplyTo: "reply-q", CorrelationId: "corr-1"}, "response")
	assert.NoError(t, err)
}

func TestExtractUser_AllHeaders(t *testing.T) {
	h := &Handlers{}

	d := amqp.Delivery{
		Headers: amqp.Table{
			"user":          "test-user",
			"Authorization": "Bearer token",
		},
	}
	ctx := h.extractUser(context.Background(), d)
	assert.Equal(t, "test-user", ctx.Value(domain.UserContextKey))
	assert.Equal(t, "Bearer token", ctx.Value(domain.AuthContextKey))
}

func TestExtractUser_NoHeaders(t *testing.T) {
	h := &Handlers{}
	ctx := h.extractUser(context.Background(), amqp.Delivery{})
	assert.Nil(t, ctx.Value(domain.UserContextKey))
	assert.Nil(t, ctx.Value(domain.AuthContextKey))
}

// ========== Listener GetHandler expanded tests ==========

func TestListener_GetHandler_FinalizeDeletion(t *testing.T) {
	l, _ := NewListener(nil)
	h := &Handlers{}
	handler, ok := l.GetHandler(CmdPartyFinalizeDeletion, h)
	assert.True(t, ok)
	assert.NotNil(t, handler)
}

func TestListener_GetHandler_CancelDeletion(t *testing.T) {
	l, _ := NewListener(nil)
	h := &Handlers{}
	handler, ok := l.GetHandler(CmdPartyCancelDeletion, h)
	assert.True(t, ok)
	assert.NotNil(t, handler)
}

// ========== Create with full sub-resources ==========

func TestHandleCreateParty_IndividualWithAllSubResources(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("CreateIndividual", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, testifymock.Anything, testifymock.Anything).Return(nil)

	payload := map[string]interface{}{
		"@type":      "Individual",
		"id":         "sub-res-1",
		"givenName":  "Sub",
		"familyName": "Res",
		"href":       "http://example.com/sub-res-1",
		"contactMediums": []map[string]interface{}{
			{"id": "", "mediumType": "email", "value": "a@b.com", "preferred": true},
			{"id": "cm-2", "mediumType": "phone", "value": "555-1234"},
		},
		"identifications": []map[string]interface{}{
			{"identificationType": "passport", "identificationId": "AB123", "issuingAuthority": "GOV", "issuingDate": "2024-01-01T00:00:00Z"},
			{"id": "id-2", "identificationType": "id-card", "identificationId": "CD456"},
		},
		"relatedParties": []map[string]interface{}{
			{"id": "rp-1", "relatedPartyId": "rel-1", "role": "owner", "permissions": []string{"read", "write"}},
		},
		"characteristics": []map[string]interface{}{
			{"name": "lang", "value": "en", "valueType": "string"},
		},
		"externalReferences": []map[string]interface{}{
			{"externalSystemId": "CRM", "externalReferenceId": "CRM-001"},
		},
		"taxExemptions": []map[string]interface{}{
			{"certificateNumber": "TX001", "issuingJurisdiction": "EU", "validForStart": "2024-01-01T00:00:00Z", "validForEnd": "2025-01-01T00:00:00Z"},
		},
		"attachments": []map[string]interface{}{
			{"name": "photo.png", "mimeType": "image/png", "url": "https://s3.example.com/photo.png"},
		},
	}
	body, _ := json.Marshal(payload)
	err := h.HandleCreateParty(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err)
}

func TestHandleCreateParty_OrgWithAllSubResources(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("CreateOrganization", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, testifymock.Anything, testifymock.Anything).Return(nil)

	payload := map[string]interface{}{
		"@type":           "Organization",
		"id":              "sub-res-2",
		"tradingName":     "SubCorp",
		"isLegalEntity":   true,
		"href":            "http://example.com/sub-res-2",
		"contactMediums":  []map[string]interface{}{{"mediumType": "email", "value": "org@test.com"}},
		"identifications": []map[string]interface{}{{"identificationType": "reg", "identificationId": "REG001"}},
		"characteristics": []map[string]interface{}{{"id": "ch-1", "name": "region", "value": "EU"}},
		"taxExemptions":   []map[string]interface{}{{"certificateNumber": "TX2"}},
		"attachments":     []map[string]interface{}{{"name": "doc.pdf", "refType": "S3", "refId": "s3://doc.pdf"}},
	}
	body, _ := json.Marshal(payload)
	err := h.HandleCreateParty(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err)
}

// ========== Update with status change & no status change ==========

func TestHandleUpdateParty_IndividualSameStatus(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "upd-same").Return(&domain.Party{
		ID: "upd-same", Type: domain.PartyTypeIndividual, Status: "Active",
	}, nil)
	repo.On("UpdateIndividual", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, EvtPartyUpdated, testifymock.Anything).Return(nil)

	payload := map[string]interface{}{
		"@type": "Individual", "id": "upd-same", "givenName": "Same", "status": "Active",
	}
	body, _ := json.Marshal(payload)
	err := h.HandleUpdateParty(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err)
	// Should NOT publish state change since status is the same
	pub.AssertNotCalled(t, "Publish", ctx, EventExchange, EvtPartyStateChange, testifymock.Anything)
}

func TestHandleUpdateParty_OrgStatusChange(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "upd-org-sc").Return(&domain.Party{
		ID: "upd-org-sc", Type: domain.PartyTypeOrganization, Status: "Initialized",
	}, nil)
	repo.On("UpdateOrganization", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, testifymock.Anything, testifymock.Anything).Return(nil)

	payload := map[string]interface{}{
		"@type": "Organization", "id": "upd-org-sc", "tradingName": "NewCorp", "status": "Active",
	}
	body, _ := json.Marshal(payload)
	err := h.HandleUpdateParty(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err)
}

// ========== Patch with status change ==========

func TestHandlePatchParty_IndividualStatusChange(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "ptch-sc").Return(&domain.Party{ID: "ptch-sc", Type: domain.PartyTypeIndividual}, nil)
	repo.On("GetIndividual", ctx, "ptch-sc").Return(&domain.Individual{
		Party: domain.Party{ID: "ptch-sc", Status: "Initialized"},
	}, nil)
	repo.On("UpdateIndividual", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, testifymock.Anything, testifymock.Anything).Return(nil)

	newStatus := "Active"
	newName := "PatchedName"
	payload := PatchPartyPayload{ID: "ptch-sc", Status: &newStatus, GivenName: &newName}
	body, _ := json.Marshal(payload)
	err := h.HandlePatchParty(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err)
}

func TestHandlePatchParty_OrgStatusChange(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "ptch-org-sc").Return(&domain.Party{ID: "ptch-org-sc", Type: domain.PartyTypeOrganization}, nil)
	repo.On("GetOrganization", ctx, "ptch-org-sc").Return(&domain.Organization{
		Party: domain.Party{ID: "ptch-org-sc", Status: "Active"},
	}, nil)
	repo.On("UpdateOrganization", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, testifymock.Anything, testifymock.Anything).Return(nil)

	newStatus := "Suspended"
	payload := PatchPartyPayload{ID: "ptch-org-sc", Status: &newStatus}
	body, _ := json.Marshal(payload)
	err := h.HandlePatchParty(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err)
}

func TestHandlePatchParty_IndividualFamilyName(t *testing.T) {
	h, repo, pub := setupHandlerWithMocks()
	ctx := context.Background()

	repo.On("GetParty", ctx, "ptch-fn").Return(&domain.Party{ID: "ptch-fn", Type: domain.PartyTypeIndividual}, nil)
	repo.On("GetIndividual", ctx, "ptch-fn").Return(&domain.Individual{
		Party: domain.Party{ID: "ptch-fn", Status: "Active"}, FamilyName: "Old",
	}, nil)
	repo.On("UpdateIndividual", ctx, testifymock.Anything).Return(nil)
	pub.On("Publish", ctx, EventExchange, testifymock.Anything, testifymock.Anything).Return(nil)

	newFam := "NewFamily"
	payload := PatchPartyPayload{ID: "ptch-fn", FamilyName: &newFam}
	body, _ := json.Marshal(payload)
	err := h.HandlePatchParty(ctx, amqp.Delivery{Body: body})
	assert.NoError(t, err)
}
