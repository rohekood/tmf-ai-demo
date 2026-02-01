package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"tmf/pkg/rabbitmq"
	"tmf/services/product-catalog-management/internal/core/ports"
)

// CatalogRPCHandler handles RPC requests for product catalog
type CatalogRPCHandler struct {
	offeringRepo ports.ProductOfferingRepository
	publisher    rabbitmq.Publisher
	logger       *slog.Logger
}

// NewCatalogRPCHandler creates a new RPC handler
func NewCatalogRPCHandler(offeringRepo ports.ProductOfferingRepository, publisher rabbitmq.Publisher, logger *slog.Logger) *CatalogRPCHandler {
	return &CatalogRPCHandler{
		offeringRepo: offeringRepo,
		publisher:    publisher,
		logger:       logger,
	}
}

// HandleGetOffering handles the query.catalog.offering.get RPC request
func (h *CatalogRPCHandler) HandleGetOffering(ctx context.Context, payload []byte) error {
	var req struct {
		OfferingID string `json:"offeringId"`
	}

	if err := json.Unmarshal(payload, &req); err != nil {
		return fmt.Errorf("failed to unmarshal request: %w", err)
	}

	h.logger.InfoContext(ctx, "Getting offering for pricing", "offeringId", req.OfferingID)

	// Get offering from repository
	offering, err := h.offeringRepo.Get(ctx, req.OfferingID)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to get offering", "offeringId", req.OfferingID, "error", err)
		return fmt.Errorf("offering not found: %w", err)
	}

	// Extract base price from offering prices
	var basePrice float64 = 100.0 // Default
	var currency string = "EUR"   // Default

	if len(offering.ProductOfferingPrice) > 0 {
		// Use the first price as base price
		basePrice = offering.ProductOfferingPrice[0].Price.Value
		currency = offering.ProductOfferingPrice[0].Price.Unit
	}

	// Create response with base price
	response := map[string]interface{}{
		"id":        offering.ID,
		"name":      offering.Name,
		"basePrice": basePrice,
		"currency":  currency,
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

// BindRPCHandlers binds RPC handlers to the consumer
func (h *CatalogRPCHandler) BindRPCHandlers(consumer rabbitmq.Consumer) error {
	// Bind query.catalog.offering.get
	if err := consumer.Subscribe("query.catalog.offering.get", h.HandleGetOffering); err != nil {
		return fmt.Errorf("failed to bind query.catalog.offering.get: %w", err)
	}

	h.logger.Info("Catalog RPC handlers bound successfully")
	return nil
}
