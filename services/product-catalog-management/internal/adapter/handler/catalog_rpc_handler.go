package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"tmf/pkg/rabbitmq"
	"tmf/services/product-catalog-management/internal/core/ports"
)

type CatalogRPCHandler struct {
	offeringRepo ports.ProductOfferingRepository
	publisher    rabbitmq.Publisher
	logger       *slog.Logger
}

func NewCatalogRPCHandler(offeringRepo ports.ProductOfferingRepository, publisher rabbitmq.Publisher, logger *slog.Logger) *CatalogRPCHandler {
	return &CatalogRPCHandler{
		offeringRepo: offeringRepo,
		publisher:    publisher,
		logger:       logger,
	}
}

func (h *CatalogRPCHandler) HandleGetOffering(ctx context.Context, payload []byte) error {
	var req struct {
		OfferingID string `json:"offeringId"`
	}

	if err := json.Unmarshal(payload, &req); err != nil {
		return fmt.Errorf("failed to unmarshal request: %w", err)
	}

	h.logger.InfoContext(ctx, "Getting offering for pricing", "offeringId", req.OfferingID)

	offering, err := h.offeringRepo.Get(ctx, req.OfferingID)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to get offering", "offeringId", req.OfferingID, "error", err)
		return h.replyError(ctx, err)
	}

	var basePrice = 100.0
	var currency = "EUR"

	if len(offering.ProductOfferingPrice) > 0 {
		basePrice = offering.ProductOfferingPrice[0].Price.Value
		currency = offering.ProductOfferingPrice[0].Price.Unit
	}

	response := map[string]interface{}{
		"id":        offering.ID,
		"name":      offering.Name,
		"basePrice": basePrice,
		"currency":  currency,
	}

	replyTo := ctx.Value(rabbitmq.ContextKeyReplyTo)
	correlationID := ctx.Value(rabbitmq.ContextKeyAMQPCorrelationID)

	if replyTo == nil || correlationID == nil {
		return fmt.Errorf("missing replyTo or correlationId in context")
	}

	return h.publisher.PublishToQueue(ctx, replyTo.(string), correlationID.(string), response)
}

func (h *CatalogRPCHandler) HandleGetOffersByCategory(ctx context.Context, payload []byte) error {
	var req struct {
		Category string `json:"category"`
	}
	if err := json.Unmarshal(payload, &req); err != nil {
		return fmt.Errorf("failed to unmarshal request: %w", err)
	}

	h.logger.InfoContext(ctx, "Getting offers by category", "category", req.Category)

	filters := map[string]interface{}{
		"category": req.Category,
	}

	offerings, err := h.offeringRepo.List(ctx, filters)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to list offerings", "category", req.Category, "error", err)
		return fmt.Errorf("failed to list offerings: %w", err)
	}

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
			Characteristics: map[string]string{},
		})
	}

	replyTo := ctx.Value(rabbitmq.ContextKeyReplyTo)
	correlationID := ctx.Value(rabbitmq.ContextKeyAMQPCorrelationID)

	if replyTo == nil || correlationID == nil {
		return fmt.Errorf("missing replyTo or correlationId in context")
	}

	return h.publisher.PublishToQueue(ctx, replyTo.(string), correlationID.(string), response)
}

func (h *CatalogRPCHandler) replyError(ctx context.Context, err error) error {
	replyTo := ctx.Value(rabbitmq.ContextKeyReplyTo)
	correlationID := ctx.Value(rabbitmq.ContextKeyAMQPCorrelationID)

	if replyTo == nil || correlationID == nil {
		return fmt.Errorf("missing replyTo or correlationId in context")
	}

	errResponse := map[string]string{"error": err.Error()}
	return h.publisher.PublishToQueue(ctx, replyTo.(string), correlationID.(string), errResponse)
}

func (h *CatalogRPCHandler) BindRPCHandlers(consumer rabbitmq.Consumer) error {
	if err := consumer.Subscribe("query.catalog.offering.#", h.HandleDispatch); err != nil {
		return fmt.Errorf("failed to bind query.catalog.offering.#: %w", err)
	}

	h.logger.Info("Catalog RPC handlers bound successfully")
	return nil
}

func (h *CatalogRPCHandler) HandleDispatch(ctx context.Context, payload []byte) error {
	routingKey, ok := ctx.Value(rabbitmq.ContextKeyRoutingKey).(string)
	if !ok {
		return fmt.Errorf("routing key not found in context")
	}

	switch routingKey {
	case "query.catalog.offering.get":
		return h.HandleGetOffering(ctx, payload)
	case "query.catalog.offering.by_category":
		return h.HandleGetOffersByCategory(ctx, payload)
	default:
		h.logger.DebugContext(ctx, "Ignoring unknown RPC routing key", "key", routingKey)
		return nil
	}
}
