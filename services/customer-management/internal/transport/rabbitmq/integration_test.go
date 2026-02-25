package rabbitmq

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"tmf/pkg/rabbitmq"
	"tmf/services/customer-management/internal/domain"
	"tmf/services/customer-management/internal/infrastructure/postgres"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	pgContainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	rabbitContainer "github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"github.com/testcontainers/testcontainers-go/wait"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	sharedDB             *gorm.DB
	sharedRepo           *postgres.CustomerRepository
	sharedConn           *amqp.Connection
	sharedPublisher      rabbitmq.Publisher
	sharedTM             *postgres.TransactionManager
	sharedEventPublisher *postgres.OutboxPublisher
	pgInstance           testcontainers.Container
	rabbitInstance       testcontainers.Container
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// 1. Start Postgres container
	pg, err := pgContainer.Run(ctx,
		"postgres:15",
		pgContainer.WithDatabase("testdb"),
		pgContainer.WithUsername("postgres"),
		pgContainer.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		log.Fatalf("failed to start postgres: %v", err)
	}
	pgInstance = pg

	pgConnStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get postgres connection string: %v", err)
	}

	// Run migrations
	_, filename, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(filename), "..", "..", "infrastructure", "postgres", "migrations")

	mig, err := migrate.New("file://"+migrationsPath, pgConnStr)
	if err != nil {
		log.Fatalf("failed to create migrate: %v", err)
	}
	if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("failed to run migrations: %v", err)
	}

	sharedDB, err = gorm.Open(gormPostgres.Open(pgConnStr), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	sharedRepo = postgres.NewCustomerRepository(sharedDB)

	// 2. Start RabbitMQ container
	rabbit, err := rabbitContainer.Run(ctx,
		"rabbitmq:3-management",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Server startup complete").
				WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		log.Fatalf("failed to start rabbitmq: %v", err)
	}
	rabbitInstance = rabbit

	rabbitURL, err := rabbit.AmqpURL(ctx)
	if err != nil {
		log.Fatalf("failed to get rabbitmq URL: %v", err)
	}

	sharedConn, err = amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}

	// Declare exchange to avoid channel closure errors
	ch, err := sharedConn.Channel()
	if err != nil {
		log.Fatalf("failed to open channel to declare exchange: %v", err)
	}
	err = ch.ExchangeDeclare(
		"tmf.events", // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		log.Fatalf("failed to declare exchange: %v", err)
	}
	_ = ch.Close()

	sharedPublisher, err = rabbitmq.NewPublisherWithConnection(sharedConn)
	if err != nil {
		log.Fatalf("failed to create publisher: %v", err)
	}
	// Exchange is already declared above manually (lines 112-123) which is fine,
	// OR we can use DeclareTopicExchange helper.
	// Since tests declared it manually, we can keep it or replace it.
	// Let's replace the manual declaration with helper later if needed, but for now just fix the NewPublisherWithConnection call.
	// Actually, the manual declaration code (lines 112-123) is redundant now if we use helper, but it's safe.
	// But lines 126 uses "tmf.events" in NewPublisherWithConnection, which logic is removed.
	// The manual declaration uses "tmf.events".
	// The implementation of NewPublisherWithConnection NO LONGER declares exchange.
	// So the MANUAL declaration (lines 108-124) is now NECESSARY if we don't call DeclareTopicExchange.
	// Wait, existing code lines 108-124 ALREADY declare "tmf.events". So we just update the constructor.

	sharedTM = postgres.NewTransactionManager(sharedDB)
	outboxRepo := postgres.NewOutboxRepository(sharedDB)
	sharedEventPublisher = postgres.NewOutboxPublisher(outboxRepo)
	worker := postgres.NewOutboxWorker(outboxRepo, sharedPublisher, slog.Default())
	go worker.Start(ctx)
	defer worker.Stop()

	// Run tests
	code := m.Run()

	// Cleanup
	_ = sharedConn.Close()
	_ = pgInstance.Terminate(ctx)
	_ = rabbitInstance.Terminate(ctx)

	os.Exit(code)
}

