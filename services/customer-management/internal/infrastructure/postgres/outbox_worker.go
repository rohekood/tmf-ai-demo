package postgres

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"
	"tmf/pkg/rabbitmq"
)

type OutboxWorker struct {
	repo      *OutboxRepository
	publisher rabbitmq.Publisher
	logger    *slog.Logger
	stop      chan struct{}
}

func NewOutboxWorker(repo *OutboxRepository, publisher rabbitmq.Publisher, logger *slog.Logger) *OutboxWorker {
	return &OutboxWorker{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
		stop:      make(chan struct{}),
	}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond) // Poll every 500ms
	defer ticker.Stop()

	w.logger.Info("Starting Outbox Worker")

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Stopping Outbox Worker via context")
			return
		case <-w.stop:
			w.logger.Info("Stopping Outbox Worker via signal")
			return
		case <-ticker.C:
			w.processEvents(ctx)
		}
	}
}

func (w *OutboxWorker) Stop() {
	close(w.stop)
}

func (w *OutboxWorker) processEvents(ctx context.Context) {
	events, err := w.repo.FetchPending(ctx, 10)
	if err != nil {
		w.logger.Error("Failed to fetch pending events", "error", err)
		return
	}

	for _, event := range events {
		var payload any
		err := json.Unmarshal(event.Payload, &payload)
		if err != nil {
			w.logger.Error("Failed to unmarshal event payload", "id", event.ID, "error", err)
			continue
		}

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
			}
		}

		err = w.publisher.Publish(publishCtx, "customer.events", event.RoutingKey, payload)
		if err != nil {
			w.logger.Error("Failed to publish event", "id", event.ID, "error", err)
			continue
		}

		if err := w.repo.MarkAsProcessed(ctx, event.ID); err != nil {
			w.logger.Error("Failed to mark event as processed", "id", event.ID, "error", err)
		}
	}
}
