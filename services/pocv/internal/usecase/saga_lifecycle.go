package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"tmf/pkg/rabbitmq"
	"tmf/services/pocv/internal/core/domain"
	"tmf/services/pocv/internal/core/ports"

	"github.com/google/uuid"
)

type sagaUseCase struct {
	repo       ports.SagaRepository
	cartClient ports.CartClient
}

func NewSagaUseCase(repo ports.SagaRepository, cartClient ports.CartClient) ports.SagaUseCase {
	return &sagaUseCase{
		repo:       repo,
		cartClient: cartClient,
	}
}

func (u *sagaUseCase) StartSaga(ctx context.Context, cartID string) error {
	existing, err := u.repo.GetByCartID(ctx, cartID)
	if err != nil {
		return err
	}
	if existing != nil {
		slog.InfoContext(ctx, "Saga already exists", "sagaId", existing.ID)
		return nil
	}

	cart, err := u.cartClient.GetCart(ctx, cartID)
	if err != nil {
		return fmt.Errorf("failed to fetch cart: %w", err)
	}

	cartJSON, _ := json.Marshal(cart)
	sagaID := uuid.New().String()

	// Create Command: Reserve Inventory
	cmdPayload := map[string]interface{}{
		"sagaId": sagaID,
		"items":  cart["items"],
	}
	cmdJSON, _ := json.Marshal(cmdPayload)

	events := []domain.OutboxEvent{{
		ID:        uuid.New().String(),
		Topic:     rabbitmq.CmdInventoryResourceReserve,
		Payload:   cmdJSON,
		Status:    "PENDING",
		CreatedAt: time.Now().UTC(),
	}}

	saga := &domain.SagaInstance{
		ID:          sagaID,
		CartID:      cartID,
		Status:      domain.SagaStatusInProgress,
		CurrentStep: domain.StepInventory,
		Payload:     cartJSON,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	return u.repo.Create(ctx, saga, events)
}

func (u *sagaUseCase) HandleInventoryReserved(ctx context.Context, sagaID string) error {
	saga, err := u.repo.Get(ctx, sagaID)
	if err != nil || saga == nil {
		return err
	}

	if saga.CurrentStep != domain.StepInventory {
		return nil // Idempotency
	}

	// Transition to Payment
	saga.CurrentStep = domain.StepPayment
	saga.UpdatedAt = time.Now().UTC()

	// Command: Auth Payment
	// Calculate total from Payload (Cart)
	var cart map[string]interface{}
	_ = json.Unmarshal(saga.Payload, &cart)

	cmdPayload := map[string]interface{}{
		"sagaId": sagaID,
		"amount": cart["totalPriceAmount"],
		"token":  "tok_visa_mock", // Mock token
	}
	cmdJSON, _ := json.Marshal(cmdPayload)

	events := []domain.OutboxEvent{{
		ID:        uuid.New().String(),
		Topic:     rabbitmq.CmdPaymentTransactionAuthorize,
		Payload:   cmdJSON,
		Status:    "PENDING",
		CreatedAt: time.Now(),
	}}

	return u.repo.Update(ctx, saga, events)
}

func (u *sagaUseCase) HandleInventoryFailed(ctx context.Context, sagaID string) error {
	saga, err := u.repo.Get(ctx, sagaID)
	if err != nil || saga == nil {
		return err
	}

	saga.Status = domain.SagaStatusFailed
	saga.UpdatedAt = time.Now().UTC()

	// No compensation needed for Inventory Fail (nothing reserved)
	return u.repo.Update(ctx, saga, nil)
}

func (u *sagaUseCase) HandlePaymentAuthorized(ctx context.Context, sagaID string) error {
	saga, err := u.repo.Get(ctx, sagaID)
	if err != nil || saga == nil {
		return err
	}

	if saga.CurrentStep != domain.StepPayment {
		return nil
	}

	// Transition to Order Creation
	saga.CurrentStep = domain.StepOrderCreation
	saga.UpdatedAt = time.Now().UTC()

	// Command: Create Order
	// Pass full cart payload to Ordering Service
	events := []domain.OutboxEvent{{
		ID:        uuid.New().String(),
		Topic:     rabbitmq.CmdOrderManagementCreate,
		Payload:   saga.Payload, // Cart is the order payload
		Status:    "PENDING",
		CreatedAt: time.Now().UTC(),
	}}

	return u.repo.Update(ctx, saga, events)
}

func (u *sagaUseCase) HandlePaymentDeclined(ctx context.Context, sagaID string) error {
	saga, err := u.repo.Get(ctx, sagaID)
	if err != nil || saga == nil {
		return err
	}

	saga.Status = domain.SagaStatusCompensating
	saga.UpdatedAt = time.Now().UTC()

	// Compensation: Release Inventory
	cmdPayload := map[string]interface{}{
		"sagaId": sagaID,
		"reason": "PaymentDeclined",
	}
	cmdJSON, _ := json.Marshal(cmdPayload)

	events := []domain.OutboxEvent{{
		ID:        uuid.New().String(),
		Topic:     rabbitmq.CmdInventoryResourceRelease,
		Payload:   cmdJSON,
		Status:    "PENDING",
		CreatedAt: time.Now(),
	}}

	return u.repo.Update(ctx, saga, events)
}

func (u *sagaUseCase) HandleOrderCreated(ctx context.Context, sagaID string) error {
	saga, err := u.repo.Get(ctx, sagaID)
	if err != nil || saga == nil {
		return err
	}

	saga.Status = domain.SagaStatusCompleted
	saga.UpdatedAt = time.Now().UTC()

	return u.repo.Update(ctx, saga, nil)
}

func (u *sagaUseCase) GetSaga(ctx context.Context, id string) (*domain.SagaInstance, error) {
	return u.repo.Get(ctx, id)
}