// 1. Onboard Customer Use Case
func TestUseCase_OnboardNewCustomer(t *testing.T) {
	ctx := context.Background()
	handlers := NewHandlers(sharedRepo, sharedPublisher, sharedTM, sharedEventPublisher)

	payload := OnboardCustomerPayload{
		ID:        "cust-onboard-1",
		Name:      "Full Profile Customer",
		PartyID:   "party-onboard-1",
		PartyType: "Individual",
		Accounts: []CustomerAccountDTO{
			{ID: "acc-1", Name: "Savings", AccountStatus: "active", AccountType: "savings"},
		},
		CreditProfiles: []CreditProfileDTO{
			{ID: "cp-1", CreditScore: 750, CreditRiskScore: 10},
		},
		ContactMediums: []ContactMediumDTO{
			{ID: "cm-1", MediumType: "email", Value: "test@example.com", Preferred: true},
		},
		Characteristics: []CharacteristicDTO{
			{ID: "char-1", Name: "Segment", Value: "VIP"},
		},

		PrivacyConsents: []PrivacyConsentDTO{
			{ID: "pc-1", ConsentType: "Marketing", Status: "given", ValidForStart: time.Now().Format(time.RFC3339)},
		},
	}
	body, _ := json.Marshal(payload)

	err := handlers.HandleOnboardCustomer(ctx, amqp.Delivery{
		Body:     body,
		Exchange: "tmf.events",
	})
	require.NoError(t, err)

	// Verify DB state
	saved, err := sharedRepo.GetCustomer(ctx, "cust-onboard-1")
	require.NoError(t, err)
	assert.Equal(t, "Full Profile Customer", saved.Name)
	assert.Equal(t, domain.CustomerStatusActive, saved.Status)

	// Verify sub-resources
	require.Len(t, saved.Accounts, 1)
	assert.Equal(t, "Savings", saved.Accounts[0].Name)

	require.Len(t, saved.CreditProfiles, 1)
	assert.Equal(t, 750, saved.CreditProfiles[0].CreditScore)

	require.Len(t, saved.ContactMediums, 1)
	assert.Equal(t, "test@example.com", saved.ContactMediums[0].Value)

	require.Len(t, saved.Characteristics, 1)
	assert.Equal(t, "VIP", saved.Characteristics[0].Value)

	require.Len(t, saved.PrivacyConsents, 1)
	assert.Equal(t, "given", saved.PrivacyConsents[0].Status)
}

func TestUseCase_Onboard_AutoGenerateIDs(t *testing.T) {
	ctx := context.Background()
	handlers := NewHandlers(sharedRepo, sharedPublisher, sharedTM, sharedEventPublisher)

	// Payload with NO IDs for sub-resources or main customer
	// Note: We provide main ID in payload struct usually, so we'll leave it empty to test auto-gen if handler supports it,
	// or provide one to easily fetch it. The handler: "if payload.ID == "" { payload.ID = uuid.New().String() }"
	// Let's rely on that but we need a way to find it.
	// Actually, easier to test sub-resource ID generation with a known main ID.
	custID := "cust-autogen-1"

	payload := OnboardCustomerPayload{
		ID:        custID,
		Name:      "Auto Gen IDs Test",
		PartyID:   "p-autogen-1",
		PartyType: "Individual",
		Accounts: []CustomerAccountDTO{
			{Name: "No ID Account", AccountStatus: "Active"},
		},
	}
	body, _ := json.Marshal(payload)

	err := handlers.HandleOnboardCustomer(ctx, amqp.Delivery{
		Body:     body,
		Exchange: "tmf.events",
	})
	require.NoError(t, err)

	saved, err := sharedRepo.GetCustomer(ctx, custID)
	require.NoError(t, err)

	require.Len(t, saved.Accounts, 1)
	assert.NotEmpty(t, saved.Accounts[0].ID)

}

