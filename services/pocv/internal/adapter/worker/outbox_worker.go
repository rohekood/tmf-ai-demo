package worker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"tmf/pkg/rabbitmq"
	"tmf/services/pocv/internal/adapter/repository"

	"gorm.io/gorm"
)

type OutboxWorker struct {
	db           *gorm.DB
	publisher    rabbitmq.Publisher
	exchangeName string
}

func NewOutboxWorker(db *gorm.DB, pub rabbitmq.Publisher, exchangeName string) *OutboxWorker {
	return &OutboxWorker{
		db:           db,
		publisher:    pub,
		exchangeName: exchangeName,
	}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *OutboxWorker) processBatch(ctx context.Context) {
	var events []repository.OutboxEventModel
	if err := w.db.WithContext(ctx).
		Where("status = ?", repository.StatusPending).
		Order("created_at ASC").
		Limit(10).
		Find(&events).Error; err != nil {
		return
	}

	if len(events) == 0 {
		return
	}

	for _, evt := range events {
		w.publishEvent(ctx, &evt)
	}
}

func (w *OutboxWorker) publishEvent(ctx context.Context, evt *repository.OutboxEventModel) {
	pubCtx := ctx
	if len(evt.Headers) > 0 {
		var headers map[string]string
		if err := json.Unmarshal(evt.Headers, &headers); err == nil {
			for k, v := range headers {
				pubCtx = context.WithValue(pubCtx, k, v) //nolint:staticcheck
			}
		}
	}

	err := w.publisher.Publish(pubCtx, w.exchangeName, evt.Topic, json.RawMessage(evt.Payload))

	now := time.Now()
	if err != nil {
		log.Printf("POCV: Failed to publish event %s: %v", evt.ID, err)
		evt.Status = repository.StatusFailed
	} else {
		evt.Status = repository.StatusPublished
		evt.ProcessedAt = &now
	}
	w.db.Save(evt)
}
