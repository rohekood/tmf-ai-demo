package handler

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"tmf/services/product-catalog-management/internal/core/ports"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (h *RabbitMQHandler) handleGetCatalog(d amqp.Delivery) {
	type GetPayload struct {
		ID string `json:"id"`
	}
	var payload GetPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		log.Printf("Error unmarshalling event: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := h.getCatalogUC.Execute(ctx, ports.GetCatalogInput{ID: payload.ID})
	if err != nil {
		log.Printf("Error executing GetCatalog: %v", err)
		return
	}

	h.publishResponse(ctx, d, result)
}

func (h *RabbitMQHandler) handleListCategories(d amqp.Delivery) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results, err := h.listCategoriesUC.Execute(ctx, ports.ListCategoriesInput{Filters: nil})
	if err != nil {
		log.Printf("Error executing ListCategories: %v", err)
		return
	}

	h.publishResponse(ctx, d, results)
}

func (h *RabbitMQHandler) handleGetCategory(d amqp.Delivery) {
	type GetPayload struct {
		ID string `json:"id"`
	}
	var payload GetPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		log.Printf("Error unmarshalling event: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := h.getCategoryUC.Execute(ctx, ports.GetCategoryInput{ID: payload.ID})
	if err != nil {
		log.Printf("Error executing GetCategory: %v", err)
		return
	}

	h.publishResponse(ctx, d, result)
}

func (h *RabbitMQHandler) handleListProductSpecifications(d amqp.Delivery) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results, err := h.listProductSpecificationsUC.Execute(ctx, ports.ListProductSpecificationsInput{Filters: nil})
	if err != nil {
		log.Printf("Error executing ListProductSpecifications: %v", err)
		return
	}

	h.publishResponse(ctx, d, results)
}

func (h *RabbitMQHandler) handleGetProductSpecification(d amqp.Delivery) {
	type GetPayload struct {
		ID string `json:"id"`
	}
	var payload GetPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		log.Printf("Error unmarshalling event: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := h.getProductSpecificationUC.Execute(ctx, ports.GetProductSpecificationInput{ID: payload.ID})
	if err != nil {
		log.Printf("Error executing GetProductSpecification: %v", err)
		return
	}

	h.publishResponse(ctx, d, result)
}

func (h *RabbitMQHandler) handleListProductOfferings(d amqp.Delivery) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	type ListPayload struct {
		Name     *string  `json:"name"`
		Category *string  `json:"category"`
		MinPrice *float64 `json:"minPrice"`
		MaxPrice *float64 `json:"maxPrice"`
	}

	var payload ListPayload
	if len(d.Body) > 0 {
		if err := json.Unmarshal(d.Body, &payload); err != nil {
			log.Printf("Error unmarshalling event: %v", err)
			return
		}
	}

	filters := ports.ProductOfferingFilters{
		Name:     payload.Name,
		Category: payload.Category,
		MinPrice: payload.MinPrice,
		MaxPrice: payload.MaxPrice,
	}

	results, err := h.listProductOfferingsUC.Execute(ctx, ports.ListProductOfferingsInput{Filters: filters})
	if err != nil {
		log.Printf("Error executing ListProductOfferings: %v", err)
		return
	}

	h.publishResponse(ctx, d, results)
}

func (h *RabbitMQHandler) handleGetProductOffering(d amqp.Delivery) {
	type GetPayload struct {
		ID         string `json:"id"`
		OfferingID string `json:"offeringId"` // Support alternate key
		Enrich     bool   `json:"enrich"`
	}
	var payload GetPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		log.Printf("Error unmarshalling event: %v", err)
		return
	}

	id := payload.ID
	if id == "" {
		id = payload.OfferingID
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := h.getProductOfferingUC.Execute(ctx, ports.GetProductOfferingInput{ID: id, Enrich: payload.Enrich})
	if err != nil {
		log.Printf("Error executing GetProductOffering: %v", err)
		return
	}

	h.publishResponse(ctx, d, result)
}

func (h *RabbitMQHandler) publishResponse(ctx context.Context, d amqp.Delivery, response interface{}) {
	responseBody, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling response: %v", err)
		return
	}

	err = h.channel.PublishWithContext(ctx,
		"",        // exchange
		d.ReplyTo, // routing key (reply queue)
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: d.CorrelationId,
			Body:          responseBody,
		},
	)
	if err != nil {
		log.Printf("Failed to publish query response: %v", err)
	}
}