func TestUseCase_Onboard_NewTMFFeatures(t *testing.T) {
	ctx := context.Background()
	handlers := NewHandlers(sharedRepo, sharedPublisher, sharedTM, sharedEventPublisher)

	payload := OnboardCustomerPayload{
		ID:        "cust-tmf-1",
		Name:      "TMF Features Customer",
		PartyID:   "party-tmf-1",
		PartyType: "Organization",
		Accounts: []CustomerAccountDTO{
			{Name: "Detailed Account", AccountStatus: "active", BillFormat: "Email", BillingCycle: "Weekly"},
		},
		PrivacyConsents: []PrivacyConsentDTO{
			{ID: "pc-1", ConsentType: "Marketing", Status: "given", ValidForStart: time.Now().Format(time.RFC3339)},
		},
		RelatedParties: []RelatedPartyDTO{
			{RelatedPartyID: "rp-1", Role: "ParentCompany", Name: "Big Corp", ValidForStart: time.Now().Format(time.RFC3339), ValidForEnd: time.Now().Add(24 * time.Hour).Format(time.RFC3339)},
		},
		PaymentMethods: []PaymentMethodDTO{
			{Type: "CreditCard", Token: "tok_123", IsDefault: true, Details: "{}", ValidForStart: time.Now().Format(time.RFC3339), ValidForEnd: time.Now().Add(24 * time.Hour).Format(time.RFC3339)},
		},
		MarketSegments: []MarketSegmentDTO{
			{Name: "Enterprise", Category: "B2B"},
		},
	}
	body, _ := json.Marshal(payload)

	err := handlers.HandleOnboardCustomer(ctx, amqp.Delivery{
		Body:     body,
		Exchange: "tmf.events",
	})
	require.NoError(t, err)

	// Verify DB state
	saved, err := sharedRepo.GetCustomer(ctx, "cust-tmf-1")
	require.NoError(t, err)

	// Verify Account Logic
	require.Len(t, saved.Accounts, 1)
	assert.Equal(t, "Email", saved.Accounts[0].BillFormat)
	assert.Equal(t, "Weekly", saved.Accounts[0].BillingCycle)

	// Verify Related Parties
	require.Len(t, saved.RelatedParties, 1)
	assert.Equal(t, "Big Corp", saved.RelatedParties[0].Name)
	assert.Equal(t, "ParentCompany", saved.RelatedParties[0].Role)

	// Verify Payment Methods
	require.Len(t, saved.PaymentMethods, 1)
	assert.Equal(t, "CreditCard", saved.PaymentMethods[0].Type)
	assert.Equal(t, "tok_123", saved.PaymentMethods[0].Token)

	// Verify Market Segments
	require.Len(t, saved.MarketSegments, 1)
	assert.Equal(t, "Enterprise", saved.MarketSegments[0].Name)

}

