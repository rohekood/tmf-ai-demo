package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"tmf/pkg/rabbitmq"
	"tmf/services/pocv/internal/core/ports"
)

// RPCHandler handles RPC requests for POCV saga queries.
type RPCHandler struct {
	useCase   ports.SagaUseCase
	publisher rabbitmq.Publisher
	logger    *slog.Logger
}

// NewRPCHandler creates a new RPCHandler.
func NewRPCHandler(uc ports.SagaUseCase, pub rabbitmq.Publisher, logger *slog.Logger) *RPCHandler {
	return &RPCHandler{
		useCase:   uc,
		publisher: pub,
		logger:    logger,
	}
}

// HandleGetSaga handles the query.pocv.saga.get RPC request.
func (h *RPCHandler) HandleGetSaga(ctx context.Context, payload []byte) error {
	var req struct {
		SagaID string `json:"sagaId"`
	}

	if err := json.Unmarshal(payload, &req); err != nil {
		return fmt.Errorf("failed to unmarshal request: %w", err)
	}

	h.logger.InfoContext(ctx, "Getting POCV saga", "sagaId", req.SagaID)

	saga, err := h.useCase.GetSaga(ctx, req.SagaID)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to get saga", "sagaId", req.SagaID, "error", err)
		return h.replyError(ctx, "saga not found")
	}

	replyTo, correlationID, err := h.rpcContextValues(ctx)
	if err != nil {
		return err
	}

	return h.publisher.PublishToQueue(ctx, replyTo, correlationID, saga)
}

// replyError sends an error response to the caller.
func (h *RPCHandler) replyError(ctx context.Context, msg string) error {
	replyTo, correlationID, err := h.rpcContextValues(ctx)
	if err != nil {
		return err
	}

	errResponse := map[string]string{"error": msg}
	return h.publisher.PublishToQueue(ctx, replyTo, correlationID, errResponse)
}

// rpcContextValues extracts the RPC reply-to queue and correlation ID from ctx.
func (h *RPCHandler) rpcContextValues(ctx context.Context) (replyTo, correlationID string, err error) {
	replyTo, _ = ctx.Value(rabbitmq.ContextKeyReplyTo).(string)
	correlationID, _ = ctx.Value(rabbitmq.ContextKeyAMQPCorrelationID).(string)

	if replyTo == "" {
		return "", "", fmt.Errorf("missing replyTo in context")
	}
	return replyTo, correlationID, nil
}

// BindRPCHandlers registers all RPC handlers on the given consumer.
func (h *RPCHandler) BindRPCHandlers(consumer rabbitmq.Consumer) error {
	if err := consumer.Subscribe("query.pocv.saga.get", h.HandleGetSaga); err != nil {
		return fmt.Errorf("failed to bind query.pocv.saga.get: %w", err)
	}

	h.logger.Info("POCV RPC handlers bound successfully")
	return nil
}
