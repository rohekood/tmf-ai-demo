package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"tmf/pkg/rabbitmq"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"
)

type RabbitMQHandler struct {
	publisher                    rabbitmq.Publisher
	createCatalogUC              ports.CreateCatalogUseCase
	updateCatalogUC              ports.UpdateCatalogUseCase
	deleteCatalogUC              ports.DeleteCatalogUseCase
	listCatalogsUC               ports.ListCatalogsUseCase
	getCatalogUC                 ports.GetCatalogUseCase
	createCategoryUC             ports.CreateCategoryUseCase
	updateCategoryUC             ports.UpdateCategoryUseCase
	deleteCategoryUC             ports.DeleteCategoryUseCase
	getCategoryUC                ports.GetCategoryUseCase
	listCategoriesUC             ports.ListCategoriesUseCase
	createProductSpecificationUC ports.CreateProductSpecificationUseCase
	updateProductSpecificationUC ports.UpdateProductSpecificationUseCase
	deleteProductSpecificationUC ports.DeleteProductSpecificationUseCase
	getProductSpecificationUC    ports.GetProductSpecificationUseCase
	listProductSpecificationsUC  ports.ListProductSpecificationsUseCase
	createProductOfferingUC      ports.CreateProductOfferingUseCase
	updateProductOfferingUC      ports.UpdateProductOfferingUseCase
	deleteProductOfferingUC      ports.DeleteProductOfferingUseCase
	getProductOfferingUC         ports.GetProductOfferingUseCase
	listProductOfferingsUC       ports.ListProductOfferingsUseCase
}

func NewRabbitMQHandler(
	publisher rabbitmq.Publisher,
	createCatalogUC ports.CreateCatalogUseCase,
	updateCatalogUC ports.UpdateCatalogUseCase,
	deleteCatalogUC ports.DeleteCatalogUseCase,
	listCatalogsUC ports.ListCatalogsUseCase,
	getCatalogUC ports.GetCatalogUseCase,
	createCategoryUC ports.CreateCategoryUseCase,
	updateCategoryUC ports.UpdateCategoryUseCase,
	deleteCategoryUC ports.DeleteCategoryUseCase,
	getCategoryUC ports.GetCategoryUseCase,
	listCategoriesUC ports.ListCategoriesUseCase,
	createProductSpecificationUC ports.CreateProductSpecificationUseCase,
	updateProductSpecificationUC ports.UpdateProductSpecificationUseCase,
	deleteProductSpecificationUC ports.DeleteProductSpecificationUseCase,
	getProductSpecificationUC ports.GetProductSpecificationUseCase,
	listProductSpecificationsUC ports.ListProductSpecificationsUseCase,
	createProductOfferingUC ports.CreateProductOfferingUseCase,
	updateProductOfferingUC ports.UpdateProductOfferingUseCase,
	deleteProductOfferingUC ports.DeleteProductOfferingUseCase,
	getProductOfferingUC ports.GetProductOfferingUseCase,
	listProductOfferingsUC ports.ListProductOfferingsUseCase,
) *RabbitMQHandler {
	return &RabbitMQHandler{
		publisher:                    publisher,
		createCatalogUC:              createCatalogUC,
		updateCatalogUC:              updateCatalogUC,
		deleteCatalogUC:              deleteCatalogUC,
		listCatalogsUC:               listCatalogsUC,
		getCatalogUC:                 getCatalogUC,
		createCategoryUC:             createCategoryUC,
		updateCategoryUC:             updateCategoryUC,
		deleteCategoryUC:             deleteCategoryUC,
		getCategoryUC:                getCategoryUC,
		listCategoriesUC:             listCategoriesUC,
		createProductSpecificationUC: createProductSpecificationUC,
		updateProductSpecificationUC: updateProductSpecificationUC,
		deleteProductSpecificationUC: deleteProductSpecificationUC,
		getProductSpecificationUC:    getProductSpecificationUC,
		listProductSpecificationsUC:  listProductSpecificationsUC,
		createProductOfferingUC:      createProductOfferingUC,
		updateProductOfferingUC:      updateProductOfferingUC,
		deleteProductOfferingUC:      deleteProductOfferingUC,
		getProductOfferingUC:         getProductOfferingUC,
		listProductOfferingsUC:       listProductOfferingsUC,
	}
}

