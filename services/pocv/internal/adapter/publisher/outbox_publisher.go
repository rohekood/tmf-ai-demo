package publisher

import (
	"context"
	"tmf/services/pocv/internal/adapter/repository"
	"tmf/services/pocv/internal/core/domain"

	"gorm.io/gorm"
)

type OutboxPublisher struct {
	db *gorm.DB
}

func NewOutboxPublisher(db *gorm.DB) *OutboxPublisher {
	return &OutboxPublisher{db: db}
}

func (p *OutboxPublisher) getDB(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value("tx").(*gorm.DB)
	if ok {
		return tx
	}
	return p.db.WithContext(ctx)
}

func (p *OutboxPublisher) saveEvent(ctx context.Context, routingKey string, payload interface{}) error {
	// Add Standard Headers if needed (e.g., from context)
	headers := make(map[string]string)
	if user, ok := ctx.Value("user").(string); ok {
		headers["user"] = user
	}
	if auth, ok := ctx.Value("Authorization").(string); ok {
		headers["Authorization"] = auth
	}

	evt, err := repository.NewOutboxEvent(routingKey, payload, headers)
	if err != nil {
		return err
	}
	return p.getDB(ctx).Create(evt).Error
}

func (p *OutboxPublisher) PublishOrderCreated(ctx context.Context, evt domain.OrderCreatedEvent) error {
	return p.saveEvent(ctx, "evt.order.created", evt)
}

func (p *OutboxPublisher) PublishReserveInventory(ctx context.Context, cmd domain.ReserveInventoryCommand) error {
	return p.saveEvent(ctx, "cmd.inventory.reserve", cmd)
}

func (p *OutboxPublisher) PublishOrderCompleted(ctx context.Context, evt domain.OrderCompletedEvent) error {
	return p.saveEvent(ctx, "evt.order.completed", evt)
}

func (p *OutboxPublisher) PublishOrderFailed(ctx context.Context, evt domain.OrderFailedEvent) error {
	return p.saveEvent(ctx, "evt.order.failed", evt)
}
