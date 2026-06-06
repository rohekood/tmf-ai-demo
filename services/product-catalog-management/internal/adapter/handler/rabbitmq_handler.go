package handler

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQHandler struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	replyChannel *amqp.Channel
	replyMu      sync.Mutex
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
	conn *amqp.Connection,
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
) (*RabbitMQHandler, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	replyCh, err := conn.Channel()
	if err != nil {
		_ = ch.Close()
		return nil, err
	}
	return &RabbitMQHandler{
		conn:                         conn,
		channel:                      ch,
		replyChannel:                 replyCh,
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
	}, nil
}

func (h *RabbitMQHandler) Start(ctx context.Context) error {

	if err := h.setupConsumers(); err != nil {
		return err
	}

	if err := h.setupQueryConsumers(); err != nil {
		return err
	}

	log.Println("RabbitMQ Handler started")
	<-ctx.Done()
	if err := h.channel.Close(); err != nil {
		log.Printf("Error closing RabbitMQ channel: %v", err)
	}
	h.replyMu.Lock()
	if err := h.replyChannel.Close(); err != nil {
		log.Printf("Error closing RabbitMQ reply channel: %v", err)
	}
	h.replyMu.Unlock()
	return nil
}

func (h *RabbitMQHandler) setupConsumers() error {
	q, err := h.channel.QueueDeclare(
		"catalog_commands", // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		return err
	}

	err = h.channel.QueueBind(
		q.Name,                  // queue name
		"cmd.catalog.catalog.*", // routing key
		"catalog_events",        // exchange
		false,
		nil,
	)
	if err != nil {
		return err
	}

	err = h.channel.QueueBind(
		q.Name,                   // queue name
		"cmd.catalog.category.*", // routing key
		"catalog_events",         // exchange
		false,
		nil,
	)
	if err != nil {
		return err
	}

	err = h.channel.QueueBind(
		q.Name,                        // queue name
		"cmd.catalog.specification.*", // routing key
		"catalog_events",              // exchange
		false,
		nil,
	)
	if err != nil {
		return err
	}

	err = h.channel.QueueBind(
		q.Name,                   // queue name
		"cmd.catalog.offering.*", // routing key
		"catalog_events",         // exchange
		false,
		nil,
	)
	if err != nil {
		return err
	}

	msgs, err := h.channel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack (should be false in production usually)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			h.handleMessage(d)
		}
	}()

	return nil
}

func (h *RabbitMQHandler) setupQueryConsumers() error {
	q, err := h.channel.QueueDeclare(
		"catalog_queries", // name
		true,              // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		return err
	}

	err = h.channel.QueueBind(
		q.Name,            // queue name
		"query.catalog.#", // routing key wildcard
		"catalog_events",  // exchange
		false,
		nil,
	)
	if err != nil {
		return err
	}

	msgs, err := h.channel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			h.handleQuery(d)
		}
	}()

	return nil
}

func (h *RabbitMQHandler) handleMessage(d amqp.Delivery) {
	switch d.RoutingKey {
	case "cmd.catalog.catalog.create":
		h.handleCreateCatalog(d)
	case "cmd.catalog.catalog.update":
		h.handleUpdateCatalog(d)
	case "cmd.catalog.catalog.delete":
		h.handleDeleteCatalog(d)
	case "cmd.catalog.category.create":
		h.handleCreateCategory(d)
	case "cmd.catalog.category.update":
		h.handleUpdateCategory(d)
	case "cmd.catalog.category.delete":
		h.handleDeleteCategory(d)
	case "cmd.catalog.specification.create":
		h.handleCreateProductSpecification(d)
	case "cmd.catalog.specification.update":
		h.handleUpdateProductSpecification(d)
	case "cmd.catalog.specification.delete":
		h.handleDeleteProductSpecification(d)
	case "cmd.catalog.offering.create":
		h.handleCreateProductOffering(d)
	case "cmd.catalog.offering.update":
		h.handleUpdateProductOffering(d)
	case "cmd.catalog.offering.delete":
		h.handleDeleteProductOffering(d)
	default:
		log.Printf("Unknown routing key: %s", d.RoutingKey)
	}
}

