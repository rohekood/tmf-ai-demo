package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"tmf/pkg/rabbitmq"
	"tmf/services/qualification/internal/core/ports"
)

// RPCHandler handles RPC requests for qualification sessions
type RPCHandler struct {
	sessionRepo ports.SessionRepository
	publisher   rabbitmq.Publisher
	logger      *slog.Logger
}

// NewRPCHandler creates a new RPC handler
func NewRPCHandler(sessionRepo ports.SessionRepository, publisher rabbitmq.Publisher, logger *slog.Logger) *RPCHandler {
	return &RPCHandler{
		sessionRepo: sessionRepo,
		publisher:   publisher,
		logger:      logger,
	}
}

// HandleGetSession handles the query.qual.session.get RPC request
func (h *RPCHandler) HandleGetSession(ctx context.Context, payload []byte) error {
	var req struct {
		SessionID string `json:"sessionId"`
	}

	if err := json.Unmarshal(payload, &req); err != nil {
		return fmt.Errorf("failed to unmarshal request: %w", err)
	}

	h.logger.InfoContext(ctx, "Getting qualification session", "sessionId", req.SessionID)

	// Get session from database
	session, err := h.sessionRepo.Get(ctx, req.SessionID)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to get session", "sessionId", req.SessionID, "error", err)
		return h.replyError(ctx, err)
	}

	// Validate not expired
	if time.Now().UTC().After(session.ExpiresAt) {
		h.logger.WarnContext(ctx, "Session expired",
			"sessionId", req.SessionID,
			"expiresAt", session.ExpiresAt,
			"now", time.Now().UTC(),
		)
		return h.replyError(ctx, fmt.Errorf("session expired: now=%v, expires=%v", time.Now().UTC(), session.ExpiresAt))
	}

	// Send RPC reply
	replyTo := ctx.Value(rabbitmq.ContextKeyReplyTo)
	correlationID := ctx.Value(rabbitmq.ContextKeyAMQPCorrelationID)

	if replyTo == nil || correlationID == nil {
		return fmt.Errorf("missing replyTo or correlationId in context")
	}

	return h.publisher.PublishToQueue(ctx, replyTo.(string), correlationID.(string), session)
}

// replyError sends an error response
func (h *RPCHandler) replyError(ctx context.Context, err error) error {
	replyTo := ctx.Value(rabbitmq.ContextKeyReplyTo)
	correlationID := ctx.Value(rabbitmq.ContextKeyAMQPCorrelationID)

	if replyTo == nil || correlationID == nil {
		return fmt.Errorf("missing replyTo or correlationId in context")
	}

	errResponse := map[string]string{"error": err.Error()}
	return h.publisher.PublishToQueue(ctx, replyTo.(string), correlationID.(string), errResponse)
}

// BindRPCHandlers binds RPC handlers to the consumer
func (h *RPCHandler) BindRPCHandlers(consumer rabbitmq.Consumer) error {
	// Bind query.qual.session.get
	if err := consumer.Subscribe("query.qual.session.get", h.HandleGetSession); err != nil {
		return fmt.Errorf("failed to bind query.qual.session.get: %w", err)
	}

	h.logger.Info("RPC handlers bound successfully")
	return nil
}
