package ports

import (
	"context"
	"tmf/services/pocv/internal/core/domain"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}

type EventPublisher interface {
	PublishOrderCreated(ctx context.Context, evt domain.OrderCreatedEvent) error
	PublishReserveInventory(ctx context.Context, cmd domain.ReserveInventoryCommand) error
	PublishOrderCompleted(ctx context.Context, evt domain.OrderCompletedEvent) error
	PublishOrderFailed(ctx context.Context, evt domain.OrderFailedEvent) error
}

type OrderUseCase interface {
	SubmitOrder(ctx context.Context, cmd domain.SubmitOrderCommand) (*domain.Order, error)
	HandleInventoryReserved(ctx context.Context, orderID string) error
	HandleInventoryFailed(ctx context.Context, orderID string, reason string) error
}

type SagaRepository interface {
	GetByCartID(ctx context.Context, cartID string) (*domain.SagaInstance, error)
	Create(ctx context.Context, saga *domain.SagaInstance, events []domain.OutboxEvent) error
	Get(ctx context.Context, id string) (*domain.SagaInstance, error)
	Update(ctx context.Context, saga *domain.SagaInstance, events []domain.OutboxEvent) error
}

type CartClient interface {
	GetCart(ctx context.Context, id string) (map[string]interface{}, error)
}

type SagaUseCase interface {
	StartSaga(ctx context.Context, cartID string) error
	HandleInventoryReserved(ctx context.Context, sagaID string) error
	HandleInventoryFailed(ctx context.Context, sagaID string) error
	HandlePaymentAuthorized(ctx context.Context, sagaID string) error
	HandlePaymentDeclined(ctx context.Context, sagaID string) error
	HandleOrderCreated(ctx context.Context, sagaID string) error
	GetSaga(ctx context.Context, id string) (*domain.SagaInstance, error)
}
