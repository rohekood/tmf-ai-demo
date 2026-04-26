package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"tmf/services/pocv/internal/core/domain"
)

type mockSagaRepository struct {
	getByCartID func(ctx context.Context, cartID string) (*domain.SagaInstance, error)
	create      func(ctx context.Context, saga *domain.SagaInstance, events []domain.OutboxEvent) error
	get         func(ctx context.Context, id string) (*domain.SagaInstance, error)
	update      func(ctx context.Context, saga *domain.SagaInstance, events []domain.OutboxEvent) error
}

func (m *mockSagaRepository) GetByCartID(ctx context.Context, cartID string) (*domain.SagaInstance, error) {
	if m.getByCartID != nil {
		return m.getByCartID(ctx, cartID)
	}
	return nil, nil
}

func (m *mockSagaRepository) Create(ctx context.Context, saga *domain.SagaInstance, events []domain.OutboxEvent) error {
	if m.create != nil {
		return m.create(ctx, saga, events)
	}
	return nil
}

func (m *mockSagaRepository) Get(ctx context.Context, id string) (*domain.SagaInstance, error) {
	if m.get != nil {
		return m.get(ctx, id)
	}
	return nil, nil
}

func (m *mockSagaRepository) Update(ctx context.Context, saga *domain.SagaInstance, events []domain.OutboxEvent) error {
	if m.update != nil {
		return m.update(ctx, saga, events)
	}
	return nil
}

type mockCartClient struct {
	getCart func(ctx context.Context, id string) (map[string]interface{}, error)
}

func (m *mockCartClient) GetCart(ctx context.Context, id string) (map[string]interface{}, error) {
	if m.getCart != nil {
		return m.getCart(ctx, id)
	}
	return nil, nil
}