// 2. Update Customer Use Case
func TestUseCase_UpdateCustomerProfile(t *testing.T) {
	ctx := context.Background()
	handlers := NewHandlers(sharedRepo, sharedPublisher, sharedTM, sharedEventPublisher)

	// Setup: Create initial customer
	custID := "cust-update-1"
	initialCust := &domain.Customer{
		ID:      custID,
		Name:    "Original Name",
		Status:  domain.CustomerStatusActive,
		PartyID: "p-update-1",
		// Initial sub-resources to verify replacement
		Accounts:       []domain.CustomerAccount{{ID: "acc-old", Name: "Old Acc", CustomerID: custID}},
		CreditProfiles: []domain.CreditProfile{{ID: "cp-old", CreditScore: 600, CustomerID: custID}},
	}
	require.NoError(t, sharedRepo.CreateCustomer(ctx, initialCust))

	// Update Payload
	payload := UpdateCustomerPayload{
		ID:     custID,
		Name:   "Updated Name",
		Status: domain.CustomerStatusSuspended,
		// Update Accounts (Replace list)
		Accounts: []CustomerAccountDTO{
			{ID: "acc-new", Name: "New Acc", AccountStatus: "active", AccountType: "checking"},
		},
		// Update Credit Profile (Replace list)
		CreditProfiles: []CreditProfileDTO{
			{ID: "cp-new", CreditScore: 800, CreditRiskScore: 5},
		},
		// Update Contact Mediums
		ContactMediums: []ContactMediumDTO{
			{ID: "cm-new", MediumType: "email", Value: "new@example.com"},
		},
		// Update Characteristics
		Characteristics: []CharacteristicDTO{
			{ID: "char-new", Name: "Level", Value: "Pro"},
		},
		// Update Related Parties
		RelatedParties: []RelatedPartyDTO{
			{ID: "rp-new", Name: "Sibling", Role: "Family"},
		},
		// Update Payment Methods
		PaymentMethods: []PaymentMethodDTO{
			{ID: "pm-new", Type: "Debit", Token: "tok_new"},
		},
		// Update Market Segments
		MarketSegments: []MarketSegmentDTO{
			{ID: "ms-new", Name: "SME", Category: "Biz"},
		},
		// Update Privacy Consent (New list)
		PrivacyConsents: []PrivacyConsentDTO{
			{ConsentType: "Marketing", Status: "withdrawn", ValidForStart: time.Now().Format(time.RFC3339)},
		},
	}
	body, _ := json.Marshal(payload)

	err := handlers.HandleUpdateCustomer(ctx, amqp.Delivery{Body: body})
	require.NoError(t, err)

	// Verify Updates
	updated, err := sharedRepo.GetCustomer(ctx, custID)
	require.NoError(t, err)

	// Basic Info
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, domain.CustomerStatusSuspended, updated.Status)

	// Accounts (Should be replaced)
	require.Len(t, updated.Accounts, 1)
	assert.Equal(t, "New Acc", updated.Accounts[0].Name)
	assert.Equal(t, "acc-new", updated.Accounts[0].ID)

	// Credit Profile (Should be replaced)
	require.Len(t, updated.CreditProfiles, 1)
	assert.Equal(t, 800, updated.CreditProfiles[0].CreditScore)
	assert.Equal(t, "cp-new", updated.CreditProfiles[0].ID)

	// Tax Exemptions

	// Privacy Consents
	require.Len(t, updated.PrivacyConsents, 1)
	assert.Equal(t, "withdrawn", updated.PrivacyConsents[0].Status)
}

// 3. Get Customer Use Case (RPC)
func TestUseCase_RetrieveCustomer(t *testing.T) {
	ctx := context.Background()
	handlers := NewHandlers(sharedRepo, sharedPublisher, sharedTM, sharedEventPublisher)

	// Setup: Complete customer
	custID := "cust-get-1"
	fullCust := &domain.Customer{
		ID:             custID,
		Name:           "Get Test",
		Status:         domain.CustomerStatusActive,
		PartyID:        "p-get-1",
		Accounts:       []domain.CustomerAccount{{ID: "a1", Name: "A1", CustomerID: custID}},
		CreditProfiles: []domain.CreditProfile{{ID: "cp1", CreditScore: 700, CustomerID: custID}},

		PrivacyConsents: []domain.PrivacyConsent{{ID: "pc1", ConsentType: "All", CustomerID: custID}},
	}
	require.NoError(t, sharedRepo.CreateCustomer(ctx, fullCust))

	// Prepare RPC reply queue
	ch, err := sharedConn.Channel()
	require.NoError(t, err)
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()
	replyQueue, _ := ch.QueueDeclare("", false, true, true, false, nil)
	msgs, _ := ch.Consume(replyQueue.Name, "", true, true, false, false, nil)

	// Execute HandleGetCustomer
	corrID := "corr-get-1"
	payload, _ := json.Marshal(GetCustomerPayload{ID: custID})
	err = handlers.HandleGetCustomer(ctx, amqp.Delivery{
		Body: payload, ReplyTo: replyQueue.Name, CorrelationId: corrID,
	})
	require.NoError(t, err)

	// Test errors
	_ = handlers.HandleGetCustomer(ctx, amqp.Delivery{Body: []byte("{invalid json}")})
	_ = handlers.HandleGetCustomer(ctx, amqp.Delivery{Body: []byte(`{"id":"non-existent"}`)})

	// Verify Reply
	select {
	case msg := <-msgs:
		assert.Equal(t, corrID, msg.CorrelationId)
		var resp domain.Customer
		err = json.Unmarshal(msg.Body, &resp)
		require.NoError(t, err)

		assert.Equal(t, custID, resp.ID)
		assert.Equal(t, "Get Test", resp.Name)

		// Verify all sub-resources present
		require.Len(t, resp.Accounts, 1)
		assert.Equal(t, "A1", resp.Accounts[0].Name)

		require.Len(t, resp.CreditProfiles, 1)
		assert.Equal(t, 700, resp.CreditProfiles[0].CreditScore)

		require.Len(t, resp.PrivacyConsents, 1)
		assert.Equal(t, "All", resp.PrivacyConsents[0].ConsentType)

	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for RPC reply")
	}
}

