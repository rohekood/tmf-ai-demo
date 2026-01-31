package postgres

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"tmf/pkg/rabbitmq"
	"tmf/services/party-management/internal/domain"
)

type OutboxWorker struct {
	repo            *OutboxRepository
	rabbitPublisher rabbitmq.Publisher
	batchSize       int
	pollInterval    time.Duration
}

func NewOutboxWorker(repo *OutboxRepository, rabbitPublisher rabbitmq.Publisher) *OutboxWorker {
	return &OutboxWorker{
		repo:            repo,
		rabbitPublisher: rabbitPublisher,
		batchSize:       50,
		pollInterval:    500 * time.Millisecond,
	}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	slog.Info("Starting Outbox Worker...")
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping Outbox Worker...")
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *OutboxWorker) processBatch(ctx context.Context) {
	events, err := w.repo.FetchPending(ctx, w.batchSize)
	if err != nil {
		slog.Error("Error processing outbox batch", "error", err)
		return
	}

	if len(events) == 0 {
		return
	}

	for _, event := range events {
		w.processEvent(ctx, &event)
	}
}

func (w *OutboxWorker) processEvent(ctx context.Context, event *domain.OutboxEvent) {
	// Unmarshal wrapper payload to raw json bytes if needed, but it's already json
	var payload interface{} = event.Payload
	exchange := "tmf.events"

	publishCtx := ctx
	if len(event.Headers) > 0 {
		var headers map[string]string
		if err := json.Unmarshal(event.Headers, &headers); err == nil {
			if user, ok := headers["user"]; ok {
				publishCtx = context.WithValue(publishCtx, rabbitmq.ContextKeyUser, user)
			}
			if auth, ok := headers["Authorization"]; ok {
				publishCtx = context.WithValue(publishCtx, rabbitmq.Key("Authorization"), auth)
			}
		} else {
			slog.Warn("Failed to unmarshal headers for event", "id", event.ID, "error", err)
		}
	}

	err := w.rabbitPublisher.Publish(publishCtx, exchange, event.RoutingKey, payload)

	if err != nil {
		slog.Error("Failed to publish outbox event", "id", event.ID, "error", err)
		// Retry logic handled by polling if status remains PENDING, but for now we might want to log or backoff.
		// If we don't update status, it will be picked up again.
	} else {
		// Use detached context for DB update to ensure it completes even if worker shuts down
		updateCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := w.repo.MarkAsProcessed(updateCtx, event.ID); err != nil {
			slog.Error("Failed to update status for outbox event", "id", event.ID, "error", err)
		}
	}
}
