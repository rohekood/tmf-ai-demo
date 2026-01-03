package publisher

import (
	"context"
	"encoding/json"
	"log"

	"tmf/services/product-catalog-management/internal/core/domain"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher struct {
	channel *amqp.Channel
}

func NewRabbitMQPublisher(conn *amqp.Connection) (*RabbitMQPublisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	// Declare exchange if not exists (Idempotent)
	err = ch.ExchangeDeclare(
		"catalog_events", // name
		"topic",          // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		return nil, err
	}
	return &RabbitMQPublisher{channel: ch}, nil
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
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	err = p.channel.PublishWithContext(ctx,
		"catalog_events", // exchange
		routingKey,       // routing key
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		log.Printf("Failed to publish event %s: %v", routingKey, err)
		return err
	}
	log.Printf("Published event %s", routingKey)
	return nil
}
