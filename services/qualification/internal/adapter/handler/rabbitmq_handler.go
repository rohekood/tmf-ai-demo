package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"tmf/pkg/rabbitmq"
	"tmf/services/qualification/internal/core/domain"
	"tmf/services/qualification/internal/core/ports"
)

type RabbitMQHandler struct {
	useCase ports.CheckEligibilityUseCase
	logger  *slog.Logger
}

func NewRabbitMQHandler(uc ports.CheckEligibilityUseCase, logger *slog.Logger) *RabbitMQHandler {
	return &RabbitMQHandler{
		useCase: uc,
		logger:  logger,
	}
}

func (h *RabbitMQHandler) HandleCheckCommand(ctx context.Context, payload []byte) error {
	var cmd domain.CheckEligibilityCommand
	if err := json.Unmarshal(payload, &cmd); err != nil {
		h.logger.Error("Invalid JSON payload", "error", err)
		return nil
	}

	if val, ok := ctx.Value(rabbitmq.ContextKeyCorrelationID).(string); ok && cmd.CorrelationID == "" {
		cmd.CorrelationID = val
	}

	h.logger.Info("Received Check Command", "address_street", cmd.Address.Street, "address_city", cmd.Address.City)
	return h.useCase.Execute(ctx, cmd)
}
