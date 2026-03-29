package repository

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"tmf/services/pocv/internal/core/domain"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupIntegrationDB starts a real PostgreSQL container and runs migrations.
// It skips the test if testcontainers is unavailable (e.g. no Docker).
func setupIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	ctx := context.Background()

	defer func() {
		if r := recover(); r != nil {
			t.Skipf("Skipping integration test (testcontainers panic): %v", r)
		}
	}()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:15",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		t.Skipf("Skipping integration test (cannot start postgres): %v", err)
		return nil
	}

	t.Cleanup(func() {
		_ = pgContainer.Terminate(ctx)
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Skipf("Skipping integration test (cannot get connection string): %v", err)
		return nil
	}

	// Run migrations
	// From internal/adapter/repository/ → up 3 dirs → services/pocv/ → internal/infrastructure/postgres/migrations
	_, filename, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(filename), "..", "..", "..", "internal", "infrastructure", "postgres", "migrations")

	m, err := migrate.New("file://"+migrationsPath, connStr)
	if err != nil {
		t.Skipf("Skipping integration test (cannot create migrator): %v", err)
		return nil
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Skipf("Skipping integration test (migration failed): %v", err)
		return nil
	}

	db, err := gorm.Open(gormPostgres.Open(connStr), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping integration test (cannot connect to postgres): %v", err)
		return nil
	}

	return db
}

// TestSagaRepository_GetByCartID_DBError covers the unexpected DB error branch
// (not ErrRecordNotFound) by closing the DB connection.
func TestSagaRepository_GetByCartID_DBError(t *testing.T) {
	db := setupIntegrationDB(t)
	if db == nil {
		return
	}

	repo := NewSagaRepository(db)
	ctx := context.Background()

	// Close the underlying SQL connection to simulate a DB error
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB: %v", err)
	}
	_ = sqlDB.Close()

	_, err = repo.GetByCartID(ctx, uuid.New().String())
	if err == nil {
		t.Error("expected error when DB is closed, got nil")
	}
}

// TestSagaRepository_Get_DBError covers the unexpected DB error branch in Get.
func TestSagaRepository_Get_DBError(t *testing.T) {
	db := setupIntegrationDB(t)
	if db == nil {
		return
	}

	repo := NewSagaRepository(db)
	ctx := context.Background()

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB: %v", err)
	}
	_ = sqlDB.Close()

	_, err = repo.Get(ctx, uuid.New().String())
	if err == nil {
		t.Error("expected error when DB is closed, got nil")
	}
}