// 4. Search Customer Use Case
func TestUseCase_SearchCustomers(t *testing.T) {
	ctx := context.Background()
	handlers := NewHandlers(sharedRepo, sharedPublisher, sharedTM, sharedEventPublisher)

	// Setup
	require.NoError(t, sharedRepo.CreateCustomer(ctx, &domain.Customer{
		ID: "cust-search-1", Name: "UniqueTarget", Status: domain.CustomerStatusActive, PartyID: "some-party",
	}))

	// RPC Setup
	ch, err := sharedConn.Channel()
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()
	replyQueue, _ := ch.QueueDeclare("", false, true, true, false, nil)
	msgs, _ := ch.Consume(replyQueue.Name, "", true, true, false, false, nil)

	// Execute
	corrID := "corr-search-1"
	payload, _ := json.Marshal(SearchCustomerPayload{Name: "UniqueTarget", ID: "cust-search-1", Search: "Uniq", Status: "Active", PartyID: "some-party"})
	err = handlers.HandleSearchCustomer(ctx, amqp.Delivery{
		Body: payload, ReplyTo: replyQueue.Name, CorrelationId: corrID,
	})
	require.NoError(t, err)

	// Verify
	select {
	case msg := <-msgs:
		assert.Equal(t, corrID, msg.CorrelationId)
		var results []domain.Customer
		err = json.Unmarshal(msg.Body, &results)
		require.NoError(t, err)

		assert.NotEmpty(t, results)
		assert.Equal(t, "UniqueTarget", results[0].Name)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout")
	}
}

// 5. Delete Customer Use Case
func TestUseCase_DeleteCustomer(t *testing.T) {
	ctx := context.Background()
	handlers := NewHandlers(sharedRepo, sharedPublisher, sharedTM, sharedEventPublisher)

	// Setup
	custID := "cust-del-1"
	require.NoError(t, sharedRepo.CreateCustomer(ctx, &domain.Customer{ID: custID, Name: "To Delete", Status: domain.CustomerStatusActive}))

	// Execute
	ch, err := sharedConn.Channel()
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()
	replyQueue, _ := ch.QueueDeclare("", false, true, true, false, nil)
	msgs, _ := ch.Consume(replyQueue.Name, "", true, true, false, false, nil)

	payload, _ := json.Marshal(DeleteCustomerPayload{ID: custID})
	err = handlers.HandleDeleteCustomer(ctx, amqp.Delivery{Body: payload, Exchange: "tmf.events", ReplyTo: replyQueue.Name})
	require.NoError(t, err)

	<-msgs // wait for reply

	// Verify Gone
	_, err = sharedRepo.GetCustomer(ctx, custID)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrNotFound, err)
}

