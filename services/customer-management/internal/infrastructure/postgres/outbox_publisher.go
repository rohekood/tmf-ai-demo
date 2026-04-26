package postgres

import (
	"context"
	"encoding/json"
	"time"
	"tmf/services/customer-management/internal/domain"

	"github.com/google/uuid"
)

type OutboxPublisher struct {
	repo *OutboxRepository
}

func NewOutboxPublisher(repo *OutboxRepository) *OutboxPublisher {
	return &OutboxPublisher{repo: repo}
}

func (p *OutboxPublisher) Publish(ctx context.Context, routingKey string, payload any) error {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	headers := make(map[string]string)
	if user, ok := ctx.Value(domain.UserContextKey).(string); ok {
		headers["user"] = user
	}
	if auth, ok := ctx.Value(domain.AuthContextKey).(string); ok {
		headers["Authorization"] = auth
	}
	headerBytes, err := json.Marshal(headers)
	if err != nil {
		return err
	}

	event := &domain.OutboxEvent{
		ID:         uuid.New().String(),
		RoutingKey: routingKey,
		Payload:    bytes,
		Headers:    headerBytes,
		Status:     "PENDING",
		CreatedAt:  time.Now().UTC(),
	}

	return p.repo.Save(ctx, event)
}
