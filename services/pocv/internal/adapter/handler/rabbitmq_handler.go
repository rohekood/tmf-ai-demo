package handler

import (
	"context"
	"encoding/json"
	"log"

	"tmf/pkg/rabbitmq"
	"tmf/services/pocv/internal/core/domain"
	"tmf/services/pocv/internal/core/ports"
)

type RabbitMQHandler struct {
	useCase ports.SagaUseCase
}

func NewRabbitMQHandler(uc ports.SagaUseCase) *RabbitMQHandler {
	return &RabbitMQHandler{useCase: uc}
}

// Dispatcher
func (h *RabbitMQHandler) HandleSagaEvent(ctx context.Context, payload []byte) error {
	routingKey, _ := ctx.Value(rabbitmq.ContextKeyRoutingKey).(string)
	// log.Printf("POCV: Dispatching event: %s", routingKey)

	switch routingKey {
	case rabbitmq.CmdOrderCheckoutSubmit:
		return h.HandleSubmitOrder(ctx, payload)
	case rabbitmq.EvtInventoryResourceReserved:
		return h.HandleInventoryReserved(ctx, payload)
	case rabbitmq.EvtInventoryResourceFailed:
		return h.HandleInventoryFailed(ctx, payload)
	case rabbitmq.EvtPaymentTransactionAuthorized:
		return h.HandlePaymentAuthorized(ctx, payload)
	case rabbitmq.EvtPaymentTransactionDeclined:
		return h.HandlePaymentDeclined(ctx, payload)
	case rabbitmq.EvtOrderManagementCreated:
		return h.HandleOrderCreated(ctx, payload)
	}
	return nil
}

func (h *RabbitMQHandler) HandleSubmitOrder(ctx context.Context, payload []byte) error {
	var cmd domain.SubmitOrderCommand
	if err := json.Unmarshal(payload, &cmd); err != nil {
		log.Printf("POCV: Error unmarshaling order command: %v", err)
		return nil // Don't retry malformed
	}

	// Start Saga
	if err := h.useCase.StartSaga(ctx, cmd.CartID); err != nil {
		log.Printf("POCV: Error starting saga: %v", err)
		return err // Retry
	}

	log.Printf("POCV: Saga Started for Cart: %s", cmd.CartID)
	return nil
}

func (h *RabbitMQHandler) HandleInventoryReserved(ctx context.Context, payload []byte) error {
	type InventoryReservedEvent struct {
		OrderID string `json:"orderId"` // Assuming OrderID maps to SagaID
	}
	var evt InventoryReservedEvent
	if err := json.Unmarshal(payload, &evt); err != nil {
		log.Printf("POCV: Error unmarshaling inventory event: %v", err)
		return nil
	}

	if err := h.useCase.HandleInventoryReserved(ctx, evt.OrderID); err != nil {
		log.Printf("POCV: Error processing inventory reserved: %v", err)
		return err
	}

	log.Printf("POCV: Inventory Processed for Saga: %s", evt.OrderID)
	return nil
}

func (h *RabbitMQHandler) HandleInventoryFailed(ctx context.Context, payload []byte) error {
	type InventoryFailedEvent struct {
		OrderID string `json:"orderId"`
		Reason  string `json:"reason"`
	}
	var evt InventoryFailedEvent
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil
	}

	// Note: SagaUseCase.HandleInventoryFailed only takes sagaID in ports.go?
	// Let's check ports.go: HandleInventoryFailed(ctx context.Context, sagaID string) error
	// It ignores reason? Or maybe I should log it.

	if err := h.useCase.HandleInventoryFailed(ctx, evt.OrderID); err != nil {
		return err
	}
	return nil
}

func (h *RabbitMQHandler) HandlePaymentAuthorized(ctx context.Context, payload []byte) error {
	type PaymentAuthorizedEvent struct {
		SagaID string `json:"sagaId"`
	}
	var evt PaymentAuthorizedEvent
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil
	}

	if err := h.useCase.HandlePaymentAuthorized(ctx, evt.SagaID); err != nil {
		return err
	}
	return nil
}

func (h *RabbitMQHandler) HandlePaymentDeclined(ctx context.Context, payload []byte) error {
	type PaymentDeclinedEvent struct {
		SagaID string `json:"sagaId"`
	}
	var evt PaymentDeclinedEvent
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil
	}

	if err := h.useCase.HandlePaymentDeclined(ctx, evt.SagaID); err != nil {
		return err
	}
	return nil
}

func (h *RabbitMQHandler) HandleOrderCreated(ctx context.Context, payload []byte) error {
	type OrderCreatedEvent struct {
		OrderID string `json:"orderId"`
	}
	var evt OrderCreatedEvent
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil
	}

	if err := h.useCase.HandleOrderCreated(ctx, evt.OrderID); err != nil {
		return err
	}
	return nil
}
