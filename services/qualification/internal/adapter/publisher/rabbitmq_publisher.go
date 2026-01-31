package publisher

import (
	"context"
	"tmf/pkg/rabbitmq"
	"tmf/services/qualification/internal/core/domain"
)

type EventPublisher struct {
	publisher    rabbitmq.Publisher
	exchangeName string
}

func NewEventPublisher(pub rabbitmq.Publisher, exchangeName string) *EventPublisher {
	return &EventPublisher{
		publisher:    pub,
		exchangeName: exchangeName,
	}
}

func (p *EventPublisher) PublishEligibilityChecked(ctx context.Context, result domain.EligibilityResult) error {
	return p.publisher.Publish(ctx, p.exchangeName, "evt.qual.checked", result)
}
