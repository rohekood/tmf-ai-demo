package publisher

import (
	"context"
	"encoding/json"
	"log"

	"tmf/pkg/rabbitmq"
	"tmf/services/product-catalog-management/internal/core/domain"
)

type RabbitMQPublisher struct {
	publisher    rabbitmq.Publisher
	exchangeName string
}

func NewRabbitMQPublisher(publisher rabbitmq.Publisher, exchangeName string) (*RabbitMQPublisher, error) {
	return &RabbitMQPublisher{
		publisher:    publisher,
		exchangeName: exchangeName,
	}, nil
}

func (p *RabbitMQPublisher) PublishCatalogCreated(ctx context.Context, event domain.CatalogCreatedEvent) error {
	return p.publish(ctx, "evt.catalog.catalog.created", event)
}

func (p *RabbitMQPublisher) PublishCategoryCreated(ctx context.Context, event domain.CategoryCreatedEvent) error {
	return p.publish(ctx, "evt.catalog.category.created", event)
}

func (p *RabbitMQPublisher) PublishProductSpecificationCreated(ctx context.Context, event domain.ProductSpecificationCreatedEvent) error {
	return p.publish(ctx, "evt.catalog.specification.created", event)
}

func (p *RabbitMQPublisher) PublishProductOfferingCreated(ctx context.Context, event domain.ProductOfferingCreatedEvent) error {
	return p.publish(ctx, "evt.catalog.offering.created", event)
}

func (p *RabbitMQPublisher) PublishCatalogUpdated(ctx context.Context, event domain.CatalogUpdatedEvent) error {
	return p.publish(ctx, "evt.catalog.catalog.updated", event)
}

func (p *RabbitMQPublisher) PublishCatalogDeleted(ctx context.Context, event domain.CatalogDeletedEvent) error {
	return p.publish(ctx, "evt.catalog.catalog.deleted", event)
}

func (p *RabbitMQPublisher) PublishCategoryUpdated(ctx context.Context, event domain.CategoryUpdatedEvent) error {
	return p.publish(ctx, "evt.catalog.category.updated", event)
}

func (p *RabbitMQPublisher) PublishCategoryDeleted(ctx context.Context, event domain.CategoryDeletedEvent) error {
	return p.publish(ctx, "evt.catalog.category.deleted", event)
}

func (p *RabbitMQPublisher) PublishProductSpecificationUpdated(ctx context.Context, event domain.ProductSpecificationUpdatedEvent) error {
	return p.publish(ctx, "evt.catalog.specification.updated", event)
}

func (p *RabbitMQPublisher) PublishProductSpecificationDeleted(ctx context.Context, event domain.ProductSpecificationDeletedEvent) error {
	return p.publish(ctx, "evt.catalog.specification.deleted", event)
}

func (p *RabbitMQPublisher) PublishProductOfferingUpdated(ctx context.Context, event domain.ProductOfferingUpdatedEvent) error {
	return p.publish(ctx, "evt.catalog.offering.updated", event)
}

func (p *RabbitMQPublisher) PublishProductOfferingDeleted(ctx context.Context, event domain.ProductOfferingDeletedEvent) error {
	return p.publish(ctx, "evt.catalog.offering.deleted", event)
}

func (p *RabbitMQPublisher) publish(ctx context.Context, routingKey string, event interface{}) error {
	return p.publisher.Publish(ctx, p.exchangeName, routingKey, event)
}

func (p *RabbitMQPublisher) PublishRaw(ctx context.Context, routingKey string, body []byte) error {
	err := p.publisher.Publish(ctx, p.exchangeName, routingKey, json.RawMessage(body))
	if err != nil {
		log.Printf("Failed to publish event %s: %v", routingKey, err)
		return err
	}
	log.Printf("Published event %s", routingKey)
	return nil
}