// BindHandlers subscribes all command and query handlers to their respective consumers.
// Both consumers must already be connected to the catalog_events exchange.
func (h *RabbitMQHandler) BindHandlers(cmdConsumer, queryConsumer rabbitmq.Consumer) error {
	cmds := []struct {
		topic string
		fn    rabbitmq.ConsumerHandler
	}{
		{"cmd.catalog.catalog.create", h.handleCreateCatalog},
		{"cmd.catalog.catalog.update", h.handleUpdateCatalog},
		{"cmd.catalog.catalog.delete", h.handleDeleteCatalog},
		{"cmd.catalog.category.create", h.handleCreateCategory},
		{"cmd.catalog.category.update", h.handleUpdateCategory},
		{"cmd.catalog.category.delete", h.handleDeleteCategory},
		{"cmd.catalog.specification.create", h.handleCreateProductSpecification},
		{"cmd.catalog.specification.update", h.handleUpdateProductSpecification},
		{"cmd.catalog.specification.delete", h.handleDeleteProductSpecification},
		{"cmd.catalog.offering.create", h.handleCreateProductOffering},
		{"cmd.catalog.offering.update", h.handleUpdateProductOffering},
		{"cmd.catalog.offering.delete", h.handleDeleteProductOffering},
	}
	for _, c := range cmds {
		if err := cmdConsumer.Subscribe(c.topic, c.fn); err != nil {
			return fmt.Errorf("subscribe %s: %w", c.topic, err)
		}
	}

	queries := []struct {
		topic string
		fn    rabbitmq.ConsumerHandler
	}{
		{"query.catalog.catalog.list", h.handleListCatalogs},
		{"query.catalog.catalog.get", h.handleGetCatalog},
		{"query.catalog.category.list", h.handleListCategories},
		{"query.catalog.category.get", h.handleGetCategory},
		{"query.catalog.specification.list", h.handleListProductSpecifications},
		{"query.catalog.specification.get", h.handleGetProductSpecification},
		{"query.catalog.offering.list", h.handleListProductOfferings},
		{"query.catalog.offering.get", h.handleGetProductOffering},
	}
	for _, q := range queries {
		if err := queryConsumer.Subscribe(q.topic, q.fn); err != nil {
			return fmt.Errorf("subscribe %s: %w", q.topic, err)
		}
	}

	slog.Info("Catalog CRUD handlers bound successfully")
	return nil
}

func (h *RabbitMQHandler) reply(ctx context.Context, response any) {
	replyTo, _ := ctx.Value(rabbitmq.ContextKeyReplyTo).(string)
	if replyTo == "" {
		return
	}
	correlationID, _ := ctx.Value(rabbitmq.ContextKeyAMQPCorrelationID).(string)
	if err := h.publisher.PublishToQueue(ctx, replyTo, correlationID, response); err != nil {
		slog.ErrorContext(ctx, "Failed to publish response", "error", err)
	}
}