// 6. Party Events Use Cases
func TestUseCase_PartyEvent_Update(t *testing.T) {
	ctx := context.Background()
	handlers := NewHandlers(sharedRepo, sharedPublisher, sharedTM, sharedEventPublisher)

	// Setup: Customer linked to a party
	custID := "cust-pe-upd-1"
	partyID := "party-pe-1"
	require.NoError(t, sharedRepo.CreateCustomer(ctx, &domain.Customer{
		ID: custID, Name: "Old Name", PartyID: partyID, Status: domain.CustomerStatusActive,
	}))

	// Event: Party Updated (Individual)
	evtPayload := PartyEventPayload{
		ID: partyID, Type: "Individual", GivenName: "John", FamilyName: "Updated",
	}
	body, _ := json.Marshal(evtPayload)

	err := handlers.HandlePartyEvent(ctx, amqp.Delivery{
		Body: body, RoutingKey: EvtPartyUpdated,
	})
	require.NoError(t, err)

	// Verify Customer Name Updated
	updated, err := sharedRepo.GetCustomer(ctx, custID)
	require.NoError(t, err)
	assert.Equal(t, "John Updated", updated.Name)

	// Test Organization
	orgEvt := PartyEventPayload{ID: partyID, Type: "Organization", TradingName: "Acme Corp"}
	bodyOrg, _ := json.Marshal(orgEvt)
	_ = handlers.HandlePartyEvent(ctx, amqp.Delivery{Body: bodyOrg, RoutingKey: EvtPartyUpdated})

	updatedOrg, _ := sharedRepo.GetCustomer(ctx, custID)
	assert.Equal(t, "Acme Corp", updatedOrg.Name)
}

func TestUseCase_PartyEvent_Delete(t *testing.T) {
	ctx := context.Background()
	handlers := NewHandlers(sharedRepo, sharedPublisher, sharedTM, sharedEventPublisher)

	// Setup
	custID := "cust-pe-del-1"
	partyID := "party-pe-del-1"
	require.NoError(t, sharedRepo.CreateCustomer(ctx, &domain.Customer{
		ID: custID, Name: "To Close", PartyID: partyID, Status: domain.CustomerStatusActive,
	}))

	// Event: Party Deleted
	evtPayload := PartyEventPayload{ID: partyID, Type: "Individual"}
	body, _ := json.Marshal(evtPayload)

	err := handlers.HandlePartyEvent(ctx, amqp.Delivery{
		Body: body, RoutingKey: EvtPartyDeleted,
	})
	require.NoError(t, err)

	// Verify Customer Closed
	closed, err := sharedRepo.GetCustomer(ctx, custID)
	require.NoError(t, err)
	assert.Equal(t, domain.CustomerStatusClosed, closed.Status)
	assert.Contains(t, closed.StatusReason, "Linked party was deleted")
}

// 7. Audit Logging Use Case
func TestUseCase_AuditLogging(t *testing.T) {
	ctx := context.Background()
	handlers := NewHandlers(sharedRepo, sharedPublisher, sharedTM, sharedEventPublisher)

	// Test Onboard with User Header
	userID := "audit-tester"
	payload := OnboardCustomerPayload{ID: "cust-audit-1", Name: "Audit Me"}
	body, _ := json.Marshal(payload)

	err := handlers.HandleOnboardCustomer(ctx, amqp.Delivery{
		Body: body, Headers: amqp.Table{"user": userID}, Exchange: "tmf.events",
	})
	require.NoError(t, err)

	// Verify Audit Log
	type LoggedAction struct {
		TableName string `gorm:"column:table_name"`
		UserName  string `gorm:"column:user_name"`
		Action    string `gorm:"column:action"`
	}

	var auditLog LoggedAction
	err = sharedDB.Table("audit.logged_actions").
		Where("table_name = ? AND action = ? AND user_name = ?", "customers", "I", userID).
		Order("action_tstamp_clk DESC").
		First(&auditLog).Error

	require.NoError(t, err)
	assert.Equal(t, "customers", auditLog.TableName)
	assert.Equal(t, "I", auditLog.Action)
	assert.Equal(t, userID, auditLog.UserName)
}

