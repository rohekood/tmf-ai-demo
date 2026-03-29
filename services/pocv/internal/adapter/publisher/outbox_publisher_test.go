package publisher

import (
	"context"
	"testing"

	"tmf/services/pocv/internal/adapter/repository"
	"tmf/services/pocv/internal/core/domain"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	err = db.AutoMigrate(&repository.OutboxEventModel{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}
	return db
}

type contextKey string

const (
	userKey contextKey = "user"
	authKey contextKey = "Authorization"
	txKey   contextKey = "tx"
)

func TestOutboxPublisher_PublishEvents(t *testing.T) {
	db := setupTestDB(t)
	pub := NewOutboxPublisher(db)

	ctx := context.WithValue(context.Background(), userKey, "test-user")
	ctx = context.WithValue(ctx, authKey, "Bearer token")

	t.Run("OrderCreated", func(t *testing.T) {
		evt := domain.OrderCreatedEvent{
			OrderID: "order-1",
		}
		err := pub.PublishOrderCreated(ctx, evt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var dbEvt repository.OutboxEventModel
		if err := db.Order("created_at desc").First(&dbEvt).Error; err != nil {
			t.Fatalf("failed to find created outbox event: %v", err)
		}
		if dbEvt.Topic != "evt.order.created" {
			t.Errorf("expected topic evt.order.created, got %s", dbEvt.Topic)
		}
	})

	t.Run("ReserveInventory", func(t *testing.T) {
		cmd := domain.ReserveInventoryCommand{
			OrderID: "order-2",
		}
		err := pub.PublishReserveInventory(ctx, cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var dbEvt repository.OutboxEventModel
		if err := db.Order("created_at desc").First(&dbEvt).Error; err != nil {
			t.Fatalf("failed to find created outbox event: %v", err)
		}
		if dbEvt.Topic != "cmd.inventory.reserve" {
			t.Errorf("expected topic cmd.inventory.reserve, got %s", dbEvt.Topic)
		}
	})

	t.Run("OrderCompleted", func(t *testing.T) {
		evt := domain.OrderCompletedEvent{
			OrderID: "order-3",
		}
		err := pub.PublishOrderCompleted(ctx, evt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var dbEvt repository.OutboxEventModel
		if err := db.Order("created_at desc").First(&dbEvt).Error; err != nil {
			t.Fatalf("failed to find created outbox event: %v", err)
		}
		if dbEvt.Topic != "evt.order.completed" {
			t.Errorf("expected topic evt.order.completed, got %s", dbEvt.Topic)
		}
	})

	t.Run("OrderFailed", func(t *testing.T) {
		evt := domain.OrderFailedEvent{
			OrderID: "order-4",
			Reason:  "failed",
		}
		err := pub.PublishOrderFailed(ctx, evt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var dbEvt repository.OutboxEventModel
		if err := db.Order("created_at desc").First(&dbEvt).Error; err != nil {
			t.Fatalf("failed to find created outbox event: %v", err)
		}
		if dbEvt.Topic != "evt.order.failed" {
			t.Errorf("expected topic evt.order.failed, got %s", dbEvt.Topic)
		}
	})

	t.Run("WithTransaction", func(t *testing.T) {
		tx := db.Begin()
		ctxTx := context.WithValue(ctx, txKey, tx)

		evt := domain.OrderFailedEvent{
			OrderID: "order-5",
			Reason:  "failed tx",
		}
		err := pub.PublishOrderFailed(ctxTx, evt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		tx.Commit()

		var dbEvt repository.OutboxEventModel
		if err := db.Order("created_at desc").First(&dbEvt).Error; err != nil {
			t.Fatalf("failed to find created outbox event: %v", err)
		}
		if dbEvt.Topic != "evt.order.failed" {
			t.Errorf("expected topic evt.order.failed, got %s", dbEvt.Topic)
		}
	})

	t.Run("InvalidPayload", func(t *testing.T) {
		err := pub.saveEvent(ctx, "test", make(chan int))
		if err == nil {
			t.Errorf("expected error for invalid payload")
		}
	})
}