// TestSagaRepository_Create_WithEmptyEvents covers the empty events branch in Create
// (the if len(outbox) > 0 check is false, so outbox insert is skipped).
func TestSagaRepository_Create_WithEmptyEvents(t *testing.T) {
	db := setupIntegrationDB(t)
	if db == nil {
		return
	}

	repo := NewSagaRepository(db)
	ctx := context.Background()

	sagaID := uuid.New().String()
	cartID := uuid.New().String()
	customerID := uuid.New().String()

	saga := &domain.SagaInstance{
		ID:          sagaID,
		CartID:      cartID,
		CustomerID:  &customerID,
		CurrentStep: domain.StepInventory,
		Status:      domain.SagaStatusInProgress,
		Payload:     []byte(`{}`),
		History:     []byte(`[]`),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Pass nil events — takes the "no outbox" branch (len(outbox) > 0 is false)
	err := repo.Create(ctx, saga, nil)
	if err != nil {
		t.Fatalf("unexpected error creating saga with no events: %v", err)
	}

	// Verify it was persisted
	found, err := repo.Get(ctx, sagaID)
	if err != nil {
		t.Fatalf("unexpected error fetching saga: %v", err)
	}
	if found == nil || found.ID != sagaID {
		t.Errorf("expected saga %s, got %v", sagaID, found)
	}
}

// TestSagaRepository_Update_WithEmptyEvents covers the empty events branch in Update.
func TestSagaRepository_Update_WithEmptyEvents(t *testing.T) {
	db := setupIntegrationDB(t)
	if db == nil {
		return
	}

	repo := NewSagaRepository(db)
	ctx := context.Background()

	sagaID := uuid.New().String()
	cartID := uuid.New().String()
	customerID := uuid.New().String()

	// First create a saga to update
	saga := &domain.SagaInstance{
		ID:          sagaID,
		CartID:      cartID,
		CustomerID:  &customerID,
		CurrentStep: domain.StepInventory,
		Status:      domain.SagaStatusInProgress,
		Payload:     []byte(`{}`),
		History:     []byte(`[]`),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := repo.Create(ctx, saga, nil); err != nil {
		t.Fatalf("unexpected error creating saga: %v", err)
	}

	// Now update with no events — takes the "no outbox" branch
	saga.CurrentStep = domain.StepPayment
	saga.Status = domain.SagaStatusCompleted
	saga.UpdatedAt = time.Now()

	err := repo.Update(ctx, saga, nil)
	if err != nil {
		t.Fatalf("unexpected error updating saga with no events: %v", err)
	}

	// Verify update was persisted
	found, err := repo.Get(ctx, sagaID)
	if err != nil {
		t.Fatalf("unexpected error fetching updated saga: %v", err)
	}
	if found == nil || found.CurrentStep != domain.StepPayment {
		t.Errorf("expected step PAYMENT, got %v", found)
	}
}

// TestSagaRepository_Create_OutboxInsertError covers the outbox insert failure in Create
// by pre-inserting an event with the same UUID to force a constraint violation.
func TestSagaRepository_Create_OutboxInsertError(t *testing.T) {
	db := setupIntegrationDB(t)
	if db == nil {
		return
	}

	repo := NewSagaRepository(db)
	ctx := context.Background()

	// Pre-insert an outbox event with a known UUID
	dupEvtID := uuid.New().String()
	existingOutbox := &OutboxTable{
		ID:        dupEvtID,
		Topic:     "test.topic",
		Payload:   []byte(`{}`),
		Status:    "PENDING",
		CreatedAt: time.Now(),
	}
	if err := db.Create(existingOutbox).Error; err != nil {
		t.Fatalf("failed to pre-insert outbox event: %v", err)
	}

	// Try to create a saga with the same outbox event UUID — outbox insert should fail
	sagaID := uuid.New().String()
	cartID := uuid.New().String()
	customerID := uuid.New().String()

	saga := &domain.SagaInstance{
		ID:          sagaID,
		CartID:      cartID,
		CustomerID:  &customerID,
		CurrentStep: domain.StepInventory,
		Status:      domain.SagaStatusInProgress,
		Payload:     []byte(`{}`),
		History:     []byte(`[]`),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	events := []domain.OutboxEvent{
		{
			ID:        dupEvtID, // duplicate UUID — will trigger primary key violation
			Topic:     "test.topic",
			Payload:   []byte(`{}`),
			Status:    "PENDING",
			CreatedAt: time.Now(),
		},
	}

	err := repo.Create(ctx, saga, events)
	if err == nil {
		t.Error("expected error due to duplicate outbox UUID, got nil")
	}

	// Verify the saga was NOT created (transaction rolled back)
	found, _ := repo.Get(ctx, sagaID)
	if found != nil {
		t.Error("saga should not have been persisted after outbox insert failure")
	}
}

// TestSagaRepository_Update_OutboxInsertError covers the outbox insert failure in Update.
func TestSagaRepository_Update_OutboxInsertError(t *testing.T) {
	db := setupIntegrationDB(t)
	if db == nil {
		return
	}

	repo := NewSagaRepository(db)
	ctx := context.Background()

	sagaID := uuid.New().String()
	cartID := uuid.New().String()
	customerID := uuid.New().String()

	// Create base saga
	saga := &domain.SagaInstance{
		ID:          sagaID,
		CartID:      cartID,
		CustomerID:  &customerID,
		CurrentStep: domain.StepInventory,
		Status:      domain.SagaStatusInProgress,
		Payload:     []byte(`{}`),
		History:     []byte(`[]`),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := repo.Create(ctx, saga, nil); err != nil {
		t.Fatalf("failed to create base saga: %v", err)
	}

	// Pre-insert an outbox event that will conflict
	dupEvtID := uuid.New().String()
	existingOutbox := &OutboxTable{
		ID:        dupEvtID,
		Topic:     "test.topic",
		Payload:   []byte(`{}`),
		Status:    "PENDING",
		CreatedAt: time.Now(),
	}
	if err := db.Create(existingOutbox).Error; err != nil {
		t.Fatalf("failed to pre-insert outbox event: %v", err)
	}

	// Try update with duplicate outbox event UUID
	saga.CurrentStep = domain.StepPayment
	events := []domain.OutboxEvent{
		{
			ID:        dupEvtID, // duplicate — triggers primary key violation
			Topic:     "test.topic",
			Payload:   []byte(`{}`),
			Status:    "PENDING",
			CreatedAt: time.Now(),
		},
	}

	err := repo.Update(ctx, saga, events)
	if err == nil {
		t.Error("expected error due to duplicate outbox UUID in update, got nil")
	}

	// Verify saga step was NOT updated (transaction rolled back)
	found, _ := repo.Get(ctx, sagaID)
	if found != nil && found.CurrentStep != domain.StepInventory {
		t.Errorf("saga step should not have changed after update failure; got %v", found.CurrentStep)
	}
}
