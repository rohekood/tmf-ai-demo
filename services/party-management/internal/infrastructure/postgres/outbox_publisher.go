package postgres

import (
	"context"
	"encoding/json"
	"tmf/services/party-management/internal/domain"

	"github.com/google/uuid"
)

type OutboxPublisher struct {
	repo *OutboxRepository
}

func NewOutboxPublisher(repo *OutboxRepository) *OutboxPublisher {
	return &OutboxPublisher{repo: repo}
}

// Publish implements the EventPublisher interface but saves to outbox
func (p *OutboxPublisher) Publish(ctx context.Context, exchange, routingKey string, event any) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Extract headers (user, auth)
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

	outboxEvent := &domain.OutboxEvent{
		ID:         uuid.New().String(),
		RoutingKey: routingKey,
		Payload:    payload,
		Headers:    headerBytes,
		Status:     domain.StatusPending,
	}

	return p.repo.Save(ctx, outboxEvent)
}
