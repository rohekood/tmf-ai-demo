package handler

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"tmf/pkg/rabbitmq"
	"tmf/services/qualification/internal/core/domain"
)

type mockUseCase struct {
	called bool
	cmd    domain.CheckEligibilityCommand
}

func (m *mockUseCase) Execute(ctx context.Context, cmd domain.CheckEligibilityCommand) error {
	m.called = true
	m.cmd = cmd
	return nil
}

func TestRabbitMQHandler_HandleCheckCommand(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("Valid JSON", func(t *testing.T) {
		mockUC := &mockUseCase{}
		h := NewRabbitMQHandler(mockUC, logger)

		cmd := domain.CheckEligibilityCommand{
			CorrelationID: "test-123",
			Address:       domain.Address{City: "Berlin"},
		}
		payload, _ := json.Marshal(cmd)

		err := h.HandleCheckCommand(context.Background(), payload)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if !mockUC.called {
			t.Error("expected usecase to be called")
		}
		if mockUC.cmd.CorrelationID != "test-123" {
			t.Errorf("expected correlationID test-123, got %s", mockUC.cmd.CorrelationID)
		}
	})

	t.Run("Correlation ID From Context", func(t *testing.T) {
		mockUC := &mockUseCase{}
		h := NewRabbitMQHandler(mockUC, logger)

		//nolint:staticcheck // key is a string in this legacy system
		ctx := context.WithValue(context.Background(), rabbitmq.ContextKeyCorrelationID, "ctx-456")
		cmd := domain.CheckEligibilityCommand{Address: domain.Address{City: "Berlin"}}
		payload, _ := json.Marshal(cmd)

		err := h.HandleCheckCommand(ctx, payload)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if mockUC.cmd.CorrelationID != "ctx-456" {
			t.Errorf("expected correlationID ctx-456, got %s", mockUC.cmd.CorrelationID)
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		mockUC := &mockUseCase{}
		h := NewRabbitMQHandler(mockUC, logger)

		payload := []byte("{invalid-json")

		err := h.HandleCheckCommand(context.Background(), payload)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if mockUC.called {
			t.Error("expected usecase NOT to be called")
		}
	})
}