func (h *RabbitMQHandler) handleQuery(d amqp.Delivery) {
	switch d.RoutingKey {
	case "query.catalog.catalog.list":
		h.handleListCatalogs(d)
	case "query.catalog.catalog.get":
		h.handleGetCatalog(d)
	case "query.catalog.category.list":
		h.handleListCategories(d)
	case "query.catalog.category.get":
		h.handleGetCategory(d)
	case "query.catalog.specification.list":
		h.handleListProductSpecifications(d)
	case "query.catalog.specification.get":
		h.handleGetProductSpecification(d)
	case "query.catalog.offering.list":
		h.handleListProductOfferings(d)
	case "query.catalog.offering.get":
		h.handleGetProductOffering(d)
	default:
		log.Printf("Unknown query routing key: %s", d.RoutingKey)
	}
}

func (h *RabbitMQHandler) handleListCatalogs(d amqp.Delivery) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	catalogs, err := h.listCatalogsUC.Execute(ctx, ports.ListCatalogsInput{Filters: nil})
	if err != nil {
		log.Printf("Error executing ListCatalogs: %v", err)
		h.reply(ctx, d, map[string]string{"error": err.Error()})
		return
	}

	h.reply(ctx, d, catalogs)
}

func (h *RabbitMQHandler) handleCreateCatalog(d amqp.Delivery) {
	var event domain.CatalogCreateEvent
	if err := json.Unmarshal(d.Body, &event); err != nil {
		log.Printf("Error unmarshalling event: %v", err)
		return
	}

	input := ports.CreateCatalogInput{
		Name:            event.Name,
		Description:     event.Description,
		ValidFor:        event.ValidFor,
		LifecycleStatus: event.LifecycleStatus,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if result, err := h.createCatalogUC.Execute(ctx, input); err != nil {
		log.Printf("Error executing CreateCatalog: %v", err)
		h.reply(ctx, d, map[string]string{"error": err.Error()})
	} else {
		log.Printf("Catalog created successfully")
		h.reply(ctx, d, result)
	}
}

func (h *RabbitMQHandler) handleUpdateCatalog(d amqp.Delivery) {
	type UpdatePayload struct {
		ID              string             `json:"id"`
		Name            *string            `json:"name"`
		Description     *string            `json:"description"`
		ValidFor        *domain.TimePeriod `json:"validFor"`
		LifecycleStatus *string            `json:"lifecycleStatus"`
	}

	var payload UpdatePayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		log.Printf("Error unmarshalling event: %v", err)
		return
	}

	input := ports.UpdateCatalogInput{
		ID:              payload.ID,
		Name:            payload.Name,
		Description:     payload.Description,
		ValidFor:        payload.ValidFor,
		LifecycleStatus: payload.LifecycleStatus,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if result, err := h.updateCatalogUC.Execute(ctx, input); err != nil {
		log.Printf("Error executing UpdateCatalog: %v", err)
		h.reply(ctx, d, map[string]string{"error": err.Error()})
	} else {
		log.Printf("Catalog updated successfully")
		h.reply(ctx, d, result)
	}
}

func (h *RabbitMQHandler) handleDeleteCatalog(d amqp.Delivery) {
	type DeletePayload struct {
		ID string `json:"id"`
	}
	var payload DeletePayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		log.Printf("Error unmarshalling event: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.deleteCatalogUC.Execute(ctx, ports.DeleteCatalogInput{ID: payload.ID}); err != nil {
		log.Printf("Error executing DeleteCatalog: %v", err)
		h.reply(ctx, d, map[string]string{"error": err.Error()})
	} else {
		log.Printf("Catalog deleted successfully")
		h.reply(ctx, d, map[string]string{"status": "deleted", "id": payload.ID})
	}
}

func (h *RabbitMQHandler) handleCreateCategory(d amqp.Delivery) {
	var event domain.CategoryCreateEvent
	if err := json.Unmarshal(d.Body, &event); err != nil {
		log.Printf("Error unmarshalling event: %v", err)
		return
	}

	input := ports.CreateCategoryInput{
		Name:            event.Name,
		Description:     event.Description,
		ParentID:        event.ParentID,
		IsRoot:          event.IsRoot,
		CatalogID:       event.CatalogID,
		ValidFor:        event.ValidFor,
		LifecycleStatus: event.LifecycleStatus,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if result, err := h.createCategoryUC.Execute(ctx, input); err != nil {
		log.Printf("Error executing CreateCategory: %v", err)
		h.reply(ctx, d, map[string]string{"error": err.Error()})
	} else {
		log.Printf("Category created successfully")
		h.reply(ctx, d, result)
	}
}

func (h *RabbitMQHandler) handleUpdateCategory(d amqp.Delivery) {
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
	var payload UpdatePayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		log.Printf("Error unmarshalling event: %v", err)
		return
	}

	input := ports.UpdateCategoryInput{
		ID:              payload.ID,
		Name:            payload.Name,
		Description:     payload.Description,
		ParentID:        payload.ParentID,
		IsRoot:          payload.IsRoot,
		CatalogID:       payload.CatalogID,
		ValidFor:        payload.ValidFor,
		LifecycleStatus: payload.LifecycleStatus,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if result, err := h.updateCategoryUC.Execute(ctx, input); err != nil {
		log.Printf("Error executing UpdateCategory: %v", err)
		h.reply(ctx, d, map[string]string{"error": err.Error()})
	} else {
		log.Printf("Category updated successfully")
		h.reply(ctx, d, result)
	}
}

func (h *RabbitMQHandler) handleDeleteCategory(d amqp.Delivery) {
	type DeletePayload struct {
		ID string `json:"id"`
	}
	var payload DeletePayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		log.Printf("Error unmarshalling event: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.deleteCategoryUC.Execute(ctx, ports.DeleteCategoryInput{ID: payload.ID}); err != nil {
		log.Printf("Error executing DeleteCategory: %v", err)
		h.reply(ctx, d, map[string]string{"error": err.Error()})
	} else {
		log.Printf("Category deleted successfully")
		h.reply(ctx, d, map[string]string{"status": "deleted", "id": payload.ID})
	}
}

func (h *RabbitMQHandler) handleCreateProductSpecification(d amqp.Delivery) {
	var event domain.ProductSpecificationCreateEvent
	if err := json.Unmarshal(d.Body, &event); err != nil {
		log.Printf("Error unmarshalling event: %v", err)
		return
	}

	input := ports.CreateProductSpecificationInput{
		Name:            event.Name,
		ProductNumber:   event.ProductNumber,
		Description:     event.Description,
		IsBundle:        event.IsBundle,
		LifecycleStatus: event.LifecycleStatus,
		ValidFor:        event.ValidFor,
		Characteristics: event.Characteristics,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if result, err := h.createProductSpecificationUC.Execute(ctx, input); err != nil {
		log.Printf("Error executing CreateProductSpecification: %v", err)
		h.reply(ctx, d, map[string]string{"error": err.Error()})
	} else {
		log.Printf("ProductSpecification created successfully")
		h.reply(ctx, d, result)
	}
}

func (h *RabbitMQHandler) handleUpdateProductSpecification(d amqp.Delivery) {
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
	var payload UpdatePayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		log.Printf("Error unmarshalling event: %v", err)
		return
	}

	input := ports.UpdateProductSpecificationInput{
		ID:              payload.ID,
		Name:            payload.Name,
		Description:     payload.Description,
		ProductNumber:   payload.ProductNumber,
		IsBundle:        payload.IsBundle,
		LifecycleStatus: payload.LifecycleStatus,
		ValidFor:        payload.ValidFor,
		Characteristics: payload.Characteristics,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if result, err := h.updateProductSpecificationUC.Execute(ctx, input); err != nil {
		log.Printf("Error executing UpdateProductSpecification: %v", err)
		h.reply(ctx, d, map[string]string{"error": err.Error()})
	} else {
		log.Printf("ProductSpecification updated successfully")
		h.reply(ctx, d, result)
	}
}

func (h *RabbitMQHandler) handleDeleteProductSpecification(d amqp.Delivery) {
	type DeletePayload struct {
		ID string `json:"id"`
	}
	var payload DeletePayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		log.Printf("Error unmarshalling event: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.deleteProductSpecificationUC.Execute(ctx, ports.DeleteProductSpecificationInput{ID: payload.ID}); err != nil {
		log.Printf("Error executing DeleteProductSpecification: %v", err)
		h.reply(ctx, d, map[string]string{"error": err.Error()})
	} else {
		log.Printf("ProductSpecification deleted successfully")
		h.reply(ctx, d, map[string]string{"status": "deleted", "id": payload.ID})
	}
}

func (h *RabbitMQHandler) handleCreateProductOffering(d amqp.Delivery) {
	var event domain.ProductOfferingCreateEvent
	if err := json.Unmarshal(d.Body, &event); err != nil {
		log.Printf("Error unmarshalling event: %v", err)
		return
	}

	input := ports.CreateProductOfferingInput{
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
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if result, err := h.createProductOfferingUC.Execute(ctx, input); err != nil {
		log.Printf("Error executing CreateProductOffering: %v", err)
		h.reply(ctx, d, map[string]string{"error": err.Error()})
	} else {
		log.Printf("ProductOffering created successfully")
		h.reply(ctx, d, result)
	}
}

func (h *RabbitMQHandler) handleUpdateProductOffering(d amqp.Delivery) {
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
	var payload UpdatePayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		log.Printf("Error unmarshalling event: %v", err)
		return
	}

	input := ports.UpdateProductOfferingInput{
		ID:              payload.ID,
		Name:            payload.Name,
		Description:     payload.Description,
		LifecycleStatus: payload.LifecycleStatus,
		ValidFor:        payload.ValidFor,
		IsBundle:        payload.IsBundle,
		IsSellable:      payload.IsSellable,
		CategoryIDs:     payload.CategoryIDs,
		Prices:          payload.Prices,
		Attachments:     payload.Attachments,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if result, err := h.updateProductOfferingUC.Execute(ctx, input); err != nil {
		log.Printf("Error executing UpdateProductOffering: %v", err)
		h.reply(ctx, d, map[string]string{"error": err.Error()})
	} else {
		log.Printf("ProductOffering updated successfully")
		h.reply(ctx, d, result)
	}
}

func (h *RabbitMQHandler) handleDeleteProductOffering(d amqp.Delivery) {
	type DeletePayload struct {
		ID string `json:"id"`
	}
	var payload DeletePayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		log.Printf("Error unmarshalling event: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.deleteProductOfferingUC.Execute(ctx, ports.DeleteProductOfferingInput{ID: payload.ID}); err != nil {
		log.Printf("Error executing DeleteProductOffering: %v", err)
		h.reply(ctx, d, map[string]string{"error": err.Error()})
	} else {
		log.Printf("ProductOffering deleted successfully")
		h.reply(ctx, d, map[string]string{"status": "deleted", "id": payload.ID})
	}
}

func (h *RabbitMQHandler) reply(ctx context.Context, d amqp.Delivery, response any) {
	responseBody, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling response: %v", err)
		return
	}

	h.replyMu.Lock()
	err = h.replyChannel.PublishWithContext(ctx,
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
	h.replyMu.Unlock()
	if err != nil {
		log.Printf("Failed to publish response: %v", err)
	}
}
