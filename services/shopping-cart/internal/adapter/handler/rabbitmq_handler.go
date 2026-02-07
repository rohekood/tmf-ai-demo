package handler

import (
	"context"
	"encoding/json"
	"log/slog"

	"tmf/pkg/rabbitmq"
	"tmf/services/shopping-cart/internal/core/ports"
)

type CartHandler struct {
	manageUC  ports.ManageItemsUseCase
	priceUC   ports.UpdatePriceUseCase
	syncUC    ports.SyncCatalogUseCase
	repo      ports.CartRepository // Added for Query
	publisher rabbitmq.Publisher   // Added for RPC Reply
}

func NewCartHandler(
	manageUC ports.ManageItemsUseCase,
	priceUC ports.UpdatePriceUseCase,
	syncUC ports.SyncCatalogUseCase,
	repo ports.CartRepository,
	pub rabbitmq.Publisher,
) *CartHandler {
	return &CartHandler{
		manageUC:  manageUC,
		priceUC:   priceUC,
		syncUC:    syncUC,
		repo:      repo,
		publisher: pub,
	}
}

// HandleAddItem handles cmd.cart.item.add
func (h *CartHandler) HandleAddItem(ctx context.Context, payload []byte) error {
	var cmd struct {
		CartID                 string `json:"cartId"`
		OfferingID             string `json:"offeringId"`
		Quantity               int    `json:"quantity"`
		QualificationSessionID string `json:"qualificationSessionId,omitempty"`
	}
	if err := json.Unmarshal(payload, &cmd); err != nil {
		return err
	}

	slog.InfoContext(ctx, "Adding item to cart",
		"cartId", cmd.CartID,
		"offeringId", cmd.OfferingID,
		"qualificationSessionId", cmd.QualificationSessionID)
	return h.manageUC.AddItem(ctx, cmd.CartID, cmd.OfferingID, cmd.QualificationSessionID, cmd.Quantity)
}

// HandleUpdatePrice (Deprecated/Internal)
func (h *CartHandler) HandleUpdatePrice(ctx context.Context, payload []byte) error {
	var cmd ports.UpdateCartPriceCommand
	if err := json.Unmarshal(payload, &cmd); err != nil {
		return err
	}
	return h.priceUC.UpdatePrice(ctx, cmd)
}

// HandleCatalogEvent handles evt.catalog.offering.*
func (h *CartHandler) HandleCatalogEvent(ctx context.Context, payload []byte) error {
	var event struct {
		ID    string `json:"id"`
		Price struct {
			Amount   float64 `json:"amount"`
			Currency string  `json:"currency"`
		} `json:"price"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		return err
	}
	return h.syncUC.SyncOffering(ctx, event.ID, event.Price.Amount, event.Price.Currency)
}

// HandleGetCart (RPC Server)
func (h *CartHandler) HandleGetCart(ctx context.Context, payload []byte) error {
	// 1. Parse Request
	var req struct {
		CartID string `json:"cartId"`
	}
	if err := json.Unmarshal(payload, &req); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal GetCart", "error", err)
		return err
	}

	slog.InfoContext(ctx, "Processing RPC: GetCart", "cartId", req.CartID)

	// 2. Fetch Data
	cart, err := h.repo.Get(ctx, req.CartID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to fetch cart", "error", err)
		return err
	}

	// 3. Send Reply
	replyTo, _ := ctx.Value(rabbitmq.ContextKeyReplyTo).(string)

	correlationID, _ := ctx.Value(rabbitmq.ContextKeyAMQPCorrelationID).(string)

	if replyTo == "" {
		slog.WarnContext(ctx, "Missing ReplyTo in RPC request - cannot reply", "cartId", req.CartID)
		return nil
	}

	// Publish Reply
	return h.publisher.PublishToQueue(ctx, replyTo, correlationID, cart)
}
