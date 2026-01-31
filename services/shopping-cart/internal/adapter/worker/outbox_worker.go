package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"tmf/pkg/rabbitmq"
	"tmf/services/shopping-cart/internal/adapter/repository"

	"gorm.io/gorm"
)

type OutboxWorker struct {
	db           *gorm.DB
	publisher    rabbitmq.Publisher
	exchangeName string
}

func NewOutboxWorker(db *gorm.DB, pub rabbitmq.Publisher, exchangeName string) *OutboxWorker {
	return &OutboxWorker{db: db, publisher: pub, exchangeName: exchangeName}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processEvents(ctx)
		}
	}
}

func (w *OutboxWorker) processEvents(ctx context.Context) {
	// Use the Table struct (DAO) which has GORM tags
	var events []repository.OutboxTable

	err := w.db.Transaction(func(tx *gorm.DB) error {
		// Use DAO
		if err := tx.Where("status = ?", "PENDING").Limit(10).Find(&events).Error; err != nil {
			return err
		}

		if len(events) == 0 {
			return nil
		}

		for _, evt := range events {
			// Publish
			if err := w.publisher.Publish(ctx, w.exchangeName, evt.Topic, json.RawMessage(evt.Payload)); err != nil {
				slog.Error("Failed to publish event", "id", evt.ID, "error", err)
				continue
			}

			// Update Status using ID
			if err := tx.Model(&repository.OutboxTable{}).Where("id = ?", evt.ID).Update("status", "PUBLISHED").Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		slog.Error("Outbox worker error", "error", err)
	}
}
