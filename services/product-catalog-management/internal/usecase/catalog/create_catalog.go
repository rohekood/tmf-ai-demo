package catalog

import (
	"context"
	"time"
	"tmf/services/product-catalog-management/internal/adapter/repository"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"github.com/google/uuid"
)

type CreateCatalog struct {
	repo      ports.CatalogRepository
	publisher ports.EventPublisher
	tm        repository.TransactionManager
}

func NewCreateCatalog(repo ports.CatalogRepository, publisher ports.EventPublisher, tm repository.TransactionManager) *CreateCatalog { // Added TM
	return &CreateCatalog{
		repo:      repo,
		publisher: publisher,
		tm:        tm, // Inject TM
	}
}

func (uc *CreateCatalog) Execute(ctx context.Context, input ports.CreateCatalogInput) (*domain.Catalog, error) {
	lifecycleStatus := input.LifecycleStatus
	if lifecycleStatus == "" {
		lifecycleStatus = "Draft"
	}

	// 1. Create Domain Entity
	catalog := &domain.Catalog{
		ID:              uuid.New().String(),
		Name:            input.Name,
		Description:     input.Description,
		ValidFor:        input.ValidFor,
		LastUpdate:      time.Now().UTC(),
		LifecycleStatus: lifecycleStatus,
	}

	// 2. Validate
	if err := catalog.Validate(); err != nil {
		return nil, err
	}

	// 3. Persist
	// 3. Persist & Publish in Transaction
	if err := uc.tm.Run(ctx, func(ctx context.Context) error {
		if err := uc.repo.Create(ctx, catalog); err != nil {
			return err
		}
		if err := uc.publisher.PublishCatalogCreated(ctx, domain.CatalogCreatedEvent{Catalog: catalog}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// 4. Return
	return catalog, nil
}