func (h *RabbitMQHandler) handleCreateCatalog(ctx context.Context, payload []byte) error {
	var event domain.CatalogCreateEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := h.createCatalogUC.Execute(ctx, ports.CreateCatalogInput{
		Name:            event.Name,
		Description:     event.Description,
		ValidFor:        event.ValidFor,
		LifecycleStatus: event.LifecycleStatus,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Error executing CreateCatalog", "error", err)
		h.reply(ctx, map[string]string{"error": err.Error()})
		return nil
	}
	h.reply(ctx, result)
	return nil
}

func (h *RabbitMQHandler) handleUpdateCatalog(ctx context.Context, payload []byte) error {
	type UpdatePayload struct {
		ID              string             `json:"id"`
		Name            *string            `json:"name"`
		Description     *string            `json:"description"`
		ValidFor        *domain.TimePeriod `json:"validFor"`
		LifecycleStatus *string            `json:"lifecycleStatus"`
	}
	var p UpdatePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := h.updateCatalogUC.Execute(ctx, ports.UpdateCatalogInput{
		ID:              p.ID,
		Name:            p.Name,
		Description:     p.Description,
		ValidFor:        p.ValidFor,
		LifecycleStatus: p.LifecycleStatus,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Error executing UpdateCatalog", "error", err)
		h.reply(ctx, map[string]string{"error": err.Error()})
		return nil
	}
	h.reply(ctx, result)
	return nil
}

func (h *RabbitMQHandler) handleDeleteCatalog(ctx context.Context, payload []byte) error {
	var p struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := h.deleteCatalogUC.Execute(ctx, ports.DeleteCatalogInput{ID: p.ID}); err != nil {
		slog.ErrorContext(ctx, "Error executing DeleteCatalog", "error", err)
		h.reply(ctx, map[string]string{"error": err.Error()})
		return nil
	}
	h.reply(ctx, map[string]string{"status": "deleted", "id": p.ID})
	return nil
}

func (h *RabbitMQHandler) handleCreateCategory(ctx context.Context, payload []byte) error {
	var event domain.CategoryCreateEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := h.createCategoryUC.Execute(ctx, ports.CreateCategoryInput{
		Name:            event.Name,
		Description:     event.Description,
		ParentID:        event.ParentID,
		IsRoot:          event.IsRoot,
		CatalogID:       event.CatalogID,
		ValidFor:        event.ValidFor,
		LifecycleStatus: event.LifecycleStatus,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Error executing CreateCategory", "error", err)
		h.reply(ctx, map[string]string{"error": err.Error()})
		return nil
	}
	h.reply(ctx, result)
	return nil
}

func (h *RabbitMQHandler) handleUpdateCategory(ctx context.Context, payload []byte) error {
	type UpdatePayload struct {
		ID              string             `json:"id"`
		Name            *string            `json:"name"`
		Description     *string            `json:"description"`
		ParentID        *string            `json:"parentId"`
		IsRoot          *bool              `json:"isRoot"`
		CatalogID       *string            `json:"catalogId"`
		ValidFor        *domain.TimePeriod `json:"validFor"`
		LifecycleStatus *string            `json:"lifecycleStatus"`
	}
	var p UpdatePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := h.updateCategoryUC.Execute(ctx, ports.UpdateCategoryInput{
		ID:              p.ID,
		Name:            p.Name,
		Description:     p.Description,
		ParentID:        p.ParentID,
		IsRoot:          p.IsRoot,
		CatalogID:       p.CatalogID,
		ValidFor:        p.ValidFor,
		LifecycleStatus: p.LifecycleStatus,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Error executing UpdateCategory", "error", err)
		h.reply(ctx, map[string]string{"error": err.Error()})
		return nil
	}
	h.reply(ctx, result)
	return nil
}

func (h *RabbitMQHandler) handleDeleteCategory(ctx context.Context, payload []byte) error {
	var p struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := h.deleteCategoryUC.Execute(ctx, ports.DeleteCategoryInput{ID: p.ID}); err != nil {
		slog.ErrorContext(ctx, "Error executing DeleteCategory", "error", err)
		h.reply(ctx, map[string]string{"error": err.Error()})
		return nil
	}
	h.reply(ctx, map[string]string{"status": "deleted", "id": p.ID})
	return nil
}

func (h *RabbitMQHandler) handleCreateProductSpecification(ctx context.Context, payload []byte) error {
	var event domain.ProductSpecificationCreateEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := h.createProductSpecificationUC.Execute(ctx, ports.CreateProductSpecificationInput{
		Name:            event.Name,
		ProductNumber:   event.ProductNumber,
		Description:     event.Description,
		IsBundle:        event.IsBundle,
		LifecycleStatus: event.LifecycleStatus,
		ValidFor:        event.ValidFor,
		Characteristics: event.Characteristics,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Error executing CreateProductSpecification", "error", err)
		h.reply(ctx, map[string]string{"error": err.Error()})
		return nil
	}
	h.reply(ctx, result)
	return nil
}

func (h *RabbitMQHandler) handleUpdateProductSpecification(ctx context.Context, payload []byte) error {
	type UpdatePayload struct {
		ID              string                                      `json:"id"`
		Name            *string                                     `json:"name"`
		Description     *string                                     `json:"description"`
		ProductNumber   *string                                     `json:"productNumber"`
		IsBundle        *bool                                       `json:"isBundle"`
		LifecycleStatus *string                                     `json:"lifecycleStatus"`
		ValidFor        *domain.TimePeriod                          `json:"validFor"`
		Characteristics map[string]domain.ProductSpecCharacteristic `json:"characteristics"`
	}
	var p UpdatePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := h.updateProductSpecificationUC.Execute(ctx, ports.UpdateProductSpecificationInput{
		ID:              p.ID,
		Name:            p.Name,
		Description:     p.Description,
		ProductNumber:   p.ProductNumber,
		IsBundle:        p.IsBundle,
		LifecycleStatus: p.LifecycleStatus,
		ValidFor:        p.ValidFor,
		Characteristics: p.Characteristics,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Error executing UpdateProductSpecification", "error", err)
		h.reply(ctx, map[string]string{"error": err.Error()})
		return nil
	}
	h.reply(ctx, result)
	return nil
}

func (h *RabbitMQHandler) handleDeleteProductSpecification(ctx context.Context, payload []byte) error {
	var p struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := h.deleteProductSpecificationUC.Execute(ctx, ports.DeleteProductSpecificationInput{ID: p.ID}); err != nil {
		slog.ErrorContext(ctx, "Error executing DeleteProductSpecification", "error", err)
		h.reply(ctx, map[string]string{"error": err.Error()})
		return nil
	}
	h.reply(ctx, map[string]string{"status": "deleted", "id": p.ID})
	return nil
}

func (h *RabbitMQHandler) handleCreateProductOffering(ctx context.Context, payload []byte) error {
	var event domain.ProductOfferingCreateEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := h.createProductOfferingUC.Execute(ctx, ports.CreateProductOfferingInput{
		Name:            event.Name,
		Description:     event.Description,
		IsBundle:        event.IsBundle,
		IsSellable:      event.IsSellable,
		LifecycleStatus: event.LifecycleStatus,
		ValidFor:        event.ValidFor,
		ProductSpecID:   event.ProductSpecID,
		CategoryIDs:     event.CategoryIDs,
		Prices:          event.Prices,
		Attachments:     event.Attachments,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Error executing CreateProductOffering", "error", err)
		h.reply(ctx, map[string]string{"error": err.Error()})
		return nil
	}
	h.reply(ctx, result)
	return nil
}

func (h *RabbitMQHandler) handleUpdateProductOffering(ctx context.Context, payload []byte) error {
	type UpdatePayload struct {
		ID              string                        `json:"id"`
		Name            *string                       `json:"name"`
		Description     *string                       `json:"description"`
		LifecycleStatus *string                       `json:"lifecycleStatus"`
		ValidFor        *domain.TimePeriod            `json:"validFor"`
		IsBundle        *bool                         `json:"isBundle"`
		IsSellable      *bool                         `json:"isSellable"`
		CategoryIDs     []string                      `json:"categoryIds"`
		Prices          []domain.ProductOfferingPrice `json:"prices"`
		Attachments     []domain.Attachment           `json:"attachments"`
	}
	var p UpdatePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := h.updateProductOfferingUC.Execute(ctx, ports.UpdateProductOfferingInput{
		ID:              p.ID,
		Name:            p.Name,
		Description:     p.Description,
		LifecycleStatus: p.LifecycleStatus,
		ValidFor:        p.ValidFor,
		IsBundle:        p.IsBundle,
		IsSellable:      p.IsSellable,
		CategoryIDs:     p.CategoryIDs,
		Prices:          p.Prices,
		Attachments:     p.Attachments,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Error executing UpdateProductOffering", "error", err)
		h.reply(ctx, map[string]string{"error": err.Error()})
		return nil
	}
	h.reply(ctx, result)
	return nil
}

func (h *RabbitMQHandler) handleDeleteProductOffering(ctx context.Context, payload []byte) error {
	var p struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := h.deleteProductOfferingUC.Execute(ctx, ports.DeleteProductOfferingInput{ID: p.ID}); err != nil {
		slog.ErrorContext(ctx, "Error executing DeleteProductOffering", "error", err)
		h.reply(ctx, map[string]string{"error": err.Error()})
		return nil
	}
	h.reply(ctx, map[string]string{"status": "deleted", "id": p.ID})
	return nil
}
