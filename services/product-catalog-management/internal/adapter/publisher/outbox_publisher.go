package publisher

import (
	"context"
	"encoding/json"
	"tmf/services/product-catalog-management/internal/adapter/repository"
	"tmf/services/product-catalog-management/internal/core/domain"

	"gorm.io/gorm"
)

type OutboxPublisher struct {
	db *gorm.DB
}

func NewOutboxPublisher(db *gorm.DB) *OutboxPublisher {
	return &OutboxPublisher{db: db}
}

func (p *OutboxPublisher) saveEvent(ctx context.Context, routingKey string, event interface{}) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Extract headers
	headers := make(map[string]string)
	// Using hardcoded keys as domain constants might not be exported similarly
	if user, ok := ctx.Value("user").(string); ok {
		headers["user"] = user
	}
	if auth, ok := ctx.Value("Authorization").(string); ok {
		headers["Authorization"] = auth
	}
	// Fallback/standard keys if common package is used
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

	outboxEvent := &repository.OutboxEventModel{
		RoutingKey: routingKey,
		Payload:    payload,
		Headers:    headerBytes,
		Status:     repository.StatusPending,
	}

	tx := repository.GetTx(ctx, p.db)
	return tx.Create(outboxEvent).Error
}

func (p *OutboxPublisher) PublishCatalogCreated(ctx context.Context, event domain.CatalogCreatedEvent) error {
	return p.saveEvent(ctx, "evt.catalog.catalog.created", event)
}

func (p *OutboxPublisher) PublishCategoryCreated(ctx context.Context, event domain.CategoryCreatedEvent) error {
	return p.saveEvent(ctx, "evt.catalog.category.created", event)
}

func (p *OutboxPublisher) PublishProductSpecificationCreated(ctx context.Context, event domain.ProductSpecificationCreatedEvent) error {
	return p.saveEvent(ctx, "evt.catalog.specification.created", event)
}

func (p *OutboxPublisher) PublishProductOfferingCreated(ctx context.Context, event domain.ProductOfferingCreatedEvent) error {
	return p.saveEvent(ctx, "evt.catalog.offering.created", event)
}

func (p *OutboxPublisher) PublishCatalogUpdated(ctx context.Context, event domain.CatalogUpdatedEvent) error {
	return p.saveEvent(ctx, "evt.catalog.catalog.updated", event)
}

func (p *OutboxPublisher) PublishCatalogDeleted(ctx context.Context, event domain.CatalogDeletedEvent) error {
	return p.saveEvent(ctx, "evt.catalog.catalog.deleted", event)
}

func (p *OutboxPublisher) PublishCategoryUpdated(ctx context.Context, event domain.CategoryUpdatedEvent) error {
	return p.saveEvent(ctx, "evt.catalog.category.updated", event)
}

func (p *OutboxPublisher) PublishCategoryDeleted(ctx context.Context, event domain.CategoryDeletedEvent) error {
	return p.saveEvent(ctx, "evt.catalog.category.deleted", event)
}

func (p *OutboxPublisher) PublishProductSpecificationUpdated(ctx context.Context, event domain.ProductSpecificationUpdatedEvent) error {
	return p.saveEvent(ctx, "evt.catalog.specification.updated", event)
}

func (p *OutboxPublisher) PublishProductSpecificationDeleted(ctx context.Context, event domain.ProductSpecificationDeletedEvent) error {
	return p.saveEvent(ctx, "evt.catalog.specification.deleted", event)
}

func (p *OutboxPublisher) PublishProductOfferingUpdated(ctx context.Context, event domain.ProductOfferingUpdatedEvent) error {
	return p.saveEvent(ctx, "evt.catalog.offering.updated", event)
}

func (p *OutboxPublisher) PublishProductOfferingDeleted(ctx context.Context, event domain.ProductOfferingDeletedEvent) error {
	return p.saveEvent(ctx, "evt.catalog.offering.deleted", event)
}
