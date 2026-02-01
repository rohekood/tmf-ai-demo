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
		return fmt.Errorf("session not found: %w", err)
	}

	// Validate not expired
	if time.Now().After(session.ExpiresAt) {
		h.logger.WarnContext(ctx, "Session expired", "sessionId", req.SessionID, "expiresAt", session.ExpiresAt)
		return fmt.Errorf("session expired")
	}

	// Marshal response
	respBytes, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Send RPC reply
	replyTo := ctx.Value(rabbitmq.ContextKeyReplyTo)
	correlationID := ctx.Value(rabbitmq.ContextKeyAMQPCorrelationID)

	if replyTo == nil || correlationID == nil {
		return fmt.Errorf("missing replyTo or correlationId in context")
	}

	return h.publisher.PublishToQueue(ctx, replyTo.(string), correlationID.(string), respBytes)
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
