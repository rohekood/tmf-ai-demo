package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"tmf/pkg/rabbitmq"
	"tmf/services/customer-management/internal/domain"
)

// RPCHandler handles RPC requests for customer management
type RPCHandler struct {
	repo      domain.Repository
	publisher rabbitmq.Publisher
	logger    *slog.Logger
}

// NewRPCHandler creates a new RPC handler
func NewRPCHandler(repo domain.Repository, publisher rabbitmq.Publisher, logger *slog.Logger) *RPCHandler {
	return &RPCHandler{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

// HandleGetCustomer handles the query.customer.get RPC request
func (h *RPCHandler) HandleGetCustomer(ctx context.Context, payload []byte) error {
	var req struct {
		CustomerID string `json:"customerId"`
	}

	if err := json.Unmarshal(payload, &req); err != nil {
		return fmt.Errorf("failed to unmarshal request: %w", err)
	}

	h.logger.InfoContext(ctx, "Getting customer for pricing", "customerId", req.CustomerID)

	// Get customer from repository
	customer, err := h.repo.GetCustomer(ctx, req.CustomerID)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to get customer", "customerId", req.CustomerID, "error", err)
		return fmt.Errorf("customer not found: %w", err)
	}

	// Extract tier from customer characteristics
	tier := extractTier(customer)
	segment := extractSegment(customer)

	// Create response with tier information
	response := map[string]interface{}{
		"id":      customer.ID,
		"tier":    tier,
		"segment": segment,
		"status":  customer.Status,
	}

	// Marshal response
	respBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Send RPC reply
	replyTo := ctx.Value(rabbitmq.ContextKeyReplyTo)
	correlationID := ctx.Value(rabbitmq.ContextKeyAMQPCorrelationID)

	if replyTo == nil || correlationID == nil {
		return fmt.Errorf("missing replyTo or correlationId in context")
	}

	return h.publisher.PublishToQueue(ctx, replyTo.(string), correlationID.(string), respBytes)
}

// extractTier extracts tier from customer characteristics
func extractTier(customer *domain.Customer) string {
	for _, char := range customer.Characteristics {
		if char.Name == "tier" {
			return char.Value
		}
	}
	return "Standard" // Default tier
}

// extractSegment extracts segment from customer market segments
func extractSegment(customer *domain.Customer) string {
	if len(customer.MarketSegments) > 0 {
		return customer.MarketSegments[0].Name
	}
	return "Residential" // Default segment
}

// BindRPCHandlers binds RPC handlers to the consumer
func (h *RPCHandler) BindRPCHandlers(consumer rabbitmq.Consumer) error {
	// Bind query.customer.get
	if err := consumer.Subscribe("query.customer.get", h.HandleGetCustomer); err != nil {
		return fmt.Errorf("failed to bind query.customer.get: %w", err)
	}

	h.logger.Info("Customer RPC handlers bound successfully")
	return nil
}
