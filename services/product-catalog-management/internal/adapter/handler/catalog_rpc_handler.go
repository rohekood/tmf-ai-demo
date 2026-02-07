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
	// Send RPC reply
	replyTo := ctx.Value(rabbitmq.ContextKeyReplyTo)
	correlationID := ctx.Value(rabbitmq.ContextKeyAMQPCorrelationID)

	if replyTo == nil || correlationID == nil {
		return fmt.Errorf("missing replyTo or correlationId in context")
	}

	return h.publisher.PublishToQueue(ctx, replyTo.(string), correlationID.(string), response)
}

// HandleGetOffersByCategory handles query.catalog.offering.by_category
func (h *CatalogRPCHandler) HandleGetOffersByCategory(ctx context.Context, payload []byte) error {
	var req struct {
		Category string `json:"category"`
	}
	if err := json.Unmarshal(payload, &req); err != nil {
		return fmt.Errorf("failed to unmarshal request: %w", err)
	}

	h.logger.InfoContext(ctx, "Getting offers by category", "category", req.Category)

	// Filter by category
	filters := map[string]interface{}{
		"category": req.Category,
	}

	offerings, err := h.offeringRepo.List(ctx, filters)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to list offerings", "category", req.Category, "error", err)
		return fmt.Errorf("failed to list offerings: %w", err)
	}

	// Map to response format (simplify to compatible structure)
	// Qualification expects: []{id, name, characteristics}
	type offerResponse struct {
		ID              string            `json:"id"`
		Name            string            `json:"name"`
		Characteristics map[string]string `json:"characteristics"`
	}

	var response []offerResponse
	for _, o := range offerings {
		response = append(response, offerResponse{
			ID:              o.ID,
			Name:            o.Name,
			Characteristics: map[string]string{}, // Empty for now as ProductOffering doesn't hold them directly
		})
	}

	// Send RPC reply
	replyTo := ctx.Value(rabbitmq.ContextKeyReplyTo)
	correlationID := ctx.Value(rabbitmq.ContextKeyAMQPCorrelationID)

	if replyTo == nil || correlationID == nil {
		return fmt.Errorf("missing replyTo or correlationId in context")
	}

	return h.publisher.PublishToQueue(ctx, replyTo.(string), correlationID.(string), response)
}

// BindRPCHandlers binds RPC handlers to the consumer
func (h *CatalogRPCHandler) BindRPCHandlers(consumer rabbitmq.Consumer) error {
	// Bind query.catalog.offering.get
	if err := consumer.Subscribe("query.catalog.offering.get", h.HandleGetOffering); err != nil {
		return fmt.Errorf("failed to bind query.catalog.offering.get: %w", err)
	}

	// Bind query.catalog.offering.by_category
	if err := consumer.Subscribe("query.catalog.offering.by_category", h.HandleGetOffersByCategory); err != nil {
		return fmt.Errorf("failed to bind query.catalog.offering.by_category: %w", err)
	}

	h.logger.Info("Catalog RPC handlers bound successfully")
	return nil
}
