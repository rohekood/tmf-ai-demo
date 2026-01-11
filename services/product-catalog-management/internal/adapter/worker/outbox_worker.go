package worker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"tmf/services/product-catalog-management/internal/adapter/publisher"
	"tmf/services/product-catalog-management/internal/adapter/repository"
	"tmf/services/product-catalog-management/internal/core/domain"

	"gorm.io/gorm"
)

type OutboxWorker struct {
	db              *gorm.DB
	rabbitPublisher *publisher.RabbitMQPublisher
	batchSize       int
	pollInterval    time.Duration
}

func NewOutboxWorker(db *gorm.DB, rabbitPublisher *publisher.RabbitMQPublisher) *OutboxWorker {
	return &OutboxWorker{
		db:              db,
		rabbitPublisher: rabbitPublisher,
		batchSize:       10,
		pollInterval:    1 * time.Second, // Could be configurable
	}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	log.Println("Starting Outbox Worker...")
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Outbox Worker...")
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *OutboxWorker) processBatch(ctx context.Context) {
	var events []repository.OutboxEventModel

	// Fetch pending events
	// Use a transaction to lock rows if specific DB supports SKIP LOCKED,
	// for now simplistic fetch
	err := w.db.WithContext(ctx).
		Where("status = ?", repository.StatusPending).
		Order("created_at asc").
		Limit(w.batchSize).
		Find(&events).Error

	if err != nil {
		log.Printf("Error processing outbox batch: %v", err)
		return
	}

	if len(events) == 0 {
		return
	}

	for _, event := range events {
		w.processEvent(ctx, &event)
	}
}

func (w *OutboxWorker) processEvent(ctx context.Context, event *repository.OutboxEventModel) {
	publishCtx := ctx
	if len(event.Headers) > 0 {
		var headers map[string]string
		if err := json.Unmarshal(event.Headers, &headers); err == nil {
			if user, ok := headers["user"]; ok {
				publishCtx = context.WithValue(publishCtx, domain.UserContextKey, user)
			}
			if auth, ok := headers["Authorization"]; ok {
				publishCtx = context.WithValue(publishCtx, domain.AuthContextKey, auth)
			}
		} else {
			log.Printf("Failed to unmarshal headers for event %s: %v", event.ID, err)
		}
	}

	err := w.rabbitPublisher.PublishRaw(publishCtx, event.RoutingKey, event.Payload)
	now := time.Now()

	if err != nil {
		log.Printf("Failed to publish outbox event %s: %v", event.ID, err)
		event.Status = repository.StatusFailed // Or retry logic
	} else {
		event.Status = repository.StatusPublished
		event.ProcessedAt = &now
	}

	if err := w.db.Save(event).Error; err != nil {
		log.Printf("Failed to update status for outbox event %s: %v", event.ID, err)
	}
}