// 8. Party Deletion Saga Use Cases
func TestUseCase_PartyEvent_DeletionInitiated_ActiveCustomer(t *testing.T) {
	ctx := context.Background()
	handlers := NewHandlers(sharedRepo, sharedPublisher, sharedTM, sharedEventPublisher)

	// Setup: Active Customer linked to party
	custID := "cust-saga-active-1"
	partyID := "party-saga-active-1"
	require.NoError(t, sharedRepo.CreateCustomer(ctx, &domain.Customer{
		ID: custID, Name: "Active Customer", PartyID: partyID, Status: domain.CustomerStatusActive,
	}))

	// Setup: Mock Party Exchange and Queue to catch commands
	ch, err := sharedConn.Channel()
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()

	partyExchange := "tmf.party"
	err = ch.ExchangeDeclare(partyExchange, "topic", true, false, false, false, nil)
	require.NoError(t, err)

	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	require.NoError(t, err)

	err = ch.QueueBind(q.Name, CmdPartyCancelDeletion, partyExchange, false, nil)
	require.NoError(t, err)

	msgs, err := ch.Consume(q.Name, "", true, true, false, false, nil)
	require.NoError(t, err)

	// Execute: Event Party Deletion Initiated
	evtPayload := PartyEventPayload{ID: partyID, Type: "Individual"}
	body, _ := json.Marshal(evtPayload)

	err = handlers.HandlePartyEvent(ctx, amqp.Delivery{
		Body: body, RoutingKey: EvtPartyDeletionInitiated,
	})
	require.NoError(t, err)

	// Test Empty ID
	_ = handlers.HandlePartyEvent(ctx, amqp.Delivery{
		Body: []byte(`{}`), RoutingKey: EvtPartyDeletionInitiated,
	})

	// Verify: Should receive Cancel Command
	select {
	case msg := <-msgs:
		assert.Equal(t, CmdPartyCancelDeletion, msg.RoutingKey)
		var p map[string]string
		err = json.Unmarshal(msg.Body, &p)
		require.NoError(t, err)
		assert.Equal(t, partyID, p["id"])
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for cancel command")
	}
}

func TestUseCase_PartyEvent_DeletionInitiated_NoCustomer(t *testing.T) {
	ctx := context.Background()
	handlers := NewHandlers(sharedRepo, sharedPublisher, sharedTM, sharedEventPublisher)

	// Setup: No active customers for this party
	partyID := "party-saga-none-1"

	// Setup: Mock Party Exchange and Queue to catch commands
	ch, err := sharedConn.Channel()
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()

	partyExchange := "tmf.party"
	err = ch.ExchangeDeclare(partyExchange, "topic", true, false, false, false, nil)
	require.NoError(t, err)

	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	require.NoError(t, err)

	err = ch.QueueBind(q.Name, CmdPartyFinalizeDeletion, partyExchange, false, nil)
	require.NoError(t, err)

	msgs, err := ch.Consume(q.Name, "", true, true, false, false, nil)
	require.NoError(t, err)

	// Execute: Event Party Deletion Initiated
	evtPayload := PartyEventPayload{ID: partyID, Type: "Individual"}
	body, _ := json.Marshal(evtPayload)

	err = handlers.HandlePartyEvent(ctx, amqp.Delivery{
		Body: body, RoutingKey: EvtPartyDeletionInitiated,
	})
	require.NoError(t, err)

	// Test Empty ID
	_ = handlers.HandlePartyEvent(ctx, amqp.Delivery{
		Body: []byte(`{}`), RoutingKey: EvtPartyDeletionInitiated,
	})

	// Verify: Should receive Finalize Command
	select {
	case msg := <-msgs:
		assert.Equal(t, CmdPartyFinalizeDeletion, msg.RoutingKey)
		var p map[string]string
		err = json.Unmarshal(msg.Body, &p)
		require.NoError(t, err)
		assert.Equal(t, partyID, p["id"])
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for finalize command")
	}
}
