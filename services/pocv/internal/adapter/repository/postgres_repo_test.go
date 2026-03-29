package repository

import (
	"context"
	"testing"
	"time"

	"tmf/services/pocv/internal/core/domain"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	err = db.AutoMigrate(&SagaTable{}, &OutboxTable{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}
	return db
}

func TestSagaRepository_GetByCartID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSagaRepository(db)
	ctx := context.Background()

	// Setup data
	customerID := "cust-1"
	dao := &SagaTable{
		ID:          "saga-1",
		CartID:      "cart-1",
		CustomerID:  &customerID,
		CurrentStep: "INVENTORY",
		Status:      "IN_PROGRESS",
		Payload:     []byte(`{}`),
		History:     []byte(`[]`),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	db.Create(dao)

	t.Run("found", func(t *testing.T) {
		res, err := repo.GetByCartID(ctx, "cart-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil || res.ID != "saga-1" {
			t.Errorf("expected saga-1, got %v", res)
		}
	})

	t.Run("not found", func(t *testing.T) {
		res, err := repo.GetByCartID(ctx, "cart-2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res != nil {
			t.Errorf("expected nil, got %v", res)
		}
	})

	// Inject custom context with tx
	tx := db.Begin()
	ctxTx := WithTx(ctx, tx)
	t.Run("with transaction context", func(t *testing.T) {
		res, err := repo.GetByCartID(ctxTx, "cart-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil || res.ID != "saga-1" {
			t.Errorf("expected saga-1, got %v", res)
		}
	})
	tx.Commit()
}

func TestSagaRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSagaRepository(db)
	ctx := context.Background()

	customerID := "cust-1"
	saga := &domain.SagaInstance{
		ID:          "saga-2",
		CartID:      "cart-2",
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
			ID:        "evt-1",
			Topic:     "test",
			Payload:   []byte(`{}`),
			Status:    "PENDING",
			CreatedAt: time.Now(),
		},
	}

	err := repo.Create(ctx, saga, events)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify saga
	var dbSaga SagaTable
	if err := db.First(&dbSaga, "id = ?", "saga-2").Error; err != nil {
		t.Fatalf("failed to find created saga: %v", err)
	}

	// Verify outbox
	var dbOutbox OutboxTable
	if err := db.First(&dbOutbox, "id = ?", "evt-1").Error; err != nil {
		t.Fatalf("failed to find created outbox event: %v", err)
	}
}

func TestSagaRepository_Get(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSagaRepository(db)
	ctx := context.Background()

	dao := &SagaTable{
		ID:          "saga-3",
		CartID:      "cart-3",
		CurrentStep: "INVENTORY",
		Status:      "IN_PROGRESS",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	db.Create(dao)

	res, err := repo.Get(ctx, "saga-3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil || res.ID != "saga-3" {
		t.Errorf("expected saga-3, got %v", res)
	}

	// Not found
	res, err = repo.Get(ctx, "saga-not-found")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != nil {
		t.Errorf("expected nil, got %v", res)
	}
}

func TestDBFromContext(t *testing.T) {
	db := setupTestDB(t)
	tx := db.Begin()
	defer tx.Rollback()

	if got, ok := DBFromContext(context.Background()); ok || got != nil {
		t.Fatalf("expected no transaction in background context, got %v, %v", got, ok)
	}

	ctx := WithTx(context.Background(), tx)
	got, ok := DBFromContext(ctx)
	if !ok {
		t.Fatal("expected transaction in typed context")
	}
	if got != tx {
		t.Fatalf("expected same transaction pointer")
	}
}

func TestSagaRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSagaRepository(db)
	ctx := context.Background()

	dao := &SagaTable{
		ID:          "saga-4",
		CartID:      "cart-4",
		CurrentStep: "INVENTORY",
		Status:      "IN_PROGRESS",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	db.Create(dao)

	saga := &domain.SagaInstance{
		ID:          "saga-4",
		CartID:      "cart-4",
		CurrentStep: domain.StepPayment,
		Status:      domain.SagaStatusCompleted,
		Payload:     []byte(`{}`),
		History:     []byte(`[]`),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	events := []domain.OutboxEvent{
		{
			ID:        "evt-2",
			Topic:     "test",
			Payload:   []byte(`{}`),
			Status:    "PENDING",
			CreatedAt: time.Now(),
		},
	}

	err := repo.Update(ctx, saga, events)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify saga
	var dbSaga SagaTable
	if err := db.First(&dbSaga, "id = ?", "saga-4").Error; err != nil {
		t.Fatalf("failed to find updated saga: %v", err)
	}
	if dbSaga.CurrentStep != "PAYMENT" || dbSaga.Status != "COMPLETED" {
		t.Errorf("saga not updated properly")
	}

	// Verify outbox
	var dbOutbox OutboxTable
	if err := db.First(&dbOutbox, "id = ?", "evt-2").Error; err != nil {
		t.Fatalf("failed to find created outbox event: %v", err)
	}
}

func TestSagaRepository_CreateAndUpdate_WithoutOutbox(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSagaRepository(db)
	ctx := context.Background()

	var initialOutboxCount int64
	if err := db.Model(&OutboxTable{}).Count(&initialOutboxCount).Error; err != nil {
		t.Fatalf("failed to count initial outbox rows: %v", err)
	}

	saga := &domain.SagaInstance{
		ID:          "saga-no-outbox",
		CartID:      "cart-no-outbox",
		CurrentStep: domain.StepInventory,
		Status:      domain.SagaStatusInProgress,
		Payload:     []byte(`{}`),
		History:     []byte(`[]`),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := repo.Create(ctx, saga, nil); err != nil {
		t.Fatalf("unexpected create error without outbox: %v", err)
	}

	var created SagaTable
	if err := db.First(&created, "id = ?", saga.ID).Error; err != nil {
		t.Fatalf("failed to load created saga: %v", err)
	}

	var outboxCount int64
	if err := db.Model(&OutboxTable{}).Count(&outboxCount).Error; err != nil {
		t.Fatalf("failed to count outbox rows: %v", err)
	}
	if outboxCount != initialOutboxCount {
		t.Fatalf("expected outbox count to remain %d, got %d", initialOutboxCount, outboxCount)
	}

	saga.CurrentStep = domain.StepPayment
	saga.Status = domain.SagaStatusCompleted
	saga.UpdatedAt = time.Now()

	if err := repo.Update(ctx, saga, nil); err != nil {
		t.Fatalf("unexpected update error without outbox: %v", err)
	}

	var updated SagaTable
	if err := db.First(&updated, "id = ?", saga.ID).Error; err != nil {
		t.Fatalf("failed to load updated saga: %v", err)
	}
	if updated.CurrentStep != string(domain.StepPayment) || updated.Status != string(domain.SagaStatusCompleted) {
		t.Fatalf("expected updated saga state, got step=%s status=%s", updated.CurrentStep, updated.Status)
	}

	if err := db.Model(&OutboxTable{}).Count(&outboxCount).Error; err != nil {
		t.Fatalf("failed to count outbox rows after update: %v", err)
	}
	if outboxCount != initialOutboxCount {
		t.Fatalf("expected outbox count to remain %d after update, got %d", initialOutboxCount, outboxCount)
	}
}
