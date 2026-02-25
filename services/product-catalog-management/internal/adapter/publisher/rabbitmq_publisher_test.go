package publisher

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"tmf/services/product-catalog-management/internal/core/domain"
)

type MockRabbitMQPublisher struct {
	mock.Mock
}

func (m *MockRabbitMQPublisher) Publish(ctx context.Context, exchange, routingKey string, body interface{}) error {
	args := m.Called(ctx, exchange, routingKey, body)
	return args.Error(0)
}

func (m *MockRabbitMQPublisher) PublishToQueue(ctx context.Context, queueName string, correlationID string, body interface{}) error {
	args := m.Called(ctx, queueName, correlationID, body)
	return args.Error(0)
}

func (m *MockRabbitMQPublisher) DeclareTopicExchange(name string, durable, autoDelete, internal, noWait bool) error {
	args := m.Called(name, durable, autoDelete, internal, noWait)
	return args.Error(0)
}

func (m *MockRabbitMQPublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestRabbitMQPublisher_Events(t *testing.T) {
	mockPub := new(MockRabbitMQPublisher)
	exchange := "test_exchange"
	pub, err := NewRabbitMQPublisher(mockPub, exchange)
	assert.NoError(t, err)

	ctx := context.Background()

	// 1. CatalogCreated
	catEvt := domain.CatalogCreatedEvent{Catalog: &domain.Catalog{ID: "1"}}
	mockPub.On("Publish", ctx, exchange, "evt.catalog.catalog.created", catEvt).Return(nil).Once()
	err = pub.PublishCatalogCreated(ctx, catEvt)
	assert.NoError(t, err)

	// 2. CategoryCreated
	categoryEvt := domain.CategoryCreatedEvent{Category: &domain.Category{ID: "2"}}
	mockPub.On("Publish", ctx, exchange, "evt.catalog.category.created", categoryEvt).Return(nil).Once()
	err = pub.PublishCategoryCreated(ctx, categoryEvt)
	assert.NoError(t, err)

	// 3. ProductSpecificationCreated
	specEvt := domain.ProductSpecificationCreatedEvent{ProductSpecification: &domain.ProductSpecification{ID: "3"}}
	mockPub.On("Publish", ctx, exchange, "evt.catalog.specification.created", specEvt).Return(nil).Once()
	err = pub.PublishProductSpecificationCreated(ctx, specEvt)
	assert.NoError(t, err)

	// 4. ProductOfferingCreated
	offEvt := domain.ProductOfferingCreatedEvent{ProductOffering: &domain.ProductOffering{ID: "4"}}
	mockPub.On("Publish", ctx, exchange, "evt.catalog.offering.created", offEvt).Return(nil).Once()
	err = pub.PublishProductOfferingCreated(ctx, offEvt)
	assert.NoError(t, err)

	// 5. CatalogUpdated
	catUpdEvt := domain.CatalogUpdatedEvent{Catalog: &domain.Catalog{ID: "1"}}
	mockPub.On("Publish", ctx, exchange, "evt.catalog.catalog.updated", catUpdEvt).Return(nil).Once()
	err = pub.PublishCatalogUpdated(ctx, catUpdEvt)
	assert.NoError(t, err)

	// 6. CatalogDeleted
	catDelEvt := domain.CatalogDeletedEvent{ID: "1"}
	mockPub.On("Publish", ctx, exchange, "evt.catalog.catalog.deleted", catDelEvt).Return(nil).Once()
	err = pub.PublishCatalogDeleted(ctx, catDelEvt)
	assert.NoError(t, err)

	// 7. CategoryUpdated
	categoryUpdEvt := domain.CategoryUpdatedEvent{Category: &domain.Category{ID: "2"}}
	mockPub.On("Publish", ctx, exchange, "evt.catalog.category.updated", categoryUpdEvt).Return(nil).Once()
	err = pub.PublishCategoryUpdated(ctx, categoryUpdEvt)
	assert.NoError(t, err)

	// 8. CategoryDeleted
	categoryDelEvt := domain.CategoryDeletedEvent{ID: "2"}
	mockPub.On("Publish", ctx, exchange, "evt.catalog.category.deleted", categoryDelEvt).Return(nil).Once()
	err = pub.PublishCategoryDeleted(ctx, categoryDelEvt)
	assert.NoError(t, err)

	// 9. ProductSpecificationUpdated
	specUpdEvt := domain.ProductSpecificationUpdatedEvent{ProductSpecification: &domain.ProductSpecification{ID: "3"}}
	mockPub.On("Publish", ctx, exchange, "evt.catalog.specification.updated", specUpdEvt).Return(nil).Once()
	err = pub.PublishProductSpecificationUpdated(ctx, specUpdEvt)
	assert.NoError(t, err)

	// 10. ProductSpecificationDeleted
	specDelEvt := domain.ProductSpecificationDeletedEvent{ID: "3"}
	mockPub.On("Publish", ctx, exchange, "evt.catalog.specification.deleted", specDelEvt).Return(nil).Once()
	err = pub.PublishProductSpecificationDeleted(ctx, specDelEvt)
	assert.NoError(t, err)

	// 11. ProductOfferingUpdated
	offUpdEvt := domain.ProductOfferingUpdatedEvent{ProductOffering: &domain.ProductOffering{ID: "4"}}
	mockPub.On("Publish", ctx, exchange, "evt.catalog.offering.updated", offUpdEvt).Return(nil).Once()
	err = pub.PublishProductOfferingUpdated(ctx, offUpdEvt)
	assert.NoError(t, err)

	// 12. ProductOfferingDeleted
	offDelEvt := domain.ProductOfferingDeletedEvent{ID: "4"}
	mockPub.On("Publish", ctx, exchange, "evt.catalog.offering.deleted", offDelEvt).Return(nil).Once()
	err = pub.PublishProductOfferingDeleted(ctx, offDelEvt)
	assert.NoError(t, err)

	// 13. PublishRaw
	rawBody := []byte(`{"hello":"world"}`)
	mockPub.On("Publish", ctx, exchange, "test.raw", json.RawMessage(rawBody)).Return(nil).Once()
	err = pub.PublishRaw(ctx, "test.raw", rawBody)
	assert.NoError(t, err)

	mockPub.AssertExpectations(t)
}
