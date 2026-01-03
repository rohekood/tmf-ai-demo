package catalog

import (
	"context"
	"time"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"github.com/google/uuid"
)

type CreateCatalog struct {
	repo      ports.CatalogRepository
	publisher ports.EventPublisher
}

func NewCreateCatalog(repo ports.CatalogRepository, publisher ports.EventPublisher) *CreateCatalog {
	return &CreateCatalog{
		repo:      repo,
		publisher: publisher,
	}
}

func (uc *CreateCatalog) Execute(ctx context.Context, input ports.CreateCatalogInput) (*domain.Catalog, error) {
	// 1. Create Domain Entity
	catalog := &domain.Catalog{
		ID:              uuid.New().String(),
		Name:            input.Name,
		Description:     input.Description,
		ValidFor:        input.ValidFor,
		LastUpdate:      time.Now(),
		LifecycleStatus: "Active", // Default to Active for now, or "Draft"
	}

	// 2. Validate
	if err := catalog.Validate(); err != nil {
		return nil, err
	}

	// 3. Persist
	if err := uc.repo.Create(ctx, catalog); err != nil {
		return nil, err
	}

	// Publish Event
	// We ignore error here to not fail the request if publishing fails (or we could log it)?
	// Ideally we want transactional outbox, but for now just log error if strictly needed,
	// or return error if we want strong consistency (but that risks partial state if DB committed).
	// Let's just try to publish and log if error (standard MVP approach).
	// Re-reading: The Execute returns (*Catalog, error).
	// I'll return the error if publish fails? No, the catalog is created.
	// I will just return nil error, but maybe log it.
	// Actually, let's keep it simple: publish and if fail, return error?
	// The problem is the DB transaction is already committed (implicit in repo.Create).
	// So returning error would be confusing "Created but failed".
	// I will just return nil, but try to publish.
	_ = uc.publisher.PublishCatalogCreated(ctx, domain.CatalogCreatedEvent{Catalog: catalog})

	// 4. Return
	return catalog, nil
}