func TestStartSaga(t *testing.T) {
	ctx := context.Background()
	cartID := "cart-123"

	tests := []struct {
		name        string
		repo        *mockSagaRepository
		cartClient  *mockCartClient
		expectError bool
	}{
		{
			name: "Success",
			repo: &mockSagaRepository{
				getByCartID: func(ctx context.Context, cartID string) (*domain.SagaInstance, error) {
					return nil, nil
				},
				create: func(ctx context.Context, saga *domain.SagaInstance, events []domain.OutboxEvent) error {
					return nil
				},
			},
			cartClient: &mockCartClient{
				getCart: func(ctx context.Context, id string) (map[string]interface{}, error) {
					return map[string]interface{}{
						"items": []interface{}{"item1", "item2"},
					}, nil
				},
			},
			expectError: false,
		},
		{
			name: "Existing Saga",
			repo: &mockSagaRepository{
				getByCartID: func(ctx context.Context, cartID string) (*domain.SagaInstance, error) {
					return &domain.SagaInstance{ID: "existing-id"}, nil
				},
			},
			cartClient:  &mockCartClient{},
			expectError: false,
		},
		{
			name: "GetByCartID Error",
			repo: &mockSagaRepository{
				getByCartID: func(ctx context.Context, cartID string) (*domain.SagaInstance, error) {
					return nil, errors.New("db error")
				},
			},
			cartClient:  &mockCartClient{},
			expectError: true,
		},
		{
			name: "GetCart Error",
			repo: &mockSagaRepository{
				getByCartID: func(ctx context.Context, cartID string) (*domain.SagaInstance, error) {
					return nil, nil
				},
			},
			cartClient: &mockCartClient{
				getCart: func(ctx context.Context, id string) (map[string]interface{}, error) {
					return nil, errors.New("client error")
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewSagaUseCase(tt.repo, tt.cartClient)
			err := uc.StartSaga(ctx, cartID)
			if (err != nil) != tt.expectError {
				t.Errorf("StartSaga() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestHandleInventoryReserved(t *testing.T) {
	ctx := context.Background()
	sagaID := "saga-123"

	tests := []struct {
		name        string
		repo        *mockSagaRepository
		expectError bool
	}{
		{
			name: "Success",
			repo: &mockSagaRepository{
				get: func(ctx context.Context, id string) (*domain.SagaInstance, error) {
					payload, _ := json.Marshal(map[string]interface{}{"totalPriceAmount": float64(100)})
					return &domain.SagaInstance{
						ID:          sagaID,
						CurrentStep: domain.StepInventory,
						Payload:     payload,
					}, nil
				},
				update: func(ctx context.Context, saga *domain.SagaInstance, events []domain.OutboxEvent) error {
					if saga.CurrentStep != domain.StepPayment {
						return errors.New("wrong step")
					}
					if len(events) != 1 {
						return errors.New("missing event")
					}
					return nil
				},
			},
			expectError: false,
		},
		{
			name: "Get Error",
			repo: &mockSagaRepository{
				get: func(ctx context.Context, id string) (*domain.SagaInstance, error) {
					return nil, errors.New("db error")
				},
			},
			expectError: true,
		},
		{
			name: "Wrong Step (Idempotency)",
			repo: &mockSagaRepository{
				get: func(ctx context.Context, id string) (*domain.SagaInstance, error) {
					return &domain.SagaInstance{
						ID:          sagaID,
						CurrentStep: domain.StepPayment,
					}, nil
				},
			},
			expectError: false, // returns nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewSagaUseCase(tt.repo, nil)
			err := uc.HandleInventoryReserved(ctx, sagaID)
			if (err != nil) != tt.expectError {
				t.Errorf("HandleInventoryReserved() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestHandleInventoryFailed(t *testing.T) {
	ctx := context.Background()
	sagaID := "saga-123"

	tests := []struct {
		name        string
		repo        *mockSagaRepository
		expectError bool
	}{
		{
			name: "Success",
			repo: &mockSagaRepository{
				get: func(ctx context.Context, id string) (*domain.SagaInstance, error) {
					return &domain.SagaInstance{ID: sagaID}, nil
				},
				update: func(ctx context.Context, saga *domain.SagaInstance, events []domain.OutboxEvent) error {
					return nil
				},
			},
			expectError: false,
		},
		{
			name: "Get Error",
			repo: &mockSagaRepository{
				get: func(ctx context.Context, id string) (*domain.SagaInstance, error) {
					return nil, errors.New("db error")
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewSagaUseCase(tt.repo, nil)
			err := uc.HandleInventoryFailed(ctx, sagaID)
			if (err != nil) != tt.expectError {
				t.Errorf("HandleInventoryFailed() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestHandlePaymentAuthorized(t *testing.T) {
	ctx := context.Background()
	sagaID := "saga-123"

	tests := []struct {
		name        string
		repo        *mockSagaRepository
		expectError bool
	}{
		{
			name: "Success",
			repo: &mockSagaRepository{
				get: func(ctx context.Context, id string) (*domain.SagaInstance, error) {
					return &domain.SagaInstance{
						ID:          sagaID,
						CurrentStep: domain.StepPayment,
						Payload:     []byte(`{}`),
					}, nil
				},
				update: func(ctx context.Context, saga *domain.SagaInstance, events []domain.OutboxEvent) error {
					return nil
				},
			},
			expectError: false,
		},
		{
			name: "Wrong Step",
			repo: &mockSagaRepository{
				get: func(ctx context.Context, id string) (*domain.SagaInstance, error) {
					return &domain.SagaInstance{
						ID:          sagaID,
						CurrentStep: domain.StepInventory,
					}, nil
				},
			},
			expectError: false,
		},
		{
			name: "Get Error",
			repo: &mockSagaRepository{
				get: func(ctx context.Context, id string) (*domain.SagaInstance, error) {
					return nil, errors.New("db error")
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewSagaUseCase(tt.repo, nil)
			err := uc.HandlePaymentAuthorized(ctx, sagaID)
			if (err != nil) != tt.expectError {
				t.Errorf("HandlePaymentAuthorized() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestHandlePaymentDeclined(t *testing.T) {
	ctx := context.Background()
	sagaID := "saga-123"

	tests := []struct {
		name        string
		repo        *mockSagaRepository
		expectError bool
	}{
		{
			name: "Success",
			repo: &mockSagaRepository{
				get: func(ctx context.Context, id string) (*domain.SagaInstance, error) {
					return &domain.SagaInstance{ID: sagaID}, nil
				},
				update: func(ctx context.Context, saga *domain.SagaInstance, events []domain.OutboxEvent) error {
					return nil
				},
			},
			expectError: false,
		},
		{
			name: "Get Error",
			repo: &mockSagaRepository{
				get: func(ctx context.Context, id string) (*domain.SagaInstance, error) {
					return nil, errors.New("db error")
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewSagaUseCase(tt.repo, nil)
			err := uc.HandlePaymentDeclined(ctx, sagaID)
			if (err != nil) != tt.expectError {
				t.Errorf("HandlePaymentDeclined() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestHandleOrderCreated(t *testing.T) {
	ctx := context.Background()
	sagaID := "saga-123"

	tests := []struct {
		name        string
		repo        *mockSagaRepository
		expectError bool
	}{
		{
			name: "Success",
			repo: &mockSagaRepository{
				get: func(ctx context.Context, id string) (*domain.SagaInstance, error) {
					return &domain.SagaInstance{ID: sagaID}, nil
				},
				update: func(ctx context.Context, saga *domain.SagaInstance, events []domain.OutboxEvent) error {
					return nil
				},
			},
			expectError: false,
		},
		{
			name: "Get Error",
			repo: &mockSagaRepository{
				get: func(ctx context.Context, id string) (*domain.SagaInstance, error) {
					return nil, errors.New("db error")
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewSagaUseCase(tt.repo, nil)
			err := uc.HandleOrderCreated(ctx, sagaID)
			if (err != nil) != tt.expectError {
				t.Errorf("HandleOrderCreated() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestGetSaga(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		id          string
		repo        *mockSagaRepository
		expectError bool
		expectNil   bool
	}{
		{
			name: "Success",
			id:   "s1",
			repo: &mockSagaRepository{
				get: func(ctx context.Context, id string) (*domain.SagaInstance, error) {
					return &domain.SagaInstance{ID: id}, nil
				},
			},
			expectError: false,
			expectNil:   false,
		},
		{
			name: "NotFound",
			id:   "s2",
			repo: &mockSagaRepository{
				get: func(ctx context.Context, id string) (*domain.SagaInstance, error) {
					return nil, nil
				},
			},
			expectError: false,
			expectNil:   true,
		},
		{
			name: "Error",
			id:   "s3",
			repo: &mockSagaRepository{
				get: func(ctx context.Context, id string) (*domain.SagaInstance, error) {
					return nil, errors.New("db error")
				},
			},
			expectError: true,
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewSagaUseCase(tt.repo, nil)
			res, err := uc.GetSaga(ctx, tt.id)
			if (err != nil) != tt.expectError {
				t.Errorf("GetSaga() error = %v, expectError %v", err, tt.expectError)
			}
			if tt.expectNil && res != nil {
				t.Errorf("GetSaga() expected nil result, got %v", res)
			}
			if !tt.expectNil && res == nil {
				t.Errorf("GetSaga() expected non-nil result")
			}
		})
	}
}
