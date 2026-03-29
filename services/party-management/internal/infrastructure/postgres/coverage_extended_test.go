package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
	"tmf/services/party-management/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========== TransactionManager Tests ==========

func TestNewTransactionManager(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	tm := NewTransactionManager(db)
	require.NotNil(t, tm)
}

func TestTransactionManager_Run_Success(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	tm := NewTransactionManager(db)
	repo := NewPartyRepository(db)

	err := tm.Run(context.Background(), func(ctx context.Context) error {
		ind := &domain.Individual{
			Party: domain.Party{
				ID:     "tm-test-1",
				Type:   domain.PartyTypeIndividual,
				Status: "Active",
			},
			GivenName:  "TxTest",
			FamilyName: "User",
		}
		return repo.CreateIndividual(ctx, ind)
	})
	assert.NoError(t, err)

	// Verify the party was created
	party, err := repo.GetParty(context.Background(), "tm-test-1")
	assert.NoError(t, err)
	assert.Equal(t, "tm-test-1", party.ID)
}

func TestTransactionManager_Run_Rollback(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	tm := NewTransactionManager(db)
	repo := NewPartyRepository(db)

	err := tm.Run(context.Background(), func(ctx context.Context) error {
		ind := &domain.Individual{
			Party: domain.Party{
				ID:     "tm-rollback-1",
				Type:   domain.PartyTypeIndividual,
				Status: "Active",
			},
			GivenName: "Rollback",
		}
		if err := repo.CreateIndividual(ctx, ind); err != nil {
			return err
		}
		return errors.New("intentional error")
	})
	assert.Error(t, err)

	// Verify rollback - party should not exist
	_, err = repo.GetParty(context.Background(), "tm-rollback-1")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestGetTx_WithoutTransaction(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	tx := GetTx(context.Background(), db)
	assert.NotNil(t, tx)
}

func TestGetTx_WithTransaction(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	tm := NewTransactionManager(db)
	_ = tm.Run(context.Background(), func(ctx context.Context) error {
		tx := GetTx(ctx, db)
		assert.NotNil(t, tx)
		return nil
	})
}

// ========== PartyTable Tests ==========

func TestPartyTable_TableName(t *testing.T) {
	pt := PartyTable{}
	assert.Equal(t, "parties", pt.TableName())
}

// ========== PartyRepository.GetParty Tests ==========

func TestGetParty_NotFound(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewPartyRepository(db)
	_, err := repo.GetParty(context.Background(), "non-existent-id")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestGetParty_WithPreloads(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewPartyRepository(db)

	erID := uuid.New().String()
	teID := uuid.New().String()

	// Create party with sub-resources
	ind := &domain.Individual{
		Party: domain.Party{
			ID:     "get-party-preload",
			Type:   domain.PartyTypeIndividual,
			Status: "Active",
			ContactMediums: []domain.ContactMedium{
				{ID: "cm-gp-1", PartyID: "get-party-preload", MediumType: "email", Value: "test@test.com"},
			},
			Identifications: []domain.Identification{
				{ID: "id-gp-1", PartyID: "get-party-preload", IdentificationType: "passport", IdentificationID: "AB123"},
			},
			ExternalReferences: []domain.ExternalReference{
				{ID: erID, PartyID: "get-party-preload", ExternalSystemID: "CRM", ExternalReferenceID: "CRM-001"},
			},
			Characteristics: []domain.PartyCharacteristic{
				{ID: "ch-gp-1", PartyID: "get-party-preload", Name: "lang", Value: "en"},
			},
			TaxExemptions: []domain.TaxExemption{
				{ID: teID, PartyID: "get-party-preload", CertificateNumber: "TX001"},
			},
		},
		GivenName:  "Get",
		FamilyName: "Party",
	}
	err := repo.CreateIndividual(context.Background(), ind)
	require.NoError(t, err)

	party, err := repo.GetParty(context.Background(), "get-party-preload")
	assert.NoError(t, err)
	assert.Equal(t, "get-party-preload", party.ID)
	assert.Len(t, party.ContactMediums, 1)
	assert.Len(t, party.Identifications, 1)
	assert.Len(t, party.ExternalReferences, 1)
	assert.Len(t, party.Characteristics, 1)
	assert.Len(t, party.TaxExemptions, 1)
}

// ========== DeleteParty Tests ==========

func TestDeleteParty_Individual_Extended(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewPartyRepository(db)

	ind := &domain.Individual{
		Party: domain.Party{
			ID:     "del-ind-1",
			Type:   domain.PartyTypeIndividual,
			Status: "Active",
		},
		GivenName: "Delete",
	}
	require.NoError(t, repo.CreateIndividual(context.Background(), ind))

	err := repo.DeleteParty(context.Background(), "del-ind-1")
	assert.NoError(t, err)

	_, err = repo.GetParty(context.Background(), "del-ind-1")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestDeleteParty_Organization_Extended(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewPartyRepository(db)

	org := &domain.Organization{
		Party: domain.Party{
			ID:     "del-org-1",
			Type:   domain.PartyTypeOrganization,
			Status: "Active",
		},
		TradingName: "DeleteCorp",
	}
	require.NoError(t, repo.CreateOrganization(context.Background(), org))

	err := repo.DeleteParty(context.Background(), "del-org-1")
	assert.NoError(t, err)

	_, err = repo.GetParty(context.Background(), "del-org-1")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestDeleteParty_NotFound(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewPartyRepository(db)
	err := repo.DeleteParty(context.Background(), "non-existent")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

// ========== UpdateIndividual/Organization Tests ==========

func TestUpdateIndividual_NotFound(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewPartyRepository(db)
	err := repo.UpdateIndividual(context.Background(), &domain.Individual{
		Party: domain.Party{ID: "upd-notfound", Status: "Active"},
	})
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestUpdateOrganization_NotFound(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewPartyRepository(db)
	err := repo.UpdateOrganization(context.Background(), &domain.Organization{
		Party: domain.Party{ID: "upd-org-notfound", Status: "Active"},
	})
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestUpdateIndividual_WithSubResources(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewPartyRepository(db)

	ind := &domain.Individual{
		Party: domain.Party{
			ID:     "upd-ind-sub",
			Type:   domain.PartyTypeIndividual,
			Status: "Active",
		},
		GivenName:  "Original",
		FamilyName: "Name",
	}
	require.NoError(t, repo.CreateIndividual(context.Background(), ind))

	// Update with new sub-resources
	ind.GivenName = "Updated"
	ind.Status = "Modified"
	ind.ContactMediums = []domain.ContactMedium{
		{ID: "cm-upd-1", PartyID: "upd-ind-sub", MediumType: "phone", Value: "555-1234"},
	}
	err := repo.UpdateIndividual(context.Background(), ind)
	assert.NoError(t, err)

	// Verify
	updated, err := repo.GetIndividual(context.Background(), "upd-ind-sub")
	assert.NoError(t, err)
	assert.Equal(t, "Updated", updated.GivenName)
	assert.Len(t, updated.ContactMediums, 1)
}

func TestUpdateOrganization_WithSubResources(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewPartyRepository(db)

	org := &domain.Organization{
		Party: domain.Party{
			ID:     "upd-org-sub",
			Type:   domain.PartyTypeOrganization,
			Status: "Active",
		},
		TradingName: "OriginalCorp",
	}
	require.NoError(t, repo.CreateOrganization(context.Background(), org))

	org.TradingName = "UpdatedCorp"
	org.Characteristics = []domain.PartyCharacteristic{
		{ID: "ch-upd-1", PartyID: "upd-org-sub", Name: "region", Value: "EU"},
	}
	err := repo.UpdateOrganization(context.Background(), org)
	assert.NoError(t, err)

	updated, err := repo.GetOrganization(context.Background(), "upd-org-sub")
	assert.NoError(t, err)
	assert.Equal(t, "UpdatedCorp", updated.TradingName)
}

// ========== SearchParties Extended Tests ==========

func TestSearchParties_ByID(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewPartyRepository(db)

	ind := &domain.Individual{
		Party:     domain.Party{ID: "search-id-1", Type: domain.PartyTypeIndividual, Status: "Active"},
		GivenName: "SearchableID",
	}
	require.NoError(t, repo.CreateIndividual(context.Background(), ind))

	results, err := repo.SearchParties(context.Background(), map[string]interface{}{"id": "search-id-1"})
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "search-id-1", results[0].ID)
}

func TestSearchParties_ByExternalReference_Extended(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewPartyRepository(db)
	erID := uuid.New().String()

	ind := &domain.Individual{
		Party: domain.Party{
			ID:     "search-ext-1",
			Type:   domain.PartyTypeIndividual,
			Status: "Active",
			ExternalReferences: []domain.ExternalReference{
				{ID: erID, PartyID: "search-ext-1", ExternalSystemID: "ERP", ExternalReferenceID: "ERP-X1"},
			},
		},
		GivenName: "ExtRef",
	}
	require.NoError(t, repo.CreateIndividual(context.Background(), ind))

	results, err := repo.SearchParties(context.Background(), map[string]interface{}{"externalReference": "ERP-X1"})
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestSearchParties_ByGivenName_Extended(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewPartyRepository(db)

	ind := &domain.Individual{
		Party:     domain.Party{ID: "search-gn-1", Type: domain.PartyTypeIndividual, Status: "Active"},
		GivenName: "UniqueGivenName",
	}
	require.NoError(t, repo.CreateIndividual(context.Background(), ind))

	results, err := repo.SearchParties(context.Background(), map[string]interface{}{"given_name": "UniqueGivenName"})
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestSearchParties_ByFamilyName(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewPartyRepository(db)

	ind := &domain.Individual{
		Party:      domain.Party{ID: "search-fn-1", Type: domain.PartyTypeIndividual, Status: "Active"},
		GivenName:  "First",
		FamilyName: "UniqueFamilyName",
	}
	require.NoError(t, repo.CreateIndividual(context.Background(), ind))

	results, err := repo.SearchParties(context.Background(), map[string]interface{}{"family_name": "UniqueFamilyName"})
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestSearchParties_ByIsLegalEntity(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewPartyRepository(db)

	org := &domain.Organization{
		Party:         domain.Party{ID: "search-le-1", Type: domain.PartyTypeOrganization, Status: "Active"},
		TradingName:   "LegalCorp",
		IsLegalEntity: true,
	}
	require.NoError(t, repo.CreateOrganization(context.Background(), org))

	results, err := repo.SearchParties(context.Background(), map[string]interface{}{
		"is_legal_entity": true,
		"trading_name":    "LegalCorp",
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1)
}

func TestSearchParties_ByNameCriteria(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewPartyRepository(db)

	ind := &domain.Individual{
		Party:      domain.Party{ID: "search-name-1", Type: domain.PartyTypeIndividual, Status: "Active"},
		GivenName:  "SearchByName",
		FamilyName: "Test",
	}
	require.NoError(t, repo.CreateIndividual(context.Background(), ind))

	results, err := repo.SearchParties(context.Background(), map[string]interface{}{"name": "SearchByName"})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1)
}

func TestSearchParties_GenericSearch(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewPartyRepository(db)

	ind := &domain.Individual{
		Party:     domain.Party{ID: "search-gen-1", Type: domain.PartyTypeIndividual, Status: "Active"},
		GivenName: "GenericSearchUnique",
	}
	require.NoError(t, repo.CreateIndividual(context.Background(), ind))

	results, err := repo.SearchParties(context.Background(), map[string]interface{}{"search": "GenericSearchUnique"})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1)
}

// ========== OutboxRepository Tests ==========

func TestOutboxRepository_SaveAndFetchPending(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewOutboxRepository(db)
	ctx := context.Background()

	event := &domain.OutboxEvent{
		ID:         uuid.New().String(),
		RoutingKey: "evt.party.created",
		Payload:    []byte(`{"id":"test"}`),
		Headers:    []byte(`{"user":"test-user"}`),
		Status:     domain.StatusPending,
	}
	err := repo.Save(ctx, event)
	assert.NoError(t, err)

	events, err := repo.FetchPending(ctx, 10)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(events), 1)
}

func TestOutboxRepository_MarkAsProcessed(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewOutboxRepository(db)
	ctx := context.Background()

	outboxID := uuid.New().String()
	event := &domain.OutboxEvent{
		ID:         outboxID,
		RoutingKey: "evt.party.updated",
		Payload:    []byte(`{}`),
		Headers:    []byte(`{}`),
		Status:     domain.StatusPending,
	}
	require.NoError(t, repo.Save(ctx, event))

	err := repo.MarkAsProcessed(ctx, outboxID)
	assert.NoError(t, err)

	// Verify it's no longer pending
	events, err := repo.FetchPending(ctx, 10)
	assert.NoError(t, err)
	for _, e := range events {
		assert.NotEqual(t, outboxID, e.ID)
	}
}

// ========== OutboxPublisher Tests ==========

func TestOutboxPublisher_Publish(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	outboxRepo := NewOutboxRepository(db)
	publisher := NewOutboxPublisher(outboxRepo)
	ctx := context.Background()
	ctx = context.WithValue(ctx, domain.UserContextKey, "test-user")
	ctx = context.WithValue(ctx, domain.AuthContextKey, "Bearer token")

	err := publisher.Publish(ctx, "tmf.events", "evt.party.created", map[string]string{"id": "pub-test-1"})
	assert.NoError(t, err)

	events, err := outboxRepo.FetchPending(ctx, 10)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(events), 1)
}

func TestOutboxPublisher_MarshalError(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	outboxRepo := NewOutboxRepository(db)
	publisher := NewOutboxPublisher(outboxRepo)
	ctx := context.Background()

	// Unmarshalable payload
	err := publisher.Publish(ctx, "tmf.events", "test.key", make(chan int))
	assert.Error(t, err)
}

// ========== OutboxWorker Tests ==========

type mockRabbitPublisher struct {
	publishFn func(ctx context.Context, exchange, routingKey string, msg interface{}) error
}

func (m *mockRabbitPublisher) Publish(ctx context.Context, exchange, routingKey string, msg interface{}) error {
	return m.publishFn(ctx, exchange, routingKey, msg)
}

func (m *mockRabbitPublisher) Close() error { return nil }

func (m *mockRabbitPublisher) DeclareTopicExchange(name string, durable, autoDelete, internal, noWait bool) error {
	return nil
}

func (m *mockRabbitPublisher) PublishToQueue(ctx context.Context, queueName string, correlationID string, body interface{}) error {
	return nil
}

func TestNewOutboxWorker(t *testing.T) {
	w := NewOutboxWorker(nil, nil)
	assert.NotNil(t, w)
	assert.Equal(t, 50, w.batchSize)
	assert.Equal(t, 500*time.Millisecond, w.pollInterval)
}

func TestOutboxWorker_Start_ContextCancel(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	outboxRepo := NewOutboxRepository(db)
	pub := &mockRabbitPublisher{
		publishFn: func(ctx context.Context, exchange, routingKey string, msg interface{}) error {
			return nil
		},
	}

	worker := NewOutboxWorker(outboxRepo, pub)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		worker.Start(ctx)
		close(done)
	}()

	// Let it run one tick
	time.Sleep(600 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Good - worker stopped
	case <-time.After(5 * time.Second):
		t.Fatal("Worker did not stop after context cancellation")
	}
}

func TestOutboxWorker_ProcessEvent_WithHeaders(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	outboxRepo := NewOutboxRepository(db)
	ctx := context.Background()

	var capturedCtx context.Context
	pub := &mockRabbitPublisher{
		publishFn: func(ctx context.Context, exchange, routingKey string, msg interface{}) error {
			capturedCtx = ctx
			return nil
		},
	}

	// Create an event with headers
	headers := map[string]string{"user": "worker-user", "Authorization": "Bearer abc"}
	headerBytes, _ := json.Marshal(headers)
	workerID := uuid.New().String()
	event := &domain.OutboxEvent{
		ID:         workerID,
		RoutingKey: "evt.test",
		Payload:    []byte(`{"data":"test"}`),
		Headers:    headerBytes,
		Status:     domain.StatusPending,
	}
	require.NoError(t, outboxRepo.Save(ctx, event))

	worker := NewOutboxWorker(outboxRepo, pub)
	worker.processBatch(ctx)

	// Verify headers were propagated
	assert.NotNil(t, capturedCtx)
}

func TestOutboxWorker_ProcessEvent_PublishError(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	outboxRepo := NewOutboxRepository(db)
	ctx := context.Background()

	pub := &mockRabbitPublisher{
		publishFn: func(ctx context.Context, exchange, routingKey string, msg interface{}) error {
			return errors.New("publish failed")
		},
	}

	workerErrID := uuid.New().String()
	event := &domain.OutboxEvent{
		ID:         workerErrID,
		RoutingKey: "evt.test",
		Payload:    []byte(`{}`),
		Headers:    []byte(`{}`),
		Status:     domain.StatusPending,
	}
	require.NoError(t, outboxRepo.Save(ctx, event))

	worker := NewOutboxWorker(outboxRepo, pub)
	worker.processBatch(ctx) // Should not panic, just log error

	// Event should still be pending
	events, err := outboxRepo.FetchPending(ctx, 10)
	assert.NoError(t, err)
	found := false
	for _, e := range events {
		if e.ID == workerErrID {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestOutboxWorker_ProcessBatch_NoPendingEvents(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	db.Exec("DELETE FROM outbox_events")

	outboxRepo := NewOutboxRepository(db)
	pub := &mockRabbitPublisher{
		publishFn: func(ctx context.Context, exchange, routingKey string, msg interface{}) error {
			t.Fatal("Should not be called with no events")
			return nil
		},
	}

	worker := NewOutboxWorker(outboxRepo, pub)
	// Should not panic with empty batch
	worker.processBatch(context.Background())
}

func TestOutboxWorker_ProcessBatch_FetchError(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	outboxRepo := NewOutboxRepository(db)
	worker := NewOutboxWorker(outboxRepo, nil)

	// A canceled context will cause outboxRepo.FetchPending to return an error immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should safely log and return without panicking
	worker.processBatch(ctx)
}

func TestOutboxWorker_ProcessEvent_InvalidHeaders(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	outboxRepo := NewOutboxRepository(db)
	ctx := context.Background()

	pub := &mockRabbitPublisher{
		publishFn: func(ctx context.Context, exchange, routingKey string, msg interface{}) error {
			return nil
		},
	}

	// Use valid JSON that won't unmarshal into map[string]string (array not object)
	workerBadHdrID := uuid.New().String()
	event := &domain.OutboxEvent{
		ID:         workerBadHdrID,
		RoutingKey: "evt.test",
		Payload:    []byte(`{}`),
		Headers:    []byte(`[1,2,3]`),
		Status:     domain.StatusPending,
	}
	require.NoError(t, outboxRepo.Save(ctx, event))

	worker := NewOutboxWorker(outboxRepo, pub)
	worker.processBatch(ctx) // Should not panic, just log warning
}

// ========== loadAttachmentContents Tests ==========

func TestLoadAttachmentContents_InternalAttachment(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewPartyRepository(db)

	// Create a party with internal attachment content
	attID := uuid.New().String()
	attContentID := uuid.New().String()
	ind := &domain.Individual{
		Party: domain.Party{
			ID:     "att-int-1",
			Type:   domain.PartyTypeIndividual,
			Status: "Active",
			Attachments: []domain.Attachment{
				{
					ID:       attID,
					OwnerID:  "att-int-1",
					Name:     "test.txt",
					MimeType: "text/plain",
					RefType:  "Internal",
					RefID:    attContentID,
				},
			},
		},
		GivenName: "AttachTest",
	}
	require.NoError(t, repo.CreateIndividual(context.Background(), ind))

	// Manually insert content
	content := domain.AttachmentContent{
		ID:   attContentID,
		Data: []byte("hello world"),
	}
	require.NoError(t, db.Create(&content).Error)

	// Fetch and verify content is loaded
	fetched, err := repo.GetIndividual(context.Background(), "att-int-1")
	assert.NoError(t, err)
	assert.Len(t, fetched.Attachments, 1)
	assert.Equal(t, []byte("hello world"), fetched.Attachments[0].ContentData)
}

// ========== withUser Tests ==========

func TestWithUser_WithUserContext(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewPartyRepository(db)
	ctx := context.WithValue(context.Background(), domain.UserContextKey, "audit-user")

	ind := &domain.Individual{
		Party: domain.Party{
			ID:     "user-ctx-1",
			Type:   domain.PartyTypeIndividual,
			Status: "Active",
		},
		GivenName: "WithUser",
	}
	err := repo.CreateIndividual(ctx, ind)
	assert.NoError(t, err)
}

func TestWithUser_WithExistingTransaction(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewPartyRepository(db)
	tm := NewTransactionManager(db)

	err := tm.Run(context.Background(), func(ctx context.Context) error {
		ind := &domain.Individual{
			Party: domain.Party{
				ID:     "user-tx-1",
				Type:   domain.PartyTypeIndividual,
				Status: "Active",
			},
			GivenName: "WithTx",
		}
		return repo.CreateIndividual(ctx, ind)
	})
	assert.NoError(t, err)
}
