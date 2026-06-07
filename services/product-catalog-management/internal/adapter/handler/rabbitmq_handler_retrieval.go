package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"tmf/services/product-catalog-management/internal/core/ports"
)

func (h *RabbitMQHandler) handleGetCatalog(ctx context.Context, payload []byte) error {
	var p struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := h.getCatalogUC.Execute(ctx, ports.GetCatalogInput{ID: p.ID})
	if err != nil {
		slog.ErrorContext(ctx, "Error executing GetCatalog", "error", err)
		h.reply(ctx, map[string]string{"error": err.Error()})
		return nil
	}
	h.reply(ctx, result)
	return nil
}

func (h *RabbitMQHandler) handleListCatalogs(ctx context.Context, _ []byte) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	results, err := h.listCatalogsUC.Execute(ctx, ports.ListCatalogsInput{Filters: nil})
	if err != nil {
		slog.ErrorContext(ctx, "Error executing ListCatalogs", "error", err)
		h.reply(ctx, map[string]string{"error": err.Error()})
		return nil
	}
	h.reply(ctx, results)
	return nil
}

func (h *RabbitMQHandler) handleListCategories(ctx context.Context, _ []byte) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	results, err := h.listCategoriesUC.Execute(ctx, ports.ListCategoriesInput{Filters: nil})
	if err != nil {
		slog.ErrorContext(ctx, "Error executing ListCategories", "error", err)
		h.reply(ctx, map[string]string{"error": err.Error()})
		return nil
	}
	h.reply(ctx, results)
	return nil
}

func (h *RabbitMQHandler) handleGetCategory(ctx context.Context, payload []byte) error {
	var p struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := h.getCategoryUC.Execute(ctx, ports.GetCategoryInput{ID: p.ID})
	if err != nil {
		slog.ErrorContext(ctx, "Error executing GetCategory", "error", err)
		h.reply(ctx, map[string]string{"error": err.Error()})
		return nil
	}
	h.reply(ctx, result)
	return nil
}

func (h *RabbitMQHandler) handleListProductSpecifications(ctx context.Context, _ []byte) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	results, err := h.listProductSpecificationsUC.Execute(ctx, ports.ListProductSpecificationsInput{Filters: nil})
	if err != nil {
		slog.ErrorContext(ctx, "Error executing ListProductSpecifications", "error", err)
		h.reply(ctx, map[string]string{"error": err.Error()})
		return nil
	}
	h.reply(ctx, results)
	return nil
}

func (h *RabbitMQHandler) handleGetProductSpecification(ctx context.Context, payload []byte) error {
	var p struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := h.getProductSpecificationUC.Execute(ctx, ports.GetProductSpecificationInput{ID: p.ID})
	if err != nil {
		slog.ErrorContext(ctx, "Error executing GetProductSpecification", "error", err)
		h.reply(ctx, map[string]string{"error": err.Error()})
		return nil
	}
	h.reply(ctx, result)
	return nil
}

func (h *RabbitMQHandler) handleListProductOfferings(ctx context.Context, payload []byte) error {
	var p struct {
		Name     *string  `json:"name"`
		Category *string  `json:"category"`
		MinPrice *float64 `json:"minPrice"`
		MaxPrice *float64 `json:"maxPrice"`
	}
	if len(payload) > 0 {
		if err := json.Unmarshal(payload, &p); err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	results, err := h.listProductOfferingsUC.Execute(ctx, ports.ListProductOfferingsInput{
		Filters: ports.ProductOfferingFilters{
			Name:     p.Name,
			Category: p.Category,
			MinPrice: p.MinPrice,
			MaxPrice: p.MaxPrice,
		},
	})
	if err != nil {
		slog.ErrorContext(ctx, "Error executing ListProductOfferings", "error", err)
		h.reply(ctx, map[string]string{"error": err.Error()})
		return nil
	}
	h.reply(ctx, results)
	return nil
}

func (h *RabbitMQHandler) handleGetProductOffering(ctx context.Context, payload []byte) error {
	var p struct {
		ID         string `json:"id"`
		OfferingID string `json:"offeringId"`
		Enrich     bool   `json:"enrich"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	id := p.ID
	if id == "" {
		id = p.OfferingID
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := h.getProductOfferingUC.Execute(ctx, ports.GetProductOfferingInput{ID: id, Enrich: p.Enrich})
	if err != nil {
		slog.ErrorContext(ctx, "Error executing GetProductOffering", "error", err)
		h.reply(ctx, map[string]string{"error": err.Error()})
		return nil
	}
	h.reply(ctx, result)
	return nil
}
